package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"kubectl-multi/pkg/cluster"
	"kubectl-multi/pkg/core/operations"
)

func newApplyCommand() *cobra.Command {
	var filename string
	var recursive bool
	var dryRun string

	cmd := &cobra.Command{
		Use:     "apply (-f FILENAME | --filename=FILENAME)",
		Aliases: []string{"app"},
		Short:   "Apply a configuration to resources across all managed clusters",
		Long: `Apply a configuration to resources across all KubeStellar managed clusters.
This command applies manifests to all discovered clusters.`,
		Example: `# Apply a deployment to all managed clusters
kubestellar apply -f deployment.yaml

# Apply resources from a directory to all clusters
kubestellar apply -f dir/

# Apply with dry-run to see what would be applied
kubestellar apply -f deployment.yaml --dry-run=client

# Apply resources recursively from a directory
kubestellar apply -f dir/ -R`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if filename == "" {
				return fmt.Errorf("filename is required")
			}

			kubeconfig, remoteCtx, _, namespace, allNamespaces := GetGlobalFlags()

			// Discover clusters
			clusters, err := cluster.DiscoverClusters(kubeconfig, remoteCtx)
			if err != nil {
				return fmt.Errorf("failed to discover clusters: %v", err)
			}

			// Build options and execute
			opts := operations.ApplyOptions{
				Clusters:      clusters,
				Filename:      filename,
				Recursive:     recursive,
				DryRun:        dryRun,
				Namespace:     namespace,
				AllNamespaces: allNamespaces,
				Kubeconfig:    kubeconfig,
			}

			return operations.ExecuteApply(opts)
		},
	}

	cmd.Flags().StringVarP(&filename, "filename", "f", "", "filename, directory, or URL to files to use to apply the resource")
	cmd.Flags().BoolVarP(&recursive, "recursive", "R", false, "process the directory used in -f, --filename recursively")
	cmd.Flags().StringVar(&dryRun, "dry-run", "none", "must be \"none\", \"server\", or \"client\"")

	return cmd
}