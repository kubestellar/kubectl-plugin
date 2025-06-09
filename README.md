# kubectl-multi

A comprehensive kubectl plugin for multi-cluster operations with KubeStellar. This plugin extends kubectl to work seamlessly across all KubeStellar managed clusters, providing unified views and operations while filtering out workflow staging clusters (WDS).

## Table of Contents

- [Overview](#overview)
- [Tech Stack](#tech-stack)
- [Architecture](#architecture)
- [How It Works](#how-it-works)
- [Installation](#installation)
- [Usage](#usage)
- [Code Organization](#code-organization)
- [Technical Implementation](#technical-implementation)
- [Examples](#examples)
- [Development](#development)
- [Contributing](#contributing)

## Overview

kubectl-multi is a kubectl plugin written in Go that automatically discovers KubeStellar managed clusters and executes kubectl commands across all of them simultaneously. It provides a unified tabular output with cluster context information, making it easy to monitor and manage resources across multiple clusters.

### Key Features

- **Multi-cluster resource viewing**: Get resources from all managed clusters with unified output
- **Cluster context identification**: Each resource shows which cluster it belongs to
- **All kubectl commands**: Supports all major kubectl commands across clusters
- **KubeStellar integration**: Automatically discovers managed clusters via KubeStellar APIs
- **WDS filtering**: Automatically excludes Workload Description Space clusters
- **Familiar syntax**: Uses the same command structure as kubectl

## Tech Stack

### Languages & Frameworks
- **Go 1.21+**: Primary language for the plugin
- **Cobra**: CLI framework for command structure and parsing
- **Kubernetes client-go**: Official Kubernetes Go client library
- **KubeStellar APIs**: For managed cluster discovery

### Key Dependencies
```go
require (
	github.com/spf13/cobra v1.8.0           // CLI framework
	k8s.io/api v0.29.0                      // Kubernetes API types
	k8s.io/apimachinery v0.29.0             // Kubernetes API machinery
	k8s.io/client-go v0.29.0                // Kubernetes Go client
	k8s.io/kubectl v0.29.0                  // kubectl utilities
)
```

### Build System
- **Make**: Build automation and installation
- **Go modules**: Dependency management
- **Static binary**: Single executable for easy distribution

## Architecture

### High-Level Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────────┐
│   kubectl CLI   │    │  kubectl-multi   │──▶│  KubeStellar ITS    │
│                 │──▶│     Plugin       │──▶│    (Discovery)      │
│ kubectl multi   │    │                  │    │                     │
│   get pods      │    │  Cluster Disco.  │    │ ManagedCluster CRDs │
└─────────────────┘    │  Command Exec.   │    └─────────────────────┘
                       │  Output Format   │
                       └──────────────────┘
                              │
                              ▼
                    ┌─────────────────────┐
                    │  Managed Clusters   │
                    │                     │
                    │ ┌─────────────────┐ │
                    │ │    cluster1     │ │
                    │ │    cluster2     │ │
                    │ │       ...       │ │
                    │ └─────────────────┘ │
                    └─────────────────────┘
```

### Plugin Architecture

```
kubectl-multi/
├── main.go                 # Entry point - delegates to cmd package
├── pkg/
│   ├── cmd/               # Command structure (Cobra-based)
│   │   ├── root.go        # Root command & global flags
│   │   ├── get.go         # Get command implementation
│   │   ├── describe.go    # Describe command
│   │   └── ...           # Other kubectl commands
│   ├── cluster/           # Cluster discovery & management
│   │   └── discovery.go   # KubeStellar cluster discovery
│   └── util/              # Utility functions
│       └── formatting.go  # Resource formatting & helpers
```

## How It Works

### 1. Cluster Discovery Process

The plugin discovers clusters through a multi-step process:

```go
func DiscoverClusters(kubeconfig, remoteCtx string) ([]ClusterInfo, error) {
	// 1. Connect to ITS cluster (e.g., "its1")
	// 2. List ManagedCluster CRDs using dynamic client
	// 3. Filter out WDS clusters (wds1, wds2, etc.)
	// 4. Build clients for each workload cluster
	// 5. Return slice of ClusterInfo with all clients
}
```

**ManagedCluster Discovery:**
```go
// Uses KubeStellar's ManagedCluster CRDs
gvr := schema.GroupVersionResource{
	Group:    "cluster.open-cluster-management.io",
	Version:  "v1", 
	Resource: "managedclusters",
}
```

**WDS Filtering:**
```go
func isWDSCluster(clusterName string) bool {
	lowerName := strings.ToLower(clusterName)
	return strings.HasPrefix(lowerName, "wds") || 
		   strings.Contains(lowerName, "-wds-") || 
		   strings.Contains(lowerName, "_wds_")
}
```

### 2. Command Processing Flow

```
User Input: kubectl multi get pods -n kube-system
     │
     ▼
┌──────────────────────────────────────────────────────────┐
│  1. Parse Command & Flags (Cobra)                       │
│     - Resource type: "pods"                             │
│     - Namespace: "kube-system"                          │
│     - Other flags: selector, output format, etc.       │
└──────────────────────────────────────────────────────────┘
     │
     ▼
┌──────────────────────────────────────────────────────────┐
│  2. Discover Clusters                                    │
│     - Connect to ITS cluster                            │
│     - List ManagedCluster CRDs                          │
│     - Filter out WDS clusters                           │
│     - Build clients for each cluster                    │
└──────────────────────────────────────────────────────────┘
     │
     ▼
┌──────────────────────────────────────────────────────────┐
│  3. Route to Resource Handler                            │
│     - handlePodsGet() for pods                          │
│     - handleNodesGet() for nodes                        │
│     - handleGenericGet() for other resources            │
└──────────────────────────────────────────────────────────┘
     │
     ▼
┌──────────────────────────────────────────────────────────┐
│  4. Execute Across All Clusters                         │
│     - Print header once                                 │
│     - For each cluster:                                 │
│       * List resources using appropriate client         │
│       * Format and append to output                     │
└──────────────────────────────────────────────────────────┘
     │
     ▼
┌──────────────────────────────────────────────────────────┐
│  5. Unified Output                                       │
│     - Single table with CONTEXT and CLUSTER columns     │
│     - Resources from all clusters combined              │
└──────────────────────────────────────────────────────────┘
```

### 3. Resource Type Handling

The plugin handles different resource types through a sophisticated routing system:

#### **Built-in Resource Handlers**
```go
switch strings.ToLower(resourceType) {
case "nodes", "node", "no":
	return handleNodesGet(...)
case "pods", "pod", "po":
	return handlePodsGet(...)
case "services", "service", "svc":
	return handleServicesGet(...)
case "deployments", "deployment", "deploy":
	return handleDeploymentsGet(...)
// ... more specific handlers
default:
	return handleGenericGet(...) // Uses dynamic client for discovery
}
```

#### **Dynamic Resource Discovery**
For unknown resource types, the plugin uses Kubernetes API discovery:

```go
func DiscoverGVR(discoveryClient discovery.DiscoveryInterface, resourceType string) (schema.GroupVersionResource, bool, error) {
	// 1. Get all API resources from the cluster
	_, apiResourceLists, err := discoveryClient.ServerGroupsAndResources()
	
	// 2. Normalize resource type (handle aliases like "po" -> "pods")
	normalizedType := normalizeResourceType(resourceType)
	
	// 3. Search through all API resources for matches
	// 4. Return GroupVersionResource + whether it's namespaced
}
```

### 4. Output Formatting

The plugin generates unified tabular output with cluster context:

#### **Single Header Strategy**
```go
// Print header only once at the top
fmt.Fprintf(tw, "CONTEXT\tCLUSTER\tNAME\tSTATUS\tROLES\tAGE\tVERSION\n")

// Then iterate through all clusters and resources
for _, clusterInfo := range clusters {
	for _, resource := range resources {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			clusterInfo.Context, clusterInfo.Name, ...)
	}
}
```

#### **Namespace Handling**
The plugin intelligently handles namespace-scoped vs cluster-scoped resources:

```go
// For namespace-scoped resources with -A flag
if allNamespaces {
	fmt.Fprintf(tw, "CONTEXT\tCLUSTER\tNAMESPACE\tNAME\t...\n")
} else {
	fmt.Fprintf(tw, "CONTEXT\tCLUSTER\tNAME\t...\n")  // No namespace column
}
```

## Installation

### Prerequisites
- Go 1.21 or later
- kubectl installed and configured
- Access to KubeStellar managed clusters

### Build and Install

```bash
# Clone the repository
git clone <repository-url>
cd kubectl-multi

# Build and install as kubectl plugin
make install

# Or install system-wide
make install-system
```

### Manual Installation

```bash
# Build binary
make build

# Copy to PATH
cp bin/kubectl-multi ~/.local/bin/
chmod +x ~/.local/bin/kubectl-multi

# Verify installation
kubectl plugin list | grep multi
```

## Usage

### Basic Commands

```bash
# Get nodes from all managed clusters
kubectl multi get nodes

# Get pods from all clusters in all namespaces
kubectl multi get pods -A

# Get services in specific namespace
kubectl multi get services -n kube-system

# Use label selectors
kubectl multi get pods -l app=nginx -A

# Show labels
kubectl multi get pods --show-labels -n kube-system
```

### Global Flags

- `--kubeconfig string`: Path to kubeconfig file
- `--remote-context string`: Remote hosting context (default: "its1")
- `--all-clusters`: Operate on all managed clusters (default: true)
- `-n, --namespace string`: Target namespace
- `-A, --all-namespaces`: List resources across all namespaces

## Code Organization

### Entry Point (`main.go`)
```go
package main

import "kubectl-multi/pkg/cmd"

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
```

### Command Structure (`pkg/cmd/`)

#### Root Command (`root.go`)
- Sets up global flags and command structure
- Uses Cobra for CLI framework
- Initializes all subcommands

#### Get Command (`get.go`)
- Handles all `kubectl get` operations
- Routes to specific resource handlers
- Manages output formatting

#### Resource Handlers
Each resource type has specialized handling:

```go
func handlePodsGet(tw *tabwriter.Writer, clusters []cluster.ClusterInfo, ...) error {
	// 1. Print header once
	// 2. Iterate through clusters
	// 3. List pods in each cluster
	// 4. Format output with cluster context
}
```

### Cluster Management (`pkg/cluster/`)

#### Discovery (`discovery.go`)
```go
type ClusterInfo struct {
	Name            string                    // Cluster name
	Context         string                    // kubectl context
	Client          *kubernetes.Clientset    // Typed client
	DynamicClient   dynamic.Interface         // Dynamic client
	DiscoveryClient discovery.DiscoveryInterface // API discovery
	RestConfig      *rest.Config             // REST configuration
}

func DiscoverClusters(kubeconfig, remoteCtx string) ([]ClusterInfo, error)
func buildClusterClient(kcfg, ctxOverride string) (...)
func listManagedClusters(kubeconfig, remoteCtx string) ([]string, error)
```

### Utilities (`pkg/util/`)

#### Formatting (`formatting.go`)
```go
// Resource-specific formatters
func GetNodeStatus(node corev1.Node) string
func GetPodReadyContainers(pod *corev1.Pod) int32
func GetServiceExternalIP(svc *corev1.Service) string
func FormatLabels(labels map[string]string) string

// Dynamic resource discovery
func DiscoverGVR(discoveryClient discovery.DiscoveryInterface, resourceType string) (...)
```

## Technical Implementation

### Client Management

The plugin maintains multiple Kubernetes clients for each cluster:

```go
type ClusterInfo struct {
	Client          *kubernetes.Clientset    // For typed operations (pods, services, etc.)
	DynamicClient   dynamic.Interface         // For custom resources and generic operations
	DiscoveryClient discovery.DiscoveryInterface // For API resource discovery
}
```

### Error Handling Strategy

The plugin uses graceful error handling to ensure partial failures don't break the entire operation:

```go
for _, clusterInfo := range clusters {
	resources, err := clusterInfo.Client.CoreV1().Pods(ns).List(...)
	if err != nil {
		fmt.Printf("Warning: failed to list pods in cluster %s: %v\n", clusterInfo.Name, err)
		continue  // Continue with other clusters
	}
	// Process resources...
}
```

### Resource Type Discovery

For unknown resource types, the plugin uses Kubernetes API discovery:

1. **Normalization**: Converts aliases (`po` → `pods`, `svc` → `services`)
2. **API Discovery**: Queries the cluster for available resources
3. **Matching**: Finds resources by name, singular name, or short names
4. **Fallback**: Uses sensible defaults for common resources

### Namespace Scope Detection

The plugin automatically detects whether resources are namespace-scoped:

```go
gvr, isNamespaced, err := util.DiscoverGVR(clusterInfo.DiscoveryClient, resourceType)

if isNamespaced && !allNamespaces && targetNS != "" {
	// List in specific namespace
	list, err = clusterInfo.DynamicClient.Resource(gvr).Namespace(targetNS).List(...)
} else {
	// List cluster-wide or all namespaces
	list, err = clusterInfo.DynamicClient.Resource(gvr).List(...)
}
```

## Examples

### Sample Input and Output

#### Input: Get Nodes
```bash
kubectl multi get nodes
```

#### Output:
```
CONTEXT  CLUSTER       NAME                    STATUS  ROLES          AGE    VERSION
its1     cluster1      cluster1-control-plane  Ready   control-plane  6d23h  v1.33.1
its1     cluster2      cluster2-control-plane  Ready   control-plane  6d23h  v1.33.1
its1     its1-cluster  kubeflex-control-plane  Ready   <none>         6d23h  v1.27.2+k3s1
```

#### Input: Get Pods with Namespace
```bash
kubectl multi get pods -n kube-system
```

#### Output:
```
CONTEXT  CLUSTER       NAME                                            READY  STATUS   RESTARTS  AGE
its1     cluster1      coredns-674b8bbfcf-6k7vc                        1/1    Running  2         6d23h
its1     cluster1      etcd-cluster1-control-plane                     1/1    Running  2         6d23h
its1     cluster1      kube-apiserver-cluster1-control-plane           1/1    Running  2         6d23h
its1     cluster2      coredns-674b8bbfcf-5c46s                        1/1    Running  2         6d23h
its1     cluster2      etcd-cluster2-control-plane                     1/1    Running  2         6d23h
its1     its1-cluster  coredns-68559449b6-g8kpn                        1/1    Running  14        6d23h
```

#### Input: Get Services with All Namespaces
```bash
kubectl multi get services -A
```

#### Output:
```
CONTEXT  CLUSTER       NAMESPACE    NAME          TYPE       CLUSTER-IP    EXTERNAL-IP  PORT(S)                 AGE
its1     cluster1      default      kubernetes    ClusterIP  10.96.0.1     <none>       443/TCP                 6d23h
its1     cluster1      kube-system  kube-dns      ClusterIP  10.96.0.10    <none>       53/UDP,53/TCP,9153/TCP  6d23h
its1     cluster2      default      kubernetes    ClusterIP  10.96.0.1     <none>       443/TCP                 6d23h
its1     cluster2      kube-system  kube-dns      ClusterIP  10.96.0.10    <none>       53/UDP,53/TCP,9153/TCP  6d23h
```

#### Input: Label Selector
```bash
kubectl multi get pods -l k8s-app=kube-dns -A
```

#### Output:
```
CONTEXT  CLUSTER       NAMESPACE    NAME                      READY  STATUS   RESTARTS  AGE
its1     cluster1      kube-system  coredns-674b8bbfcf-6k7vc  1/1    Running  2         6d23h
its1     cluster1      kube-system  coredns-674b8bbfcf-vhh9g  1/1    Running  2         6d23h
its1     cluster2      kube-system  coredns-674b8bbfcf-5c46s  1/1    Running  2         6d23h
its1     cluster2      kube-system  coredns-674b8bbfcf-7gft4  1/1    Running  2         6d23h
```

### Resource Type Support

#### Cluster-Scoped Resources
- `nodes` - Kubernetes nodes
- `namespaces` - Kubernetes namespaces  
- `persistentvolumes` (pv) - Persistent volumes
- `storageclasses` - Storage classes
- `clusterroles` - RBAC cluster roles

#### Namespace-Scoped Resources  
- `pods` (po) - Kubernetes pods
- `services` (svc) - Kubernetes services
- `deployments` (deploy) - Kubernetes deployments
- `configmaps` (cm) - Configuration maps
- `secrets` - Kubernetes secrets
- `persistentvolumeclaims` (pvc) - PV claims

#### Custom Resources
- Any CRD installed in clusters (auto-discovered)
- KubeStellar resources (managedclusters, etc.)

## Development

### Building from Source

```bash
# Clone repository
git clone <repository-url>
cd kubectl-multi

# Download dependencies
make deps

# Build binary
make build

# Run tests
make test

# Format code
make fmt

# Run all checks
make check
```

### Project Structure

```
kubectl-multi/
├── main.go                 # Entry point
├── go.mod                  # Go module definition
├── go.sum                  # Dependency checksums
├── Makefile               # Build automation
├── README.md              # This documentation
├── pkg/
│   ├── cmd/               # Command implementations
│   │   ├── root.go        # Root command & CLI setup
│   │   ├── get.go         # Get command (fully implemented)
│   │   ├── describe.go    # Describe command (basic)
│   │   ├── apply.go       # Apply command (placeholder)
│   │   └── delete.go      # Other commands (placeholders)
│   ├── cluster/           # Cluster discovery & management
│   │   └── discovery.go   # KubeStellar cluster discovery
│   └── util/              # Utility functions
│       └── formatting.go  # Resource formatting utilities
└── bin/                   # Build output directory
    └── kubectl-multi      # Compiled binary
```

### Adding New Commands

To add a new kubectl command (e.g., `logs`):

1. **Create command file**: `pkg/cmd/logs.go`
```go
func newLogsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logs [-f] [-p] POD [-c CONTAINER]",
		Short: "Print logs for a container in a pod across managed clusters",
		RunE: func(cmd *cobra.Command, args []string) error {
			return handleLogsCommand(args, ...)
		},
	}
	return cmd
}
```

2. **Register in root command**: `pkg/cmd/root.go`
```go
func init() {
	rootCmd.AddCommand(newLogsCommand())
}
```

3. **Implement handler**: Handle multi-cluster logic

### Testing Strategy

```bash
# Unit tests
go test ./pkg/...

# Integration testing with real clusters
kubectl multi get nodes
kubectl multi get pods -A
kubectl multi get services -n kube-system

# Edge cases
kubectl multi get nonexistent-resource
kubectl multi get pods -n nonexistent-namespace
```

## Contributing

### Development Workflow

1. **Fork the repository**
2. **Create feature branch**: `git checkout -b feature/new-command`
3. **Make changes**: Follow Go best practices
4. **Add tests**: Test new functionality
5. **Run checks**: `make check`
6. **Submit PR**: With detailed description

### Code Style

- **Go formatting**: Use `gofmt` and `go vet`
- **Error handling**: Always handle errors gracefully
- **Documentation**: Comment exported functions
- **Testing**: Add tests for new functionality

### Architecture Principles

1. **Graceful degradation**: Continue operation if some clusters fail
2. **Unified output**: Maintain consistent tabular format
3. **kubectl compatibility**: Support all standard kubectl flags
4. **Performance**: Parallel operations where possible
5. **User experience**: Clear error messages and help text

## Related Projects

- [KubeStellar](https://github.com/kubestellar/kubestellar) - Multi-cluster configuration management
- [kubectl](https://kubernetes.io/docs/reference/kubectl/) - Kubernetes command-line tool
- [Cobra](https://github.com/spf13/cobra) - CLI framework for Go
- [client-go](https://github.com/kubernetes/client-go) - Official Kubernetes Go client

## Support

For issues and questions:
- File an issue in this repository  
- Check the KubeStellar documentation
- Join the KubeStellar community discussions

## License

This project is licensed under the Apache License 2.0. See the LICENSE file for details. 