# Interactive CLI Guide

The KubeStellar interactive CLI provides a rich, user-friendly interface for managing multi-cluster environments. This guide covers all aspects of using the interactive mode.

## Starting Interactive Mode

Launch the interactive CLI by running the standalone command without arguments:

```bash
kubestellar
```

This will display the beautiful KubeStellar banner and enter interactive mode.

## Welcome Banner

Upon startup, you'll see the KubeStellar ASCII art banner:

```
╭─────────────────────────────────────────────────────────────────────────────────────────────╮
│                                                                                             │
│  ██╗  ██╗██╗   ██╗██████╗ ███████╗███████╗████████╗███████╗██╗     ██╗      █████╗ ██████╗  │
│  ██║ ██╔╝██║   ██║██╔══██╗██╔════╝██╔════╝╚══██╔══╝██╔════╝██║     ██║     ██╔══██╗██╔══██╗ │
│  █████╔╝ ██║   ██║██████╔╝█████╗  ███████╗   ██║   █████╗  ██║     ██║     ███████║██████╔╝ │
│  ██╔═██╗ ██║   ██║██╔══██╗██╔══╝  ╚════██║   ██║   ██╔══╝  ██║     ██║     ██╔══██║██╔══██╗ │
│  ██║  ██╗╚██████╔╝██████╔╝███████╗███████║   ██║   ███████╗███████╗███████╗██║  ██║██║  ██║ │
│  ╚═╝  ╚═╝ ╚═════╝ ╚═════╝ ╚══════╝╚══════╝   ╚═╝   ╚══════╝╚══════╝╚══════╝╚═╝  ╚═╝╚═╝  ╚═╝ │
│                                                                                             │
│                       🌟 Multi-Cluster Kubernetes Management System 🌟                       │
╰─────────────────────────────────────────────────────────────────────────────────────────────╯
```

## Available Commands

The interactive CLI supports all core KubeStellar functionality organized into intuitive command groups:

### Core Resource Operations

- `get <resource>` - Get resources from all clusters
- `apply <file>` - Apply configurations to all clusters
- `delete <resource>` - Delete resources from all clusters

Example:
```bash
kubestellar> get pods
kubestellar> apply deployment.yaml
kubestellar> delete pod nginx-pod
```

### Cluster Management

- `clusters list` - List all registered clusters
- `clusters add <name> <endpoint>` - Register a new cluster
- `clusters remove <name>` - Remove a cluster from ITS
- `clusters info <name>` - Show detailed cluster information
- `clusters label <name> <labels>` - Update cluster labels

Example:
```bash
kubestellar> clusters list
kubestellar> clusters add my-cluster https://cluster.example.com
kubestellar> clusters info my-cluster
```

### BindingPolicy Management

- `bindingpolicy list` - List all binding policies
- `bindingpolicy create <name>` - Create a new binding policy
- `bindingpolicy delete <name>` - Delete a binding policy
- `bindingpolicy add-clusters <policy> <clusters...>` - Add clusters to policy
- `bindingpolicy remove-clusters <policy> <clusters...>` - Remove clusters from policy

Example:
```bash
kubestellar> bindingpolicy list
kubestellar> bindingpolicy create app-policy
kubestellar> bindingpolicy add-clusters app-policy cluster1 cluster2
```

### Helm Operations

- `helm install <release> <chart>` - Install a Helm chart
- `helm upgrade <release> <chart>` - Upgrade a Helm release
- `helm list` - List Helm releases
- `helm uninstall <release>` - Uninstall a Helm release
- `helm values <release>` - Show values for a release
- `helm rollback <release> <revision>` - Rollback to a previous revision

Example:
```bash
kubestellar> helm install nginx nginx/nginx
kubestellar> helm list
kubestellar> helm upgrade nginx nginx/nginx
```

### Configuration Management

- `config set <key> <value>` - Set a configuration value
- `config get <key>` - Get a configuration value
- `config list` - List all configurations
- `config apply <file>` - Apply configuration from YAML file
- `config export <file>` - Export configuration to YAML file

Example:
```bash
kubestellar> config set its-context its1
kubestellar> config list
kubestellar> config apply kubestellar-config.yaml
```

## Command Help

### General Help

- `help` - Show all available commands
- `help <command>` - Show help for a specific command
- `?` - Alias for help

### Context-Sensitive Help

Each command group has its own help system:

```bash
kubestellar> help clusters
kubestellar> help bindingpolicy
kubestellar> help helm
kubestellar> help config
```

## Interactive Features

### Command Prompt

The interactive prompt shows the current context:

```bash
kubestellar>
```

### Color-Coded Output

- **Green**: Success messages and prompts
- **Yellow**: Informational messages and headers
- **Red**: Error messages
- **Cyan**: Banner and decorative elements

### Special Commands

- `clear` or `cls` - Clear the screen and redisplay banner
- `exit`, `quit`, or `q` - Exit the interactive mode

## Configuration

### Default Settings

The interactive CLI uses these default configurations:

- **ITS Context**: `its1`
- **WDS Context**: `wds1`
- **Default Namespace**: `default`

### Persistent Configuration

Settings can be persisted using the config commands:

```bash
kubestellar> config set its-context my-its
kubestellar> config set default-namespace production
kubestellar> config set wds-context my-wds
```

### Environment Variables

The CLI respects these environment variables:

- `KUBECONFIG` - Path to kubeconfig file
- `KUBESTELLAR_ITS_CONTEXT` - Default ITS context
- `KUBESTELLAR_DEFAULT_NAMESPACE` - Default namespace

## Examples

### Basic Workflow

```bash
# Start interactive mode
kubestellar

# List available clusters
kubestellar> clusters list

# Get pods from all clusters
kubestellar> get pods

# Create a binding policy
kubestellar> bindingpolicy create my-app-policy

# Add clusters to the policy
kubestellar> bindingpolicy add-clusters my-app-policy cluster1 cluster2

# Apply a deployment
kubestellar> apply my-deployment.yaml

# Install a Helm chart
kubestellar> helm install nginx nginx/nginx

# Check the status
kubestellar> get deployments
kubestellar> helm list

# Exit
kubestellar> exit
```

### Multi-Step Operations

```bash
# Start interactive mode
kubestellar

# Check cluster status
kubestellar> clusters list

# Create and configure a binding policy
kubestellar> bindingpolicy create web-app
kubestellar> bindingpolicy add-clusters web-app production-east production-west

# Deploy application with Helm
kubestellar> helm install web-app mycompany/web-app --set replicas=3

# Monitor deployment
kubestellar> get pods -l app=web-app
kubestellar> helm status web-app

# Configure monitoring
kubestellar> config set monitoring.enabled true
kubestellar> config apply monitoring-config.yaml
```

## Tips and Best Practices

### 1. Use Tab Completion

While not implemented in the current version, future versions will support tab completion for:
- Command names
- Resource types
- Cluster names
- Policy names

### 2. Leverage Help System

Always use the help system to discover available options:

```bash
kubestellar> help bindingpolicy
kubestellar> help helm install
```

### 3. Start with Cluster Listing

Begin sessions by checking cluster status:

```bash
kubestellar> clusters list
```

### 4. Use Configuration Management

Set up your preferred defaults:

```bash
kubestellar> config set its-context my-its
kubestellar> config set default-namespace my-namespace
```

### 5. Clear Screen for Fresh Start

Use clear to refresh the interface:

```bash
kubestellar> clear
```

## Troubleshooting

### Common Issues

1. **Connection Errors**
   - Check kubeconfig file path
   - Verify ITS context is accessible
   - Ensure clusters are properly registered

2. **Permission Errors**
   - Verify RBAC permissions in target clusters
   - Check service account configurations

3. **Discovery Issues**
   - Confirm ManagedCluster CRDs exist
   - Verify cluster labels and selectors

### Debug Mode

Enable verbose output by setting environment variables:

```bash
export KUBESTELLAR_DEBUG=true
export KUBESTELLAR_LOG_LEVEL=debug
kubestellar
```

### Log Files

Interactive sessions are logged to:
- `~/.kubestellar/logs/interactive.log`
- `/tmp/kubestellar-session.log`

## Future Enhancements

Planned improvements for the interactive CLI include:

- **Auto-completion**: Tab completion for commands and resources
- **History**: Command history with up/down arrow navigation
- **Real-time Status**: Live cluster and resource status updates
- **Advanced Filtering**: Complex queries and filtering options
- **Scripting Support**: Batch command execution from files
- **Plugin System**: Custom command extensions

## See Also

- [Installation Guide](installation_guide.md) - Setting up KubeStellar CLI
- [BindingPolicy Management](bindingpolicy.md) - Managing workload distribution
- [Cluster Management](cluster_management.md) - Cluster registration and management
- [Helm Integration](helm_integration.md) - Multi-cluster Helm operations