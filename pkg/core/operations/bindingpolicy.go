package operations

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
)

// BindingPolicyGVR defines the GroupVersionResource for KubeStellar BindingPolicy
var BindingPolicyGVR = schema.GroupVersionResource{
	Group:    "control.kubestellar.io",
	Version:  "v1alpha1",
	Resource: "bindingpolicies",
}

// BindingPolicyOptions contains options for binding policy operations
type BindingPolicyOptions struct {
	Kubeconfig              string
	WDSContext              string // WDS (Workload Description Space) context
	PolicyName              string
	Namespace               string
	ClusterNames            []string // List of clusters to bind
	ClusterMatchLabels      map[string]string
	ClusterMatchExpressions []LabelSelectorRequirement
	WorkloadMatchLabels     map[string]string
	WorkloadMatchExpressions []LabelSelectorRequirement
	ResourceTypes           []ResourceSpec
	Namespaced              bool
	AllClusters             bool
}

// LabelSelectorRequirement represents a label selector requirement
type LabelSelectorRequirement struct {
	Key      string
	Operator string // In, NotIn, Exists, DoesNotExist, Gt, Lt
	Values   []string
}

// ResourceSpec represents a resource type specification for downsync
type ResourceSpec struct {
	APIVersion  string
	Kind        string
	Namespaced  bool
	NamePattern string // Optional name pattern (e.g., "nginx-*")
	LabelSelector map[string]string // Optional label selector for workloads
}

// CreateBindingPolicy creates a new binding policy to distribute workloads to specified clusters
func CreateBindingPolicy(opts BindingPolicyOptions) error {
	// Load kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", opts.Kubeconfig)
	if err != nil {
		return fmt.Errorf("failed to build config: %v", err)
	}

	// Create dynamic client
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create dynamic client: %v", err)
	}

	// Build the binding policy object
	bindingPolicy := buildBindingPolicy(opts)

	// Create the binding policy
	created, err := dynamicClient.Resource(BindingPolicyGVR).
		Namespace(opts.Namespace).
		Create(context.TODO(), bindingPolicy, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create binding policy: %v", err)
	}

	fmt.Printf("BindingPolicy '%s' created successfully\n", created.GetName())
	return nil
}

// buildBindingPolicy constructs the binding policy unstructured object
func buildBindingPolicy(opts BindingPolicyOptions) *unstructured.Unstructured {
	bindingPolicy := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "control.kubestellar.io/v1alpha1",
			"kind":       "BindingPolicy",
			"metadata": map[string]interface{}{
				"name":      opts.PolicyName,
				"namespace": opts.Namespace,
			},
			"spec": map[string]interface{}{
				"clusterSelectors": buildClusterSelectors(opts),
				"downsync":         buildDownsyncSpec(opts),
			},
		},
	}

	return bindingPolicy
}

// buildClusterSelectors creates the cluster selector specification with advanced label support
func buildClusterSelectors(opts BindingPolicyOptions) []interface{} {
	selectors := []interface{}{}

	if opts.AllClusters {
		// Select all clusters
		selectors = append(selectors, map[string]interface{}{
			"matchLabels": map[string]interface{}{},
		})
	} else if len(opts.ClusterNames) > 0 {
		// Select specific clusters by name
		for _, clusterName := range opts.ClusterNames {
			selectors = append(selectors, map[string]interface{}{
				"matchLabels": map[string]interface{}{
					"name": clusterName,
				},
			})
		}
	} else {
		// Build advanced label-based selector
		selector := map[string]interface{}{}
		
		// Add matchLabels if provided
		if len(opts.ClusterMatchLabels) > 0 {
			selector["matchLabels"] = opts.ClusterMatchLabels
		}
		
		// Add matchExpressions if provided
		if len(opts.ClusterMatchExpressions) > 0 {
			expressions := []interface{}{}
			for _, expr := range opts.ClusterMatchExpressions {
				expressions = append(expressions, map[string]interface{}{
					"key":      expr.Key,
					"operator": expr.Operator,
					"values":   expr.Values,
				})
			}
			selector["matchExpressions"] = expressions
		}
		
		// Only add selector if it has content
		if len(selector) > 0 {
			selectors = append(selectors, selector)
		}
	}

	return selectors
}

// buildDownsyncSpec creates the downsync specification for resources with label-based workload selection
func buildDownsyncSpec(opts BindingPolicyOptions) []interface{} {
	downsync := []interface{}{}

	// Use custom resource types if provided
	if len(opts.ResourceTypes) > 0 {
		for _, resourceSpec := range opts.ResourceTypes {
			resource := map[string]interface{}{
				"apiVersion": resourceSpec.APIVersion,
				"kind":       resourceSpec.Kind,
				"namespaced": resourceSpec.Namespaced,
			}
			
			// Add name pattern if specified
			if resourceSpec.NamePattern != "" {
				resource["name"] = resourceSpec.NamePattern
			}
			
			// Add label selector if specified
			if len(resourceSpec.LabelSelector) > 0 {
				resource["labelSelector"] = map[string]interface{}{
					"matchLabels": resourceSpec.LabelSelector,
				}
			}
			
			// Add namespace if specified and resource is namespaced
			if opts.Namespaced && resourceSpec.Namespaced {
				resource["namespace"] = opts.Namespace
			}
			
			downsync = append(downsync, resource)
		}
	} else {
		// Use default common Kubernetes resources with workload label selectors
		commonResources := []map[string]interface{}{
			{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"namespaced": true,
			},
			{
				"apiVersion": "v1",
				"kind":       "Service",
				"namespaced": true,
			},
			{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"namespaced": true,
			},
			{
				"apiVersion": "v1",
				"kind":       "Secret",
				"namespaced": true,
			},
			{
				"apiVersion": "apps/v1",
				"kind":       "StatefulSet",
				"namespaced": true,
			},
			{
				"apiVersion": "apps/v1",
				"kind":       "DaemonSet",
				"namespaced": true,
			},
		}

		for _, resource := range commonResources {
			// Add workload label selector if provided
			if len(opts.WorkloadMatchLabels) > 0 {
				labelSelector := map[string]interface{}{
					"matchLabels": opts.WorkloadMatchLabels,
				}
				
				// Add match expressions if provided
				if len(opts.WorkloadMatchExpressions) > 0 {
					expressions := []interface{}{}
					for _, expr := range opts.WorkloadMatchExpressions {
						expressions = append(expressions, map[string]interface{}{
							"key":      expr.Key,
							"operator": expr.Operator,
							"values":   expr.Values,
						})
					}
					labelSelector["matchExpressions"] = expressions
				}
				
				resource["labelSelector"] = labelSelector
			}
			
			// Add namespace if specified and resource is namespaced
			if opts.Namespaced {
				resource["namespace"] = opts.Namespace
			}
			
			downsync = append(downsync, resource)
		}
	}

	return downsync
}

// ListBindingPolicies lists all binding policies in the WDS
func ListBindingPolicies(kubeconfig, wdsContext, namespace string) error {
	// Build config with WDS context
	configOverrides := &clientcmd.ConfigOverrides{}
	if wdsContext != "" {
		configOverrides.CurrentContext = wdsContext
	}

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if kubeconfig != "" {
		loadingRules.ExplicitPath = kubeconfig
	}

	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	config, err := clientConfig.ClientConfig()
	if err != nil {
		return fmt.Errorf("failed to build config: %v", err)
	}

	// Create dynamic client
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create dynamic client: %v", err)
	}

	// List binding policies
	var policies *unstructured.UnstructuredList
	if namespace != "" {
		policies, err = dynamicClient.Resource(BindingPolicyGVR).
			Namespace(namespace).
			List(context.TODO(), metav1.ListOptions{})
	} else {
		policies, err = dynamicClient.Resource(BindingPolicyGVR).
			List(context.TODO(), metav1.ListOptions{})
	}

	if err != nil {
		return fmt.Errorf("failed to list binding policies: %v", err)
	}

	// Display policies
	fmt.Printf("NAMESPACE\tNAME\tCLUSTERS\n")
	for _, policy := range policies.Items {
		ns := policy.GetNamespace()
		if ns == "" {
			ns = "default"
		}
		
		// Extract cluster info from spec
		clusterInfo := extractClusterInfo(policy.Object)
		fmt.Printf("%s\t%s\t%s\n", ns, policy.GetName(), clusterInfo)
	}

	return nil
}

// extractClusterInfo extracts cluster selector information from binding policy
func extractClusterInfo(policy map[string]interface{}) string {
	spec, ok := policy["spec"].(map[string]interface{})
	if !ok {
		return "unknown"
	}

	selectors, ok := spec["clusterSelectors"].([]interface{})
	if !ok || len(selectors) == 0 {
		return "none"
	}

	// Check if it's selecting all clusters
	if len(selectors) == 1 {
		selector := selectors[0].(map[string]interface{})
		if matchLabels, ok := selector["matchLabels"].(map[string]interface{}); ok {
			if len(matchLabels) == 0 {
				return "all"
			}
		}
	}

	return fmt.Sprintf("%d selector(s)", len(selectors))
}

// DeleteBindingPolicy deletes a binding policy
func DeleteBindingPolicy(kubeconfig, wdsContext, namespace, policyName string) error {
	// Build config with WDS context
	configOverrides := &clientcmd.ConfigOverrides{}
	if wdsContext != "" {
		configOverrides.CurrentContext = wdsContext
	}

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if kubeconfig != "" {
		loadingRules.ExplicitPath = kubeconfig
	}

	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	config, err := clientConfig.ClientConfig()
	if err != nil {
		return fmt.Errorf("failed to build config: %v", err)
	}

	// Create dynamic client
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create dynamic client: %v", err)
	}

	// Delete the binding policy
	err = dynamicClient.Resource(BindingPolicyGVR).
		Namespace(namespace).
		Delete(context.TODO(), policyName, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete binding policy: %v", err)
	}

	fmt.Printf("BindingPolicy '%s' deleted successfully\n", policyName)
	return nil
}

// UpdateBindingPolicyWithClusters updates a binding policy to add or remove clusters
func UpdateBindingPolicyWithClusters(kubeconfig, wdsContext, namespace, policyName string, addClusters, removeClusters []string) error {
	// Build config with WDS context
	configOverrides := &clientcmd.ConfigOverrides{}
	if wdsContext != "" {
		configOverrides.CurrentContext = wdsContext
	}

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if kubeconfig != "" {
		loadingRules.ExplicitPath = kubeconfig
	}

	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	config, err := clientConfig.ClientConfig()
	if err != nil {
		return fmt.Errorf("failed to build config: %v", err)
	}

	// Create dynamic client
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create dynamic client: %v", err)
	}

	// Get existing binding policy
	policy, err := dynamicClient.Resource(BindingPolicyGVR).
		Namespace(namespace).
		Get(context.TODO(), policyName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get binding policy: %v", err)
	}

	// Update cluster selectors
	spec, ok := policy.Object["spec"].(map[string]interface{})
	if !ok {
		spec = map[string]interface{}{}
		policy.Object["spec"] = spec
	}

	selectors, ok := spec["clusterSelectors"].([]interface{})
	if !ok {
		selectors = []interface{}{}
	}

	// Add new clusters
	for _, clusterName := range addClusters {
		found := false
		for _, sel := range selectors {
			if selector, ok := sel.(map[string]interface{}); ok {
				if matchLabels, ok := selector["matchLabels"].(map[string]interface{}); ok {
					if name, ok := matchLabels["name"].(string); ok && name == clusterName {
						found = true
						break
					}
				}
			}
		}
		if !found {
			selectors = append(selectors, map[string]interface{}{
				"matchLabels": map[string]interface{}{
					"name": clusterName,
				},
			})
		}
	}

	// Remove clusters (filter out matching selectors)
	if len(removeClusters) > 0 {
		newSelectors := []interface{}{}
		for _, sel := range selectors {
			if selector, ok := sel.(map[string]interface{}); ok {
				if matchLabels, ok := selector["matchLabels"].(map[string]interface{}); ok {
					if name, ok := matchLabels["name"].(string); ok {
						// Check if this cluster should be removed
						shouldRemove := false
						for _, removeCluster := range removeClusters {
							if name == removeCluster {
								shouldRemove = true
								break
							}
						}
						if !shouldRemove {
							newSelectors = append(newSelectors, sel)
						}
					} else {
						// Keep non-name based selectors
						newSelectors = append(newSelectors, sel)
					}
				}
			}
		}
		selectors = newSelectors
	}

	spec["clusterSelectors"] = selectors

	// Update the binding policy
	updated, err := dynamicClient.Resource(BindingPolicyGVR).
		Namespace(namespace).
		Update(context.TODO(), policy, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update binding policy: %v", err)
	}

	fmt.Printf("BindingPolicy '%s' updated successfully\n", updated.GetName())
	return nil
}

// CreateLabelBasedBindingPolicy creates a binding policy using label-based cluster and workload selection
func CreateLabelBasedBindingPolicy(opts BindingPolicyOptions) error {
	// Load kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", opts.Kubeconfig)
	if err != nil {
		return fmt.Errorf("failed to build config: %v", err)
	}

	// Create dynamic client
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create dynamic client: %v", err)
	}

	// Build the binding policy object with label-based selectors
	bindingPolicy := buildAdvancedBindingPolicy(opts)

	// Create the binding policy
	created, err := dynamicClient.Resource(BindingPolicyGVR).
		Namespace(opts.Namespace).
		Create(context.TODO(), bindingPolicy, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create binding policy: %v", err)
	}

	fmt.Printf("Label-based BindingPolicy '%s' created successfully\n", created.GetName())
	fmt.Printf("Cluster selectors: %d label-based selector(s)\n", len(opts.ClusterMatchExpressions)+1)
	if len(opts.WorkloadMatchLabels) > 0 || len(opts.WorkloadMatchExpressions) > 0 {
		fmt.Printf("Workload selectors: label-based workload filtering enabled\n")
	}
	return nil
}

// buildAdvancedBindingPolicy constructs a binding policy with advanced label selectors
func buildAdvancedBindingPolicy(opts BindingPolicyOptions) *unstructured.Unstructured {
	bindingPolicy := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "control.kubestellar.io/v1alpha1",
			"kind":       "BindingPolicy",
			"metadata": map[string]interface{}{
				"name":      opts.PolicyName,
				"namespace": opts.Namespace,
				"labels": map[string]interface{}{
					"managed-by": "kubestellar-cli",
					"version":    "v1alpha1",
				},
				"annotations": map[string]interface{}{
					"kubestellar.io/created-by": "kubestellar-cli",
					"kubestellar.io/description": "Label-based binding policy for workload distribution",
				},
			},
			"spec": map[string]interface{}{
				"clusterSelectors": buildAdvancedClusterSelectors(opts),
				"downsync":         buildAdvancedDownsyncSpec(opts),
			},
		},
	}

	return bindingPolicy
}

// buildAdvancedClusterSelectors creates cluster selectors with full label expression support
func buildAdvancedClusterSelectors(opts BindingPolicyOptions) []interface{} {
	selectors := []interface{}{}

	if opts.AllClusters {
		// Select all WEC clusters (exclude WDS clusters)
		selectors = append(selectors, map[string]interface{}{
			"matchLabels": map[string]interface{}{
				"type": "wec",
			},
		})
	} else if len(opts.ClusterNames) > 0 {
		// Select specific clusters by name
		for _, clusterName := range opts.ClusterNames {
			selectors = append(selectors, map[string]interface{}{
				"matchLabels": map[string]interface{}{
					"name": clusterName,
					"type": "wec", // Ensure we only target WEC clusters
				},
			})
		}
	} else {
		// Build comprehensive label-based selector
		selector := map[string]interface{}{}
		
		// Start with basic cluster type filtering
		matchLabels := map[string]interface{}{
			"type": "wec", // Always ensure we target WEC clusters
		}
		
		// Add custom match labels
		if len(opts.ClusterMatchLabels) > 0 {
			for k, v := range opts.ClusterMatchLabels {
				matchLabels[k] = v
			}
		}
		selector["matchLabels"] = matchLabels
		
		// Add match expressions if provided
		if len(opts.ClusterMatchExpressions) > 0 {
			expressions := []interface{}{}
			for _, expr := range opts.ClusterMatchExpressions {
				expressions = append(expressions, map[string]interface{}{
					"key":      expr.Key,
					"operator": expr.Operator,
					"values":   expr.Values,
				})
			}
			selector["matchExpressions"] = expressions
		}
		
		selectors = append(selectors, selector)
	}

	return selectors
}

// buildAdvancedDownsyncSpec creates downsync specifications with workload label filtering
func buildAdvancedDownsyncSpec(opts BindingPolicyOptions) []interface{} {
	downsync := []interface{}{}

	// Use custom resource types if provided
	if len(opts.ResourceTypes) > 0 {
		for _, resourceSpec := range opts.ResourceTypes {
			resource := map[string]interface{}{
				"apiVersion": resourceSpec.APIVersion,
				"kind":       resourceSpec.Kind,
				"namespaced": resourceSpec.Namespaced,
			}
			
			// Add name pattern if specified
			if resourceSpec.NamePattern != "" {
				resource["name"] = resourceSpec.NamePattern
			}
			
			// Build comprehensive label selector
			if len(resourceSpec.LabelSelector) > 0 || len(opts.WorkloadMatchLabels) > 0 || len(opts.WorkloadMatchExpressions) > 0 {
				labelSelector := map[string]interface{}{}
				
				// Combine resource-specific and global workload labels
				matchLabels := map[string]interface{}{}
				for k, v := range opts.WorkloadMatchLabels {
					matchLabels[k] = v
				}
				for k, v := range resourceSpec.LabelSelector {
					matchLabels[k] = v
				}
				
				if len(matchLabels) > 0 {
					labelSelector["matchLabels"] = matchLabels
				}
				
				// Add workload match expressions
				if len(opts.WorkloadMatchExpressions) > 0 {
					expressions := []interface{}{}
					for _, expr := range opts.WorkloadMatchExpressions {
						expressions = append(expressions, map[string]interface{}{
							"key":      expr.Key,
							"operator": expr.Operator,
							"values":   expr.Values,
						})
					}
					labelSelector["matchExpressions"] = expressions
				}
				
				resource["labelSelector"] = labelSelector
			}
			
			// Add namespace targeting
			if opts.Namespaced && resourceSpec.Namespaced {
				resource["namespace"] = opts.Namespace
			}
			
			downsync = append(downsync, resource)
		}
	} else {
		// Use enhanced default resources with workload filtering
		defaultResources := []ResourceSpec{
			{APIVersion: "apps/v1", Kind: "Deployment", Namespaced: true},
			{APIVersion: "v1", Kind: "Service", Namespaced: true},
			{APIVersion: "v1", Kind: "ConfigMap", Namespaced: true},
			{APIVersion: "v1", Kind: "Secret", Namespaced: true},
			{APIVersion: "apps/v1", Kind: "StatefulSet", Namespaced: true},
			{APIVersion: "apps/v1", Kind: "DaemonSet", Namespaced: true},
			{APIVersion: "batch/v1", Kind: "Job", Namespaced: true},
			{APIVersion: "batch/v1", Kind: "CronJob", Namespaced: true},
		}

		for _, resourceSpec := range defaultResources {
			resource := map[string]interface{}{
				"apiVersion": resourceSpec.APIVersion,
				"kind":       resourceSpec.Kind,
				"namespaced": resourceSpec.Namespaced,
			}
			
			// Add workload label selector if provided
			if len(opts.WorkloadMatchLabels) > 0 || len(opts.WorkloadMatchExpressions) > 0 {
				labelSelector := map[string]interface{}{}
				
				if len(opts.WorkloadMatchLabels) > 0 {
					labelSelector["matchLabels"] = opts.WorkloadMatchLabels
				}
				
				if len(opts.WorkloadMatchExpressions) > 0 {
					expressions := []interface{}{}
					for _, expr := range opts.WorkloadMatchExpressions {
						expressions = append(expressions, map[string]interface{}{
							"key":      expr.Key,
							"operator": expr.Operator,
							"values":   expr.Values,
						})
					}
					labelSelector["matchExpressions"] = expressions
				}
				
				resource["labelSelector"] = labelSelector
			}
			
			// Add namespace targeting
			if opts.Namespaced {
				resource["namespace"] = opts.Namespace
			}
			
			downsync = append(downsync, resource)
		}
	}

	return downsync
}

// UpdateBindingPolicyLabels updates cluster and workload label selectors for a binding policy
func UpdateBindingPolicyLabels(kubeconfig, wdsContext, namespace, policyName string, 
	clusterLabels, workloadLabels map[string]string,
	clusterExpressions, workloadExpressions []LabelSelectorRequirement) error {
	
	// Build config with WDS context
	configOverrides := &clientcmd.ConfigOverrides{}
	if wdsContext != "" {
		configOverrides.CurrentContext = wdsContext
	}

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if kubeconfig != "" {
		loadingRules.ExplicitPath = kubeconfig
	}

	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	config, err := clientConfig.ClientConfig()
	if err != nil {
		return fmt.Errorf("failed to build config: %v", err)
	}

	// Create dynamic client
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create dynamic client: %v", err)
	}

	// Get existing binding policy
	policy, err := dynamicClient.Resource(BindingPolicyGVR).
		Namespace(namespace).
		Get(context.TODO(), policyName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get binding policy: %v", err)
	}

	// Update cluster selectors with new labels
	spec, ok := policy.Object["spec"].(map[string]interface{})
	if !ok {
		spec = map[string]interface{}{}
		policy.Object["spec"] = spec
	}

	// Update cluster selectors
	if len(clusterLabels) > 0 || len(clusterExpressions) > 0 {
		selectors := []interface{}{}
		selector := map[string]interface{}{}
		
		// Ensure we always target WEC clusters
		matchLabels := map[string]interface{}{
			"type": "wec",
		}
		for k, v := range clusterLabels {
			matchLabels[k] = v
		}
		selector["matchLabels"] = matchLabels
		
		// Add match expressions
		if len(clusterExpressions) > 0 {
			expressions := []interface{}{}
			for _, expr := range clusterExpressions {
				expressions = append(expressions, map[string]interface{}{
					"key":      expr.Key,
					"operator": expr.Operator,
					"values":   expr.Values,
				})
			}
			selector["matchExpressions"] = expressions
		}
		
		selectors = append(selectors, selector)
		spec["clusterSelectors"] = selectors
	}

	// Update downsync with workload label selectors
	if len(workloadLabels) > 0 || len(workloadExpressions) > 0 {
		if downsync, ok := spec["downsync"].([]interface{}); ok {
			for _, item := range downsync {
				if resource, ok := item.(map[string]interface{}); ok {
					labelSelector := map[string]interface{}{}
					
					if len(workloadLabels) > 0 {
						labelSelector["matchLabels"] = workloadLabels
					}
					
					if len(workloadExpressions) > 0 {
						expressions := []interface{}{}
						for _, expr := range workloadExpressions {
							expressions = append(expressions, map[string]interface{}{
								"key":      expr.Key,
								"operator": expr.Operator,
								"values":   expr.Values,
							})
						}
						labelSelector["matchExpressions"] = expressions
					}
					
					resource["labelSelector"] = labelSelector
				}
			}
		}
	}

	// Update the binding policy
	updated, err := dynamicClient.Resource(BindingPolicyGVR).
		Namespace(namespace).
		Update(context.TODO(), policy, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update binding policy: %v", err)
	}

	fmt.Printf("BindingPolicy '%s' label selectors updated successfully\n", updated.GetName())
	if len(clusterLabels) > 0 || len(clusterExpressions) > 0 {
		fmt.Printf("Updated cluster label selectors\n")
	}
	if len(workloadLabels) > 0 || len(workloadExpressions) > 0 {
		fmt.Printf("Updated workload label selectors\n")
	}
	return nil
}