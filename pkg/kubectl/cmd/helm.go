package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"kubectl-multi/pkg/cluster"
	"kubectl-multi/pkg/core/operations"
)

func newHelmCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "helm",
		Aliases: []string{"h"},
		Short:   "Manage Helm releases across multiple clusters",
		Long: `Manage Helm releases across all KubeStellar managed clusters.

This command provides multi-cluster Helm operations including install, upgrade,
list, uninstall, and rollback operations that execute across all managed clusters.`,
		Example: `# Install a chart across all clusters
kubectl multi helm install nginx nginx/nginx

# Upgrade a release across all clusters
kubectl multi helm upgrade nginx nginx/nginx --set replicas=3

# List releases across all clusters
kubectl multi helm list

# Uninstall a release from all clusters
kubectl multi helm uninstall nginx

# Rollback a release across all clusters
kubectl multi h rollback nginx 1`,
	}

	cmd.AddCommand(newHelmInstallCommand())
	cmd.AddCommand(newHelmUpgradeCommand())
	cmd.AddCommand(newHelmListCommand())
	cmd.AddCommand(newHelmUninstallCommand())
	cmd.AddCommand(newHelmRollbackCommand())
	cmd.AddCommand(newHelmStatusCommand())
	cmd.AddCommand(newHelmValuesCommand())

	return cmd
}

func newHelmInstallCommand() *cobra.Command {
	var setValues []string
	var valuesFiles []string
	var createNamespace bool
	var wait bool
	var timeout string

	cmd := &cobra.Command{
		Use:     "install RELEASE_NAME CHART",
		Aliases: []string{"i"},
		Short:   "Install a Helm chart across all managed clusters",
		Long: `Install a Helm chart across all KubeStellar managed clusters.

The chart will be installed with the same release name and configuration
on all discovered clusters.`,
		Example: `# Install nginx chart
kubectl multi helm install nginx nginx/nginx

# Install with custom values
kubectl multi helm install webapp ./charts/webapp \
  --set image.tag=v2.0 \
  --set replicas=3

# Install with values file
kubectl multi helm install app ./chart --values values.yaml

# Install and create namespace
kubectl multi h install monitoring prometheus/kube-prometheus-stack \
  --namespace monitoring --create-namespace`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			kubeconfig, remoteCtx, _, namespace, _ := GetGlobalFlags()

			// Discover clusters
			clusters, err := cluster.DiscoverClusters(kubeconfig, remoteCtx)
			if err != nil {
				return fmt.Errorf("failed to discover clusters: %v", err)
			}

			// Parse set values
			variables := make(map[string]string)
			for _, setValue := range setValues {
				parts := strings.SplitN(setValue, "=", 2)
				if len(parts) == 2 {
					variables[parts[0]] = parts[1]
				}
			}

			opts := operations.HelmOptions{
				Clusters:        clusters,
				ReleaseName:     args[0],
				ChartName:       args[1],
				Variables:       variables,
				ValuesFiles:     valuesFiles,
				Namespace:       namespace,
				CreateNamespace: createNamespace,
				Wait:           wait,
				Timeout:        timeout,
			}

			return operations.HelmInstall(opts)
		},
	}

	cmd.Flags().StringSliceVar(&setValues, "set", []string{}, "Set values for the chart (format: key=value)")
	cmd.Flags().StringSliceVarP(&valuesFiles, "values", "f", []string{}, "Values files to use")
	cmd.Flags().BoolVar(&createNamespace, "create-namespace", false, "Create the release namespace if not present")
	cmd.Flags().BoolVar(&wait, "wait", false, "Wait for the installation to complete")
	cmd.Flags().StringVar(&timeout, "timeout", "5m0s", "Time to wait for completion")

	return cmd
}

func newHelmUpgradeCommand() *cobra.Command {
	var setValues []string
	var valuesFiles []string
	var install bool
	var wait bool
	var timeout string

	cmd := &cobra.Command{
		Use:     "upgrade RELEASE_NAME CHART",
		Aliases: []string{"up"},
		Short:   "Upgrade a Helm release across all managed clusters",
		Long: `Upgrade a Helm release across all KubeStellar managed clusters.

The release will be upgraded with the same configuration on all clusters
where it exists.`,
		Example: `# Upgrade a release
kubectl multi helm upgrade nginx nginx/nginx

# Upgrade with new values
kubectl multi helm upgrade webapp ./charts/webapp \
  --set image.tag=v2.1 \
  --set replicas=5

# Upgrade or install if not exists
kubectl multi h upgrade app ./chart --install`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			kubeconfig, remoteCtx, _, namespace, _ := GetGlobalFlags()

			// Discover clusters
			clusters, err := cluster.DiscoverClusters(kubeconfig, remoteCtx)
			if err != nil {
				return fmt.Errorf("failed to discover clusters: %v", err)
			}

			// Parse set values
			variables := make(map[string]string)
			for _, setValue := range setValues {
				parts := strings.SplitN(setValue, "=", 2)
				if len(parts) == 2 {
					variables[parts[0]] = parts[1]
				}
			}

			opts := operations.HelmOptions{
				Clusters:     clusters,
				ReleaseName:  args[0],
				ChartName:    args[1],
				Variables:    variables,
				ValuesFiles:  valuesFiles,
				Namespace:    namespace,
				Install:      install,
				Wait:        wait,
				Timeout:     timeout,
			}

			return operations.HelmUpgrade(opts)
		},
	}

	cmd.Flags().StringSliceVar(&setValues, "set", []string{}, "Set values for the chart (format: key=value)")
	cmd.Flags().StringSliceVarP(&valuesFiles, "values", "f", []string{}, "Values files to use")
	cmd.Flags().BoolVarP(&install, "install", "i", false, "Install the release if it does not exist")
	cmd.Flags().BoolVar(&wait, "wait", false, "Wait for the upgrade to complete")
	cmd.Flags().StringVar(&timeout, "timeout", "5m0s", "Time to wait for completion")

	return cmd
}

func newHelmListCommand() *cobra.Command {
	var allNamespaces bool
	var deployed bool
	var failed bool
	var pending bool

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls", "l"},
		Short:   "List Helm releases across all managed clusters",
		Long: `List Helm releases across all KubeStellar managed clusters.

Shows release information including name, namespace, revision, status,
chart, and app version for each cluster.`,
		Example: `# List all releases
kubectl multi helm list

# List releases in all namespaces
kubectl multi helm list --all-namespaces

# List only deployed releases
kubectl multi h list --deployed

# List failed releases
kubectl multi h ls --failed`,
		RunE: func(cmd *cobra.Command, args []string) error {
			kubeconfig, remoteCtx, _, namespace, _ := GetGlobalFlags()

			// Discover clusters
			clusters, err := cluster.DiscoverClusters(kubeconfig, remoteCtx)
			if err != nil {
				return fmt.Errorf("failed to discover clusters: %v", err)
			}

			opts := operations.HelmOptions{
				Clusters:      clusters,
				Namespace:     namespace,
				AllNamespaces: allNamespaces,
				Deployed:      deployed,
				Failed:        failed,
				Pending:       pending,
			}

			return operations.HelmList(opts)
		},
	}

	cmd.Flags().BoolVarP(&allNamespaces, "all-namespaces", "A", false, "List releases across all namespaces")
	cmd.Flags().BoolVar(&deployed, "deployed", false, "Show deployed releases")
	cmd.Flags().BoolVar(&failed, "failed", false, "Show failed releases")
	cmd.Flags().BoolVar(&pending, "pending", false, "Show pending releases")

	return cmd
}

func newHelmUninstallCommand() *cobra.Command {
	var keepHistory bool

	cmd := &cobra.Command{
		Use:     "uninstall RELEASE_NAME",
		Aliases: []string{"del", "delete", "un"},
		Short:   "Uninstall a Helm release from all managed clusters",
		Long: `Uninstall a Helm release from all KubeStellar managed clusters.

The release will be removed from all clusters where it exists.`,
		Example: `# Uninstall a release
kubectl multi helm uninstall nginx

# Uninstall and keep history
kubectl multi helm uninstall webapp --keep-history

# Short forms
kubectl multi h del old-app
kubectl multi h un deprecated-service`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			kubeconfig, remoteCtx, _, namespace, _ := GetGlobalFlags()

			// Discover clusters
			clusters, err := cluster.DiscoverClusters(kubeconfig, remoteCtx)
			if err != nil {
				return fmt.Errorf("failed to discover clusters: %v", err)
			}

			opts := operations.HelmOptions{
				Clusters:    clusters,
				ReleaseName: args[0],
				Namespace:   namespace,
				KeepHistory: keepHistory,
			}

			return operations.HelmUninstall(opts)
		},
	}

	cmd.Flags().BoolVar(&keepHistory, "keep-history", false, "Keep release history")

	return cmd
}

func newHelmRollbackCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "rollback RELEASE_NAME [REVISION]",
		Aliases: []string{"rb"},
		Short:   "Rollback a Helm release across all managed clusters",
		Long: `Rollback a Helm release to a previous revision across all KubeStellar managed clusters.

If no revision is specified, rollback to the previous revision.`,
		Example: `# Rollback to previous revision
kubectl multi helm rollback nginx

# Rollback to specific revision
kubectl multi helm rollback webapp 3

# Short form
kubectl multi h rb app 2`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			kubeconfig, remoteCtx, _, namespace, _ := GetGlobalFlags()

			// Discover clusters
			clusters, err := cluster.DiscoverClusters(kubeconfig, remoteCtx)
			if err != nil {
				return fmt.Errorf("failed to discover clusters: %v", err)
			}

			revision := ""
			if len(args) > 1 {
				revision = args[1]
			}

			opts := operations.HelmOptions{
				Clusters:    clusters,
				ReleaseName: args[0],
				Namespace:   namespace,
				Revision:    revision,
			}

			return operations.HelmRollback(opts)
		},
	}

	return cmd
}

func newHelmStatusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "status RELEASE_NAME",
		Aliases: []string{"stat"},
		Short:   "Show status of a Helm release across all managed clusters",
		Long: `Show the status of a Helm release across all KubeStellar managed clusters.

Displays release information, resources, and notes for each cluster.`,
		Example: `# Show release status
kubectl multi helm status nginx

# Short form
kubectl multi h stat webapp`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			kubeconfig, remoteCtx, _, namespace, _ := GetGlobalFlags()

			// Discover clusters
			clusters, err := cluster.DiscoverClusters(kubeconfig, remoteCtx)
			if err != nil {
				return fmt.Errorf("failed to discover clusters: %v", err)
			}

			opts := operations.HelmOptions{
				Clusters:    clusters,
				ReleaseName: args[0],
				Namespace:   namespace,
			}

			return operations.HelmStatus(opts)
		},
	}

	return cmd
}

func newHelmValuesCommand() *cobra.Command {
	var allValues bool
	var revision string

	cmd := &cobra.Command{
		Use:     "values RELEASE_NAME",
		Aliases: []string{"get-values", "vals"},
		Short:   "Show values of a Helm release across all managed clusters",
		Long: `Show the values of a Helm release across all KubeStellar managed clusters.

Displays the computed values for the release on each cluster.`,
		Example: `# Show user-supplied values
kubectl multi helm values nginx

# Show all values (computed)
kubectl multi helm values webapp --all

# Show values for specific revision
kubectl multi h vals app --revision 2`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			kubeconfig, remoteCtx, _, namespace, _ := GetGlobalFlags()

			// Discover clusters
			clusters, err := cluster.DiscoverClusters(kubeconfig, remoteCtx)
			if err != nil {
				return fmt.Errorf("failed to discover clusters: %v", err)
			}

			opts := operations.HelmOptions{
				Clusters:    clusters,
				ReleaseName: args[0],
				Namespace:   namespace,
				AllValues:   allValues,
				Revision:    revision,
			}

			return operations.HelmGetValues(opts)
		},
	}

	cmd.Flags().BoolVarP(&allValues, "all", "a", false, "Show all values (computed)")
	cmd.Flags().StringVar(&revision, "revision", "", "Get values for specific revision")

	return cmd
}