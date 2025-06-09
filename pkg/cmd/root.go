package cmd

import (
	"github.com/spf13/cobra"
)

var (
	kubeconfig    string
	remoteCtx     string
	allClusters   bool
	namespace     string
	allNamespaces bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "kubectl-multi",
	Short: "Multi-cluster kubectl operations for KubeStellar",
	Long: `kubectl-multi provides multi-cluster operations for KubeStellar managed clusters.
It executes kubectl commands across all managed clusters and presents unified output.

This plugin automatically discovers KubeStellar managed clusters and executes
kubectl operations across all of them, displaying results with cluster context
information for easy identification.`,
	Example: `# Get nodes from all managed clusters
kubectl multi get nodes

# Get pods from all clusters in default namespace
kubectl multi get pods

# Get pods from all clusters in all namespaces
kubectl multi get pods -A

# Describe a specific pod across all clusters
kubectl multi describe pod mypod

# Get services in specific namespace across all clusters
kubectl multi get services -n kube-system

# Apply a manifest to all clusters
kubectl multi apply -f deployment.yaml

# Delete resources from all clusters
kubectl multi delete deployment nginx`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&kubeconfig, "kubeconfig", "", "path to kubeconfig file (defaults to $HOME/.kube/config)")
	rootCmd.PersistentFlags().StringVar(&remoteCtx, "remote-context", "its1", "remote hosting context for ManagedCluster resources")
	rootCmd.PersistentFlags().BoolVar(&allClusters, "all-clusters", true, "operate on all managed clusters")
	rootCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "target namespace")
	rootCmd.PersistentFlags().BoolVarP(&allNamespaces, "all-namespaces", "A", false, "list resources across all namespaces")

	// Add subcommands
	rootCmd.AddCommand(newGetCommand())
	rootCmd.AddCommand(newDescribeCommand())
	rootCmd.AddCommand(newApplyCommand())
	rootCmd.AddCommand(newDeleteCommand())
	rootCmd.AddCommand(newLogsCommand())
	rootCmd.AddCommand(newExecCommand())
	rootCmd.AddCommand(newCreateCommand())
	rootCmd.AddCommand(newEditCommand())
	rootCmd.AddCommand(newPatchCommand())
	rootCmd.AddCommand(newScaleCommand())
	rootCmd.AddCommand(newRolloutCommand())
	rootCmd.AddCommand(newPortForwardCommand())
	rootCmd.AddCommand(newTopCommand())
}

// GetGlobalFlags returns the global flags that can be used by subcommands
func GetGlobalFlags() (string, string, bool, string, bool) {
	return kubeconfig, remoteCtx, allClusters, namespace, allNamespaces
}
