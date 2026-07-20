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
	out io.Writer,
	pattern string,
	openShell bool,
) error {
	if openShell {
		if err := kubernetes.EnsureNoActiveNamespaceShell(); err != nil {
			return err
		}
	}

	client, err := kubernetes.NewClient(
		kubeconfig,
		contextName,
		"",
	)
	if err != nil {
		return err
	}

	namespaceList, err := client.Clientset.
		CoreV1().
		Namespaces().
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("list namespaces: %w", err)
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
			return fmt.Errorf(
				"no namespaces matching %q found",
				pattern,
			)
		}

		return fmt.Errorf("no namespaces found")
	}

	sort.Strings(namespaces)

	selectedNamespace, err := ui.SelectNamespace(namespaces)
	if err != nil {
		return err
	}

	if openShell {
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
