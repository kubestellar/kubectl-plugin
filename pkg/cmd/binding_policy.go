package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// Custom help function for binding-policy command
func bindingPolicyHelpFunc(cmd *cobra.Command, args []string) {
	fmt.Fprintln(cmd.OutOrStdout(), `Create KubeStellar BindingPolicy for multi-cluster resource management.

BindingPolicy provides a declarative way to manage resource deployment across multiple clusters.
It uses label selectors to determine which clusters should receive which resources.

Description:
  A BindingPolicy defines:
  - clusterSelectors: Which clusters should receive resources (by cluster labels)
  - downsync: Which resources should be deployed (by resource labels)

Examples:
  # Create a basic nginx binding policy
  kubectl multi create-binding-policy nginx-policy \
    --cluster-labels="location=edge,env=prod" \
    --resource-labels="app=nginx"

  # Create a binding policy for PostgreSQL with specific namespace
  kubectl multi create-binding-policy postgres-policy \
    --cluster-labels="database=enabled" \
    --resource-labels="app.kubernetes.io/name=postgres" \
    --namespace="postgres-ns"

  # Create a binding policy with multiple cluster selectors
  kubectl multi create-binding-policy multi-env-policy \
    --cluster-labels="env=prod" \
    --cluster-labels="env=staging" \
    --resource-labels="tier=frontend"

Usage:
  kubectl multi create-binding-policy NAME [flags]

Flags:
  --cluster-labels stringArray    Labels to select target clusters (key=value format)
  --resource-labels stringArray   Labels to select resources for deployment (key=value format)
  --namespace string             Target namespace for namespaced resources
  --api-group string             API group for resource selection
  --resource string              Resource type for selection (e.g., deployments, services)
  --dry-run                      Print the policy without applying it`)
}

func newCreateBindingPolicyCommand() *cobra.Command {
	var clusterLabels []string
	var resourceLabels []string
	var targetNamespace string
	var apiGroup string
	var resourceType string
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "create-binding-policy NAME [flags]",
		Short: "Create a KubeStellar BindingPolicy for multi-cluster resource management",
		Long: `Create a KubeStellar BindingPolicy for multi-cluster resource management.

BindingPolicy provides a declarative way to manage resource deployment across multiple clusters.
It uses label selectors to determine which clusters should receive which resources.

This is the recommended approach for multi-cluster deployments instead of direct resource creation.`,
		Example: `# Create a basic nginx binding policy
kubectl multi create-binding-policy nginx-policy \
  --cluster-labels="location=edge,env=prod" \
  --resource-labels="app=nginx"

# Create a binding policy for PostgreSQL
kubectl multi create-binding-policy postgres-policy \
  --cluster-labels="database=enabled" \
  --resource-labels="app.kubernetes.io/name=postgres"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("BindingPolicy name is required")
			}
			if len(clusterLabels) == 0 {
				return fmt.Errorf("at least one --cluster-labels must be specified")
			}
			if len(resourceLabels) == 0 {
				return fmt.Errorf("at least one --resource-labels must be specified")
			}

			return handleCreateBindingPolicy(args[0], clusterLabels, resourceLabels, targetNamespace, apiGroup, resourceType, dryRun)
		},
	}

	cmd.Flags().StringArrayVar(&clusterLabels, "cluster-labels", []string{}, "Labels to select target clusters (key=value format)")
	cmd.Flags().StringArrayVar(&resourceLabels, "resource-labels", []string{}, "Labels to select resources for deployment (key=value format)")
	cmd.Flags().StringVar(&targetNamespace, "namespace", "", "Target namespace for namespaced resources")
	cmd.Flags().StringVar(&apiGroup, "api-group", "", "API group for resource selection")
	cmd.Flags().StringVar(&resourceType, "resource", "", "Resource type for selection (e.g., deployments, services)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the policy without applying it")

	// Set custom help function
	cmd.SetHelpFunc(bindingPolicyHelpFunc)

	return cmd
}

func handleCreateBindingPolicy(name string, clusterLabels, resourceLabels []string, targetNamespace, apiGroup, resourceType string, dryRun bool) error {
	// Parse cluster labels into map
	clusterSelectorLabels := make(map[string]string)
	for _, labelGroup := range clusterLabels {
		// Split comma-separated labels
		labels := strings.Split(labelGroup, ",")
		for _, label := range labels {
			parts := strings.SplitN(strings.TrimSpace(label), "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid cluster label format: %s (expected key=value)", label)
			}
			clusterSelectorLabels[parts[0]] = parts[1]
		}
	}

	// Parse resource labels into map
	resourceSelectorLabels := make(map[string]string)
	for _, labelGroup := range resourceLabels {
		// Split comma-separated labels
		labels := strings.Split(labelGroup, ",")
		for _, label := range labels {
			parts := strings.SplitN(strings.TrimSpace(label), "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid resource label format: %s (expected key=value)", label)
			}
			resourceSelectorLabels[parts[0]] = parts[1]
		}
	}

	// Generate the BindingPolicy YAML
	bindingPolicy := generateBindingPolicy(name, clusterSelectorLabels, resourceSelectorLabels, targetNamespace, apiGroup, resourceType)

	if dryRun {
		fmt.Println("# Generated BindingPolicy (dry-run mode)")
		fmt.Println(bindingPolicy)
		return nil
	}

	fmt.Println("Creating BindingPolicy...")
	fmt.Println(bindingPolicy)
	fmt.Println()
	
	// Show available clusters and their labels
	fmt.Println("=== Available Clusters ===")
	showAvailableClusters()
	fmt.Println()
	
	fmt.Println("To apply this BindingPolicy, save it to a file and run:")
	fmt.Printf("kubectl apply -f binding-policy.yaml\n")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("1. Label your clusters with the specified cluster labels")
	fmt.Println("2. Label your resources with the specified resource labels")
	fmt.Println("3. Apply the BindingPolicy to start multi-cluster deployment")
	fmt.Println()
	fmt.Println("Example cluster labeling:")
	for key, value := range clusterSelectorLabels {
		fmt.Printf("kubectl label managedcluster <cluster-name> %s=%s\n", key, value)
	}
	fmt.Println()
	fmt.Println("Example resource labeling:")
	for key, value := range resourceSelectorLabels {
		fmt.Printf("kubectl label deployment <deployment-name> %s=%s\n", key, value)
	}

	return nil
}

func showAvailableClusters() error {
	_, remoteCtx, _, _, _ := GetGlobalFlags()
	
	// Import the cluster package functionality here
	fmt.Println("Clusters discoverable through KubeStellar:")
	
	// This is a simplified version - in a real implementation you'd use the cluster discovery
	fmt.Printf("Context: %s (remote context for ManagedCluster discovery)\n", remoteCtx)
	fmt.Println("Run 'kubectl multi get nodes' to see all available clusters")
	fmt.Println("Run 'kubectl multi deploy-to --list-clusters' for detailed cluster information")
	
	return nil
}

func generateBindingPolicy(name string, clusterLabels, resourceLabels map[string]string, targetNamespace, apiGroup, resourceType string) string {
	var sb strings.Builder

	// API version and kind
	sb.WriteString("apiVersion: control.kubestellar.io/v1alpha1\n")
	sb.WriteString("kind: BindingPolicy\n")
	sb.WriteString("metadata:\n")
	sb.WriteString(fmt.Sprintf("  name: %s\n", name))
	sb.WriteString("spec:\n")

	// Cluster selectors
	sb.WriteString("  clusterSelectors:\n")
	sb.WriteString("  - matchLabels:\n")
	for key, value := range clusterLabels {
		sb.WriteString(fmt.Sprintf("      \"%s\": \"%s\"\n", key, value))
	}

	// Downsync section
	sb.WriteString("  downsync:\n")
	sb.WriteString("  - objectSelectors:\n")
	sb.WriteString("    - matchLabels:\n")
	for key, value := range resourceLabels {
		sb.WriteString(fmt.Sprintf("        \"%s\": \"%s\"\n", key, value))
	}

	// Add optional fields if specified
	if targetNamespace != "" || apiGroup != "" || resourceType != "" {
		// Remove the last part and add additional selectors
		sb.WriteString("      ")
		if targetNamespace != "" {
			sb.WriteString(fmt.Sprintf("namespace: %s\n      ", targetNamespace))
		}
		if apiGroup != "" {
			sb.WriteString(fmt.Sprintf("apiGroup: %s\n      ", apiGroup))
		}
		if resourceType != "" {
			sb.WriteString(fmt.Sprintf("resource: %s\n", resourceType))
		}
	}

	return sb.String()
}

// Helper function to create a demo binding policy
func newDemoBindingPolicyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "demo-binding-policy",
		Short: "Generate demo BindingPolicy examples",
		Long:  `Generate demo BindingPolicy examples for common use cases.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return showDemoBindingPolicies()
		},
	}

	return cmd
}

func showDemoBindingPolicies() error {
	fmt.Println("# Demo BindingPolicy Examples")
	fmt.Println()

	// Example 1: Nginx deployment
	fmt.Println("## Example 1: Nginx Deployment to Edge Clusters")
	nginxPolicy := `apiVersion: control.kubestellar.io/v1alpha1
kind: BindingPolicy
metadata:
  name: nginx-edge-policy
spec:
  clusterSelectors:
  - matchLabels:
      location: edge
      env: prod
  downsync:
  - objectSelectors:
    - matchLabels:
        app: nginx
        tier: frontend`

	fmt.Println(nginxPolicy)
	fmt.Println()

	// Example 2: Database deployment
	fmt.Println("## Example 2: Database Deployment to Database-Enabled Clusters")
	dbPolicy := `apiVersion: control.kubestellar.io/v1alpha1
kind: BindingPolicy
metadata:
  name: postgres-policy
spec:
  clusterSelectors:
  - matchLabels:
      database: enabled
      storage: ssd
  downsync:
  - objectSelectors:
    - matchLabels:
        app.kubernetes.io/name: postgres
        app.kubernetes.io/component: database`

	fmt.Println(dbPolicy)
	fmt.Println()

	// Example 3: Multi-environment
	fmt.Println("## Example 3: Multi-Environment Deployment")
	multiEnvPolicy := `apiVersion: control.kubestellar.io/v1alpha1
kind: BindingPolicy
metadata:
  name: multi-env-policy
spec:
  clusterSelectors:
  - matchLabels:
      env: staging
  - matchLabels:
      env: prod
  downsync:
  - objectSelectors:
    - matchLabels:
        app: web-frontend
        version: stable`

	fmt.Println(multiEnvPolicy)
	fmt.Println()

	fmt.Println("# Usage Instructions:")
	fmt.Println("1. Save any of the above policies to a .yaml file")
	fmt.Println("2. Apply with: kubectl apply -f binding-policy.yaml")
	fmt.Println("3. Label your clusters and resources accordingly")
	fmt.Println()
	fmt.Println("# Cluster Labeling Examples:")
	fmt.Println("kubectl label cluster edge-cluster-1 location=edge env=prod")
	fmt.Println("kubectl label cluster db-cluster-1 database=enabled storage=ssd")
	fmt.Println()
	fmt.Println("# Resource Labeling Examples:")
	fmt.Println("kubectl label deployment nginx app=nginx tier=frontend")
	fmt.Println("kubectl label deployment postgres app.kubernetes.io/name=postgres app.kubernetes.io/component=database")

	return nil
}