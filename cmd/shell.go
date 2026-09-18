package cmd

import (
	"context"
	"fmt"
	"io"

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
			cmd.InOrStdin(),
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
	in io.Reader,
	out io.Writer,
) error {

	activeShell := kubernetes.IsNamespaceShellActive()

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
		selectedNamespace, err = resolveNamespace(
			ctx,
			client,
			in,
			out,
			"",
		)
		if err != nil {
			return err
		}
	}

	if activeShell {
		return kubernetes.SwitchNamespaceShell(
			kubeconfig,
			selectedContext,
			selectedNamespace,
			out,
		)
	}

	return kubernetes.RunNamespaceShell(
		kubeconfig,
		selectedContext,
		selectedNamespace,
		out,
	)
}
