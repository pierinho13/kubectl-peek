package cmd

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/pierinho13/kubectl-peek/internal/kubernetes"
	"github.com/pierinho13/kubectl-peek/internal/ui"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	execNamespace string
	execContainer string
	execShell     string
)

var execCmd = &cobra.Command{
	Use:     "exec [pattern]",
	Aliases: []string{"x"},
	Short:   "Open an interactive shell in a Kubernetes Pod",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var pattern string
		if len(args) == 1 {
			pattern = args[0]
		}

		return runExec(
			cmd.Context(),
			cmd.OutOrStdout(),
			pattern,
		)
	},
}

func init() {
	execCmd.Flags().StringVarP(
		&execNamespace,
		"namespace",
		"n",
		"",
		"Kubernetes namespace",
	)

	execCmd.Flags().StringVarP(
		&execContainer,
		"container",
		"c",
		"",
		"Container name",
	)

	execCmd.Flags().StringVar(
		&execShell,
		"shell",
		"",
		"Shell executable to run inside the container",
	)
}

func runExec(
	ctx context.Context,
	out io.Writer,
	pattern string,
) error {
	client, err := kubernetes.NewClient(
		kubeconfig,
		contextName,
		execNamespace,
	)
	if err != nil {
		return err
	}

	podList, err := client.Clientset.
		CoreV1().
		Pods(client.Namespace).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf(
			"list Pods in namespace %q: %w",
			client.Namespace,
			err,
		)
	}

	normalizedPattern := strings.ToLower(pattern)
	pods := make([]ui.PodOption, 0, len(podList.Items))

	for i := range podList.Items {
		pod := &podList.Items[i]

		if pod.Status.Phase == corev1.PodSucceeded ||
			pod.Status.Phase == corev1.PodFailed ||
			len(pod.Spec.Containers) == 0 {
			continue
		}

		if pattern != "" &&
			!strings.Contains(
				strings.ToLower(pod.Name),
				normalizedPattern,
			) {
			continue
		}

		pods = append(
			pods,
			ui.PodOption{
				Name:       pod.Name,
				Phase:      pod.Status.Phase,
				Ready:      readyContainerCount(pod.Status.ContainerStatuses),
				Containers: len(pod.Spec.Containers),
			},
		)
	}

	if len(pods) == 0 {
		if pattern != "" {
			return fmt.Errorf(
				"no Pods matching %q found in namespace %q",
				pattern,
				client.Namespace,
			)
		}

		return fmt.Errorf(
			"no executable Pods found in namespace %q",
			client.Namespace,
		)
	}

	sort.Slice(pods, func(i, j int) bool {
		return pods[i].Name < pods[j].Name
	})

	selectedPodName, err := ui.SelectPod(client.Namespace, pods)
	if err != nil {
		return err
	}

	pod, err := client.Clientset.
		CoreV1().
		Pods(client.Namespace).
		Get(ctx, selectedPodName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf(
			"get Pod %q in namespace %q: %w",
			selectedPodName,
			client.Namespace,
			err,
		)
	}

	if pod.Status.Phase != corev1.PodRunning {
		return fmt.Errorf(
			"Pod %q is not running (current phase: %s)",
			pod.Name,
			pod.Status.Phase,
		)
	}

	containerName, err := selectExecContainer(pod, execContainer)
	if err != nil {
		return err
	}

	fmt.Fprintf(
		out,
		"Opening shell in %s/%s (container %s)\n",
		client.Namespace,
		pod.Name,
		containerName,
	)

	return kubernetes.ExecPodShell(
		ctx,
		client,
		client.Namespace,
		pod.Name,
		containerName,
		execShell,
	)
}

func selectExecContainer(
	pod *corev1.Pod,
	requested string,
) (string, error) {
	containers := make([]string, 0, len(pod.Spec.Containers))

	for _, container := range pod.Spec.Containers {
		containers = append(containers, container.Name)
	}

	if requested != "" {
		for _, container := range containers {
			if container == requested {
				return requested, nil
			}
		}

		return "", fmt.Errorf(
			"container %q not found in Pod %q",
			requested,
			pod.Name,
		)
	}

	if len(containers) == 1 {
		return containers[0], nil
	}

	sort.Strings(containers)
	return ui.SelectContainer(pod.Name, containers)
}

func readyContainerCount(
	statuses []corev1.ContainerStatus,
) int {
	ready := 0
	for _, status := range statuses {
		if status.Ready {
			ready++
		}
	}
	return ready
}
