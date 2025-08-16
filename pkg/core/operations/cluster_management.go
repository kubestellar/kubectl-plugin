package operations

import (
	"context"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
)

// ManagedClusterGVR defines the GroupVersionResource for Open Cluster Management ManagedCluster
var ManagedClusterGVR = schema.GroupVersionResource{
	Group:    "cluster.open-cluster-management.io",
	Version:  "v1",
	Resource: "managedclusters",
}

// ClusterRegistrationOptions contains options for registering clusters with ITS
type ClusterRegistrationOptions struct {
	Kubeconfig      string
	ITSContext      string // ITS (Inventory and Transport Space) context (e.g., "its1")
	ClusterName     string
	ClusterEndpoint string
	ClusterLabels   map[string]string
	Labels          map[string]string // Alias for ClusterLabels for CLI compatibility
	ClusterType     string // e.g., "wec" (Workload Execution Cluster)
	Overwrite       bool   // Overwrite existing labels
}

// RegisterClusterWithITS registers a new cluster with the ITS
func RegisterClusterWithITS(opts ClusterRegistrationOptions) error {
	// Build config with ITS context
	configOverrides := &clientcmd.ConfigOverrides{}
	if opts.ITSContext != "" {
		configOverrides.CurrentContext = opts.ITSContext
	}

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if opts.Kubeconfig != "" {
		loadingRules.ExplicitPath = opts.Kubeconfig
	}

	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	config, err := clientConfig.ClientConfig()
	if err != nil {
		return fmt.Errorf("failed to build config for ITS: %v", err)
	}

	// Create dynamic client
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create dynamic client: %v", err)
	}

	// Build the ManagedCluster object
	managedCluster := buildManagedCluster(opts)

	// Create the ManagedCluster in ITS
	created, err := dynamicClient.Resource(ManagedClusterGVR).
		Create(context.TODO(), managedCluster, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to register cluster with ITS: %v", err)
	}

	fmt.Printf("Cluster '%s' registered successfully with ITS '%s'\n", created.GetName(), opts.ITSContext)
	return nil
}

// buildManagedCluster constructs the ManagedCluster unstructured object
func buildManagedCluster(opts ClusterRegistrationOptions) *unstructured.Unstructured {
	// Prepare labels
	labels := map[string]interface{}{
		"name": opts.ClusterName,
	}
	
	// Add cluster type label if specified
	if opts.ClusterType != "" {
		labels["type"] = opts.ClusterType
		labels["cluster.kubestellar.io/type"] = opts.ClusterType
	}
	
	// Add any additional labels
	for k, v := range opts.ClusterLabels {
		labels[k] = v
	}

	managedCluster := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "cluster.open-cluster-management.io/v1",
			"kind":       "ManagedCluster",
			"metadata": map[string]interface{}{
				"name":   opts.ClusterName,
				"labels": labels,
			},
			"spec": map[string]interface{}{
				"hubAcceptsClient": true,
				"managedClusterClientConfigs": []interface{}{
					map[string]interface{}{
						"url": opts.ClusterEndpoint,
					},
				},
			},
		},
	}

	return managedCluster
}

// ListClustersInITS lists all clusters registered in the ITS
func ListClustersInITS(kubeconfig, itsContext string, includeWDS bool) error {
	// Build config with ITS context
	configOverrides := &clientcmd.ConfigOverrides{}
	if itsContext != "" {
		configOverrides.CurrentContext = itsContext
	}

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if kubeconfig != "" {
		loadingRules.ExplicitPath = kubeconfig
	}

	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	config, err := clientConfig.ClientConfig()
	if err != nil {
		return fmt.Errorf("failed to build config for ITS: %v", err)
	}

	// Create dynamic client
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create dynamic client: %v", err)
	}

	// List ManagedClusters
	clusters, err := dynamicClient.Resource(ManagedClusterGVR).
		List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list clusters in ITS: %v", err)
	}

	// Display clusters
	fmt.Printf("NAME\tTYPE\tSTATUS\tENDPOINT\n")
	for _, cluster := range clusters.Items {
		name := cluster.GetName()
		
		// Skip WDS clusters if not included
		if !includeWDS && isWDSCluster(name) {
			continue
		}

		// Extract type from labels
		clusterType := "wec" // default to WEC
		if labels := cluster.GetLabels(); labels != nil {
			if t, ok := labels["type"]; ok {
				clusterType = t
			} else if t, ok := labels["cluster.kubestellar.io/type"]; ok {
				clusterType = t
			}
		}

		// Extract status
		status := extractClusterStatus(cluster.Object)
		
		// Extract endpoint
		endpoint := extractClusterEndpoint(cluster.Object)

		fmt.Printf("%s\t%s\t%s\t%s\n", name, clusterType, status, endpoint)
	}

	return nil
}

// isWDSCluster checks if a cluster is a WDS (Workload Description Space) cluster
func isWDSCluster(clusterName string) bool {
	lowerName := strings.ToLower(clusterName)
	return strings.HasPrefix(lowerName, "wds") || 
		strings.Contains(lowerName, "-wds-") || 
		strings.Contains(lowerName, "_wds_")
}

// extractClusterStatus extracts the status from ManagedCluster object
func extractClusterStatus(cluster map[string]interface{}) string {
	status, ok := cluster["status"].(map[string]interface{})
	if !ok {
		return "Unknown"
	}

	conditions, ok := status["conditions"].([]interface{})
	if !ok || len(conditions) == 0 {
		return "Unknown"
	}

	// Look for ManagedClusterConditionAvailable condition
	for _, cond := range conditions {
		if condition, ok := cond.(map[string]interface{}); ok {
			if condType, ok := condition["type"].(string); ok && condType == "ManagedClusterConditionAvailable" {
				if condStatus, ok := condition["status"].(string); ok {
					if condStatus == "True" {
						return "Ready"
					}
					return "NotReady"
				}
			}
		}
	}

	return "Unknown"
}

// extractClusterEndpoint extracts the endpoint from ManagedCluster spec
func extractClusterEndpoint(cluster map[string]interface{}) string {
	spec, ok := cluster["spec"].(map[string]interface{})
	if !ok {
		return "none"
	}

	configs, ok := spec["managedClusterClientConfigs"].([]interface{})
	if !ok || len(configs) == 0 {
		return "none"
	}

	if config, ok := configs[0].(map[string]interface{}); ok {
		if url, ok := config["url"].(string); ok {
			return url
		}
	}

	return "none"
}

// RemoveClusterFromITS removes a cluster from ITS registration
func RemoveClusterFromITS(kubeconfig, itsContext, clusterName string) error {
	// Build config with ITS context
	configOverrides := &clientcmd.ConfigOverrides{}
	if itsContext != "" {
		configOverrides.CurrentContext = itsContext
	}

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if kubeconfig != "" {
		loadingRules.ExplicitPath = kubeconfig
	}

	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	config, err := clientConfig.ClientConfig()
	if err != nil {
		return fmt.Errorf("failed to build config for ITS: %v", err)
	}

	// Create dynamic client
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create dynamic client: %v", err)
	}

	// Delete the ManagedCluster
	err = dynamicClient.Resource(ManagedClusterGVR).
		Delete(context.TODO(), clusterName, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to remove cluster from ITS: %v", err)
	}

	fmt.Printf("Cluster '%s' removed successfully from ITS '%s'\n", clusterName, itsContext)
	return nil
}

// UpdateClusterLabels updates labels on a ManagedCluster in ITS
func UpdateClusterLabels(kubeconfig, itsContext, clusterName string, addLabels, removeLabels map[string]string) error {
	// Build config with ITS context
	configOverrides := &clientcmd.ConfigOverrides{}
	if itsContext != "" {
		configOverrides.CurrentContext = itsContext
	}

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if kubeconfig != "" {
		loadingRules.ExplicitPath = kubeconfig
	}

	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	config, err := clientConfig.ClientConfig()
	if err != nil {
		return fmt.Errorf("failed to build config for ITS: %v", err)
	}

	// Create dynamic client
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create dynamic client: %v", err)
	}

	// Get existing ManagedCluster
	cluster, err := dynamicClient.Resource(ManagedClusterGVR).
		Get(context.TODO(), clusterName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get cluster from ITS: %v", err)
	}

	// Update labels
	labels := cluster.GetLabels()
	if labels == nil {
		labels = map[string]string{}
	}

	// Add new labels
	for k, v := range addLabels {
		labels[k] = v
	}

	// Remove specified labels
	for k := range removeLabels {
		delete(labels, k)
	}

	cluster.SetLabels(labels)

	// Update the ManagedCluster
	updated, err := dynamicClient.Resource(ManagedClusterGVR).
		Update(context.TODO(), cluster, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update cluster in ITS: %v", err)
	}

	fmt.Printf("Cluster '%s' labels updated successfully in ITS '%s'\n", updated.GetName(), itsContext)
	return nil
}

// GetClusterDetails retrieves detailed information about a specific cluster from ITS
func GetClusterDetails(kubeconfig, itsContext, clusterName string) error {
	// Build config with ITS context
	configOverrides := &clientcmd.ConfigOverrides{}
	if itsContext != "" {
		configOverrides.CurrentContext = itsContext
	}

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if kubeconfig != "" {
		loadingRules.ExplicitPath = kubeconfig
	}

	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	config, err := clientConfig.ClientConfig()
	if err != nil {
		return fmt.Errorf("failed to build config for ITS: %v", err)
	}

	// Create dynamic client
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create dynamic client: %v", err)
	}

	// Get the ManagedCluster
	cluster, err := dynamicClient.Resource(ManagedClusterGVR).
		Get(context.TODO(), clusterName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get cluster from ITS: %v", err)
	}

	// Display cluster details
	fmt.Printf("Cluster: %s\n", cluster.GetName())
	fmt.Printf("ITS Context: %s\n", itsContext)
	
	// Display labels
	if labels := cluster.GetLabels(); labels != nil && len(labels) > 0 {
		fmt.Printf("Labels:\n")
		for k, v := range labels {
			fmt.Printf("  %s: %s\n", k, v)
		}
	}

	// Display spec details
	if spec, ok := cluster.Object["spec"].(map[string]interface{}); ok {
		fmt.Printf("Spec:\n")
		if hubAccepts, ok := spec["hubAcceptsClient"].(bool); ok {
			fmt.Printf("  Hub Accepts Client: %v\n", hubAccepts)
		}
		
		if configs, ok := spec["managedClusterClientConfigs"].([]interface{}); ok && len(configs) > 0 {
			fmt.Printf("  Client Configs:\n")
			for i, cfg := range configs {
				if config, ok := cfg.(map[string]interface{}); ok {
					fmt.Printf("    Config %d:\n", i+1)
					if url, ok := config["url"].(string); ok {
						fmt.Printf("      URL: %s\n", url)
					}
				}
			}
		}
	}

	// Display status
	if status, ok := cluster.Object["status"].(map[string]interface{}); ok {
		fmt.Printf("Status:\n")
		if conditions, ok := status["conditions"].([]interface{}); ok {
			fmt.Printf("  Conditions:\n")
			for _, cond := range conditions {
				if condition, ok := cond.(map[string]interface{}); ok {
					condType := condition["type"]
					condStatus := condition["status"]
					fmt.Printf("    %v: %v\n", condType, condStatus)
				}
			}
		}
	}

	return nil
}

// UnregisterClusterFromITS is a wrapper for RemoveClusterFromITS for CLI compatibility
func UnregisterClusterFromITS(opts ClusterRegistrationOptions) error {
	return RemoveClusterFromITS(opts.Kubeconfig, opts.ITSContext, opts.ClusterName)
}

// DescribeClusterInITS is a wrapper for GetClusterDetails for CLI compatibility
func DescribeClusterInITS(opts ClusterRegistrationOptions) error {
	return GetClusterDetails(opts.Kubeconfig, opts.ITSContext, opts.ClusterName)
}

// UpdateClusterLabelsInITS updates labels on a cluster using the options pattern
func UpdateClusterLabelsInITS(opts ClusterRegistrationOptions) error {
	// Use Labels field if ClusterLabels is not set
	labels := opts.ClusterLabels
	if len(labels) == 0 && len(opts.Labels) > 0 {
		labels = opts.Labels
	}
	
	// If overwrite is true, remove existing labels first
	var removeLabels map[string]string
	if opts.Overwrite {
		// TODO: Implement getting existing labels and removing them
		removeLabels = map[string]string{}
	}
	
	return UpdateClusterLabels(opts.Kubeconfig, opts.ITSContext, opts.ClusterName, labels, removeLabels)
}