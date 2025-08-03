package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"kubectl-multi/pkg/cluster"
)

// Custom help function for deploy-to command
func deployToHelpFunc(cmd *cobra.Command, args []string) {
	fmt.Fprintln(cmd.OutOrStdout(), `Deploy resources to specific clusters within KubeStellar managed clusters.

This command allows you to target specific clusters for deployment instead of deploying
to all clusters. You can specify clusters by name or by labels.

Description:
  deploy-to provides fine-grained control over which clusters receive your deployments.
  It's an alternative to creating binding policies when you need direct control.

Examples:
  # Deploy to specific clusters by name
  kubectl multi deploy-to --clusters=cluster1,cluster2 -f deployment.yaml

  # Deploy a single resource to named clusters
  kubectl multi deploy-to --clusters=cluster1 deployment nginx --image=nginx

  # Deploy with different images per cluster
  kubectl multi deploy-to --clusters=cluster1,cluster2 \
    --cluster-images="cluster1=nginx:1.20,cluster2=nginx:1.21" \
    deployment nginx

  # Deploy to clusters with global image override
  kubectl multi deploy-to --clusters=cluster1,cluster2 \
    --image=nginx:latest \
    deployment nginx

  # Deploy to clusters with specific labels (requires cluster labeling first)
  kubectl multi deploy-to --cluster-labels="env=prod,location=edge" -f app.yaml

  # List available clusters and their details
  kubectl multi deploy-to --list-clusters

  # Deploy with dry-run to see what would be deployed where
  kubectl multi deploy-to --clusters=cluster1 --dry-run -f deployment.yaml

Usage:
  kubectl multi deploy-to [flags] [RESOURCE_TYPE NAME | -f FILE]

Flags:
  --clusters stringArray         Names of specific clusters to deploy to
  --cluster-labels stringArray   Label selectors for cluster targeting (key=value)
  --list-clusters               List available clusters and their labels
  -f, --filename string         Filename, directory, or URL to files to deploy
  --dry-run                     Show what would be deployed without actually doing it
  --namespace string            Target namespace for deployment
  --image string                Override image for all deployments
  --cluster-images stringArray  Per-cluster image overrides (cluster=image format)`)
}

func newDeployToCommand() *cobra.Command {
	var targetClusters []string
	var clusterLabels []string
	var filename string
	var listClusters bool
	var dryRun bool
	var namespace string
	var image string
	var clusterImages []string

	cmd := &cobra.Command{
		Use:   "deploy-to [flags] [RESOURCE_TYPE NAME | -f FILE]",
		Short: "Deploy resources to specific clusters",
		Long: `Deploy resources to specific clusters within KubeStellar managed clusters.

This command allows you to target specific clusters for deployment instead of deploying
to all clusters. You can specify clusters by name or by labels.`,
		Example: `# Deploy to specific clusters by name
kubectl multi deploy-to --clusters=cluster1,cluster2 -f deployment.yaml

# Deploy a single resource to named clusters
kubectl multi deploy-to --clusters=cluster1 deployment nginx --image=nginx

# List available clusters
kubectl multi deploy-to --list-clusters`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if listClusters {
				return handleListClusters()
			}

			if len(targetClusters) == 0 && len(clusterLabels) == 0 {
				return fmt.Errorf("must specify either --clusters or --cluster-labels")
			}

			if filename == "" && len(args) == 0 {
				return fmt.Errorf("must specify either --filename or resource type and name")
			}

			kubeconfig, remoteCtx, _, ns, allNamespaces := GetGlobalFlags()
			if namespace != "" {
				ns = namespace
			}

			return handleDeployTo(targetClusters, clusterLabels, filename, args, dryRun, image, clusterImages, kubeconfig, remoteCtx, ns, allNamespaces)
		},
	}

	cmd.Flags().StringArrayVar(&targetClusters, "clusters", []string{}, "Names of specific clusters to deploy to")
	cmd.Flags().StringArrayVar(&clusterLabels, "cluster-labels", []string{}, "Label selectors for cluster targeting (key=value)")
	cmd.Flags().StringVarP(&filename, "filename", "f", "", "Filename, directory, or URL to files to deploy")
	cmd.Flags().BoolVar(&listClusters, "list-clusters", false, "List available clusters and their labels")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be deployed without actually doing it")
	cmd.Flags().StringVar(&namespace, "namespace", "", "Target namespace for deployment")
	cmd.Flags().StringVar(&image, "image", "", "Override image for deployments")
	cmd.Flags().StringArrayVar(&clusterImages, "cluster-images", []string{}, "Per-cluster image overrides (cluster=image format)")

	// Set custom help function
	cmd.SetHelpFunc(deployToHelpFunc)

	return cmd
}

func handleListClusters() error {
	kubeconfig, remoteCtx, _, _, _ := GetGlobalFlags()

	clusters, err := cluster.DiscoverClusters(kubeconfig, remoteCtx)
	if err != nil {
		return fmt.Errorf("failed to discover clusters: %v", err)
	}

	if len(clusters) == 0 {
		fmt.Println("No clusters discovered")
		return nil
	}

	fmt.Println("=== Available Clusters ===")
	fmt.Printf("%-20s %-25s %-10s\n", "CLUSTER NAME", "CONTEXT", "STATUS")
	fmt.Printf("%-20s %-25s %-10s\n", "------------", "-------", "------")

	for _, clusterInfo := range clusters {
		status := "Ready"
		if clusterInfo.Client == nil {
			status = "Unreachable"
		}
		fmt.Printf("%-20s %-25s %-10s\n", clusterInfo.Name, clusterInfo.Context, status)
	}

	fmt.Println()
	fmt.Println("Usage examples:")
	fmt.Printf("# Deploy to specific clusters:\n")
	for _, clusterInfo := range clusters {
		if clusterInfo.Client != nil {
			fmt.Printf("kubectl multi deploy-to --clusters=%s -f deployment.yaml\n", clusterInfo.Name)
			break
		}
	}
	fmt.Println()
	fmt.Println("# For better long-term management, consider using binding policies:")
	fmt.Println("kubectl multi create-binding-policy my-policy --cluster-labels=\"env=prod\" --resource-labels=\"app=nginx\"")

	return nil
}

func handleDeployTo(targetClusters, clusterLabels []string, filename string, args []string, dryRun bool, image string, clusterImages []string, kubeconfig, remoteCtx, namespace string, allNamespaces bool) error {
	// Discover all available clusters
	allClusters, err := cluster.DiscoverClusters(kubeconfig, remoteCtx)
	if err != nil {
		return fmt.Errorf("failed to discover clusters: %v", err)
	}

	// Filter clusters based on selection criteria
	selectedClusters := filterClusters(allClusters, targetClusters, clusterLabels)

	if len(selectedClusters) == 0 {
		fmt.Println("No clusters match the selection criteria")
		fmt.Println("Available clusters:")
		for _, c := range allClusters {
			fmt.Printf("  - %s (context: %s)\n", c.Name, c.Context)
		}
		return nil
	}

	// Show what will be deployed
	fmt.Printf("Deploying to %d cluster(s):\n", len(selectedClusters))
	for _, c := range selectedClusters {
		fmt.Printf("  - %s (context: %s)\n", c.Name, c.Context)
	}
	fmt.Println()

	if dryRun {
		fmt.Println("DRY RUN - No actual deployment will occur")
		return nil
	}

	// Build command arguments
	var cmdArgs []string
	if filename != "" {
		cmdArgs = []string{"apply", "-f", filename}
	} else {
		cmdArgs = append([]string{"create"}, args...)
	}

	if namespace != "" {
		cmdArgs = append(cmdArgs, "-n", namespace)
	}
	if allNamespaces {
		cmdArgs = append(cmdArgs, "-A")
	}

	// Parse cluster-specific images
	clusterImageMap := make(map[string]string)
	for _, clusterImage := range clusterImages {
		parts := strings.SplitN(clusterImage, "=", 2)
		if len(parts) == 2 {
			clusterImageMap[parts[0]] = parts[1]
		}
	}

	// Deploy to selected clusters
	for _, clusterInfo := range selectedClusters {
		if clusterInfo.Client == nil {
			fmt.Printf("=== Cluster: %s ===\n", clusterInfo.Name)
			fmt.Println("Error: Cluster is not reachable")
			fmt.Println()
			continue
		}

		fmt.Printf("=== Cluster: %s ===\n", clusterInfo.Name)

		clusterArgs := append([]string{}, cmdArgs...)
		
		// Add cluster-specific image override for create deployment commands
		if len(args) > 0 && args[0] == "deployment" {
			if clusterImg, exists := clusterImageMap[clusterInfo.Name]; exists {
				clusterArgs = append(clusterArgs, "--image="+clusterImg)
				fmt.Printf("Using cluster-specific image: %s\n", clusterImg)
			} else if image != "" {
				clusterArgs = append(clusterArgs, "--image="+image)
				fmt.Printf("Using global image override: %s\n", image)
			}
		}
		
		clusterArgs = append(clusterArgs, "--context", clusterInfo.Context)

		output, err := runKubectl(clusterArgs, kubeconfig)
		if err != nil {
			// Check for common error patterns and provide friendly messages
			if strings.Contains(output, "already exists") {
				fmt.Printf("❌ Resource already exists in this cluster\n")
				fmt.Printf("   Output: %s", output)
			} else if strings.Contains(output, "not found") {
				fmt.Printf("❌ Resource or cluster not accessible\n")
				fmt.Printf("   Output: %s", output)
			} else {
				fmt.Printf("❌ Error: %v\n", err)
				if output != "" {
					fmt.Printf("   Output: %s", output)
				}
			}
		} else {
			fmt.Print(output)
		}
		fmt.Println()
	}

	return nil
}

func filterClusters(allClusters []cluster.ClusterInfo, targetNames, clusterLabels []string) []cluster.ClusterInfo {
	var selected []cluster.ClusterInfo

	// If specific cluster names are provided, filter by name
	if len(targetNames) > 0 {
		nameSet := make(map[string]bool)
		for _, nameList := range targetNames {
			// Split comma-separated cluster names
			names := strings.Split(nameList, ",")
			for _, name := range names {
				nameSet[strings.TrimSpace(name)] = true
			}
		}

		for _, c := range allClusters {
			if nameSet[c.Name] || nameSet[c.Context] {
				selected = append(selected, c)
			}
		}
		return selected
	}

	// If cluster labels are provided, filter by labels
	if len(clusterLabels) > 0 {
		// Parse label selectors
		labelSelectors := make(map[string]string)
		for _, label := range clusterLabels {
			parts := strings.SplitN(label, "=", 2)
			if len(parts) == 2 {
				labelSelectors[parts[0]] = parts[1]
			}
		}

		// Note: In a real implementation, you would check cluster labels
		// For now, this is a placeholder that would need actual cluster label checking
		fmt.Println("Note: Cluster label filtering requires actual cluster labeling implementation")
		fmt.Println("For now, showing all clusters. Use --clusters to specify by name.")
		return allClusters
	}

	return selected
}

// Helper command to show cluster labeling examples
func newClusterLabelingCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "label-cluster-examples",
		Short: "Show examples of how to label clusters for KubeStellar",
		Long:  `Show examples of how to label clusters for use with KubeStellar binding policies.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return showClusterLabelingExamples()
		},
	}

	return cmd
}

func showClusterLabelingExamples() error {
	fmt.Println("# Cluster Labeling Examples for KubeStellar")
	fmt.Println()
	
	fmt.Println("## Label ManagedClusters for targeting:")
	fmt.Println()
	
	fmt.Println("# Production environment clusters")
	fmt.Println("kubectl label managedcluster cluster1 env=prod location=datacenter")
	fmt.Println("kubectl label managedcluster cluster2 env=prod location=edge")
	fmt.Println()
	
	fmt.Println("# Development environment")
	fmt.Println("kubectl label managedcluster dev-cluster env=dev location=datacenter")
	fmt.Println()
	
	fmt.Println("# Special purpose clusters")
	fmt.Println("kubectl label managedcluster db-cluster database=enabled storage=ssd")
	fmt.Println("kubectl label managedcluster gpu-cluster compute=gpu workload=ml")
	fmt.Println()
	
	fmt.Println("## View cluster labels:")
	fmt.Println("kubectl get managedcluster --show-labels")
	fmt.Println()
	
	fmt.Println("## Use labels in binding policies:")
	fmt.Println(`kubectl multi create-binding-policy prod-web-policy \
  --cluster-labels="env=prod,location=edge" \
  --resource-labels="app=nginx,tier=frontend"`)
	fmt.Println()
	
	fmt.Println("## Deploy to labeled clusters:")
	fmt.Println(`kubectl multi deploy-to \
  --cluster-labels="env=prod" \
  -f deployment.yaml`)

	return nil
}