package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"kubectl-multi/pkg/cluster"
	"kubectl-multi/pkg/util"
)

// ClusterInfo for multiget (may include extra fields for ITS)
type MultiGetClusterInfo struct {
	Name           string
	KubeconfigPath string
	Client         *kubernetes.Clientset
	DynamicClient  dynamic.Interface
	RestConfig     *rest.Config
}

func toClusterInfo(m MultiGetClusterInfo) cluster.ClusterInfo {
	return cluster.ClusterInfo{
		Name:          m.Name,
		Context:       m.Name, // Use ITS name as context
		Client:        m.Client,
		DynamicClient: m.DynamicClient,
		RestConfig:    m.RestConfig,
	}
}

func newMultiGetCommand() *cobra.Command {
	var outputFormat string
	var selector string
	var showLabels bool
	var watch bool
	var watchOnly bool

	cmd := &cobra.Command{
		Use:   "multiget [TYPE[.VERSION][.GROUP] [NAME | -l label] | TYPE[.VERSION][.GROUP]/NAME ...]",
		Short: "Display resources across all ITS clusters discovered from Kubestellar core",
		Long:  `Get resources from all ITS clusters by discovering them from the Kubestellar core cluster. No dependency on kflex or pre-imported kubeconfigs.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("resource type must be specified")
			}

			kubeconfig, _, _, namespace, allNamespaces := GetGlobalFlags()
			// Auto-discover the KubeFlex hosting cluster
			coreContext, err := discoverKubeFlexHostingCluster(kubeconfig)
			if err != nil {
				return fmt.Errorf("failed to discover KubeFlex hosting cluster: %v", err)
			}
			clusters, err := discoverITSClustersFromCore(kubeconfig, coreContext)
			if err != nil {
				return fmt.Errorf("failed to discover ITS clusters: %v", err)
			}
			return handleMultiGetCommand(args, outputFormat, selector, showLabels, watch, watchOnly, clusters, namespace, allNamespaces)
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "", "output format (json|yaml|wide|name|custom-columns=...|custom-columns-file=...|go-template=...|go-template-file=...|jsonpath=...|jsonpath-file=...)")
	cmd.Flags().StringVarP(&selector, "selector", "l", "", "selector (label query) to filter on")
	cmd.Flags().BoolVar(&showLabels, "show-labels", false, "show all labels as the last column")
	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "watch for changes to the requested object(s)")
	cmd.Flags().BoolVar(&watchOnly, "watch-only", false, "watch for changes to the requested object(s), without listing/getting first")

	return cmd
}

// discoverITSClustersFromCore discovers ITS clusters by querying ControlPlane CRDs and fetching kubeconfigs from secrets
func discoverITSClustersFromCore(coreKubeconfig, coreContext string) ([]MultiGetClusterInfo, error) {
	var clusters []MultiGetClusterInfo

	// Build dynamic client for Kubestellar core
	loading := clientcmd.NewDefaultClientConfigLoadingRules()
	if coreKubeconfig != "" {
		loading.ExplicitPath = coreKubeconfig
	}
	overrides := &clientcmd.ConfigOverrides{}
	if coreContext != "" {
		overrides.CurrentContext = coreContext
	}
	cfg := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loading, overrides)
	restCfg, err := cfg.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to build rest config for core: %v", err)
	}
	dyn, err := dynamic.NewForConfig(restCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to build dynamic client for core: %v", err)
	}
	coreClient, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to build typed client for core: %v", err)
	}

	// List ControlPlane CRDs
	gvr := schema.GroupVersionResource{
		Group:    "tenancy.kflex.kubestellar.org",
		Version:  "v1alpha1",
		Resource: "controlplanes",
	}
	cps, err := dyn.Resource(gvr).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list ControlPlane CRDs: %v", err)
	}

	for _, cp := range cps.Items {
		name := cp.GetName()
		// Only process ITS clusters (vcluster type)
		typeVal, found, _ := unstructured.NestedString(cp.Object, "spec", "type")
		if !found || typeVal != "vcluster" {
			continue
		}
		// Get secretRef
		secretName, found1, _ := unstructured.NestedString(cp.Object, "status", "secretRef", "name")
		secretNamespace, found2, _ := unstructured.NestedString(cp.Object, "status", "secretRef", "namespace")
		key, found3, _ := unstructured.NestedString(cp.Object, "status", "secretRef", "key")
		if !found1 || !found2 || !found3 {
			continue
		}
		// Fetch kubeconfig from secret
		secret, err := coreClient.CoreV1().Secrets(secretNamespace).Get(context.TODO(), secretName, metav1.GetOptions{})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to get secret %s/%s: %v\n", secretNamespace, secretName, err)
			continue
		}
		kubeconfigBytes, ok := secret.Data[key]
		if !ok {
			fmt.Fprintf(os.Stderr, "Warning: secret %s/%s missing key %s\n", secretNamespace, secretName, key)
			continue
		}
		// Write kubeconfig to temp file
		tmpFile, err := os.CreateTemp("", fmt.Sprintf("%s-kubeconfig-*.yaml", name))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to create temp kubeconfig for %s: %v\n", name, err)
			continue
		}
		if _, err := tmpFile.Write(kubeconfigBytes); err != nil {
			tmpFile.Close()
			fmt.Fprintf(os.Stderr, "Warning: failed to write kubeconfig for %s: %v\n", name, err)
			continue
		}
		tmpFile.Close()

		// Build client for ITS vcluster
		itsCfg, err := clientcmd.BuildConfigFromFlags("", tmpFile.Name())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to build rest config for ITS %s: %v\n", name, err)
			continue
		}
		itsClient, err := kubernetes.NewForConfig(itsCfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to build typed client for ITS %s: %v\n", name, err)
			continue
		}
		itsDyn, err := dynamic.NewForConfig(itsCfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to build dynamic client for ITS %s: %v\n", name, err)
			continue
		}

		// Add the ITS cluster itself to the results
		clusters = append(clusters, MultiGetClusterInfo{
			Name:           name,
			KubeconfigPath: tmpFile.Name(),
			Client:         itsClient,
			DynamicClient:  itsDyn,
			RestConfig:     itsCfg,
		})

		// Discover ManagedClusters from the ITS vcluster
		mcGVR := schema.GroupVersionResource{
			Group:    "cluster.open-cluster-management.io",
			Version:  "v1",
			Resource: "managedclusters",
		}
		mcs, err := itsDyn.Resource(mcGVR).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to list managed clusters from ITS %s: %v\n", name, err)
			continue
		}

		// For each ManagedCluster, get its kubeconfig and add to clusters list
		for _, mc := range mcs.Items {
			mcName := mc.GetName()

			// Get the kubeconfig from the ManagedCluster spec
			clientConfigs, found, _ := unstructured.NestedSlice(mc.Object, "spec", "managedClusterClientConfigs")
			if !found || len(clientConfigs) == 0 {
				fmt.Fprintf(os.Stderr, "Warning: no client configs found for managed cluster %s\n", mcName)
				continue
			}

			// Use the first client config
			clientConfig, ok := clientConfigs[0].(map[string]interface{})
			if !ok {
				fmt.Fprintf(os.Stderr, "Warning: invalid client config for managed cluster %s\n", mcName)
				continue
			}

			url, found, _ := unstructured.NestedString(clientConfig, "url")
			caBundle, found2, _ := unstructured.NestedString(clientConfig, "caBundle")
			if !found || !found2 {
				fmt.Fprintf(os.Stderr, "Warning: missing url or caBundle for managed cluster %s\n", mcName)
				continue
			}

			// Create a kubeconfig from the ManagedCluster spec
			kubeconfig := fmt.Sprintf(`apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: %s
    server: %s
  name: %s
contexts:
- context:
    cluster: %s
    user: %s
  name: %s
current-context: %s
kind: Config
users:
- name: %s
  user:
    token: ""  # We'll use in-cluster config or need to get token from somewhere
`, caBundle, url, mcName, mcName, mcName, mcName, mcName, mcName)

			// Write managed cluster kubeconfig to temp file
			mcTmpFile, err := os.CreateTemp("", fmt.Sprintf("%s-kubeconfig-*.yaml", mcName))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to create temp kubeconfig for managed cluster %s: %v\n", mcName, err)
				continue
			}
			if _, err := mcTmpFile.Write([]byte(kubeconfig)); err != nil {
				mcTmpFile.Close()
				fmt.Fprintf(os.Stderr, "Warning: failed to write kubeconfig for managed cluster %s: %v\n", mcName, err)
				continue
			}
			mcTmpFile.Close()

			// Use the existing context-based approach since we have the contexts
			loading := clientcmd.NewDefaultClientConfigLoadingRules()
			overrides := &clientcmd.ConfigOverrides{
				CurrentContext: mcName,
			}
			cfg := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loading, overrides)
			mcCfg, err := cfg.ClientConfig()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to build rest config for managed cluster %s: %v\n", mcName, err)
				continue
			}

			mcClient, err := kubernetes.NewForConfig(mcCfg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to build typed client for managed cluster %s: %v\n", mcName, err)
				continue
			}
			mcDyn, err := dynamic.NewForConfig(mcCfg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to build dynamic client for managed cluster %s: %v\n", mcName, err)
				continue
			}

			clusters = append(clusters, MultiGetClusterInfo{
				Name:           mcName,
				KubeconfigPath: mcTmpFile.Name(),
				Client:         mcClient,
				DynamicClient:  mcDyn,
				RestConfig:     mcCfg,
			})
		}
	}
	return clusters, nil
}

// discoverKubeFlexHostingCluster finds the cluster that has KubeFlex installed
func discoverKubeFlexHostingCluster(kubeconfig string) (string, error) {
	// Try common names first (most likely candidates)
	commonNames := []string{"kind-kubeflex", "kubeflex", "ks-core", "kubestellar-core"}

	for _, name := range commonNames {
		if hasKubeFlexResources(kubeconfig, name) {
			return name, nil
		}
	}

	// If not found, scan all contexts
	loading := clientcmd.NewDefaultClientConfigLoadingRules()
	if kubeconfig != "" {
		loading.ExplicitPath = kubeconfig
	}
	cfg := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loading, &clientcmd.ConfigOverrides{})
	rawCfg, err := cfg.RawConfig()
	if err != nil {
		return "", fmt.Errorf("failed to load kubeconfig: %v", err)
	}

	for contextName := range rawCfg.Contexts {
		if hasKubeFlexResources(kubeconfig, contextName) {
			return contextName, nil
		}
	}

	return "", fmt.Errorf("no KubeFlex hosting cluster found. Please ensure KubeFlex is installed in one of your clusters")
}

// hasKubeFlexResources checks if a context has the ControlPlane CRD
func hasKubeFlexResources(kubeconfig, contextName string) bool {
	// Build config for this context
	loading := clientcmd.NewDefaultClientConfigLoadingRules()
	if kubeconfig != "" {
		loading.ExplicitPath = kubeconfig
	}
	overrides := &clientcmd.ConfigOverrides{
		CurrentContext: contextName,
	}
	cfg := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loading, overrides)
	restCfg, err := cfg.ClientConfig()
	if err != nil {
		return false
	}

	// Check for ControlPlane CRD - this is the definitive indicator
	dyn, err := dynamic.NewForConfig(restCfg)
	if err != nil {
		return false
	}

	gvr := schema.GroupVersionResource{
		Group:    "tenancy.kflex.kubestellar.org",
		Version:  "v1alpha1",
		Resource: "controlplanes",
	}
	_, err = dyn.Resource(gvr).List(context.Background(), metav1.ListOptions{})
	return err == nil
}

// handleMultiGetCommand runs the get logic across all discovered ITS clusters
func handleMultiGetCommand(args []string, outputFormat, selector string, showLabels, watch, watchOnly bool, clusters []MultiGetClusterInfo, namespace string, allNamespaces bool) error {
	resourceType := args[0]
	resourceName := ""
	if len(args) > 1 {
		resourceName = args[1]
	}

	if watch || watchOnly {
		return fmt.Errorf("watch operations are not supported in multi-cluster mode")
	}

	tw := tabwriter.NewWriter(util.GetOutputStream(), 0, 0, 2, ' ', 0)
	defer tw.Flush()

	switch strings.ToLower(resourceType) {
	case "nodes", "node", "no":
		return handleNodesGetMulti(tw, clusters, resourceName, selector, showLabels, outputFormat)
	case "pods", "pod", "po":
		return handlePodsGetMulti(tw, clusters, resourceName, selector, showLabels, outputFormat, namespace, allNamespaces)
	case "services", "service", "svc":
		return handleServicesGetMulti(tw, clusters, resourceName, selector, showLabels, outputFormat, namespace, allNamespaces)
	case "deployments", "deployment", "deploy":
		return handleDeploymentsGetMulti(tw, clusters, resourceName, selector, showLabels, outputFormat, namespace, allNamespaces)
	case "namespaces", "namespace", "ns":
		return handleNamespacesGetMulti(tw, clusters, resourceName, selector, showLabels, outputFormat)
	case "configmaps", "configmap", "cm":
		return handleConfigMapsGetMulti(tw, clusters, resourceName, selector, showLabels, outputFormat, namespace, allNamespaces)
	case "secrets", "secret":
		return handleSecretsGetMulti(tw, clusters, resourceName, selector, showLabels, outputFormat, namespace, allNamespaces)
	case "serviceaccounts", "serviceaccount", "sa":
		return handleServiceAccountsGetMulti(tw, clusters, resourceName, selector, showLabels, outputFormat, namespace, allNamespaces)
	case "persistentvolumes", "persistentvolume", "pv":
		return handlePVGetMulti(tw, clusters, resourceName, selector, showLabels, outputFormat)
	case "persistentvolumeclaims", "persistentvolumeclaim", "pvc":
		return handlePVCGetMulti(tw, clusters, resourceName, selector, showLabels, outputFormat, namespace, allNamespaces)
	default:
		return handleGenericGetMulti(tw, clusters, resourceType, resourceName, selector, showLabels, outputFormat, namespace, allNamespaces)
	}
}

func handleNodesGetMulti(tw *tabwriter.Writer, clusters []MultiGetClusterInfo, resourceName, selector string, showLabels bool, outputFormat string) error {
	var infos []cluster.ClusterInfo
	for _, c := range clusters {
		infos = append(infos, toClusterInfo(c))
	}
	return handleNodesGet(tw, infos, resourceName, selector, showLabels, outputFormat)
}

func handlePodsGetMulti(tw *tabwriter.Writer, clusters []MultiGetClusterInfo, resourceName, selector string, showLabels bool, outputFormat, namespace string, allNamespaces bool) error {
	var infos []cluster.ClusterInfo
	for _, c := range clusters {
		infos = append(infos, toClusterInfo(c))
	}
	return handlePodsGet(tw, infos, resourceName, selector, showLabels, outputFormat, namespace, allNamespaces)
}

func handleServicesGetMulti(tw *tabwriter.Writer, clusters []MultiGetClusterInfo, resourceName, selector string, showLabels bool, outputFormat, namespace string, allNamespaces bool) error {
	var infos []cluster.ClusterInfo
	for _, c := range clusters {
		infos = append(infos, toClusterInfo(c))
	}
	return handleServicesGet(tw, infos, resourceName, selector, showLabels, outputFormat, namespace, allNamespaces)
}

func handleDeploymentsGetMulti(tw *tabwriter.Writer, clusters []MultiGetClusterInfo, resourceName, selector string, showLabels bool, outputFormat, namespace string, allNamespaces bool) error {
	var infos []cluster.ClusterInfo
	for _, c := range clusters {
		infos = append(infos, toClusterInfo(c))
	}
	return handleDeploymentsGet(tw, infos, resourceName, selector, showLabels, outputFormat, namespace, allNamespaces)
}

func handleNamespacesGetMulti(tw *tabwriter.Writer, clusters []MultiGetClusterInfo, resourceName, selector string, showLabels bool, outputFormat string) error {
	var infos []cluster.ClusterInfo
	for _, c := range clusters {
		infos = append(infos, toClusterInfo(c))
	}
	return handleNamespacesGet(tw, infos, resourceName, selector, showLabels, outputFormat)
}

func handleConfigMapsGetMulti(tw *tabwriter.Writer, clusters []MultiGetClusterInfo, resourceName, selector string, showLabels bool, outputFormat, namespace string, allNamespaces bool) error {
	var infos []cluster.ClusterInfo
	for _, c := range clusters {
		infos = append(infos, toClusterInfo(c))
	}
	return handleConfigMapsGet(tw, infos, resourceName, selector, showLabels, outputFormat, namespace, allNamespaces)
}

func handleSecretsGetMulti(tw *tabwriter.Writer, clusters []MultiGetClusterInfo, resourceName, selector string, showLabels bool, outputFormat, namespace string, allNamespaces bool) error {
	var infos []cluster.ClusterInfo
	for _, c := range clusters {
		infos = append(infos, toClusterInfo(c))
	}
	return handleSecretsGet(tw, infos, resourceName, selector, showLabels, outputFormat, namespace, allNamespaces)
}

func handleServiceAccountsGetMulti(tw *tabwriter.Writer, clusters []MultiGetClusterInfo, resourceName, selector string, showLabels bool, outputFormat, namespace string, allNamespaces bool) error {
	var infos []cluster.ClusterInfo
	for _, c := range clusters {
		infos = append(infos, toClusterInfo(c))
	}
	return handleServiceAccountsGet(tw, infos, resourceName, selector, showLabels, outputFormat, namespace, allNamespaces)
}

func handlePVGetMulti(tw *tabwriter.Writer, clusters []MultiGetClusterInfo, resourceName, selector string, showLabels bool, outputFormat string) error {
	var infos []cluster.ClusterInfo
	for _, c := range clusters {
		infos = append(infos, toClusterInfo(c))
	}
	return handlePVGet(tw, infos, resourceName, selector, showLabels, outputFormat)
}

func handlePVCGetMulti(tw *tabwriter.Writer, clusters []MultiGetClusterInfo, resourceName, selector string, showLabels bool, outputFormat, namespace string, allNamespaces bool) error {
	var infos []cluster.ClusterInfo
	for _, c := range clusters {
		infos = append(infos, toClusterInfo(c))
	}
	return handlePVCGet(tw, infos, resourceName, selector, showLabels, outputFormat, namespace, allNamespaces)
}

func handleGenericGetMulti(tw *tabwriter.Writer, clusters []MultiGetClusterInfo, resourceType, resourceName, selector string, showLabels bool, outputFormat, namespace string, allNamespaces bool) error {
	var infos []cluster.ClusterInfo
	for _, c := range clusters {
		infos = append(infos, toClusterInfo(c))
	}
	return handleGenericGet(tw, infos, resourceType, resourceName, selector, showLabels, outputFormat, namespace, allNamespaces)
}
