package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	kubeconfig    string
	remoteCtx     string
	localCtx      string
	namespace     string
	allNamespaces bool
)

var rootCmd = &cobra.Command{
	Use:   "kubestellar",
	Short: "KubeStellar - Multi-cluster management tool",
	Long: `KubeStellar is a powerful multi-cluster management tool that enables
you to manage multiple Kubernetes clusters from a single command line interface.

It provides unified views of resources across all your managed clusters,
making it easy to deploy, monitor, and manage applications in a multi-cluster environment.`,
	Version: "1.0.0",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&kubeconfig, "kubeconfig", "", "path to the kubeconfig file")
	rootCmd.PersistentFlags().StringVar(&remoteCtx, "remote-context", "", "remote context to use (e.g., kind-wds1 for remote WDS)")
	rootCmd.PersistentFlags().StringVar(&localCtx, "local-context", "", "local context to use for local operations")
	rootCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "namespace to operate in")
	rootCmd.PersistentFlags().BoolVarP(&allNamespaces, "all-namespaces", "A", false, "operate across all namespaces")

	// Add commands with short forms
	rootCmd.AddCommand(newGetCommand())
	rootCmd.AddCommand(newApplyCommand())
	rootCmd.AddCommand(newDeleteCommand())
	rootCmd.AddCommand(newDescribeCommand())
	rootCmd.AddCommand(newLogsCommand())
	rootCmd.AddCommand(newRolloutCommand())
	rootCmd.AddCommand(newRunCommand())
	rootCmd.AddCommand(newBindingPolicyCommand())
	rootCmd.AddCommand(newClustersCommand())
	rootCmd.AddCommand(newHelmCommand())

	// Version command formatting
	rootCmd.SetVersionTemplate(`KubeStellar {{.Version}}
`)
}

// GetGlobalFlags returns the global flag values
func GetGlobalFlags() (string, string, string, string, bool) {
	// Get kubeconfig from env if not set via flag
	if kubeconfig == "" {
		kubeconfig = os.Getenv("KUBECONFIG")
	}
	return kubeconfig, remoteCtx, localCtx, namespace, allNamespaces
}