package cluster

import (
	"context"
	"fmt"
	"sort"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// ClusterInfo contains information about a discovered cluster
type ClusterInfo struct {
	Name            string
	Context         string
	Client          *kubernetes.Clientset
	DynamicClient   dynamic.Interface
	DiscoveryClient discovery.DiscoveryInterface
	RestConfig      *rest.Config
}

// DiscoverClusters finds all clusters including the local cluster and managed clusters
func DiscoverClusters(kubeconfig, remoteCtx string) ([]ClusterInfo, error) {
	var clusters []ClusterInfo

	// Add managed clusters first (excluding WDS clusters)
	if remoteCtx != "" {
		managedClusters, err := listManagedClusters(kubeconfig, remoteCtx)
		if err != nil {
			fmt.Printf("Warning: could not list managed clusters: %v\n", err)
		} else {
			for _, mcName := range managedClusters {
				// Skip WDS clusters - they are for workflow staging, not workload execution
				if isWDSCluster(mcName) {
					continue
				}

				_, _, cs, dyn, disc, restCfg := buildClusterClient(kubeconfig, mcName)
				if cs != nil { // Only add if we can connect
					clusters = append(clusters, ClusterInfo{
						Name:            mcName,
						Context:         remoteCtx,
						Client:          cs,
						DynamicClient:   dyn,
						DiscoveryClient: disc,
						RestConfig:      restCfg,
					})
				}
			}
		}
	}

	// Add local cluster (ITS cluster) - but check if it's not already included
	localCtx, localCluster, localClient, localDynamic, localDiscovery, localRestConfig := buildClusterClient(kubeconfig, "")
	if localClient != nil && !isWDSCluster(localCluster) {
		// Check if this cluster is already in the list (avoid duplicates)
		found := false
		for _, cluster := range clusters {
			if cluster.Name == localCluster {
				found = true
				break
			}
		}
		if !found {
			clusters = append(clusters, ClusterInfo{
				Name:            localCluster,
				Context:         localCtx,
				Client:          localClient,
				DynamicClient:   localDynamic,
				DiscoveryClient: localDiscovery,
				RestConfig:      localRestConfig,
			})
		}
	}

	return clusters, nil
}

// isWDSCluster checks if a cluster name indicates it's a Workload Description Space cluster
func isWDSCluster(clusterName string) bool {
	// WDS clusters typically have names like "wds1", "wds2", etc.
	// or contain "wds" in their name
	lowerName := strings.ToLower(clusterName)
	return strings.HasPrefix(lowerName, "wds") || strings.Contains(lowerName, "-wds-") || strings.Contains(lowerName, "_wds_")
}

// buildClusterClient creates all necessary clients for a cluster
func buildClusterClient(kcfg, ctxOverride string) (string, string, *kubernetes.Clientset, dynamic.Interface, discovery.DiscoveryInterface, *rest.Config) {
	loading := clientcmd.NewDefaultClientConfigLoadingRules()
	if kcfg != "" {
		loading.ExplicitPath = kcfg
	}
	overrides := &clientcmd.ConfigOverrides{}
	if ctxOverride != "" {
		overrides.CurrentContext = ctxOverride
	}

	cfg := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loading, overrides)
	rawCfg, err := cfg.RawConfig()
	if err != nil {
		fmt.Printf("Warning: failed to load kubeconfig: %v\n", err)
		return "", "", nil, nil, nil, nil
	}

	restCfg, err := cfg.ClientConfig()
	if err != nil {
		fmt.Printf("Warning: failed to create rest config: %v\n", err)
		return "", "", nil, nil, nil, nil
	}

	cs, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		fmt.Printf("Warning: failed to create kubernetes client: %v\n", err)
		return "", "", nil, nil, nil, nil
	}

	dyn, err := dynamic.NewForConfig(restCfg)
	if err != nil {
		fmt.Printf("Warning: failed to create dynamic client: %v\n", err)
		return "", "", nil, nil, nil, nil
	}

	disc, err := discovery.NewDiscoveryClientForConfig(restCfg)
	if err != nil {
		fmt.Printf("Warning: failed to create discovery client: %v\n", err)
		return "", "", nil, nil, nil, nil
	}

	ctxName := rawCfg.CurrentContext
	clusterName := "<unknown>"
	if ctx, ok := rawCfg.Contexts[ctxName]; ok {
		clusterName = ctx.Cluster
	}

	return ctxName, clusterName, cs, dyn, disc, restCfg
}

// listManagedClusters discovers KubeStellar managed clusters
func listManagedClusters(kubeconfig, remoteCtx string) ([]string, error) {
	_, _, _, dyn, _, _ := buildClusterClient(kubeconfig, remoteCtx)
	if dyn == nil {
		return nil, fmt.Errorf("failed to create dynamic client for remote context %s", remoteCtx)
	}

	gvr := schema.GroupVersionResource{
		Group:    "cluster.open-cluster-management.io",
		Version:  "v1",
		Resource: "managedclusters",
	}

	mcs, err := dyn.Resource(gvr).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list managed clusters: %v", err)
	}

	var clusters []string
	for _, mc := range mcs.Items {
		clusterName := mc.GetName()
		// Filter out WDS clusters at the discovery level too
		if !isWDSCluster(clusterName) {
			clusters = append(clusters, clusterName)
		}
	}
	sort.Strings(clusters)
	return clusters, nil
}

// GetTargetNamespace determines the target namespace for operations
func GetTargetNamespace(namespace string) string {
	if namespace != "" {
		return namespace
	}
	// Try to get namespace from environment or kubeconfig context
	// For now, default to "default"
	return "default"
}
