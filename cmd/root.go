package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/pierinho13/kubectl-peek/internal/config"
	"github.com/pierinho13/kubectl-peek/internal/kubernetes"
	"github.com/pierinho13/kubectl-peek/internal/ui"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const usageRulesEnvVar = "KUBECTL_PEEK_RULE_FILE"

var version = "dev"

var (
	namespace      string
	contextName    string
	kubeconfig     string
	usageRulesFile string
)

var rootCmd = &cobra.Command{
	Use:           "peek [pattern]",
	Short:         "Interactively inspect Kubernetes Secrets",
	Version:       version,
	SilenceUsage:  true,
	SilenceErrors: true,
	Args:          cobra.MaximumNArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		var pattern string

		if len(args) == 1 {
			pattern = args[0]
		}

		return runPeek(
			cmd.Context(),
			cmd.OutOrStdout(),
			pattern,
		)
	},
}

func init() {
	rootCmd.Flags().StringVarP(
		&namespace,
		"namespace",
		"n",
		"",
		"Kubernetes namespace",
	)

	rootCmd.Flags().StringVar(
		&contextName,
		"context",
		"",
		"Kubernetes context",
	)

	rootCmd.Flags().StringVar(
		&kubeconfig,
		"kubeconfig",
		"",
		"Path to the kubeconfig file",
	)

	rootCmd.Flags().StringVar(
		&usageRulesFile,
		"rules",
		os.Getenv(usageRulesEnvVar),
		"Path to a YAML file containing custom Secret usage rules; defaults to $KUBECTL_PEEK_RULE_FILE",
	)
}

func Execute() error {
	return rootCmd.Execute()
}

func runPeek(
	ctx context.Context,
	out io.Writer,
	pattern string,
) error {
	client, err := kubernetes.NewClient(
		kubeconfig,
		contextName,
		namespace,
	)
	if err != nil {
		return err
	}

	usageRules, err := config.LoadUsageRules(
		usageRulesFile,
	)
	if err != nil {
		return err
	}

	client.UsageRules = usageRules

	secrets, err := client.Clientset.
		CoreV1().
		Secrets(client.Namespace).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf(
			"list Secrets in namespace %q: %w",
			client.Namespace,
			err,
		)
	}

	filteredSecrets := make(
		[]string,
		0,
		len(secrets.Items),
	)

	normalizedPattern := strings.ToLower(pattern)

	for _, secret := range secrets.Items {
		if pattern == "" ||
			strings.Contains(
				strings.ToLower(secret.Name),
				normalizedPattern,
			) {
			filteredSecrets = append(
				filteredSecrets,
				secret.Name,
			)
		}
	}

	if len(filteredSecrets) == 0 {
		if pattern != "" {
			return fmt.Errorf(
				"no Secrets matching %q found in namespace %q",
				pattern,
				client.Namespace,
			)
		}

		return fmt.Errorf(
			"no Secrets found in namespace %q",
			client.Namespace,
		)
	}

	sort.Strings(filteredSecrets)

	selectedSecret, err := ui.SelectSecret(
		client.Namespace,
		filteredSecrets,
	)
	if err != nil {
		return err
	}

	secret, err := client.Clientset.
		CoreV1().
		Secrets(client.Namespace).
		Get(
			ctx,
			selectedSecret,
			metav1.GetOptions{},
		)
	if err != nil {
		return fmt.Errorf(
			"get Secret %q: %w",
			selectedSecret,
			err,
		)
	}

	if len(secret.Data) == 0 {
		return ui.RenderEmptySecretError(secret.Name)
	}

	result, err := kubernetes.FindSecretUsagesDetailed(
		ctx,
		client,
		client.Namespace,
		secret.Name,
	)
	if err != nil {
		return fmt.Errorf(
			"find secret usages: %w",
			err,
		)
	}

	fmt.Fprintln(out)
	fmt.Fprintln(
		out,
		ui.RenderSecret(secret, result),
	)

	return nil
}
