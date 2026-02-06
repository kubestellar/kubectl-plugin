package cluster

import (
	"os"
	"path/filepath"
	"testing"

	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func TestMain(m *testing.M) {
	setupTestKubeconfig()
	code := m.Run()
	cleanupTestKubeconfig()
	os.Exit(code)
}

func setupTestKubeconfig() {
	testConfig := &clientcmdapi.Config{
		Clusters: map[string]*clientcmdapi.Cluster{
			"test-cluster": {
				Server:                "https://test-server:443",
				InsecureSkipTLSVerify: true,
			},
			"wds1-cluster": {
				Server:                "https://wds1-server:443",
				InsecureSkipTLSVerify: true,
			},
		},
		AuthInfos: map[string]*clientcmdapi.AuthInfo{
			"test-user": {
				Token: "fake-token",
			},
			"wds1-user": {
				Token: "fake-wds1-token",
			},
		},
		Contexts: map[string]*clientcmdapi.Context{
			"test-context": {
				Cluster:  "test-cluster",
				AuthInfo: "test-user",
			},
			"wds1": {
				Cluster:  "wds1-cluster",
				AuthInfo: "wds1-user",
			},
		},
		CurrentContext: "test-context",
	}

	tempDir := os.TempDir()
	testKubeconfigPath := filepath.Join(tempDir, "test-kubeconfig")

	err := clientcmd.WriteToFile(*testConfig, testKubeconfigPath)
	if err != nil {
		panic(err)
	}

	os.Setenv("KUBECONFIG", testKubeconfigPath)
}

func cleanupTestKubeconfig() {
	if kubeconfigPath := os.Getenv("KUBECONFIG"); kubeconfigPath != "" {
		os.Remove(kubeconfigPath)
	}
	os.Unsetenv("KUBECONFIG")
}

func TestDiscoverClusters(t *testing.T) {
	clusters, err := DiscoverClusters(os.Getenv("KUBECONFIG"), "")
	if err != nil {
		t.Errorf("Got error: %v, expected nil", err)
	}
	if len(clusters) == 0 {
		t.Errorf("Got 0 clusters, expected length > 0")
	}
}

func TestIsWDSCluster(t *testing.T) {
	tests := []struct {
		name        string
		clusterName string
		expected    bool
	}{
		{
			name:        "wds prefix lowercase",
			clusterName: "wds1",
			expected:    true,
		},
		{
			name:        "wds prefix uppercase",
			clusterName: "WDS1",
			expected:    true,
		},
		{
			name:        "wds contains -wds-",
			clusterName: "cluster-wds-1",
			expected:    true,
		},
		{
			name:        "wds contains _wds_",
			clusterName: "cluster_wds_1",
			expected:    true,
		},
		{
			name:        "cluster name without wds",
			clusterName: "cluster1",
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isWDSCluster(tt.clusterName)
			if result != tt.expected {
				t.Errorf("isWDSCluster(%q) = %v, expected %v", tt.clusterName, result, tt.expected)
			}
		})
	}
}

func TestGetTargetNamespace(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		expected  string
	}{
		{
			name:      "namespace is not empty",
			namespace: "test",
			expected:  "test",
		},
		{
			name:      "namespace is empty",
			namespace: "",
			expected:  "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetTargetNamespace(tt.namespace)
			if result != tt.expected {
				t.Errorf("GetTargetNamespace(%q) = %v, expected %v", tt.namespace, result, tt.expected)
			}
		})
	}

}

func TestBuildClusterClient(t *testing.T) {
	tests := []struct {
		name                string
		kubeconfig          string
		ctxOverride         string
		expectedNil         bool
		expectedCtxName     string
		expectedClusterName string
	}{
		{
			name:                "kubeconfig doesn't exist",
			kubeconfig:          "./randomlocation",
			ctxOverride:         "",
			expectedNil:         true,
			expectedCtxName:     "",
			expectedClusterName: "",
		},
		{
			name:                "exists kubeconfig and context override wds1",
			kubeconfig:          os.Getenv("KUBECONFIG"),
			ctxOverride:         "wds1",
			expectedNil:         false,
			expectedCtxName:     "wds1",
			expectedClusterName: "wds1-cluster",
		},
		{
			name:                "exists kubeconfig and context override test-context",
			kubeconfig:          os.Getenv("KUBECONFIG"),
			ctxOverride:         "test-context",
			expectedNil:         false,
			expectedCtxName:     "test-context",
			expectedClusterName: "test-cluster",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctxName, clusterName, cs, dyn, disc, restCfg := buildClusterClient(tt.kubeconfig, tt.ctxOverride)

			if tt.expectedNil {
				if cs != nil || dyn != nil || disc != nil || restCfg != nil {
					t.Errorf("All clients should be nil for invalid config but got non nil values")
				}
			} else {
				if cs == nil || dyn == nil || disc == nil || restCfg == nil {
					t.Errorf("All clients should be non nil for valid config but got nil values")
				}
				if ctxName != tt.expectedCtxName {
					t.Errorf("Expected ctxName to be %q, but got %q", tt.expectedCtxName, ctxName)
				}
				if clusterName != tt.expectedClusterName {
					t.Errorf("Expected clusterName to be %q, but got %q", tt.expectedClusterName, clusterName)
				}
			}
		})
	}
}
