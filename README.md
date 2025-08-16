# KubeStellar CLI Tools

A comprehensive multi-cluster management solution for KubeStellar with both a kubectl plugin and standalone CLI. These tools extend Kubernetes operations across all KubeStellar managed clusters, providing unified views and operations while intelligently filtering workflow staging clusters.

## Overview

The KubeStellar CLI tools provide two interfaces for the same powerful multi-cluster functionality:

- **`kubectl multi`** - A kubectl plugin for users who prefer kubectl integration
- **`kubestellar`** - A standalone CLI with enhanced KubeStellar-specific features

Both tools share the same core functionality through a unified architecture, ensuring feature parity and consistent behavior.

## 🌟 Key Features

### Multi-Cluster Operations
- **Unified Resource Views**: Get resources from all managed clusters with consolidated output
- **Cluster Context Identification**: Each resource shows its origin cluster
- **Intelligent Filtering**: Automatically excludes ITS and WDS clusters from workload operations
- **Parallel Execution**: Execute commands across multiple clusters simultaneously

### KubeStellar Integration
- **Automatic Discovery**: Discovers managed clusters via KubeStellar APIs
- **BindingPolicy Management**: Create and manage binding policies for workload distribution
- **Cluster Registration**: Register and manage clusters in ITS (Inventory and Transport Space)
- **Helm Multi-Cluster Support**: Deploy Helm charts across multiple clusters

### Enhanced User Experience
- **Interactive Mode**: Beautiful interactive CLI with KubeStellar branding
- **Dual CLI Support**: Choose between kubectl plugin or standalone CLI
- **Rich Help System**: Context-aware help with kubectl command integration
- **Configuration Management**: Persistent configuration and preferences

## 🚀 Quick Start

### Installation

```bash
# Build both CLIs
make build

# Install both CLIs locally
make install

# Or install system-wide (requires sudo)
make install-system
```

### Basic Usage

#### Using the kubectl plugin:
```bash
# Get nodes from all managed clusters
kubectl multi get nodes

# Apply deployment to all clusters
kubectl multi apply -f deployment.yaml

# List pods across all clusters
kubectl multi get pods -A
```

#### Using the standalone CLI:
```bash
# Get nodes from all managed clusters
kubestellar get nodes

# Start interactive mode
kubestellar

# Manage binding policies
kubestellar bindingpolicy create app-policy

# Deploy Helm charts
kubestellar helm install nginx nginx/nginx
```

## 📖 Documentation

### User Guides
- **[Installation Guide](docs/installation_guide.md)** - Complete installation and setup instructions
- **[Usage Guide](docs/usage_guide.md)** - Detailed usage examples and best practices
- **[Interactive CLI Guide](docs/interactive_cli.md)** - Using the interactive mode features

### Technical Documentation
- **[Architecture Guide](docs/architecture_guide.md)** - Technical architecture and design decisions
- **[Development Guide](docs/development_guide.md)** - Contributing and development workflow
- **[API Reference](docs/api_reference.md)** - Code organization and technical implementation

### KubeStellar Integration
- **[BindingPolicy Management](docs/bindingpolicy.md)** - Managing workload distribution policies
- **[Cluster Management](docs/cluster_management.md)** - Registering and managing clusters
- **[Helm Integration](docs/helm_integration.md)** - Multi-cluster Helm deployments

## 🏗️ Architecture

```
KubeStellar CLI Tools
├── kubectl-multi          # kubectl plugin binary
├── kubestellar             # standalone CLI binary
├── pkg/
│   ├── core/operations/    # Shared core functionality
│   │   ├── get.go         # Multi-cluster get operations
│   │   ├── apply.go       # Multi-cluster apply operations
│   │   ├── helm.go        # Helm multi-cluster operations
│   │   ├── bindingpolicy.go   # BindingPolicy management
│   │   └── cluster_management.go # Cluster registration
│   ├── kubectl/cmd/       # kubectl plugin commands
│   ├── cli/cmd/           # standalone CLI commands
│   ├── cli/interactive/   # Interactive CLI mode
│   ├── cluster/           # Cluster discovery
│   └── util/              # Shared utilities
```

## 🛠️ Available Commands

### Core Operations (Both CLIs)

| kubectl multi | kubestellar | Description |
|--------------|-------------|-------------|
| `kubectl multi get` | `kubestellar get` | Get resources from all clusters |
| `kubectl multi apply` | `kubestellar apply` | Apply configuration to all clusters |
| `kubectl multi delete` | `kubestellar delete` | Delete resources from all clusters |
| `kubectl multi describe` | `kubestellar describe` | Describe resources across clusters |
| `kubectl multi logs` | `kubestellar logs` | Get logs from pods across clusters |

### KubeStellar-Specific Features (Standalone CLI)

| Command | Description |
|---------|-------------|
| `kubestellar clusters list` | List all registered clusters |
| `kubestellar clusters add` | Register a new cluster with ITS |
| `kubestellar bindingpolicy create` | Create a new binding policy |
| `kubestellar helm install` | Install Helm chart across clusters |
| `kubestellar config` | Manage CLI configuration |

### Interactive Mode

```bash
kubestellar
```

Launches an interactive mode with:
- Beautiful ASCII art banner
- Command completion and suggestions
- Contextual help system
- Real-time cluster status
- Configuration management

## 🔧 Configuration

### Global Flags

```bash
--kubeconfig string        Path to kubeconfig file
--remote-context string    ITS context (default: "its1")
--namespace string         Target namespace
--all-namespaces          Operate across all namespaces
```

### Environment Variables

```bash
export KUBECONFIG=/path/to/kubeconfig
export KUBESTELLAR_ITS_CONTEXT=its1
export KUBESTELLAR_DEFAULT_NAMESPACE=default
```

## 📊 Example Output

### Multi-Cluster Resource View
```
CLUSTER    NAMESPACE    NAME         READY    STATUS    RESTARTS    AGE
cluster1   default      nginx-pod    1/1      Running   0           2d
cluster2   default      nginx-pod    1/1      Running   0           2d
cluster3   default      nginx-pod    1/1      Running   0           2d
```

### BindingPolicy Status
```
NAMESPACE    NAME           CLUSTERS
default      app-policy     all
production   prod-policy    2 selector(s)
staging      test-policy    cluster1,cluster2
```

### Interactive Mode Welcome
```
╭─────────────────────────────────────────────────────────────────────────────────╮
│  ██╗  ██╗██╗   ██╗██████╗ ███████╗███████╗████████╗███████╗██╗     ██╗      █████╗ ██████╗  │
│  ██║ ██╔╝██║   ██║██╔══██╗██╔════╝██╔════╝╚══██╔══╝██╔════╝██║     ██║     ██╔══██╗██╔══██╗ │
│  █████╔╝ ██║   ██║██████╔╝█████╗  ███████╗   ██║   █████╗  ██║     ██║     ███████║██████╔╝ │
│  ██╔═██╗ ██║   ██║██╔══██╗██╔══╝  ╚════██║   ██║   ██╔══╝  ██║     ██║     ██╔══██║██╔══██╗ │
│  ██║  ██╗╚██████╔╝██████╔╝███████╗███████║   ██║   ███████╗███████╗███████╗██║  ██║██║  ██║ │
│  ╚═╝  ╚═╝ ╚═════╝ ╚═════╝ ╚══════╝╚══════╝   ╚═╝   ╚══════╝╚══════╝╚══════╝╚═╝  ╚═╝╚═╝  ╚═╝ │
│                       🌟 Multi-Cluster Kubernetes Management System 🌟                      │
╰─────────────────────────────────────────────────────────────────────────────────╮
```

## 🔗 Integration with KubeStellar

### Cluster Discovery
The CLI tools automatically discover clusters through:
- **ITS (Inventory and Transport Space)**: Central cluster registry
- **ManagedCluster CRDs**: Open Cluster Management resources
- **ControlPlane CRDs**: KubeFlex vCluster instances

### Workload Distribution
- Create and manage BindingPolicies for automated workload distribution
- Target specific clusters or use label selectors
- Support for namespaced and cluster-scoped resources

### Helm Integration
- Deploy Helm charts across multiple clusters simultaneously
- Consistent configuration management with values files
- Rollback and upgrade operations across clusters

## 🚀 Development

### Building from Source

```bash
# Clone the repository
git clone https://github.com/kubestellar/kubectl-plugin
cd kubectl-plugin

# Install dependencies
go mod tidy

# Build both CLIs
make build

# Run tests
make test

# Format code
make fmt
```

### Project Structure

```
├── cmd/
│   ├── kubectl-multi/main.go     # kubectl plugin entry point
│   └── kubestellar/main.go       # standalone CLI entry point
├── pkg/
│   ├── core/operations/          # Shared business logic
│   ├── kubectl/cmd/              # kubectl plugin commands
│   ├── cli/cmd/                  # standalone CLI commands
│   └── cli/interactive/          # Interactive mode
├── docs/                         # Documentation
├── Makefile                      # Build automation
└── README.md                     # This file
```

## 🤝 Contributing

We welcome contributions! Please see our [Development Guide](docs/development_guide.md) for details on:
- Setting up the development environment
- Code style and standards
- Testing procedures
- Submitting pull requests

## 📄 License

This project is licensed under the Apache License 2.0. See the [LICENSE](LICENSE) file for details.

## 🆘 Support

- **Issues**: [GitHub Issues](https://github.com/kubestellar/kubectl-plugin/issues)
- **Discussions**: [KubeStellar Community](https://github.com/kubestellar/kubestellar/discussions)
- **Documentation**: [KubeStellar Docs](https://docs.kubestellar.io)
- **Slack**: [KubeStellar Slack](https://kubernetes.slack.com/channels/kubestellar)

## 🔗 Related Projects

- [KubeStellar](https://github.com/kubestellar/kubestellar) - Multi-cluster configuration management
- [Open Cluster Management](https://open-cluster-management.io/) - Cluster lifecycle management
- [KubeFlex](https://github.com/kubestellar/kubeflex) - Hosting control planes as a service
- [kubectl](https://kubernetes.io/docs/reference/kubectl/) - Kubernetes command-line tool

---

⭐ **Star this repository** if you find it useful!

🐛 **Found a bug?** [Report it here](https://github.com/kubestellar/kubectl-plugin/issues/new)