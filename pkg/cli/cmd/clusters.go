package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"kubectl-multi/pkg/core/operations"
)

func newClustersCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "clusters",
		Aliases: []string{"cluster", "c"},
		Short:   "Manage cluster registration and information",
		Long: `Manage cluster registration with KubeStellar ITS and view cluster information.

This command allows you to register new clusters, list registered clusters,
update cluster labels, and manage cluster lifecycle in KubeStellar.`,
		Example: `# List all registered clusters
kubestellar clusters list
kubestellar c list

# Register a new cluster  
kubestellar clusters add my-cluster https://cluster.example.com

# Show detailed cluster information
kubestellar clusters info my-cluster

# Update cluster labels
kubestellar clusters label my-cluster env=production zone=us-east

# Remove a cluster from ITS
kubestellar c remove my-cluster`,
	}

	cmd.AddCommand(newClustersListCommand())
	cmd.AddCommand(newClustersAddCommand())
	cmd.AddCommand(newClustersRemoveCommand())
	cmd.AddCommand(newClustersInfoCommand())
	cmd.AddCommand(newClustersLabelCommand())

	return cmd
}

func newClustersListCommand() *cobra.Command {
	var includeWDS bool

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls", "get"},
		Short:   "List all registered clusters",
		Long:    "List all clusters registered with KubeStellar ITS, optionally including WDS clusters.",
		Example: `# List all WEC clusters
kubestellar clusters list

# List all clusters including WDS
kubestellar clusters list --include-wds

# Short form
kubestellar c ls`,
		RunE: func(cmd *cobra.Command, args []string) error {
			kubeconfig, itsContext, _, _, _ := GetGlobalFlags()
			return operations.ListClustersInITS(kubeconfig, itsContext, includeWDS)
		},
	}

	cmd.Flags().BoolVar(&includeWDS, "include-wds", false, "Include WDS clusters in the output")

	return cmd
}

func newClustersAddCommand() *cobra.Command {
	var clusterEndpoint string
	var clusterLabels []string

	cmd := &cobra.Command{
		Use:     "add NAME",
		Aliases: []string{"register", "new"},
		Short:   "Register a new cluster with ITS",
		Long: `Register a new cluster with KubeStellar ITS (Inventory and Transport Space).

This creates a ManagedCluster resource in the ITS that represents the cluster
and allows KubeStellar to manage workload distribution to it.`,
		Example: `# Register a cluster
kubestellar clusters add my-cluster

# Register with endpoint and labels
kubestellar clusters add prod-east \
  --endpoint https://prod-east.example.com \
  --labels env=production,zone=us-east-1

# Short form
kubestellar c add staging-cluster`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			kubeconfig, itsContext, _, _, _ := GetGlobalFlags()

			// Parse labels
			labels := make(map[string]string)
			for _, labelString := range clusterLabels {
				parts := strings.SplitN(labelString, "=", 2)
				if len(parts) == 2 {
					labels[parts[0]] = parts[1]
				}
			}

			// Default labels for WEC clusters
			if labels["type"] == "" {
				labels["type"] = "wec"
			}

			opts := operations.ClusterRegistrationOptions{
				Kubeconfig:      kubeconfig,
				ITSContext:      itsContext,
				ClusterName:     args[0],
				ClusterEndpoint: clusterEndpoint,
				Labels:          labels,
			}

			return operations.RegisterClusterWithITS(opts)
		},
	}

	cmd.Flags().StringVar(&clusterEndpoint, "endpoint", "", "Cluster API endpoint URL")
	cmd.Flags().StringSliceVar(&clusterLabels, "labels", []string{}, "Labels to apply to the cluster (format: key=value)")

	return cmd
}

func newClustersRemoveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "remove NAME",
		Aliases: []string{"delete", "rm", "unregister"},
		Short:   "Remove a cluster from ITS",
		Long: `Remove a cluster from KubeStellar ITS registration.

This deletes the ManagedCluster resource representing the cluster,
effectively removing it from KubeStellar management.`,
		Example: `# Remove a cluster
kubestellar clusters remove my-cluster

# Short forms
kubestellar c rm old-cluster
kubestellar c delete staging-cluster`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			kubeconfig, itsContext, _, _, _ := GetGlobalFlags()

			opts := operations.ClusterRegistrationOptions{
				Kubeconfig:  kubeconfig,
				ITSContext:  itsContext,
				ClusterName: args[0],
			}

			return operations.UnregisterClusterFromITS(opts)
		},
	}

	return cmd
}

func newClustersInfoCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "info NAME",
		Aliases: []string{"describe", "show"},
		Short:   "Show detailed information about a cluster",
		Long: `Show detailed information about a specific cluster registered with ITS.

Displays cluster metadata, labels, status, and registration details.`,
		Example: `# Show cluster information
kubestellar clusters info my-cluster

# Short forms
kubestellar c info prod-cluster
kubestellar c describe staging-cluster`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			kubeconfig, itsContext, _, _, _ := GetGlobalFlags()

			opts := operations.ClusterRegistrationOptions{
				Kubeconfig:  kubeconfig,
				ITSContext:  itsContext,
				ClusterName: args[0],
			}

			return operations.DescribeClusterInITS(opts)
		},
	}

	return cmd
}

func newClustersLabelCommand() *cobra.Command {
	var overwrite bool

	cmd := &cobra.Command{
		Use:     "label NAME KEY_1=VAL_1 ... KEY_N=VAL_N",
		Aliases: []string{"labels"},
		Short:   "Update labels on a cluster",
		Long: `Update labels on a cluster registered with ITS.

Labels are used by BindingPolicies to select which clusters should receive workloads.
Common labels include environment (env), zone, tier, and cluster type.`,
		Example: `# Add/update labels
kubestellar clusters label my-cluster env=production zone=us-east-1

# Update with overwrite
kubestellar clusters label my-cluster env=staging --overwrite

# Short form
kubestellar c label prod-cluster tier=web`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			kubeconfig, itsContext, _, _, _ := GetGlobalFlags()
			clusterName := args[0]

			// Parse labels from remaining arguments
			labels := make(map[string]string)
			for _, labelString := range args[1:] {
				parts := strings.SplitN(labelString, "=", 2)
				if len(parts) == 2 {
					labels[parts[0]] = parts[1]
				} else {
					return fmt.Errorf("invalid label format: %s (expected key=value)", labelString)
				}
			}

			opts := operations.ClusterRegistrationOptions{
				Kubeconfig:  kubeconfig,
				ITSContext:  itsContext,
				ClusterName: clusterName,
				Labels:      labels,
				Overwrite:   overwrite,
			}

			return operations.UpdateClusterLabelsInITS(opts)
		},
	}

	cmd.Flags().BoolVar(&overwrite, "overwrite", false, "Overwrite existing labels")

	return cmd
}