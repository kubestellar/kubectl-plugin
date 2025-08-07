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
- **Binding Policy Guidance**: Encourages label-based binding policies for better multi-cluster management
- **Multi-cluster logs**: Stream logs from containers across all clusters simultaneously
- **Selective deployment**: Deploy to specific clusters or clusters matching label selectors
- **Binding Policy creation**: Generate KubeStellar BindingPolicy YAML for declarative multi-cluster management

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

# Create resources across all clusters (use with caution)
kubectl multi create deployment nginx --image=nginx

# Apply resources from file
kubectl multi apply -f deployment.yaml

# Delete resources across all clusters
kubectl multi delete deployment nginx

# Get logs from a pod across all clusters
kubectl multi logs nginx

# Follow logs from all clusters simultaneously
kubectl multi logs -f nginx

# Get logs with container selection
kubectl multi logs nginx-app -c nginx

# Deploy to specific clusters only
kubectl multi deploy-to --clusters=cluster1,cluster2 -f deployment.yaml

# Deploy with global image override
kubectl multi deploy-to --clusters=cluster1,cluster2 --image=nginx:latest deployment web-app

# Deploy with per-cluster image overrides
kubectl multi deploy-to --clusters=cluster1,cluster2 \
  --cluster-images="cluster1=nginx:1.20" \
  --cluster-images="cluster2=nginx:1.21" \
  deployment versioned-app

# Create a KubeStellar binding policy
kubectl multi create-binding-policy nginx-policy \
  --cluster-labels="env=prod,location=edge" \
  --resource-labels="app=nginx,tier=frontend"

# List available clusters
kubectl multi deploy-to --list-clusters

# Show binding policy examples
kubectl multi demo-binding-policy

# Show cluster labeling examples
kubectl multi label-cluster-examples
```

### Global Flags

- `--kubeconfig string`: Path to kubeconfig file
- `--remote-context string`: Remote hosting context (default: "its1")
- `--all-clusters`: Operate on all managed clusters (default: true)
- `-n, --namespace string`: Target namespace
- `-A, --all-namespaces`: List resources across all namespaces

### KubeStellar Binding Policies (Recommended Approach)

For better multi-cluster management, KubeStellar recommends using binding policies instead of direct resource creation:

```yaml
# Example: Create a binding policy for nginx deployment
apiVersion: control.kubestellar.io/v1alpha1
kind: BindingPolicy
metadata:
  name: nginx-policy
spec:
  clusterSelectors:
  - matchLabels:
      location: edge      # Select clusters by label
      environment: prod
  downsync:
  - objectSelectors:
    - matchLabels:
        app: nginx       # Select resources by label
        tier: frontend
```

Benefits of binding policies:
- **Declarative**: Define once, apply everywhere automatically
- **Selective**: Target specific clusters using labels
- **Dynamic**: Automatically applies to new clusters matching criteria
- **Manageable**: Update policies to change deployment patterns

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
│   │   ├── describe.go    # Describe command
│   │   ├── apply.go       # Apply command with binding policy guidance
│   │   ├── delete.go      # Delete command
│   │   ├── create.go      # Create command with binding policy warnings
│   │   ├── logs.go        # Logs command with multi-cluster streaming
│   │   ├── binding_policy.go # KubeStellar BindingPolicy creation helper
│   │   └── deploy_to.go   # Selective deployment to specific clusters
│   ├── cluster/           # Cluster discovery & management
│   │   └── discovery.go   # KubeStellar cluster discovery
│   └── util/              # Utility functions
│       └── formatting.go  # Resource formatting utilities
└── bin/                   # Build output directory
    └── kubectl-multi      # Compiled binary
```

### Adding New Commands

To add a new kubectl command:

1. **Create command file**: e.g., `pkg/cmd/newcommand.go`
2. **Register in root command**: Add to `init()` in `pkg/cmd/root.go`
3. **Implement handler**: Handle multi-cluster logic
4. **Add help function**: Provide multi-cluster specific help
5. **Update README**: Document the new command

### Command Implementation Guidelines

#### Create Command
The `create` command includes warnings about using binding policies:
- Shows warnings when creating resources directly
- Provides examples of using binding policies instead
- Supports all standard kubectl create subcommands

#### Logs Command  
The `logs` command supports:
- Sequential log retrieval for non-follow mode
- Concurrent streaming with cluster prefixes for follow mode
- All standard kubectl logs flags (tail, since, timestamps, etc.)

#### Delete Command
The `delete` command provides:
- Support for file-based and resource-based deletion
- Label selectors and --all flag support
- Graceful error handling across clusters

#### Deploy-To Command
The `deploy-to` command offers selective deployment:
- Deploy to specific clusters by name
- Target clusters using label selectors
- List available clusters and their status
- Dry-run capability to preview deployments

#### Binding Policy Commands
KubeStellar BindingPolicy creation helpers:
- `create-binding-policy`: Generate BindingPolicy YAML with cluster and resource selectors
- `demo-binding-policy`: Show common BindingPolicy examples
- `label-cluster-examples`: Display cluster labeling patterns

#### Advanced Multi-Cluster Features

**Image Management Across Clusters:**
```bash
# Deploy same image to all specified clusters
kubectl multi deploy-to --clusters=prod-east,prod-west --image=app:v2.0 deployment myapp

# Deploy different image versions per cluster
kubectl multi deploy-to --clusters=staging,prod \
  --cluster-images="staging=app:v2.0-beta" \
  --cluster-images="prod=app:v1.5" \
  deployment myapp

# Use with existing deployment files
kubectl multi deploy-to --clusters=cluster1,cluster2 \
  --image=nginx:alpine \
  -f deployment.yaml
```

**Selective Operations:**
```bash
# Apply configurations to specific clusters only
kubectl multi deploy-to --clusters=cluster1,cluster2 -f configmap.yaml

# Get logs from specific clusters
kubectl multi deploy-to --clusters=production --dry-run logs -l app=nginx

# Delete resources from targeted clusters
kubectl multi deploy-to --clusters=staging delete deployment test-app
```

**KubeStellar Workflow Integration:**
```bash
# 1. Generate binding policy with proper selectors
kubectl multi create-binding-policy production-nginx \
  --cluster-labels="env=production,region=us-east" \
  --resource-labels="app=nginx,version=stable"

# 2. Apply the generated policy
kubectl apply -f nginx-binding-policy.yaml

# 3. Label your clusters
kubectl label managedcluster prod-east-1 env=production region=us-east
kubectl label managedcluster prod-west-1 env=production region=us-west

# 4. Deploy with proper labels (will be automatically distributed)
kubectl label deployment nginx app=nginx version=stable
kubectl apply -f nginx-deployment.yaml
```

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