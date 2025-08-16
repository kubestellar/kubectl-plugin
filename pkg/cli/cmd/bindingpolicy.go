package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"kubectl-multi/pkg/core/operations"
)

func newBindingPolicyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "bindingpolicy",
		Aliases: []string{"bp", "policy", "binding"},
		Short:   "Manage KubeStellar binding policies for workload distribution",
		Long: `Manage KubeStellar binding policies for distributing workloads across clusters.

BindingPolicies define which clusters receive workloads and how resources are selected
for distribution. This command supports advanced label-based selection for both
cluster targeting and workload filtering.`,
		Example: `# List all binding policies
kubestellar bindingpolicy list
kubestellar bp list

# Create a basic binding policy
kubestellar bindingpolicy create my-policy --clusters cluster1,cluster2

# Create policy with cluster label selectors
kubestellar bp create web-policy --cluster-labels env=production,tier=web

# Create policy with workload label selectors  
kubestellar policy create app-policy --workload-labels app=nginx,version=latest

# Create advanced policy with match expressions
kubestellar binding create advanced-policy \
  --cluster-match-expression "zone In us-east-1,us-west-1" \
  --workload-match-expression "app Exists"

# Add clusters to existing policy
kubestellar bp add-clusters my-policy cluster3 cluster4

# Remove clusters from policy
kubestellar bp remove-clusters my-policy cluster1

# Update policy with new labels
kubestellar bp update-labels my-policy \
  --cluster-labels "env=staging" \
  --workload-labels "tier=backend"

# Delete a binding policy
kubestellar bp delete my-policy`,
	}

	cmd.AddCommand(newBindingPolicyListCommand())
	cmd.AddCommand(newBindingPolicyCreateCommand())
	cmd.AddCommand(newBindingPolicyDeleteCommand())
	cmd.AddCommand(newBindingPolicyAddClustersCommand())
	cmd.AddCommand(newBindingPolicyRemoveClustersCommand())
	cmd.AddCommand(newBindingPolicyUpdateLabelsCommand())

	return cmd
}

func newBindingPolicyListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls", "get"},
		Short:   "List binding policies",
		Long:    "List all binding policies in the specified namespace or across all namespaces.",
		Example: `# List policies in default namespace
kubestellar bp list

# List policies in specific namespace
kubestellar bp list --namespace production

# List policies across all namespaces
kubestellar bp list --all-namespaces`,
		RunE: func(cmd *cobra.Command, args []string) error {
			kubeconfig, wdsContext, _, namespace, allNamespaces := GetGlobalFlags()
			
			if allNamespaces {
				namespace = ""
			}
			
			return operations.ListBindingPolicies(kubeconfig, wdsContext, namespace)
		},
	}

	return cmd
}

func newBindingPolicyCreateCommand() *cobra.Command {
	var clusterNames []string
	var clusterLabels []string
	var clusterMatchExpressions []string
	var workloadLabels []string
	var workloadMatchExpressions []string
	var resourceTypes []string
	var allClusters bool

	cmd := &cobra.Command{
		Use:     "create NAME",
		Aliases: []string{"new", "add"},
		Short:   "Create a new binding policy",
		Long: `Create a new binding policy with advanced label-based cluster and workload selection.

Supports both simple cluster names and advanced label-based selection using
matchLabels and matchExpressions for fine-grained control over workload distribution.`,
		Example: `# Create basic policy with specific clusters
kubestellar bp create my-policy --clusters cluster1,cluster2

# Create policy for all clusters
kubestellar bp create global-policy --all-clusters

# Create policy with cluster labels
kubestellar bp create prod-policy --cluster-labels env=production,zone=us-east

# Create policy with cluster match expressions
kubestellar bp create advanced-policy \
  --cluster-match-expression "env In production,staging" \
  --cluster-match-expression "zone NotIn us-west-1"

# Create policy with workload labels
kubestellar bp create app-policy \
  --workload-labels app=nginx,tier=frontend \
  --cluster-labels env=production

# Create policy with specific resource types
kubestellar bp create deployment-policy \
  --clusters cluster1,cluster2 \
  --resource-types "apps/v1:Deployment,v1:Service"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			kubeconfig, wdsContext, _, namespace, _ := GetGlobalFlags()

			opts := operations.BindingPolicyOptions{
				Kubeconfig:    kubeconfig,
				WDSContext:    wdsContext,
				PolicyName:    args[0],
				Namespace:     namespace,
				AllClusters:   allClusters,
				Namespaced:    true,
			}

			// Parse cluster names
			if len(clusterNames) > 0 {
				opts.ClusterNames = strings.Split(clusterNames[0], ",")
			}

			// Parse cluster labels
			if len(clusterLabels) > 0 {
				opts.ClusterMatchLabels = parseLabels(clusterLabels)
			}

			// Parse cluster match expressions
			if len(clusterMatchExpressions) > 0 {
				expressions, err := parseMatchExpressions(clusterMatchExpressions)
				if err != nil {
					return fmt.Errorf("failed to parse cluster match expressions: %v", err)
				}
				opts.ClusterMatchExpressions = expressions
			}

			// Parse workload labels
			if len(workloadLabels) > 0 {
				opts.WorkloadMatchLabels = parseLabels(workloadLabels)
			}

			// Parse workload match expressions
			if len(workloadMatchExpressions) > 0 {
				expressions, err := parseMatchExpressions(workloadMatchExpressions)
				if err != nil {
					return fmt.Errorf("failed to parse workload match expressions: %v", err)
				}
				opts.WorkloadMatchExpressions = expressions
			}

			// Parse resource types
			if len(resourceTypes) > 0 {
				specs, err := parseResourceTypes(resourceTypes)
				if err != nil {
					return fmt.Errorf("failed to parse resource types: %v", err)
				}
				opts.ResourceTypes = specs
			}

			// Use advanced creation if labels or expressions are provided
			if len(opts.ClusterMatchLabels) > 0 || len(opts.ClusterMatchExpressions) > 0 ||
				len(opts.WorkloadMatchLabels) > 0 || len(opts.WorkloadMatchExpressions) > 0 ||
				len(opts.ResourceTypes) > 0 {
				return operations.CreateLabelBasedBindingPolicy(opts)
			}

			return operations.CreateBindingPolicy(opts)
		},
	}

	cmd.Flags().StringSliceVar(&clusterNames, "clusters", []string{}, "Comma-separated list of cluster names")
	cmd.Flags().StringSliceVar(&clusterLabels, "cluster-labels", []string{}, "Cluster labels for selection (format: key=value)")
	cmd.Flags().StringSliceVar(&clusterMatchExpressions, "cluster-match-expression", []string{}, "Cluster match expressions (format: 'key operator value1,value2')")
	cmd.Flags().StringSliceVar(&workloadLabels, "workload-labels", []string{}, "Workload labels for selection (format: key=value)")
	cmd.Flags().StringSliceVar(&workloadMatchExpressions, "workload-match-expression", []string{}, "Workload match expressions (format: 'key operator value1,value2')")
	cmd.Flags().StringSliceVar(&resourceTypes, "resource-types", []string{}, "Resource types to include (format: apiVersion:Kind)")
	cmd.Flags().BoolVar(&allClusters, "all-clusters", false, "Target all clusters")

	return cmd
}

func newBindingPolicyDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete NAME",
		Aliases: []string{"rm", "remove"},
		Short:   "Delete a binding policy",
		Long:    "Delete a binding policy from the specified namespace.",
		Example: `# Delete a binding policy
kubestellar bp delete my-policy

# Delete policy in specific namespace
kubestellar bp delete my-policy --namespace production`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			kubeconfig, wdsContext, _, namespace, _ := GetGlobalFlags()
			return operations.DeleteBindingPolicy(kubeconfig, wdsContext, namespace, args[0])
		},
	}

	return cmd
}

func newBindingPolicyAddClustersCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add-clusters POLICY_NAME CLUSTER1 [CLUSTER2...]",
		Aliases: []string{"add", "attach"},
		Short:   "Add clusters to a binding policy",
		Long:    "Add one or more clusters to an existing binding policy.",
		Example: `# Add clusters to a policy
kubestellar bp add-clusters my-policy cluster1 cluster2

# Add cluster in specific namespace
kubestellar bp add-clusters my-policy cluster3 --namespace production`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			kubeconfig, wdsContext, _, namespace, _ := GetGlobalFlags()
			policyName := args[0]
			clusterNames := args[1:]
			return operations.UpdateBindingPolicyWithClusters(kubeconfig, wdsContext, namespace, policyName, clusterNames, []string{})
		},
	}

	return cmd
}

func newBindingPolicyRemoveClustersCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "remove-clusters POLICY_NAME CLUSTER1 [CLUSTER2...]",
		Aliases: []string{"remove", "detach", "rm-clusters"},
		Short:   "Remove clusters from a binding policy",
		Long:    "Remove one or more clusters from an existing binding policy.",
		Example: `# Remove clusters from a policy
kubestellar bp remove-clusters my-policy cluster1 cluster2

# Remove cluster in specific namespace
kubestellar bp remove-clusters my-policy cluster3 --namespace production`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			kubeconfig, wdsContext, _, namespace, _ := GetGlobalFlags()
			policyName := args[0]
			clusterNames := args[1:]
			return operations.UpdateBindingPolicyWithClusters(kubeconfig, wdsContext, namespace, policyName, []string{}, clusterNames)
		},
	}

	return cmd
}

func newBindingPolicyUpdateLabelsCommand() *cobra.Command {
	var clusterLabels []string
	var clusterMatchExpressions []string
	var workloadLabels []string
	var workloadMatchExpressions []string

	cmd := &cobra.Command{
		Use:     "update-labels POLICY_NAME",
		Aliases: []string{"update", "set-labels"},
		Short:   "Update label selectors for a binding policy",
		Long: `Update cluster and workload label selectors for an existing binding policy.

This command allows you to modify the cluster selection criteria and workload
filtering rules for a binding policy without recreating it.`,
		Example: `# Update cluster labels
kubestellar bp update-labels my-policy --cluster-labels env=staging,zone=us-east

# Update workload labels
kubestellar bp update-labels my-policy --workload-labels app=nginx,version=v2

# Update with match expressions
kubestellar bp update-labels my-policy \
  --cluster-match-expression "env In production,staging" \
  --workload-match-expression "tier NotIn deprecated"

# Update both cluster and workload selectors
kubestellar bp update-labels my-policy \
  --cluster-labels "env=production" \
  --workload-labels "app=web-server"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			kubeconfig, wdsContext, _, namespace, _ := GetGlobalFlags()
			policyName := args[0]

			// Parse cluster labels
			var clusterLabelMap map[string]string
			if len(clusterLabels) > 0 {
				clusterLabelMap = parseLabels(clusterLabels)
			}

			// Parse workload labels
			var workloadLabelMap map[string]string
			if len(workloadLabels) > 0 {
				workloadLabelMap = parseLabels(workloadLabels)
			}

			// Parse cluster match expressions
			var clusterExpressions []operations.LabelSelectorRequirement
			if len(clusterMatchExpressions) > 0 {
				expressions, err := parseMatchExpressions(clusterMatchExpressions)
				if err != nil {
					return fmt.Errorf("failed to parse cluster match expressions: %v", err)
				}
				clusterExpressions = expressions
			}

			// Parse workload match expressions
			var workloadExpressions []operations.LabelSelectorRequirement
			if len(workloadMatchExpressions) > 0 {
				expressions, err := parseMatchExpressions(workloadMatchExpressions)
				if err != nil {
					return fmt.Errorf("failed to parse workload match expressions: %v", err)
				}
				workloadExpressions = expressions
			}

			return operations.UpdateBindingPolicyLabels(
				kubeconfig, wdsContext, namespace, policyName,
				clusterLabelMap, workloadLabelMap,
				clusterExpressions, workloadExpressions,
			)
		},
	}

	cmd.Flags().StringSliceVar(&clusterLabels, "cluster-labels", []string{}, "Cluster labels for selection (format: key=value)")
	cmd.Flags().StringSliceVar(&clusterMatchExpressions, "cluster-match-expression", []string{}, "Cluster match expressions (format: 'key operator value1,value2')")
	cmd.Flags().StringSliceVar(&workloadLabels, "workload-labels", []string{}, "Workload labels for selection (format: key=value)")
	cmd.Flags().StringSliceVar(&workloadMatchExpressions, "workload-match-expression", []string{}, "Workload match expressions (format: 'key operator value1,value2')")

	return cmd
}

// Helper functions for parsing labels and expressions
func parseLabels(labelStrings []string) map[string]string {
	labels := make(map[string]string)
	for _, labelString := range labelStrings {
		parts := strings.SplitN(labelString, "=", 2)
		if len(parts) == 2 {
			labels[parts[0]] = parts[1]
		}
	}
	return labels
}

func parseMatchExpressions(expressions []string) ([]operations.LabelSelectorRequirement, error) {
	var requirements []operations.LabelSelectorRequirement
	
	for _, expr := range expressions {
		parts := strings.Fields(expr)
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid match expression format: %s", expr)
		}
		
		key := parts[0]
		operator := parts[1]
		var values []string
		
		if len(parts) > 2 {
			values = strings.Split(parts[2], ",")
		}
		
		requirements = append(requirements, operations.LabelSelectorRequirement{
			Key:      key,
			Operator: operator,
			Values:   values,
		})
	}
	
	return requirements, nil
}

func parseResourceTypes(resourceStrings []string) ([]operations.ResourceSpec, error) {
	var specs []operations.ResourceSpec
	
	for _, resourceString := range resourceStrings {
		parts := strings.SplitN(resourceString, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid resource type format: %s (expected apiVersion:Kind)", resourceString)
		}
		
		apiVersion := parts[0]
		kind := parts[1]
		
		// Determine if resource is namespaced (simplified logic)
		namespaced := true
		if strings.Contains(kind, "Cluster") || kind == "Namespace" || kind == "Node" || 
		   kind == "PersistentVolume" || kind == "ClusterRole" || kind == "ClusterRoleBinding" {
			namespaced = false
		}
		
		specs = append(specs, operations.ResourceSpec{
			APIVersion: apiVersion,
			Kind:       kind,
			Namespaced: namespaced,
		})
	}
	
	return specs, nil
}