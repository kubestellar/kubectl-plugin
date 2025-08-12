# kubectl-multi

A comprehensive kubectl plugin for multi-cluster operations with KubeStellar. This plugin extends kubectl to work seamlessly across all KubeStellar managed clusters, providing unified views and operations while filtering out workflow staging clusters (WDS).

## Overview

kubectl-multi is a kubectl plugin written in Go that automatically discovers KubeStellar managed clusters and executes kubectl commands across all of them simultaneously. It provides a unified tabular output with cluster context information, making it easy to monitor and manage resources across multiple clusters.

### Key Features

- **Multi-cluster resource viewing**: Get resources from all managed clusters with unified output
- **Cluster context identification**: Each resource shows which cluster it belongs to
- **All kubectl commands**: Supports all major kubectl commands across clusters
- **KubeStellar integration**: Automatically discovers managed clusters via KubeStellar APIs
- **WDS filtering**: Automatically excludes Workload Description Space clusters
- **Familiar syntax**: Uses the same command structure as kubectl

## Quick Start

```bash
# Install the plugin
make install

# Get nodes from all managed clusters
kubectl multi get nodes

# Get pods from all clusters in all namespaces
kubectl multi get pods -A
```

## Documentation

- **[Installation Guide](docs/installation.md)** - How to install and set up kubectl-multi
- **[Usage Guide](docs/usage.md)** - Detailed usage examples and commands
- **[Architecture Guide](docs/architecture.md)** - Technical architecture and how it works
- **[Development Guide](docs/development.md)** - Contributing and development workflow
- **[API Reference](docs/api-reference.md)** - Code organization and technical implementation

## Tech Stack

- **Go 1.21+**: Primary language for the plugin
- **Cobra**: CLI framework for command structure and parsing
- **Kubernetes client-go**: Official Kubernetes Go client library
- **KubeStellar APIs**: For managed cluster discovery

## Example Output

```
CONTEXT  CLUSTER       NAME                    STATUS  ROLES          AGE    VERSION
its1     cluster1      cluster1-control-plane  Ready   control-plane  6d23h  v1.33.1
its1     cluster2      cluster2-control-plane  Ready   control-plane  6d23h  v1.33.1
its1     its1-cluster  kubeflex-control-plane  Ready   <none>         6d23h  v1.27.2+k3s1
```

## Related Projects

- [KubeStellar](https://github.com/kubestellar/kubestellar) - Multi-cluster configuration management
- [kubectl](https://kubernetes.io/docs/reference/kubectl/) - Kubernetes command-line tool

## Support

For issues and questions:
- File an issue in this repository  
- Check the KubeStellar documentation
- Join the KubeStellar community discussions

## License

This project is licensed under the Apache License 2.0. See the LICENSE file for details.