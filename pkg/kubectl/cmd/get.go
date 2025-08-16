package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"kubectl-multi/pkg/cluster"
	"kubectl-multi/pkg/core/operations"
	"kubectl-multi/pkg/util"
)

// Custom help function for get command
func getHelpFunc(cmd *cobra.Command, args []string) {
	// Get original kubectl help using the new implementation
	cmdInfo, err := util.GetKubectlCommandInfo("get")
	if err != nil {
		// Fallback to default help if kubectl help is not available
		cmd.Help()
		return
	}

	// Multi-cluster plugin information
	multiClusterInfo := `Get resources from all managed clusters and display them in a unified view.
Supports all resource types that kubectl get supports.

The output includes cluster context information to help identify which
cluster each resource belongs to.`

	// Multi-cluster examples
	multiClusterExamples := `# List all pods in all managed clusters
kubectl multi get pods

# List all nodes in all managed clusters
kubectl multi get nodes

# List deployments in specific namespace across all clusters
kubectl multi get deployments -n production

# List all resources in all namespaces across all clusters
kubectl multi get all -A

# Get pods with labels
kubectl multi get pods -l app=nginx

# Get specific pod across all clusters
kubectl multi get pod nginx-pod

# Get services with wide output
kubectl multi get services -o wide

#get all job
kubectl multi get jobs
`

	// Multi-cluster usage
	multiClusterUsage := `kubectl multi get [TYPE[.VERSION][.GROUP] [NAME | -l label] | TYPE[.VERSION][.GROUP]/NAME ...] [flags]`

	// Format combined help using the new CommandInfo structure
	combinedHelp := util.FormatMultiClusterHelp(cmdInfo, multiClusterInfo, multiClusterExamples, multiClusterUsage)
	fmt.Fprintln(cmd.OutOrStdout(), combinedHelp)
}

func newGetCommand() *cobra.Command {
	var outputFormat string
	var selector string
	var showLabels bool
	var watch bool
	var watchOnly bool

	cmd := &cobra.Command{
		Use:   "get [TYPE[.VERSION][.GROUP] [NAME | -l label] | TYPE[.VERSION][.GROUP]/NAME ...]",
		Short: "Display one or many resources across all managed clusters",
		Long: `Get resources from all managed clusters and display them in a unified view.
Supports all resource types that kubectl get supports.

The output includes cluster context information to help identify which
cluster each resource belongs to.`,
		Example: `# List all pods in all managed clusters
kubectl multi get pods

# List all nodes in all managed clusters
kubectl multi get nodes

# List deployments in specific namespace across all clusters
kubectl multi get deployments -n production

# List all resources in all namespaces across all clusters
kubectl multi get all -A

# Get pods with labels
kubectl multi get pods -l app=nginx

# Get specific pod across all clusters
kubectl multi get pod nginx-pod

# Get services with wide output
kubectl multi get services -o wide`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("resource type must be specified")
			}

			// For watch operations, we don't support multi-cluster watch yet
			if watch || watchOnly {
				return fmt.Errorf("watch operations are not supported in multi-cluster mode")
			}

			kubeconfig, remoteCtx, _, namespace, allNamespaces := GetGlobalFlags()

			// Discover clusters
			clusters, err := cluster.DiscoverClusters(kubeconfig, remoteCtx)
			if err != nil {
				return fmt.Errorf("failed to discover clusters: %v", err)
			}

			// Parse resource type and name
			resourceType := args[0]
			resourceName := ""
			if len(args) > 1 {
				resourceName = args[1]
			}

			// Build options and execute
			opts := operations.GetOptions{
				Clusters:      clusters,
				ResourceType:  resourceType,
				ResourceName:  resourceName,
				Namespace:     namespace,
				AllNamespaces: allNamespaces,
				Selector:      selector,
				ShowLabels:    showLabels,
				OutputFormat:  outputFormat,
			}

			return operations.ExecuteGet(opts)
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "", "output format (json|yaml|wide|name|custom-columns=...|custom-columns-file=...|go-template=...|go-template-file=...|jsonpath=...|jsonpath-file=...)")
	cmd.Flags().StringVarP(&selector, "selector", "l", "", "selector (label query) to filter on")
	cmd.Flags().BoolVar(&showLabels, "show-labels", false, "show all labels as the last column")
	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "watch for changes to the requested object(s)")
	cmd.Flags().BoolVar(&watchOnly, "watch-only", false, "watch for changes to the requested object(s), without listing/getting first")

	// Set custom help function
	cmd.SetHelpFunc(getHelpFunc)

	return cmd
}