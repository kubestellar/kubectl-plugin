package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"kubectl-multi/pkg/cluster"
	"kubectl-multi/pkg/util"
)

// Custom help function for create command
func createHelpFunc(cmd *cobra.Command, args []string) {
	// Get original kubectl help using the new implementation
	cmdInfo, err := util.GetKubectlCommandInfo("create")
	if err != nil {
		// Fallback to default help if kubectl help is not available
		cmd.Help()
		return
	}

	// Multi-cluster plugin information
	multiClusterInfo := `Create resources across all managed clusters.

IMPORTANT: For better multi-cluster management, consider using label-based binding
policies instead of direct create commands. Binding policies provide:
- Better control over resource placement
- Declarative cluster selection using labels
- Automatic reconciliation and updates
- Consistent multi-cluster deployment patterns

See KubeStellar documentation for more information on binding policies.`

	// Multi-cluster examples with binding policy guidance
	multiClusterExamples := `# Create a deployment across all managed clusters (not recommended)
kubectl multi create deployment nginx --image=nginx

# Better approach: Use binding policies
# 1. First, create resources with labels in the control cluster:
kubectl label deployment nginx app=nginx env=prod

# 2. Then create a binding policy to propagate based on labels:
kubectl create -f - <<EOF
apiVersion: control.kubestellar.io/v1alpha1
kind: BindingPolicy
metadata:
  name: nginx-policy
spec:
  clusterSelectors:
  - matchLabels:
      location: edge
  downsync:
  - objectSelectors:
    - matchLabels:
        app: nginx
        env: prod
EOF

# Other create examples (use with caution in multi-cluster):
# Create a service across all clusters
kubectl multi create service clusterip my-svc --tcp=5678:8080

# Create a configmap across all clusters
kubectl multi create configmap my-config --from-literal=key1=value1

# Create a secret across all clusters
kubectl multi create secret generic my-secret --from-literal=password=secretpass`

	// Multi-cluster usage
	multiClusterUsage := `kubectl multi create -f FILENAME [flags]`

	// Format combined help using the new CommandInfo structure
	combinedHelp := util.FormatMultiClusterHelp(cmdInfo, multiClusterInfo, multiClusterExamples, multiClusterUsage)
	fmt.Fprintln(cmd.OutOrStdout(), combinedHelp)
}

func newCreateCommand() *cobra.Command {
	var filename string
	var recursive bool
	var dryRun string
	var output string

	cmd := &cobra.Command{
		Use:   "create -f FILENAME",
		Short: "Create resources across all managed clusters",
		Long: `Create resources across all managed clusters.

IMPORTANT: For better multi-cluster management, consider using label-based binding
policies instead of direct create commands. Binding policies provide:
- Better control over resource placement
- Declarative cluster selection using labels
- Automatic reconciliation and updates
- Consistent multi-cluster deployment patterns

See KubeStellar documentation for more information on binding policies.`,
		Example: `# Create a deployment across all managed clusters
kubectl multi create deployment nginx --image=nginx

# Create from a file across all clusters
kubectl multi create -f deployment.yaml

# Create a service across all clusters
kubectl multi create service clusterip my-svc --tcp=5678:8080`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Show binding policy recommendation for non-file creates
			if filename == "" && len(args) > 0 {
				fmt.Println("⚠️  WARNING: Direct resource creation across multiple clusters is not recommended.")
				fmt.Println("   Consider using KubeStellar binding policies for better multi-cluster management.")
				fmt.Println("   See: https://docs.kubestellar.io/direct/examples/binding-policy/")
				fmt.Println()
			}

			kubeconfig, remoteCtx, _, namespace, allNamespaces := GetGlobalFlags()
			return handleCreateCommand(cmd, args, filename, recursive, dryRun, output, kubeconfig, remoteCtx, namespace, allNamespaces)
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&filename, "filename", "f", "", "Filename, directory, or URL to files to use to create the resource")
	cmd.Flags().BoolVarP(&recursive, "recursive", "R", false, "Process the directory used in -f, --filename recursively")
	cmd.Flags().StringVar(&dryRun, "dry-run", "none", "Must be \"none\", \"server\", or \"client\"")
	cmd.Flags().StringVarP(&output, "output", "o", "", "Output format")

	// Set custom help function
	cmd.SetHelpFunc(createHelpFunc)

	// Add subcommands for specific resource types
	cmd.AddCommand(newCreateDeploymentCommand())
	cmd.AddCommand(newCreateServiceCommand())
	cmd.AddCommand(newCreateConfigMapCommand())
	cmd.AddCommand(newCreateSecretCommand())
	cmd.AddCommand(newCreateNamespaceCommand())

	return cmd
}

func handleCreateCommand(cmd *cobra.Command, args []string, filename string, recursive bool, dryRun, output, kubeconfig, remoteCtx, namespace string, allNamespaces bool) error {
	clusters, err := cluster.DiscoverClusters(kubeconfig, remoteCtx)
	if err != nil {
		return fmt.Errorf("failed to discover clusters: %v", err)
	}

	if len(clusters) == 0 {
		return fmt.Errorf("no clusters discovered")
	}

	// Build kubectl args
	cmdArgs := []string{"create"}

	// Add filename if provided
	if filename != "" {
		cmdArgs = append(cmdArgs, "-f", filename)
		if recursive {
			cmdArgs = append(cmdArgs, "-R")
		}
	} else {
		// Add resource type and args
		cmdArgs = append(cmdArgs, args...)
	}

	// Add common flags
	if dryRun != "none" && dryRun != "" {
		cmdArgs = append(cmdArgs, "--dry-run="+dryRun)
	}
	if output != "" {
		cmdArgs = append(cmdArgs, "-o", output)
	}
	if namespace != "" {
		cmdArgs = append(cmdArgs, "-n", namespace)
	}

	// Execute on all clusters
	for _, clusterInfo := range clusters {
		if clusterInfo.Client == nil {
			continue
		}

		fmt.Printf("=== Cluster: %s ===\n", clusterInfo.Name)

		clusterArgs := append([]string{}, cmdArgs...)
		clusterArgs = append(clusterArgs, "--context", clusterInfo.Context)

		output, err := runKubectl(clusterArgs, kubeconfig)
		if err != nil {
			// Check for common error patterns and provide friendly messages
			if strings.Contains(output, "already exists") {
				fmt.Printf("❌ Resource already exists in this cluster\n")
				fmt.Printf("   Output: %s", output)
			} else if strings.Contains(output, "not found") {
				fmt.Printf("❌ Resource or cluster not accessible\n")
				fmt.Printf("   Output: %s", output)
			} else {
				fmt.Printf("❌ Error: %v\n", err)
				if output != "" {
					fmt.Printf("   Output: %s", output)
				}
			}
		} else {
			fmt.Print(output)
		}
		fmt.Println()
	}

	return nil
}

// Subcommand for deployment
func newCreateDeploymentCommand() *cobra.Command {
	var image string
	var replicas int32
	var port int32

	cmd := &cobra.Command{
		Use:   "deployment NAME --image=image [--dry-run=server|client] [flags]",
		Short: "Create a deployment across all clusters",
		Long: `Create a deployment across all clusters.

⚠️  For multi-cluster deployments, consider using binding policies instead.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("NAME is required")
			}
			if image == "" {
				return fmt.Errorf("--image is required")
			}

			fmt.Println("⚠️  WARNING: Direct deployment creation across multiple clusters is not recommended.")
			fmt.Println("   Consider using KubeStellar binding policies for better multi-cluster management.")
			fmt.Println()

			return handleCreateDeploymentCommand(args[0], image, replicas, port)
		},
	}

	cmd.Flags().StringVar(&image, "image", "", "Image name to run")
	cmd.Flags().Int32Var(&replicas, "replicas", 1, "Number of replicas")
	cmd.Flags().Int32Var(&port, "port", 0, "Port to expose")

	return cmd
}

func handleCreateDeploymentCommand(name, image string, replicas, port int32) error {
	kubeconfig, remoteCtx, _, namespace, _ := GetGlobalFlags()

	clusters, err := cluster.DiscoverClusters(kubeconfig, remoteCtx)
	if err != nil {
		return fmt.Errorf("failed to discover clusters: %v", err)
	}

	cmdArgs := []string{"create", "deployment", name, "--image=" + image}
	if replicas > 0 {
		cmdArgs = append(cmdArgs, fmt.Sprintf("--replicas=%d", replicas))
	}
	if port > 0 {
		cmdArgs = append(cmdArgs, fmt.Sprintf("--port=%d", port))
	}
	if namespace != "" {
		cmdArgs = append(cmdArgs, "-n", namespace)
	}
	
	for _, clusterInfo := range clusters {
		fmt.Printf("=== Cluster: %s ===\n", clusterInfo.Name)

		clusterArgs := append([]string{}, cmdArgs...)
		clusterArgs = append(clusterArgs, "--context", clusterInfo.Context)

		output, err := runKubectl(clusterArgs, kubeconfig)
		if err != nil {
			// Check for common error patterns and provide friendly messages
			if strings.Contains(output, "already exists") {
				fmt.Printf("❌ Resource already exists in this cluster\n")
				fmt.Printf("   Output: %s", output)
			} else if strings.Contains(output, "not found") {
				fmt.Printf("❌ Resource or cluster not accessible\n")
				fmt.Printf("   Output: %s", output)
			} else {
				fmt.Printf("❌ Error: %v\n", err)
				if output != "" {
					fmt.Printf("   Output: %s", output)
				}
			}
		} else {
			fmt.Print(output)
		}
		fmt.Println()
	}

	return nil
}

// Subcommand for service
func newCreateServiceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "service",
		Short: "Create a service across all clusters",
		Long: `Create a service across all clusters.

⚠️  For multi-cluster services, consider using binding policies instead.`,
	}

	// Add service subtypes
	cmd.AddCommand(newCreateServiceClusterIPCommand())
	cmd.AddCommand(newCreateServiceNodePortCommand())
	cmd.AddCommand(newCreateServiceLoadBalancerCommand())

	return cmd
}

func newCreateServiceClusterIPCommand() *cobra.Command {
	var tcp []string
	var clusterIP string

	cmd := &cobra.Command{
		Use:   "clusterip NAME [--tcp=<port>:<targetPort>] [flags]",
		Short: "Create a ClusterIP service across all clusters",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("NAME is required")
			}

			kubeconfig, remoteCtx, _, namespace, _ := GetGlobalFlags()
			return handleCreateServiceCommand(args[0], "ClusterIP", tcp, clusterIP, kubeconfig, remoteCtx, namespace)
		},
	}

	cmd.Flags().StringArrayVar(&tcp, "tcp", []string{}, "Port pairs in the format port:targetPort")
	cmd.Flags().StringVar(&clusterIP, "clusterip", "", "Assign your own ClusterIP")

	return cmd
}

func newCreateServiceNodePortCommand() *cobra.Command {
	var tcp []string
	var nodePort int32

	cmd := &cobra.Command{
		Use:   "nodeport NAME [--tcp=<port>:<targetPort>] [flags]",
		Short: "Create a NodePort service across all clusters",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("NAME is required")
			}

			kubeconfig, remoteCtx, _, namespace, _ := GetGlobalFlags()
			return handleCreateServiceCommand(args[0], "NodePort", tcp, "", kubeconfig, remoteCtx, namespace)
		},
	}

	cmd.Flags().StringArrayVar(&tcp, "tcp", []string{}, "Port pairs in the format port:targetPort")
	cmd.Flags().Int32Var(&nodePort, "node-port", 0, "Port used to expose the service on each node")

	return cmd
}

func newCreateServiceLoadBalancerCommand() *cobra.Command {
	var tcp []string

	cmd := &cobra.Command{
		Use:   "loadbalancer NAME [--tcp=<port>:<targetPort>] [flags]",
		Short: "Create a LoadBalancer service across all clusters",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("NAME is required")
			}

			kubeconfig, remoteCtx, _, namespace, _ := GetGlobalFlags()
			return handleCreateServiceCommand(args[0], "LoadBalancer", tcp, "", kubeconfig, remoteCtx, namespace)
		},
	}

	cmd.Flags().StringArrayVar(&tcp, "tcp", []string{}, "Port pairs in the format port:targetPort")

	return cmd
}

func handleCreateServiceCommand(name, serviceType string, tcp []string, clusterIP, kubeconfig, remoteCtx, namespace string) error {
	clusters, err := cluster.DiscoverClusters(kubeconfig, remoteCtx)
	if err != nil {
		return fmt.Errorf("failed to discover clusters: %v", err)
	}

	cmdArgs := []string{"create", "service", strings.ToLower(serviceType), name}
	for _, t := range tcp {
		cmdArgs = append(cmdArgs, "--tcp="+t)
	}
	if clusterIP != "" {
		cmdArgs = append(cmdArgs, "--clusterip="+clusterIP)
	}
	if namespace != "" {
		cmdArgs = append(cmdArgs, "-n", namespace)
	}

	for _, clusterInfo := range clusters {
		fmt.Printf("=== Cluster: %s ===\n", clusterInfo.Name)

		clusterArgs := append([]string{}, cmdArgs...)
		clusterArgs = append(clusterArgs, "--context", clusterInfo.Context)

		output, err := runKubectl(clusterArgs, kubeconfig)
		if err != nil {
			// Check for common error patterns and provide friendly messages
			if strings.Contains(output, "already exists") {
				fmt.Printf("❌ Resource already exists in this cluster\n")
				fmt.Printf("   Output: %s", output)
			} else if strings.Contains(output, "not found") {
				fmt.Printf("❌ Resource or cluster not accessible\n")
				fmt.Printf("   Output: %s", output)
			} else {
				fmt.Printf("❌ Error: %v\n", err)
				if output != "" {
					fmt.Printf("   Output: %s", output)
				}
			}
		} else {
			fmt.Print(output)
		}
		fmt.Println()
	}

	return nil
}

// Subcommand for configmap
func newCreateConfigMapCommand() *cobra.Command {
	var fromLiteral []string
	var fromFile []string
	var fromEnvFile string

	cmd := &cobra.Command{
		Use:   "configmap NAME [--from-literal=key1=value1] [--from-file=[key=]source] [flags]",
		Short: "Create a configmap across all clusters",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("NAME is required")
			}

			kubeconfig, remoteCtx, _, namespace, _ := GetGlobalFlags()
			return handleCreateConfigMapCommand(args[0], fromLiteral, fromFile, fromEnvFile, kubeconfig, remoteCtx, namespace)
		},
	}

	cmd.Flags().StringArrayVar(&fromLiteral, "from-literal", []string{}, "Specify a key and literal value")
	cmd.Flags().StringArrayVar(&fromFile, "from-file", []string{}, "Key file can be specified using its file path")
	cmd.Flags().StringVar(&fromEnvFile, "from-env-file", "", "Specify the path to a file to read lines of key=val pairs")

	return cmd
}

func handleCreateConfigMapCommand(name string, fromLiteral, fromFile []string, fromEnvFile, kubeconfig, remoteCtx, namespace string) error {
	clusters, err := cluster.DiscoverClusters(kubeconfig, remoteCtx)
	if err != nil {
		return fmt.Errorf("failed to discover clusters: %v", err)
	}

	cmdArgs := []string{"create", "configmap", name}
	for _, literal := range fromLiteral {
		cmdArgs = append(cmdArgs, "--from-literal="+literal)
	}
	for _, file := range fromFile {
		cmdArgs = append(cmdArgs, "--from-file="+file)
	}
	if fromEnvFile != "" {
		cmdArgs = append(cmdArgs, "--from-env-file="+fromEnvFile)
	}
	if namespace != "" {
		cmdArgs = append(cmdArgs, "-n", namespace)
	}

	for _, clusterInfo := range clusters {
		fmt.Printf("=== Cluster: %s ===\n", clusterInfo.Name)

		clusterArgs := append([]string{}, cmdArgs...)
		clusterArgs = append(clusterArgs, "--context", clusterInfo.Context)

		output, err := runKubectl(clusterArgs, kubeconfig)
		if err != nil {
			// Check for common error patterns and provide friendly messages
			if strings.Contains(output, "already exists") {
				fmt.Printf("❌ Resource already exists in this cluster\n")
				fmt.Printf("   Output: %s", output)
			} else if strings.Contains(output, "not found") {
				fmt.Printf("❌ Resource or cluster not accessible\n")
				fmt.Printf("   Output: %s", output)
			} else {
				fmt.Printf("❌ Error: %v\n", err)
				if output != "" {
					fmt.Printf("   Output: %s", output)
				}
			}
		} else {
			fmt.Print(output)
		}
		fmt.Println()
	}

	return nil
}

// Subcommand for secret
func newCreateSecretCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "secret",
		Short: "Create a secret across all clusters",
	}

	cmd.AddCommand(newCreateSecretGenericCommand())

	return cmd
}

func newCreateSecretGenericCommand() *cobra.Command {
	var fromLiteral []string
	var fromFile []string

	cmd := &cobra.Command{
		Use:   "generic NAME [--from-literal=key1=value1] [--from-file=[key=]source] [flags]",
		Short: "Create a generic secret across all clusters",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("NAME is required")
			}

			kubeconfig, remoteCtx, _, namespace, _ := GetGlobalFlags()
			return handleCreateSecretCommand(args[0], fromLiteral, fromFile, kubeconfig, remoteCtx, namespace)
		},
	}

	cmd.Flags().StringArrayVar(&fromLiteral, "from-literal", []string{}, "Specify a key and literal value")
	cmd.Flags().StringArrayVar(&fromFile, "from-file", []string{}, "Key file can be specified using its file path")

	return cmd
}

func handleCreateSecretCommand(name string, fromLiteral, fromFile []string, kubeconfig, remoteCtx, namespace string) error {
	clusters, err := cluster.DiscoverClusters(kubeconfig, remoteCtx)
	if err != nil {
		return fmt.Errorf("failed to discover clusters: %v", err)
	}

	cmdArgs := []string{"create", "secret", "generic", name}
	for _, literal := range fromLiteral {
		cmdArgs = append(cmdArgs, "--from-literal="+literal)
	}
	for _, file := range fromFile {
		cmdArgs = append(cmdArgs, "--from-file="+file)
	}
	if namespace != "" {
		cmdArgs = append(cmdArgs, "-n", namespace)
	}

	for _, clusterInfo := range clusters {
		fmt.Printf("=== Cluster: %s ===\n", clusterInfo.Name)

		clusterArgs := append([]string{}, cmdArgs...)
		clusterArgs = append(clusterArgs, "--context", clusterInfo.Context)

		output, err := runKubectl(clusterArgs, kubeconfig)
		if err != nil {
			// Check for common error patterns and provide friendly messages
			if strings.Contains(output, "already exists") {
				fmt.Printf("❌ Resource already exists in this cluster\n")
				fmt.Printf("   Output: %s", output)
			} else if strings.Contains(output, "not found") {
				fmt.Printf("❌ Resource or cluster not accessible\n")
				fmt.Printf("   Output: %s", output)
			} else {
				fmt.Printf("❌ Error: %v\n", err)
				if output != "" {
					fmt.Printf("   Output: %s", output)
				}
			}
		} else {
			fmt.Print(output)
		}
		fmt.Println()
	}

	return nil
}

// Subcommand for namespace
func newCreateNamespaceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "namespace NAME [flags]",
		Short: "Create a namespace across all clusters",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("NAME is required")
			}

			kubeconfig, remoteCtx, _, _, _ := GetGlobalFlags()
			return handleCreateNamespaceCommand(args[0], kubeconfig, remoteCtx)
		},
	}

	return cmd
}

func handleCreateNamespaceCommand(name, kubeconfig, remoteCtx string) error {
	clusters, err := cluster.DiscoverClusters(kubeconfig, remoteCtx)
	if err != nil {
		return fmt.Errorf("failed to discover clusters: %v", err)
	}

	cmdArgs := []string{"create", "namespace", name}

	for _, clusterInfo := range clusters {
		fmt.Printf("=== Cluster: %s ===\n", clusterInfo.Name)

		clusterArgs := append([]string{}, cmdArgs...)
		clusterArgs = append(clusterArgs, "--context", clusterInfo.Context)

		output, err := runKubectl(clusterArgs, kubeconfig)
		if err != nil {
			// Check for common error patterns and provide friendly messages
			if strings.Contains(output, "already exists") {
				fmt.Printf("❌ Resource already exists in this cluster\n")
				fmt.Printf("   Output: %s", output)
			} else if strings.Contains(output, "not found") {
				fmt.Printf("❌ Resource or cluster not accessible\n")
				fmt.Printf("   Output: %s", output)
			} else {
				fmt.Printf("❌ Error: %v\n", err)
				if output != "" {
					fmt.Printf("   Output: %s", output)
				}
			}
		} else {
			fmt.Print(output)
		}
		fmt.Println()
	}

	return nil
}