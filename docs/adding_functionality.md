# Adding New Functionality to KubeStellar CLI

This guide explains how to extend the KubeStellar CLI tools with new features, commands, and operations. The architecture is designed to make adding functionality straightforward while maintaining consistency between the kubectl plugin and standalone CLI.

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Adding Core Operations](#adding-core-operations)
3. [Adding CLI Commands](#adding-cli-commands)
4. [Adding Interactive Features](#adding-interactive-features)
5. [Testing New Features](#testing-new-features)
6. [Documentation Requirements](#documentation-requirements)
7. [Best Practices](#best-practices)
8. [Examples](#examples)

## Architecture Overview

The KubeStellar CLI follows a layered architecture that promotes code reuse and maintainability:

```
┌─────────────────────────────────────────────────────────────┐
│                    User Interfaces                         │
├─────────────────────────┬───────────────────────────────────┤
│   kubectl-multi         │        kubestellar               │
│   (kubectl plugin)      │     (standalone CLI)             │
├─────────────────────────┴───────────────────────────────────┤
│                  Command Layer                              │
│  pkg/kubectl/cmd/       │       pkg/cli/cmd/               │
├─────────────────────────┴───────────────────────────────────┤
│                 Core Operations Layer                       │
│                pkg/core/operations/                         │
├─────────────────────────────────────────────────────────────┤
│                  Foundation Layer                           │
│     pkg/cluster/    pkg/util/    pkg/cli/interactive/       │
└─────────────────────────────────────────────────────────────┘
```

### Key Principles

1. **Shared Core Logic**: All business logic goes in `pkg/core/operations/`
2. **CLI-Specific Interfaces**: Command definitions in `pkg/kubectl/cmd/` and `pkg/cli/cmd/`
3. **Consistent API**: Both CLIs expose the same functionality with appropriate UX differences
4. **Testable Components**: Each layer can be tested independently

## Adding Core Operations

Core operations are the building blocks that both CLI interfaces use. They contain the actual business logic for multi-cluster operations.

### Step 1: Create the Operation File

Create a new file in `pkg/core/operations/` for your operation:

```go
// pkg/core/operations/newfeature.go
package operations

import (
    "context"
    "fmt"
    
    "kubectl-multi/pkg/cluster"
    // Add other necessary imports
)

// NewFeatureOptions contains all options for the new feature operation
type NewFeatureOptions struct {
    Clusters      []cluster.ClusterInfo
    ResourceType  string
    Namespace     string
    AllNamespaces bool
    // Add feature-specific options
    FeatureFlag   bool
    ConfigValue   string
}

// ExecuteNewFeature performs the new feature operation across all clusters
func ExecuteNewFeature(opts NewFeatureOptions) error {
    if len(opts.Clusters) == 0 {
        return fmt.Errorf("no clusters specified")
    }

    successCount := 0
    failureCount := 0

    for _, clusterInfo := range opts.Clusters {
        fmt.Printf("--- Executing new feature on cluster: %s ---\n", clusterInfo.Name)
        
        // Implement your feature logic here
        err := executeOnCluster(clusterInfo, opts)
        if err != nil {
            fmt.Printf("ERROR in cluster %s: %v\n", clusterInfo.Name, err)
            failureCount++
        } else {
            fmt.Printf("SUCCESS in cluster %s\n", clusterInfo.Name)
            successCount++
        }
    }

    fmt.Printf("--- Summary ---\n")
    fmt.Printf("Successfully executed on %d cluster(s)\n", successCount)
    if failureCount > 0 {
        fmt.Printf("Failed on %d cluster(s)\n", failureCount)
        return fmt.Errorf("operation failed on %d cluster(s)", failureCount)
    }

    return nil
}

// executeOnCluster implements the cluster-specific logic
func executeOnCluster(clusterInfo cluster.ClusterInfo, opts NewFeatureOptions) error {
    // Implementation specific to your feature
    // This could involve:
    // - Kubernetes API calls using clusterInfo.Client
    // - Dynamic client operations using clusterInfo.DynamicClient
    // - Custom resource operations
    // - External tool invocations
    
    return nil
}
```

### Step 2: Add Supporting Functions

Add any helper functions your operation needs:

```go
// Helper functions for your operation
func validateNewFeatureOptions(opts NewFeatureOptions) error {
    if opts.ResourceType == "" {
        return fmt.Errorf("resource type is required")
    }
    // Add other validations
    return nil
}

func buildKubernetesObject(opts NewFeatureOptions) map[string]interface{} {
    // Build Kubernetes objects if needed
    return map[string]interface{}{
        "apiVersion": "v1",
        "kind":       "ConfigMap",
        // ... other fields
    }
}
```

### Step 3: Add Tests

Create comprehensive tests for your operation:

```go
// pkg/core/operations/newfeature_test.go
package operations

import (
    "testing"
    
    "kubectl-multi/pkg/cluster"
)

func TestExecuteNewFeature(t *testing.T) {
    tests := []struct {
        name    string
        opts    NewFeatureOptions
        wantErr bool
    }{
        {
            name: "successful operation",
            opts: NewFeatureOptions{
                Clusters: []cluster.ClusterInfo{
                    {Name: "test-cluster"},
                },
                ResourceType: "configmap",
                FeatureFlag:  true,
            },
            wantErr: false,
        },
        {
            name: "no clusters specified",
            opts: NewFeatureOptions{
                ResourceType: "configmap",
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ExecuteNewFeature(tt.opts)
            if (err != nil) != tt.wantErr {
                t.Errorf("ExecuteNewFeature() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

## Adding CLI Commands

Once you have a core operation, you need to create command interfaces for both CLIs.

### Step 4: Add Kubectl Plugin Command

Create the kubectl plugin command in `pkg/kubectl/cmd/`:

```go
// pkg/kubectl/cmd/newfeature.go
package cmd

import (
    "fmt"

    "github.com/spf13/cobra"

    "kubectl-multi/pkg/cluster"
    "kubectl-multi/pkg/core/operations"
    "kubectl-multi/pkg/util"
)

func newNewFeatureCommand() *cobra.Command {
    var featureFlag bool
    var configValue string

    cmd := &cobra.Command{
        Use:   "newfeature [RESOURCE_TYPE] [flags]",
        Short: "Execute new feature operation across all managed clusters",
        Long: `Execute new feature operation across all KubeStellar managed clusters.
This command demonstrates how to add new functionality.`,
        Example: `# Execute new feature on all clusters
kubectl multi newfeature configmaps

# Execute with specific configuration
kubectl multi newfeature deployments --config-value=production --feature-flag`,
        RunE: func(cmd *cobra.Command, args []string) error {
            if len(args) == 0 {
                return fmt.Errorf("resource type must be specified")
            }

            kubeconfig, remoteCtx, _, namespace, allNamespaces := GetGlobalFlags()

            // Discover clusters
            clusters, err := cluster.DiscoverClusters(kubeconfig, remoteCtx)
            if err != nil {
                return fmt.Errorf("failed to discover clusters: %v", err)
            }

            // Build options and execute
            opts := operations.NewFeatureOptions{
                Clusters:      clusters,
                ResourceType:  args[0],
                Namespace:     namespace,
                AllNamespaces: allNamespaces,
                FeatureFlag:   featureFlag,
                ConfigValue:   configValue,
            }

            return operations.ExecuteNewFeature(opts)
        },
    }

    cmd.Flags().BoolVar(&featureFlag, "feature-flag", false, "enable feature flag")
    cmd.Flags().StringVar(&configValue, "config-value", "", "configuration value for the feature")

    // Add help function if needed
    cmd.SetHelpFunc(newFeatureHelpFunc)

    return cmd
}

func newFeatureHelpFunc(cmd *cobra.Command, args []string) {
    // Custom help implementation
    fmt.Fprintln(cmd.OutOrStdout(), "Custom help for new feature command")
}
```

### Step 5: Add Standalone CLI Command

Create the standalone CLI command in `pkg/cli/cmd/`:

```go
// pkg/cli/cmd/newfeature.go
package cmd

import (
    "fmt"

    "github.com/spf13/cobra"

    "kubectl-multi/pkg/cluster"
    "kubectl-multi/pkg/core/operations"
)

func newNewFeatureCommand() *cobra.Command {
    var featureFlag bool
    var configValue string

    cmd := &cobra.Command{
        Use:   "newfeature [RESOURCE_TYPE] [flags]",
        Short: "Execute new feature operation across all managed clusters",
        Long: `Execute new feature operation across all KubeStellar managed clusters.
This command demonstrates how to add new functionality to the standalone CLI.`,
        Example: `# Execute new feature on all clusters
kubestellar newfeature configmaps

# Execute with specific configuration
kubestellar newfeature deployments --config-value=production --feature-flag`,
        RunE: func(cmd *cobra.Command, args []string) error {
            if len(args) == 0 {
                return fmt.Errorf("resource type must be specified")
            }

            kubeconfig, remoteCtx, _, namespace, allNamespaces := GetGlobalFlags()

            // Discover clusters
            clusters, err := cluster.DiscoverClusters(kubeconfig, remoteCtx)
            if err != nil {
                return fmt.Errorf("failed to discover clusters: %v", err)
            }

            // Build options and execute
            opts := operations.NewFeatureOptions{
                Clusters:      clusters,
                ResourceType:  args[0],
                Namespace:     namespace,
                AllNamespaces: allNamespaces,
                FeatureFlag:   featureFlag,
                ConfigValue:   configValue,
            }

            return operations.ExecuteNewFeature(opts)
        },
    }

    cmd.Flags().BoolVar(&featureFlag, "feature-flag", false, "enable feature flag")
    cmd.Flags().StringVar(&configValue, "config-value", "", "configuration value for the feature")

    return cmd
}
```

### Step 6: Register Commands

Add your commands to the root commands:

```go
// In pkg/kubectl/cmd/root.go
func init() {
    // Add other commands...
    rootCmd.AddCommand(newNewFeatureCommand())
}

// In pkg/cli/cmd/root.go  
func init() {
    // Add other commands...
    rootCmd.AddCommand(newNewFeatureCommand())
}
```

## Adding Interactive Features

For the standalone CLI's interactive mode, add your command to the interactive handler:

### Step 7: Add Interactive Command Handler

Update `pkg/cli/interactive/interactive.go`:

```go
// Add to the runCommandLoop function's switch statement
case "newfeature", "nf":
    cli.handleNewFeatureCommand(args)

// Add the handler function
func (cli *InteractiveCLI) handleNewFeatureCommand(args []string) {
    if len(args) == 0 {
        cli.red.Println("Please specify a resource type. Example: newfeature configmaps")
        return
    }
    
    cli.green.Printf("Executing new feature on %s across all clusters...\n", args[0])
    
    // TODO: Call ExecuteNewFeature from operations
    // For now, show example output
    fmt.Printf("CLUSTER    RESOURCE     STATUS\n")
    fmt.Printf("cluster1   %s          Success\n", args[0])
    fmt.Printf("cluster2   %s          Success\n", args[0])
}

// Add to the showHelp function
case "newfeature", "nf":
    cli.yellow.Println("New Feature Commands:")
    fmt.Println("  newfeature <resource-type>        - Execute new feature on resource type")
    fmt.Println("  newfeature configmaps             - Execute on configmaps")
    fmt.Println("  newfeature deployments --verbose  - Execute with verbose output")
```

## Testing New Features

### Step 8: Integration Tests

Create integration tests that verify the end-to-end functionality:

```go
// tests/integration/newfeature_test.go
package integration

import (
    "testing"
    "os/exec"
    "strings"
)

func TestNewFeatureKubectlPlugin(t *testing.T) {
    // Test kubectl plugin
    cmd := exec.Command("./bin/kubectl-multi", "newfeature", "configmaps", "--help")
    output, err := cmd.CombinedOutput()
    
    if err != nil {
        t.Fatalf("Command failed: %v", err)
    }
    
    if !strings.Contains(string(output), "Execute new feature operation") {
        t.Errorf("Help output doesn't contain expected text")
    }
}

func TestNewFeatureStandaloneCLI(t *testing.T) {
    // Test standalone CLI
    cmd := exec.Command("./bin/kubestellar", "newfeature", "deployments", "--help")
    output, err := cmd.CombinedOutput()
    
    if err != nil {
        t.Fatalf("Command failed: %v", err)
    }
    
    if !strings.Contains(string(output), "Execute new feature operation") {
        t.Errorf("Help output doesn't contain expected text")
    }
}
```

### Step 9: Unit Tests

Ensure comprehensive unit test coverage:

```bash
# Run tests
go test ./pkg/core/operations/...
go test ./pkg/kubectl/cmd/...
go test ./pkg/cli/cmd/...

# Check coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Documentation Requirements

### Step 10: Add Documentation

Create documentation for your new feature:

```markdown
<!-- docs/newfeature.md -->
# New Feature Guide

## Overview

The new feature allows you to...

## Usage

### kubectl Plugin

```bash
kubectl multi newfeature configmaps
kubectl multi newfeature deployments --feature-flag
```

### Standalone CLI

```bash
kubestellar newfeature configmaps
kubestellar newfeature deployments --feature-flag
```

### Interactive Mode

```bash
kubestellar
kubestellar> newfeature configmaps
```

## Examples

[Add comprehensive examples here]

## See Also

- [Usage Guide](usage_guide.md)
- [Architecture Guide](architecture_guide.md)
```

### Step 11: Update Main Documentation

Update the main documentation files:

1. **README.md**: Add your feature to the features list
2. **docs/usage_guide.md**: Add usage examples
3. **docs/api_reference.md**: Document the API if applicable

## Best Practices

### Code Organization

1. **Single Responsibility**: Each operation should have a single, clear purpose
2. **Error Handling**: Provide clear error messages and proper error propagation
3. **Logging**: Use consistent logging patterns
4. **Configuration**: Make features configurable through options

### API Design

1. **Consistent Options**: Follow the pattern of existing `*Options` structs
2. **Validation**: Validate inputs early and provide clear error messages
3. **Extensibility**: Design options structs to be easily extended
4. **Backwards Compatibility**: Don't break existing APIs

### Testing

1. **Unit Tests**: Test individual functions and components
2. **Integration Tests**: Test the complete feature end-to-end
3. **Error Cases**: Test error conditions and edge cases
4. **Mock Dependencies**: Use mocks for external dependencies

### Documentation

1. **Inline Comments**: Document complex logic and business rules
2. **Examples**: Provide practical examples for each feature
3. **Error Scenarios**: Document common error scenarios and solutions
4. **Performance**: Document performance characteristics if relevant

## Examples

### Example 1: Simple Resource Operation

Here's a complete example of adding a "drain" operation:

```go
// pkg/core/operations/drain.go
package operations

type DrainOptions struct {
    Clusters         []cluster.ClusterInfo
    NodeName         string
    Force           bool
    IgnoreDaemonsets bool
    Timeout         time.Duration
}

func ExecuteDrain(opts DrainOptions) error {
    // Implementation
}

// pkg/kubectl/cmd/drain.go
func newDrainCommand() *cobra.Command {
    // Command implementation
}

// pkg/cli/cmd/drain.go  
func newDrainCommand() *cobra.Command {
    // Command implementation
}
```

### Example 2: Custom Resource Operation

Example for managing custom resources:

```go
// pkg/core/operations/customresource.go
package operations

import (
    "k8s.io/apimachinery/pkg/runtime/schema"
    "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type CustomResourceOptions struct {
    Clusters    []cluster.ClusterInfo
    GVR         schema.GroupVersionResource
    Object      *unstructured.Unstructured
    Operation   string // create, update, delete
}

func ExecuteCustomResource(opts CustomResourceOptions) error {
    for _, clusterInfo := range opts.Clusters {
        switch opts.Operation {
        case "create":
            _, err := clusterInfo.DynamicClient.Resource(opts.GVR).
                Namespace(opts.Object.GetNamespace()).
                Create(context.TODO(), opts.Object, metav1.CreateOptions{})
            if err != nil {
                return fmt.Errorf("failed to create in cluster %s: %v", clusterInfo.Name, err)
            }
        // Handle other operations...
        }
    }
    return nil
}
```

### Example 3: External Tool Integration

Example for integrating external tools:

```go
// pkg/core/operations/external.go
package operations

import (
    "os/exec"
    "bytes"
)

type ExternalToolOptions struct {
    Clusters []cluster.ClusterInfo
    Tool     string
    Args     []string
}

func ExecuteExternalTool(opts ExternalToolOptions) error {
    for _, clusterInfo := range opts.Clusters {
        args := append(opts.Args, "--context", clusterInfo.Context)
        
        cmd := exec.Command(opts.Tool, args...)
        var stdout, stderr bytes.Buffer
        cmd.Stdout = &stdout
        cmd.Stderr = &stderr
        
        if err := cmd.Run(); err != nil {
            fmt.Printf("Error in cluster %s: %s\n", clusterInfo.Name, stderr.String())
            continue
        }
        
        fmt.Printf("=== Cluster: %s ===\n%s\n", clusterInfo.Name, stdout.String())
    }
    return nil
}
```

## Build and Test Integration

### Step 12: Update Build System

Update the Makefile if needed:

```makefile
# Add any new build targets
test-newfeature:
	go test -v ./pkg/core/operations/newfeature_test.go

# Update existing targets
test: test-newfeature
	go test -v ./...
```

### Step 13: Continuous Integration

Ensure your feature works in CI:

```yaml
# .github/workflows/test.yml (if using GitHub Actions)
- name: Test New Feature
  run: |
    make build
    ./bin/kubectl-multi newfeature --help
    ./bin/kubestellar newfeature --help
```

## Release Process

### Step 14: Version and Release

1. **Update Version**: Bump version in appropriate files
2. **Changelog**: Add your feature to the changelog
3. **Release Notes**: Document the new functionality
4. **Migration Guide**: If breaking changes, provide migration guidance

## Getting Help

If you need assistance while adding functionality:

1. **Code Review**: Submit a draft PR for early feedback
2. **Documentation**: Check existing patterns in similar commands
3. **Testing**: Look at existing tests for patterns and best practices
4. **Community**: Reach out to the KubeStellar community for guidance

## See Also

- [Development Guide](development_guide.md) - General development practices
- [Architecture Guide](architecture_guide.md) - Understanding the system architecture
- [Testing Guide](testing_guide.md) - Comprehensive testing strategies
- [API Reference](api_reference.md) - Complete API documentation