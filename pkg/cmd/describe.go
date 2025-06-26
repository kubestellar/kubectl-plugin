package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"kubectl-multi/pkg/cluster"
	"kubectl-multi/pkg/util"
)

// Custom help function for describe command
func describeHelpFunc(cmd *cobra.Command, args []string) {
	// Get original kubectl help
	kubectlHelp, err := util.GetKubectlHelp("describe")
	if err != nil {
		// Fallback to default help if kubectl help is not available
		cmd.Help()
		return
	}

	// Multi-cluster plugin information
	multiClusterInfo := `Show details of a specific resource or group of resources across all managed clusters.
This command displays detailed information about resources similar to kubectl describe,
but across all KubeStellar managed clusters.`

	// Multi-cluster examples
	multiClusterExamples := `# Describe a specific pod across all clusters
kubectl multi describe pod nginx

# Describe all pods with a specific label across all clusters
kubectl multi describe pods -l app=nginx

# Describe a service across all clusters
kubectl multi describe service/my-service

# Describe nodes across all clusters
kubectl multi describe nodes`

	// Multi-cluster usage
	multiClusterUsage := `kubectl multi describe [TYPE[.VERSION][.GROUP] [NAME_PREFIX | -l label] | TYPE[.VERSION][.GROUP]/NAME] [flags]`

	// Format combined help
	combinedHelp := util.FormatMultiClusterHelp(kubectlHelp, multiClusterInfo, multiClusterExamples, multiClusterUsage)
	fmt.Fprintln(cmd.OutOrStdout(), combinedHelp)
}

func newDescribeCommand() *cobra.Command {
	var selector string
	var showEvents bool
	var chunkSize int

	cmd := &cobra.Command{
		Use:   "describe [TYPE[.VERSION][.GROUP] [NAME_PREFIX | -l label] | TYPE[.VERSION][.GROUP]/NAME]",
		Short: "Show details of a specific resource or group of resources across managed clusters",
		Long: `Show details of a specific resource or group of resources across all managed clusters.
This command displays detailed information about resources similar to kubectl describe,
but across all KubeStellar managed clusters.`,
		Example: `# Describe a specific pod across all clusters
kubectl multi describe pod nginx

# Describe all pods with a specific label across all clusters
kubectl multi describe pods -l app=nginx

# Describe a service across all clusters
kubectl multi describe service/my-service

# Describe nodes across all clusters
kubectl multi describe nodes`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("resource type must be specified")
			}

			kubeconfig, remoteCtx, _, namespace, allNamespaces := GetGlobalFlags()
			return handleDescribeCommand(args, selector, showEvents, chunkSize, kubeconfig, remoteCtx, namespace, allNamespaces)
		},
	}

	// Add describe-specific flags
	cmd.Flags().StringVarP(&selector, "selector", "l", "", "selector (label query) to filter on, supports '=', '==', '!=', 'in', 'notin'")
	cmd.Flags().BoolVar(&showEvents, "show-events", true, "if true, display events related to the described object")
	cmd.Flags().IntVar(&chunkSize, "chunk-size", 500, "return large lists in chunks rather than all at once")

	// Set custom help function
	cmd.SetHelpFunc(describeHelpFunc)

	return cmd
}

func handleDescribeCommand(args []string, selector string, showEvents bool, chunkSize int, kubeconfig, remoteCtx, namespace string, allNamespaces bool) error {
	clusters, err := cluster.DiscoverClusters(kubeconfig, remoteCtx)
	if err != nil {
		return fmt.Errorf("failed to discover clusters: %v", err)
	}

	if len(clusters) == 0 {
		return fmt.Errorf("no clusters discovered")
	}

	// Parse resource type and name from args
	resourceType := args[0]
	// Note: resourceName is not currently used but kept for future enhancement
	// resourceName := ""
	// if len(args) > 1 {
	// 	resourceName = args[1]
	// }

	fmt.Printf("Describing %s across %d clusters...\n\n", resourceType, len(clusters))

	// Track if any cluster had successful output
	anyOutput := false

	for _, clusterInfo := range clusters {
		if clusterInfo.Client == nil {
			fmt.Printf("Warning: skipping cluster %s (no client available)\n", clusterInfo.Name)
			continue
		}

		fmt.Printf("=== Cluster: %s (Context: %s) ===\n", clusterInfo.Name, clusterInfo.Context)

		// Build kubectl describe command
		kubectlArgs := buildDescribeArgs(args, selector, showEvents, chunkSize, namespace, allNamespaces, clusterInfo.Name)

		// Execute kubectl describe for this cluster
		output, err := executeKubectlDescribe(kubectlArgs, kubeconfig, clusterInfo.Name)
		if err != nil {
			fmt.Printf("Error describing %s in cluster %s: %v\n", resourceType, clusterInfo.Name, err)
			fmt.Printf("\n")
			continue
		}

		// If we got output, display it
		if strings.TrimSpace(output) != "" {
			fmt.Print(output)
			anyOutput = true
		} else {
			fmt.Printf("No %s found in cluster %s\n", resourceType, clusterInfo.Name)
		}

		fmt.Printf("\n")
	}

	if !anyOutput {
		fmt.Printf("No %s found in any cluster\n", resourceType)
	}

	return nil
}

// buildDescribeArgs constructs the kubectl describe command arguments
func buildDescribeArgs(args []string, selector string, showEvents bool, chunkSize int, namespace string, allNamespaces bool, clusterContext string) []string {
	var kubectlArgs []string

	// Add the describe command and resource type
	kubectlArgs = append(kubectlArgs, "describe")
	kubectlArgs = append(kubectlArgs, args...)

	// Add selector if specified
	if selector != "" {
		kubectlArgs = append(kubectlArgs, "-l", selector)
	}

	// Add namespace flags
	if allNamespaces {
		kubectlArgs = append(kubectlArgs, "-A")
	} else if namespace != "" {
		kubectlArgs = append(kubectlArgs, "-n", namespace)
	}

	// Add show-events flag
	if !showEvents {
		kubectlArgs = append(kubectlArgs, "--show-events=false")
	}

	// Add chunk-size flag
	if chunkSize != 500 {
		kubectlArgs = append(kubectlArgs, "--chunk-size", fmt.Sprintf("%d", chunkSize))
	}

	// Add context for this specific cluster
	kubectlArgs = append(kubectlArgs, "--context", clusterContext)

	return kubectlArgs
}

// executeKubectlDescribe executes kubectl describe command for a specific cluster
func executeKubectlDescribe(args []string, kubeconfig, clusterName string) (string, error) {
	// Create the command
	cmd := exec.Command("kubectl", args...)

	// Set environment variables
	cmd.Env = os.Environ()
	if kubeconfig != "" {
		cmd.Env = append(cmd.Env, "KUBECONFIG="+kubeconfig)
	}

	// Capture stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute the command
	err := cmd.Run()

	// Get the output
	output := stdout.String()
	stderrOutput := stderr.String()

	// Handle different types of errors
	if err != nil {
		// Check if it's a "not found" error (which is expected for some resources)
		if strings.Contains(stderrOutput, "not found") || strings.Contains(stderrOutput, "No resources found") {
			return "", nil // Return empty string for not found, not an error
		}

		// For other errors, return the error with context
		return "", fmt.Errorf("kubectl command failed: %v\nStderr: %s", err, stderrOutput)
	}

	// If we got stderr output but no error, it might be warnings
	if stderrOutput != "" && !strings.Contains(stderrOutput, "not found") {
		output = stderrOutput + "\n" + output
	}

	return output, nil
}
