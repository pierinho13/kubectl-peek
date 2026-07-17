package cmd

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/pierinho13/kubectl-peek/internal/ui"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kubeclient "github.com/pierinho13/kubectl-peek/internal/kubernetes"
)

var (
	namespace   string
	contextName string
	kubeconfig  string
)

var rootCmd = &cobra.Command{
	Use:           "peek [pattern]",
	Short:         "Interactively inspect Kubernetes Secrets",
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
}

func Execute() error {
	return rootCmd.Execute()
}

func runPeek(
	ctx context.Context,
	out io.Writer,
	pattern string,
) error {
	client, err := kubeclient.NewClient(
		kubeconfig,
		contextName,
		namespace,
	)
	if err != nil {
		return err
	}

	secrets, err := client.Clientset.CoreV1().
		Secrets(client.Namespace).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf(
			"list Secrets in namespace %q: %w",
			client.Namespace,
			err,
		)
	}

	filteredSecrets := make([]string, 0, len(secrets.Items))

	for _, secret := range secrets.Items {
		if pattern == "" ||
			strings.Contains(
				strings.ToLower(secret.Name),
				strings.ToLower(pattern),
			) {
			filteredSecrets = append(filteredSecrets, secret.Name)
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

	secret, err := client.Clientset.CoreV1().
		Secrets(client.Namespace).
		Get(ctx, selectedSecret, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf(
			"get Secret %q: %w",
			selectedSecret,
			err,
		)
	}

	if len(secret.Data) == 0 {
		return fmt.Errorf(
			"Secret %q contains no data",
			secret.Name,
		)
	}

	fmt.Fprintf(out, "\nSecret: %s\n", secret.Name)
	fmt.Fprintf(out, "Namespace: %s\n", secret.Namespace)
	fmt.Fprintf(out, "Type: %s\n\n", secret.Type)

	keys := make([]string, 0, len(secret.Data))

	for key := range secret.Data {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	for _, key := range keys {
		value := secret.Data[key]

		fmt.Fprintf(out, "%s:\n", key)
		fmt.Fprintf(out, "%s\n\n", string(value))
	}

	return nil
}

// func selectSecret(
// 	namespace string,
// 	secretNames []string,
// ) (string, error) {
// 	var selectedSecret string

// 	options := make([]huh.Option[string], 0, len(secretNames))

// 	for _, name := range secretNames {
// 		options = append(options, huh.NewOption(name, name))
// 	}

// 	keyMap := huh.NewDefaultKeyMap()

// 	keyMap.Select.HalfPageUp.SetKeys("left", "ctrl+u")
// 	keyMap.Select.HalfPageUp.SetHelp("←", "previous page")
// 	keyMap.Select.HalfPageUp.SetEnabled(true)

// 	keyMap.Select.HalfPageDown.SetKeys("right", "ctrl+d")
// 	keyMap.Select.HalfPageDown.SetHelp("→", "next page")
// 	keyMap.Select.HalfPageDown.SetEnabled(true)

// 	selectField := huh.NewSelect[string]().
// 		Title(fmt.Sprintf(
// 			"Select a Secret from namespace %q",
// 			namespace,
// 		)).
// 		Description("Use ↑/↓ to move, ←/→ to change page, and / to filter.").
// 		Options(options...).
// 		Height(12).
// 		Value(&selectedSecret)

// 	form := huh.NewForm(
// 		huh.NewGroup(selectField),
// 	).WithKeyMap(keyMap)

// 	if err := form.Run(); err != nil {
// 		if errors.Is(err, huh.ErrUserAborted) {
// 			return "", errors.New("selection cancelled")
// 		}

// 		return "", fmt.Errorf("select Secret: %w", err)
// 	}

// 	return selectedSecret, nil
// }
