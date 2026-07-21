package cmd

import "github.com/spf13/cobra"

var version = "dev"

var (
	contextName string
	kubeconfig  string
)

var rootCmd = &cobra.Command{
	Use:           "peek",
	Short:         "Inspect Kubernetes resources and open isolated shells",
	Version:       version,
	SilenceUsage:  true,
	SilenceErrors: true,
	Args:          cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(
		secretCmd,
		namespaceCmd,
		shellCmd,
		execCmd,
		eventsCmd,
	)

	rootCmd.PersistentFlags().StringVar(
		&contextName,
		"context",
		"",
		"Kubernetes context",
	)

	rootCmd.PersistentFlags().StringVar(
		&kubeconfig,
		"kubeconfig",
		"",
		"Path to the kubeconfig file",
	)
}

func Execute() error {
	return rootCmd.Execute()
}
