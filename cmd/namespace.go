package cmd

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/pierinho13/kubectl-peek/internal/kubernetes"
	"github.com/pierinho13/kubectl-peek/internal/ui"

	"github.com/spf13/cobra"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var namespaceShell bool

var namespaceCmd = &cobra.Command{
	Use:     "namespace [pattern]",
	Aliases: []string{"namespaces", "ns"},
	Short:   "Select the namespace for a Kubernetes context (--shell to open an isolated shell)",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var pattern string

		if len(args) == 1 {
			pattern = args[0]
		}

		return runNamespace(
			cmd.Context(),
			cmd.InOrStdin(),
			cmd.OutOrStdout(),
			pattern,
			namespaceShell,
		)
	},
}

func init() {
	namespaceCmd.Flags().BoolVar(
		&namespaceShell,
		"shell",
		false,
		"Open an isolated shell using the selected namespace",
	)
}

func runNamespace(
	ctx context.Context,
	in io.Reader,
	out io.Writer,
	pattern string,
	openShell bool,
) error {
	activeShell := openShell &&
		kubernetes.IsNamespaceShellActive()

	client, err := kubernetes.NewClient(
		kubeconfig,
		contextName,
		"",
	)
	if err != nil {
		return err
	}

	selectedNamespace, err := resolveNamespace(
		ctx,
		client,
		in,
		out,
		pattern,
	)
	if err != nil {
		return err
	}

	if openShell {
		if activeShell {
			return kubernetes.SwitchNamespaceShell(
				kubeconfig,
				contextName,
				selectedNamespace,
				out,
			)
		}

		return kubernetes.RunNamespaceShell(
			kubeconfig,
			contextName,
			selectedNamespace,
			out,
		)
	}

	changedContext, err := kubernetes.SetContextNamespace(
		kubeconfig,
		contextName,
		selectedNamespace,
	)
	if err != nil {
		return err
	}

	fmt.Fprintf(
		out,
		"Context %q now uses namespace %q\n",
		changedContext,
		selectedNamespace,
	)

	return nil
}

func resolveNamespace(
	ctx context.Context,
	client *kubernetes.Client,
	in io.Reader,
	out io.Writer,
	pattern string,
) (string, error) {
	namespaceList, err := client.Clientset.
		CoreV1().
		Namespaces().
		List(ctx, metav1.ListOptions{})
	if err != nil {
		if !apierrors.IsForbidden(err) {
			return "", fmt.Errorf("list namespaces: %w", err)
		}

		return promptNamespace(in, out)
	}

	namespaces := make([]string, 0, len(namespaceList.Items))
	normalizedPattern := strings.ToLower(pattern)

	for _, item := range namespaceList.Items {
		if pattern == "" ||
			strings.Contains(
				strings.ToLower(item.Name),
				normalizedPattern,
			) {
			namespaces = append(namespaces, item.Name)
		}
	}

	if len(namespaces) == 0 {
		if pattern != "" {
			return "", fmt.Errorf(
				"no namespaces matching %q found",
				pattern,
			)
		}

		return "", fmt.Errorf("no namespaces found")
	}

	sort.Strings(namespaces)

	return ui.SelectNamespace(namespaces)
}

func promptNamespace(in io.Reader, out io.Writer) (string, error) {
	fmt.Fprintln(
		out,
		"You don't have permission to list namespaces on this cluster.",
	)
	fmt.Fprint(out, "Enter namespace name: ")

	reader := bufio.NewReader(in)

	line, err := reader.ReadString('\n')
	if err != nil && line == "" {
		return "", fmt.Errorf("read namespace: %w", err)
	}

	namespace := strings.TrimSpace(line)
	if namespace == "" {
		return "", fmt.Errorf("no namespace entered")
	}

	return namespace, nil
}
