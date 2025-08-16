package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"kubectl-multi/pkg/cluster"
	"kubectl-multi/pkg/core/operations"
)

func newGetCommand() *cobra.Command {
	var outputFormat string
	var selector string
	var showLabels bool

	cmd := &cobra.Command{
		Use:     "get [TYPE[.VERSION][.GROUP] [NAME | -l label] | TYPE[.VERSION][.GROUP]/NAME ...]",
		Aliases: []string{"g"},
		Short:   "Display one or many resources across all managed clusters",
		Long: `Get resources from all KubeStellar managed clusters and display them in a unified view.
Supports all Kubernetes resource types.

The output includes cluster context information to help identify which
cluster each resource belongs to.`,
		Example: `# List all pods in all managed clusters
kubestellar get pods

# List all nodes in all managed clusters
kubestellar get nodes

# List deployments in specific namespace across all clusters
kubestellar get deployments -n production

# List all resources in all namespaces across all clusters
kubestellar get all -A

# Get pods with labels
kubestellar get pods -l app=nginx

# Get specific pod across all clusters
kubestellar get pod nginx-pod

# Get services with wide output
kubestellar get services -o wide

# Get all jobs
kubestellar get jobs`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("resource type must be specified")
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

	return cmd
}