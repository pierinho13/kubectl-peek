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

var (
	secretNamespace      string
	secretUsageRulesFile string
	secretShowUsage      bool
	secretShowValues     bool
)

var secretCmd = &cobra.Command{
	Use:     "secret [pattern]",
	Aliases: []string{"secrets", "sec"},
	Short:   "Inspect Kubernetes Secrets and their relationships",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var pattern string

		if len(args) == 1 {
			pattern = args[0]
		}

		return runSecret(
			cmd.Context(),
			cmd.OutOrStdout(),
			pattern,
		)
	},
}

func init() {
	secretCmd.Flags().StringVarP(
		&secretNamespace,
		"namespace",
		"n",
		"",
		"Kubernetes namespace",
	)

	secretCmd.Flags().StringVar(
		&secretUsageRulesFile,
		"rules",
		os.Getenv(usageRulesEnvVar),
		"Path to a YAML file containing custom Secret usage rules; defaults to $KUBECTL_PEEK_RULE_FILE",
	)

	secretCmd.Flags().BoolVar(
		&secretShowUsage,
		"show-usage",
		false,
		"Discover and display resources that use, produce, or reference the Secret",
	)

	secretCmd.Flags().BoolVar(
		&secretShowValues,
		"show-values",
		true,
		"Display decoded Secret values; use --show-values=false to redact them",
	)
}

func runSecret(
	ctx context.Context,
	out io.Writer,
	pattern string,
) error {
	client, err := kubernetes.NewClient(
		kubeconfig,
		contextName,
		secretNamespace,
	)
	if err != nil {
		return err
	}

	if secretShowUsage {
		usageRules, err := config.LoadUsageRules(
			secretUsageRulesFile,
		)
		if err != nil {
			return err
		}

		client.UsageRules = usageRules
	}

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

	var result kubernetes.SecretUsageResult

	if secretShowUsage {
		result, err = kubernetes.FindSecretUsagesDetailed(
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
	}

	fmt.Fprintln(out)
	fmt.Fprintln(
		out,
		ui.RenderSecret(
			secret,
			result,
			secretShowUsage,
			secretShowValues,
		),
	)

	return nil
}
