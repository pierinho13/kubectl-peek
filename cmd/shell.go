package cmd

import (
	"context"
	"fmt"
	"io"
	"sort"

	"github.com/pierinho13/kubectl-peek/internal/kubernetes"
	"github.com/pierinho13/kubectl-peek/internal/ui"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var shellNamespace string

var shellCmd = &cobra.Command{
	Use:   "shell",
	Short: "Open an isolated shell for a Kubernetes context and namespace",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runShell(
			cmd.Context(),
			cmd.OutOrStdout(),
		)
	},
}

func init() {
	shellCmd.Flags().StringVarP(
		&shellNamespace,
		"namespace",
		"n",
		"",
		"Kubernetes namespace",
	)
}

func runShell(
	ctx context.Context,
	out io.Writer,
) error {

	if err := kubernetes.EnsureNoActiveNamespaceShell(); err != nil {
		return err
	}

	selectedContext := contextName

	if selectedContext == "" {
		contextNames, err := kubernetes.ContextNames(kubeconfig)
		if err != nil {
			return err
		}

		selectedContext, err = ui.SelectContext(contextNames)
		if err != nil {
			return err
		}
	}

	client, err := kubernetes.NewClient(
		kubeconfig,
		selectedContext,
		"",
	)
	if err != nil {
		return err
	}

	selectedNamespace := shellNamespace

	if selectedNamespace != "" {
		_, err := client.Clientset.
			CoreV1().
			Namespaces().
			Get(
				ctx,
				selectedNamespace,
				metav1.GetOptions{},
			)
		if err != nil {
			return fmt.Errorf(
				"get namespace %q from context %q: %w",
				selectedNamespace,
				selectedContext,
				err,
			)
		}
	} else {
		namespaceList, err := client.Clientset.
			CoreV1().
			Namespaces().
			List(ctx, metav1.ListOptions{})
		if err != nil {
			return fmt.Errorf(
				"list namespaces from context %q: %w",
				selectedContext,
				err,
			)
		}

		namespaces := make(
			[]string,
			0,
			len(namespaceList.Items),
		)

		for _, item := range namespaceList.Items {
			namespaces = append(
				namespaces,
				item.Name,
			)
		}

		sort.Strings(namespaces)

		selectedNamespace, err = ui.SelectNamespace(namespaces)
		if err != nil {
			return err
		}
	}

	return kubernetes.RunNamespaceShell(
		kubeconfig,
		selectedContext,
		selectedNamespace,
		out,
	)
}
