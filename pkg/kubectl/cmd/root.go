package cmd

import (
	"fmt"
	"kubectl-multi/pkg/util"

	"github.com/spf13/cobra"
)

var (
	kubeconfig    string
	remoteCtx     string
	allClusters   bool
	namespace     string
	allNamespaces bool
)

// Custom help function for root command
func rootHelpFunc(cmd *cobra.Command, args []string) {
	// Get original kubectl help using the new implementation
	cmdInfo, err := util.GetKubectlRootInfo()
	if err != nil {
		// Fallback to default help if kubectl help is not available
		cmd.Help()
		return
	}

	// Multi-cluster plugin information
	multiClusterInfo := `kubectl-multi provides multi-cluster operations for KubeStellar managed clusters.
It executes kubectl commands across all managed clusters and presents unified output.

This plugin automatically discovers KubeStellar managed clusters and executes
kubectl operations across all of them, displaying results with cluster context
information for easy identification.`

	// Multi-cluster examples
	multiClusterExamples := `# Get nodes from all managed clusters
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
kubectl multi delete deployment nginx`

	// Multi-cluster usage
	multiClusterUsage := `kubectl multi [command] [flags]`

	// Format combined help using the new CommandInfo structure
	combinedHelp := util.FormatMultiClusterRootHelp(cmdInfo, multiClusterInfo, multiClusterExamples, multiClusterUsage)
	fmt.Fprintln(cmd.OutOrStdout(), combinedHelp)
}

const helpTemplate = `{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}

{{end}}{{if .Example}}Examples:
{{.Example}}

{{end}}{{if .HasAvailableFlags}}Options:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}

{{end}}Usage:
  {{.UseLine}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}
`

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
	rootCmd.SetHelpTemplate(helpTemplate)

	// Set custom help function for root command
	rootCmd.SetHelpFunc(rootHelpFunc)

	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&kubeconfig, "kubeconfig", "", "path to kubeconfig file (defaults to $HOME/.kube/config)")
	rootCmd.PersistentFlags().StringVar(&remoteCtx, "remote-context", "its1", "remote hosting context for ManagedCluster resources")
	rootCmd.PersistentFlags().BoolVar(&allClusters, "all-clusters", true, "operate on all managed clusters")
	rootCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "target namespace")
	rootCmd.PersistentFlags().BoolVarP(&allNamespaces, "all-namespaces", "A", false, "list resources across all namespaces")

	// Add subcommands with short forms
	rootCmd.AddCommand(newGetCommand())
	rootCmd.AddCommand(newDescribeCommand())
	rootCmd.AddCommand(newApplyCommand())
	rootCmd.AddCommand(newDeleteCommand())
	rootCmd.AddCommand(newLogsCommand())
	rootCmd.AddCommand(newRolloutCommand())
	rootCmd.AddCommand(newRunCommand())
	rootCmd.AddCommand(newBindingPolicyCommand())
	rootCmd.AddCommand(newClustersCommand())
	rootCmd.AddCommand(newHelmCommand())
}

// GetGlobalFlags returns the global flags that can be used by subcommands
func GetGlobalFlags() (string, string, bool, string, bool) {
	return kubeconfig, remoteCtx, allClusters, namespace, allNamespaces
}
