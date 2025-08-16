# KubeStellar CLI Architecture Guide

This document explains the technical architecture of the KubeStellar CLI tools, including both the kubectl plugin and standalone CLI implementations.

## Table of Contents

1. [High-Level Architecture](#high-level-architecture)
2. [Dual CLI Design](#dual-cli-design)
3. [Core Components](#core-components)
4. [Cluster Discovery](#cluster-discovery)
5. [Command Flow](#command-flow)
6. [Data Flow](#data-flow)
7. [Extension Points](#extension-points)
8. [Design Decisions](#design-decisions)

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              User Interfaces                                   │
├──────────────────────────────────┬──────────────────────────────────────────────┤
│         kubectl-multi            │            kubestellar                       │
│       (kubectl plugin)           │         (standalone CLI)                     │
│                                  │                                              │
│  ┌─────────────────────────────┐ │  ┌─────────────────────────────────────────┐ │
│  │    kubectl multi get        │ │  │       kubestellar get                   │ │
│  │    kubectl multi apply      │ │  │       kubestellar apply                 │ │
│  │    kubectl multi helm       │ │  │       kubestellar helm                  │ │
│  └─────────────────────────────┘ │  │       kubestellar (interactive)         │ │
│                                  │  └─────────────────────────────────────────┘ │
├──────────────────────────────────┴──────────────────────────────────────────────┤
│                             Command Interface Layer                             │
│          pkg/kubectl/cmd/                    pkg/cli/cmd/                       │
├─────────────────────────────────────────────────────────────────────────────────┤
│                           Shared Core Operations                                │
│                            pkg/core/operations/                                 │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐│
│  │    get.go   │ │  apply.go   │ │   helm.go   │ │bindingpol..││cluster_mgmt.││
│  └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘│
├─────────────────────────────────────────────────────────────────────────────────┤
│                              Foundation Layer                                   │
│     pkg/cluster/         pkg/util/         pkg/cli/interactive/                 │
└─────────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                           KubeStellar Infrastructure                            │
│                                                                                 │
│  ┌─────────────────────┐    ┌─────────────────────┐    ┌─────────────────────┐ │
│  │       ITS           │    │       WDS           │    │    WEC Clusters     │ │
│  │ (Control Plane)     │    │ (Workflow Staging)  │    │  (Workload Exec)    │ │
│  │                     │    │                     │    │                     │ │
│  │ • ManagedCluster    │    │ • BindingPolicy     │    │ • Applications      │ │
│  │   CRDs              │    │ • Workload Specs    │    │ • Services          │ │
│  │ • Cluster Registry  │    │ • Distribution      │    │ • Configurations    │ │
│  └─────────────────────┘    │   Logic             │    └─────────────────────┘ │
│                             └─────────────────────┘                            │
└─────────────────────────────────────────────────────────────────────────────────┘
```

## Dual CLI Design

The KubeStellar CLI tools implement a unique dual-interface design that provides both kubectl integration and standalone functionality.

### Design Principles

1. **Shared Core Logic**: All business logic is implemented once in `pkg/core/operations/`
2. **Interface Adaptation**: Each CLI provides an appropriate interface for its user base
3. **Feature Parity**: Both CLIs provide the same core functionality
4. **User Experience**: Each CLI optimizes UX for its specific use case

### Architecture Benefits

```
┌─────────────────────────────────────────────────────────────────┐
│                         Benefits                                │
├─────────────────────────────────────────────────────────────────┤
│  ✓ Code Reuse           │  Single implementation reduces bugs   │
│  ✓ Consistency          │  Same behavior across interfaces      │
│  ✓ Maintainability      │  Changes apply to both CLIs           │
│  ✓ Testing Efficiency   │  Test core logic once                │
│  ✓ User Choice          │  kubectl vs standalone preference     │
│  ✓ Feature Velocity     │  Faster development of new features   │
└─────────────────────────────────────────────────────────────────┘
```

## Core Components

### 1. Operations Layer (`pkg/core/operations/`)

The operations layer contains all business logic for multi-cluster operations:

```go
// Core operation interface pattern
type OperationOptions struct {
    Clusters      []cluster.ClusterInfo
    // Operation-specific fields
}

func ExecuteOperation(opts OperationOptions) error {
    // Multi-cluster logic implementation
}
```

**Key Files:**
- `get.go` - Multi-cluster resource retrieval
- `apply.go` - Multi-cluster resource application
- `helm.go` - Helm chart operations across clusters
- `bindingpolicy.go` - BindingPolicy management
- `cluster_management.go` - ITS cluster registration

### 2. Command Interface Layer

#### kubectl Plugin (`pkg/kubectl/cmd/`)

Provides kubectl-native command interface:

```go
func newGetCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "get [TYPE[.VERSION][.GROUP] [NAME | -l label]",
        Short: "Display resources across all managed clusters",
        RunE: func(cmd *cobra.Command, args []string) error {
            // Parse kubectl-style arguments
            // Call core operations
            return operations.ExecuteGet(opts)
        },
    }
    return cmd
}
```

#### Standalone CLI (`pkg/cli/cmd/`)

Provides KubeStellar-optimized interface:

```go
func newGetCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "get [TYPE] [NAME]",
        Short: "Display resources across all managed clusters", 
        Example: `kubestellar get pods
kubestellar get deployments -n production`,
        RunE: func(cmd *cobra.Command, args []string) error {
            // Parse KubeStellar-style arguments
            // Call core operations  
            return operations.ExecuteGet(opts)
        },
    }
    return cmd
}
```

### 3. Foundation Layer

#### Cluster Discovery (`pkg/cluster/`)

Implements KubeStellar-aware cluster discovery:

```go
type ClusterInfo struct {
    Name             string
    Context          string
    Client           *kubernetes.Clientset
    DynamicClient    dynamic.Interface
    DiscoveryClient  discovery.DiscoveryInterface
    RestConfig       *rest.Config
}

func DiscoverClusters(kubeconfig, remoteCtx string) ([]ClusterInfo, error) {
    // 1. Connect to ITS cluster
    // 2. Query ManagedCluster CRDs
    // 3. Filter out WDS clusters
    // 4. Build client connections
    // 5. Return cluster information
}
```

#### Utilities (`pkg/util/`)

Shared utility functions:
- Resource formatting and display
- kubectl help integration
- Error handling patterns
- Configuration management

#### Interactive Mode (`pkg/cli/interactive/`)

Rich interactive CLI experience:
- ASCII art banner
- Command parsing and routing
- Color-coded output
- Help system integration

## Cluster Discovery

### Discovery Flow

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                           Cluster Discovery Process                            │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                 │
│  1. Connect to ITS         2. Query ManagedClusters    3. Filter Clusters      │
│     ┌─────────────┐           ┌─────────────────┐        ┌─────────────────┐    │
│     │    its1     │────────▶  │ ManagedCluster  │───────▶│  Exclude WDS    │    │
│     │ (Control)   │           │      CRDs       │        │   clusters      │    │
│     └─────────────┘           └─────────────────┘        └─────────────────┘    │
│                                                                   │             │
│                                                                   ▼             │
│  4. Build Connections      5. Validate Access         6. Return ClusterInfo   │
│     ┌─────────────────┐        ┌─────────────────┐        ┌─────────────────┐  │
│     │ Create k8s      │───────▶│ Test API        │───────▶│ []ClusterInfo   │  │
│     │ clients         │        │ connectivity    │        │                 │  │
│     └─────────────────┘        └─────────────────┘        └─────────────────┘  │
└─────────────────────────────────────────────────────────────────────────────────┘
```

### Discovery Implementation

```go
func DiscoverClusters(kubeconfig, remoteCtx string) ([]ClusterInfo, error) {
    var clusters []ClusterInfo
    
    // Step 1: Connect to ITS cluster
    itsClient, err := buildITSClient(kubeconfig, remoteCtx)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to ITS: %v", err)
    }
    
    // Step 2: Query ManagedCluster CRDs
    managedClusters, err := itsClient.Resource(managedClusterGVR).
        List(context.TODO(), metav1.ListOptions{})
    if err != nil {
        return nil, fmt.Errorf("failed to list managed clusters: %v", err)
    }
    
    // Step 3: Process each cluster
    for _, mc := range managedClusters.Items {
        clusterName := mc.GetName()
        
        // Skip WDS clusters
        if isWDSCluster(clusterName) {
            continue
        }
        
        // Build cluster connection
        clusterInfo, err := buildClusterInfo(clusterName, kubeconfig)
        if err != nil {
            log.Printf("Warning: failed to connect to cluster %s: %v", clusterName, err)
            continue
        }
        
        clusters = append(clusters, clusterInfo)
    }
    
    return clusters, nil
}
```

## Command Flow

### kubectl Plugin Flow

```
kubectl multi get pods
       │
       ▼
┌─────────────────┐
│ kubectl binary  │
└─────────────────┘
       │
       ▼
┌─────────────────┐
│ kubectl-multi   │  ◀── Plugin execution
│    plugin       │
└─────────────────┘
       │
       ▼
┌─────────────────┐
│ pkg/kubectl/cmd │  ◀── Command parsing
│   /get.go       │
└─────────────────┘
       │
       ▼
┌─────────────────┐
│ pkg/core/       │  ◀── Core logic
│ operations/     │
│   get.go        │
└─────────────────┘
       │
       ▼
┌─────────────────┐
│ Multi-cluster   │  ◀── Execution
│   execution     │
└─────────────────┘
```

### Standalone CLI Flow

```
kubestellar get pods
       │
       ▼
┌─────────────────┐
│ kubestellar     │  ◀── Direct execution
│   binary        │
└─────────────────┘
       │
       ▼
┌─────────────────┐
│ pkg/cli/cmd/    │  ◀── Command parsing
│   get.go        │
└─────────────────┘
       │
       ▼
┌─────────────────┐
│ pkg/core/       │  ◀── Core logic (shared)
│ operations/     │
│   get.go        │
└─────────────────┘
       │
       ▼
┌─────────────────┐
│ Multi-cluster   │  ◀── Execution
│   execution     │
└─────────────────┘
```

### Interactive Mode Flow

```
kubestellar (no args)
       │
       ▼
┌─────────────────┐
│ pkg/cli/        │  ◀── Interactive mode
│ interactive/    │
└─────────────────┘
       │
       ▼
┌─────────────────┐
│ Display banner  │  ◀── ASCII art + welcome
│ & prompt        │
└─────────────────┘
       │
       ▼
┌─────────────────┐
│ Parse user      │  ◀── Command parsing loop
│ input           │
└─────────────────┘
       │
       ▼
┌─────────────────┐
│ Route to        │  ◀── Delegate to handlers
│ command handler │
└─────────────────┘
       │
       ▼
┌─────────────────┐
│ Execute via     │  ◀── Use shared operations
│ core operations │
└─────────────────┘
```

## Data Flow

### Resource Retrieval Flow

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                            Resource Retrieval Flow                             │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                 │
│  User Command                API Calls               Response Processing       │
│                                                                                 │
│  kubectl multi get pods     ┌─────────────────┐     ┌─────────────────────┐     │
│           │                 │    cluster1     │────▶│     Format &        │     │
│           │             ┌──▶│   GET /pods     │     │     Merge           │     │
│           ▼             │   └─────────────────┘     │    Results          │     │
│  ┌─────────────────┐    │                          └─────────────────────┘     │
│  │ Discover        │────┤   ┌─────────────────┐              │                 │
│  │ Clusters        │    │   │    cluster2     │              ▼                 │
│  └─────────────────┘    │   │   GET /pods     │     ┌─────────────────────┐     │
│                         └──▶└─────────────────┘     │    Display          │     │
│                             ┌─────────────────┐     │   Unified           │     │
│                             │    cluster3     │────▶│   Output            │     │
│                             │   GET /pods     │     └─────────────────────┘     │
│                             └─────────────────┘                                 │
└─────────────────────────────────────────────────────────────────────────────────┘
```

### Resource Application Flow

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                           Resource Application Flow                            │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                 │
│  User Command              Validation                 Parallel Application     │
│                                                                                 │
│  kubectl multi apply      ┌─────────────────┐        ┌─────────────────────┐    │
│    -f deployment.yaml     │   Parse YAML    │        │    cluster1         │    │
│           │               │      &          │    ┌──▶│ POST /deployments   │    │
│           ▼               │   Validate      │    │   └─────────────────────┘    │
│  ┌─────────────────┐      └─────────────────┘    │                              │
│  │ Discover        │               │              │   ┌─────────────────────┐    │
│  │ Clusters        │               ▼              ├──▶│    cluster2         │    │
│  └─────────────────┘      ┌─────────────────┐    │   │ POST /deployments   │    │
│                           │ Determine       │────┤   └─────────────────────┘    │
│                           │ Target          │    │                              │
│                           │ Clusters        │    │   ┌─────────────────────┐    │
│                           └─────────────────┘    └──▶│    cluster3         │    │
│                                                      │ POST /deployments   │    │
│                                                      └─────────────────────┘    │
└─────────────────────────────────────────────────────────────────────────────────┘
```

## Extension Points

The architecture provides several extension points for adding new functionality:

### 1. Core Operations

Add new operations by creating files in `pkg/core/operations/`:

```go
// pkg/core/operations/newfeature.go
package operations

type NewFeatureOptions struct {
    Clusters []cluster.ClusterInfo
    // Feature-specific options
}

func ExecuteNewFeature(opts NewFeatureOptions) error {
    // Implementation
}
```

### 2. CLI Commands

Add commands to both CLI interfaces:

```go
// pkg/kubectl/cmd/newfeature.go & pkg/cli/cmd/newfeature.go
func newNewFeatureCommand() *cobra.Command {
    return &cobra.Command{
        Use: "newfeature",
        RunE: func(cmd *cobra.Command, args []string) error {
            return operations.ExecuteNewFeature(opts)
        },
    }
}
```

### 3. Interactive Commands

Extend the interactive mode:

```go
// pkg/cli/interactive/interactive.go
case "newfeature":
    cli.handleNewFeatureCommand(args)
```

### 4. Cluster Discovery

Extend discovery mechanisms:

```go
// pkg/cluster/discovery.go
func DiscoverClustersByNewMethod() ([]ClusterInfo, error) {
    // New discovery implementation
}
```

## Design Decisions

### 1. Shared Core vs. Duplicated Code

**Decision**: Implement shared core operations layer
**Rationale**: 
- Reduces maintenance burden
- Ensures consistent behavior
- Simplifies testing
- Accelerates feature development

### 2. Dual CLI vs. Single CLI

**Decision**: Provide both kubectl plugin and standalone CLI
**Rationale**:
- kubectl plugin for kubectl users
- Standalone CLI for enhanced KubeStellar features
- Interactive mode for rich user experience
- Flexibility for different user preferences

### 3. Operations-First Design

**Decision**: Design core operations as the primary interface
**Rationale**:
- CLI commands are thin wrappers
- Core logic is reusable and testable
- Easy to add new interfaces (API, GUI, etc.)
- Clear separation of concerns

### 4. KubeStellar-Aware Architecture

**Decision**: Deep integration with KubeStellar concepts
**Rationale**:
- Automatic ITS/WDS/WEC cluster handling
- Native BindingPolicy support
- Cluster discovery via ManagedCluster CRDs
- Intelligent cluster filtering

### 5. Extensible Command Structure

**Decision**: Modular command structure with clear patterns
**Rationale**:
- Easy to add new commands
- Consistent user experience
- Maintainable codebase
- Clear documentation paths

## Performance Considerations

### 1. Parallel Execution

Operations execute across clusters in parallel where possible:

```go
func ExecuteGet(opts GetOptions) error {
    var wg sync.WaitGroup
    results := make(chan ClusterResult, len(opts.Clusters))
    
    for _, cluster := range opts.Clusters {
        wg.Add(1)
        go func(c cluster.ClusterInfo) {
            defer wg.Done()
            result := executeOnCluster(c, opts)
            results <- result
        }(cluster)
    }
    
    // Process results...
}
```

### 2. Connection Reuse

Cluster connections are established once and reused:

```go
type ClusterInfo struct {
    Client          *kubernetes.Clientset    // Reused connection
    DynamicClient   dynamic.Interface        // Reused connection
    // ...
}
```

### 3. Efficient Discovery

Discovery results are cached to avoid repeated API calls:

```go
var discoveryCache map[string][]ClusterInfo
var cacheTimestamp time.Time

func DiscoverClusters(kubeconfig, remoteCtx string) ([]ClusterInfo, error) {
    // Check cache validity
    if time.Since(cacheTimestamp) < cacheTTL {
        return discoveryCache[cacheKey], nil
    }
    // Perform discovery...
}
```

## Security Considerations

### 1. Credential Management

- Respects kubeconfig security model
- No credential storage or caching
- Uses existing authentication mechanisms

### 2. Access Control

- Operations respect RBAC permissions
- No privilege escalation
- Cluster access is user-dependent

### 3. Network Security

- Uses secure Kubernetes API connections
- Respects cluster network policies
- No additional network exposure

## Monitoring and Observability

### 1. Logging

Structured logging throughout the system:

```go
log.WithFields(log.Fields{
    "cluster": clusterInfo.Name,
    "operation": "get",
    "resource": "pods",
}).Info("Executing operation")
```

### 2. Error Handling

Comprehensive error handling with context:

```go
if err != nil {
    return fmt.Errorf("failed to get pods from cluster %s: %w", 
        clusterInfo.Name, err)
}
```

### 3. Metrics

Future: Integration with metrics systems for operational insights.

## See Also

- [Adding Functionality Guide](adding_functionality.md) - How to extend the CLI
- [Development Guide](development_guide.md) - Development practices and setup
- [Usage Guide](usage_guide.md) - User-facing documentation
- [API Reference](api_reference.md) - Complete API documentation