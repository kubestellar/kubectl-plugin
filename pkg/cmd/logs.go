package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"sync"

	"github.com/spf13/cobra"
	"kubectl-multi/pkg/cluster"
	"kubectl-multi/pkg/util"
)

// Custom help function for logs command
func logsHelpFunc(cmd *cobra.Command, args []string) {
	// Get original kubectl help using the new implementation
	cmdInfo, err := util.GetKubectlCommandInfo("logs")
	if err != nil {
		// Fallback to default help if kubectl help is not available
		cmd.Help()
		return
	}

	// Multi-cluster plugin information
	multiClusterInfo := `Print logs from containers across all managed clusters.
The output includes cluster context information to help identify which
cluster each log entry belongs to.`

	// Multi-cluster examples
	multiClusterExamples := `# Return logs from nginx pod across all clusters
kubectl multi logs nginx

# Return logs from nginx container in nginx-app pod across all clusters
kubectl multi logs nginx-app -c nginx

# Return logs from previous terminated container across all clusters
kubectl multi logs -p nginx

# Return logs from all containers in pods defined by label app=nginx
kubectl multi logs -l app=nginx --all-containers=true

# Follow logs from nginx pod across all clusters
kubectl multi logs -f nginx

# Return logs from last 10 lines across all clusters
kubectl multi logs --tail=10 nginx

# Return logs from pods with label app=nginx across all clusters
kubectl multi logs -l app=nginx`

	// Multi-cluster usage
	multiClusterUsage := `kubectl multi logs [-f] [-p] (POD | TYPE/NAME) [-c CONTAINER] [flags]`

	// Format combined help using the new CommandInfo structure
	combinedHelp := util.FormatMultiClusterHelp(cmdInfo, multiClusterInfo, multiClusterExamples, multiClusterUsage)
	fmt.Fprintln(cmd.OutOrStdout(), combinedHelp)
}

func newLogsCommand() *cobra.Command {
	var follow bool
	var previous bool
	var container string
	var allContainers bool
	var tail int64
	var sinceTime string
	var sinceSeconds int64
	var timestamps bool
	var selector string

	cmd := &cobra.Command{
		Use:   "logs [-f] [-p] (POD | TYPE/NAME) [-c CONTAINER]",
		Short: "Print logs from containers across all managed clusters",
		Long: `Print logs from containers across all managed clusters.
The output includes cluster context information to help identify which
cluster each log entry belongs to.`,
		Example: `# Return logs from nginx pod across all clusters
kubectl multi logs nginx

# Return logs from nginx container in nginx-app pod across all clusters
kubectl multi logs nginx-app -c nginx

# Return logs from previous terminated container across all clusters
kubectl multi logs -p nginx

# Return logs from all containers in pods defined by label app=nginx
kubectl multi logs -l app=nginx --all-containers=true

# Follow logs from nginx pod across all clusters
kubectl multi logs -f nginx`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 && selector == "" {
				return fmt.Errorf("POD or TYPE/NAME is required")
			}
			kubeconfig, remoteCtx, _, namespace, allNamespaces := GetGlobalFlags()
			return handleLogsCommand(args, follow, previous, container, allContainers, tail, sinceTime, sinceSeconds, timestamps, selector, kubeconfig, remoteCtx, namespace, allNamespaces)
		},
	}

	cmd.Flags().BoolVarP(&follow, "follow", "f", false, "Specify if the logs should be streamed")
	cmd.Flags().BoolVarP(&previous, "previous", "p", false, "If true, print the logs for the previous instance of the container")
	cmd.Flags().StringVarP(&container, "container", "c", "", "Print the logs of this container")
	cmd.Flags().BoolVar(&allContainers, "all-containers", false, "Get all containers' logs in the pod(s)")
	cmd.Flags().Int64Var(&tail, "tail", -1, "Lines of recent log file to display")
	cmd.Flags().StringVar(&sinceTime, "since-time", "", "Only return logs after a specific date (RFC3339)")
	cmd.Flags().Int64Var(&sinceSeconds, "since", 0, "Only return logs newer than a relative duration like 5s, 2m, or 3h")
	cmd.Flags().BoolVar(&timestamps, "timestamps", false, "Include timestamps on each line in the log output")
	cmd.Flags().StringVarP(&selector, "selector", "l", "", "Selector (label query) to filter on")

	// Set custom help function
	cmd.SetHelpFunc(logsHelpFunc)

	return cmd
}

func handleLogsCommand(args []string, follow, previous bool, container string, allContainers bool, tail int64, sinceTime string, sinceSeconds int64, timestamps bool, selector, kubeconfig, remoteCtx, namespace string, allNamespaces bool) error {
	clusters, err := cluster.DiscoverClusters(kubeconfig, remoteCtx)
	if err != nil {
		return fmt.Errorf("failed to discover clusters: %v", err)
	}

	if len(clusters) == 0 {
		return fmt.Errorf("no clusters discovered")
	}

	// For non-follow mode, execute sequentially
	if !follow {
		for _, clusterInfo := range clusters {
			if clusterInfo.Client == nil {
				continue
			}

			fmt.Printf("=== Cluster: %s ===\n", clusterInfo.Name)

			cmdArgs := []string{"logs", "--context", clusterInfo.Context}

			// Add pod/resource name if provided
			if len(args) > 0 {
				cmdArgs = append(cmdArgs, args[0])
			}

			// Add flags
			if previous {
				cmdArgs = append(cmdArgs, "-p")
			}
			if container != "" {
				cmdArgs = append(cmdArgs, "-c", container)
			}
			if allContainers {
				cmdArgs = append(cmdArgs, "--all-containers=true")
			}
			if tail >= 0 {
				cmdArgs = append(cmdArgs, fmt.Sprintf("--tail=%d", tail))
			}
			if sinceTime != "" {
				cmdArgs = append(cmdArgs, "--since-time="+sinceTime)
			}
			if sinceSeconds > 0 {
				cmdArgs = append(cmdArgs, fmt.Sprintf("--since=%ds", sinceSeconds))
			}
			if timestamps {
				cmdArgs = append(cmdArgs, "--timestamps=true")
			}
			if selector != "" {
				cmdArgs = append(cmdArgs, "-l", selector)
			}
			if namespace != "" {
				cmdArgs = append(cmdArgs, "-n", namespace)
			}
			if allNamespaces {
				cmdArgs = append(cmdArgs, "-A")
			}

			output, err := runKubectlForLogs(cmdArgs, kubeconfig)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Print(output)
			}
			fmt.Println()
		}
		return nil
	}

	// For follow mode, execute concurrently with prefixed output
	var wg sync.WaitGroup
	for _, clusterInfo := range clusters {
		if clusterInfo.Client == nil {
			continue
		}

		wg.Add(1)
		go func(c cluster.ClusterInfo) {
			defer wg.Done()

			cmdArgs := []string{"logs", "--context", c.Context, "-f"}

			// Add pod/resource name if provided
			if len(args) > 0 {
				cmdArgs = append(cmdArgs, args[0])
			}

			// Add other flags
			if container != "" {
				cmdArgs = append(cmdArgs, "-c", container)
			}
			if allContainers {
				cmdArgs = append(cmdArgs, "--all-containers=true")
			}
			if tail >= 0 {
				cmdArgs = append(cmdArgs, fmt.Sprintf("--tail=%d", tail))
			}
			if sinceTime != "" {
				cmdArgs = append(cmdArgs, "--since-time="+sinceTime)
			}
			if sinceSeconds > 0 {
				cmdArgs = append(cmdArgs, fmt.Sprintf("--since=%ds", sinceSeconds))
			}
			if timestamps {
				cmdArgs = append(cmdArgs, "--timestamps=true")
			}
			if selector != "" {
				cmdArgs = append(cmdArgs, "-l", selector)
			}
			if namespace != "" {
				cmdArgs = append(cmdArgs, "-n", namespace)
			}
			if allNamespaces {
				cmdArgs = append(cmdArgs, "-A")
			}

			runKubectlLogsWithPrefix(cmdArgs, kubeconfig, c.Name)
		}(clusterInfo)
	}

	wg.Wait()
	return nil
}

// runKubectlForLogs runs kubectl logs command and returns output
func runKubectlForLogs(args []string, kubeconfig string) (string, error) {
	cmd := exec.Command("kubectl", args...)
	if kubeconfig != "" {
		cmd.Env = append(os.Environ(), "KUBECONFIG="+kubeconfig)
	}

	output, err := cmd.CombinedOutput()
	return string(output), err
}

// runKubectlLogsWithPrefix runs kubectl logs with live output prefixed by cluster name
func runKubectlLogsWithPrefix(args []string, kubeconfig, clusterName string) {
	cmd := exec.Command("kubectl", args...)
	if kubeconfig != "" {
		cmd.Env = append(os.Environ(), "KUBECONFIG="+kubeconfig)
	}

	// Create a pipe for stdout
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Printf("[%s] Error creating stdout pipe: %v\n", clusterName, err)
		return
	}

	// Create a pipe for stderr
	stderr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Printf("[%s] Error creating stderr pipe: %v\n", clusterName, err)
		return
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		fmt.Printf("[%s] Error starting command: %v\n", clusterName, err)
		return
	}

	// Read stdout with prefix
	go util.PrefixedReader(stdout, clusterName, os.Stdout)
	go util.PrefixedReader(stderr, clusterName, os.Stderr)

	// Wait for command to finish
	if err := cmd.Wait(); err != nil {
		fmt.Printf("[%s] Command finished with error: %v\n", clusterName, err)
	}
}