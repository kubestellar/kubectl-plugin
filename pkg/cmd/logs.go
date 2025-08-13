package cmd

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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
	multiClusterInfo := `Print the logs for a container in a pod across all managed clusters.
This command retrieves and displays logs from pods across all KubeStellar managed clusters,
making it easy to troubleshoot applications running in multiple clusters.`

	// Multi-cluster examples
	multiClusterExamples := `# Print logs from a pod across all clusters
kubectl multi logs nginx-pod

# Print logs from pods matching a pattern across all clusters
kubectl multi logs transport-controller*

# Print logs from a specific container in matching pods across all clusters
kubectl multi logs nginx-pod* -c nginx-container

# Follow logs from matching pods across all clusters
kubectl multi logs app-* -f

# Print logs with timestamps from matching pods across all clusters
kubectl multi logs nginx-* --timestamps

# Print last 50 lines of logs from matching pods across all clusters
kubectl multi logs transport-* --tail=50`

	// Multi-cluster usage
	multiClusterUsage := `kubectl multi logs [-f] [-p] POD [-c CONTAINER] [flags]`

	// Format combined help using the new CommandInfo structure
	combinedHelp := util.FormatMultiClusterHelp(cmdInfo, multiClusterInfo, multiClusterExamples, multiClusterUsage)
	fmt.Fprintln(cmd.OutOrStdout(), combinedHelp)
}

func newLogsCommand() *cobra.Command {
	var follow bool
	var previous bool
	var container string
	var since string
	var sinceTime string
	var timestamps bool
	var tail int64
	var limitBytes int64

	cmd := &cobra.Command{
		Use:   "logs [-f] [-p] POD [-c CONTAINER]",
		Short: "Print the logs for a container in a pod across managed clusters",
		Long: `Print the logs for a container in a pod across all managed clusters.
This command retrieves and displays logs from pods across all KubeStellar managed clusters,
making it easy to troubleshoot applications running in multiple clusters.`,
		Example: `# Print logs from a pod across all clusters
kubectl multi logs nginx-pod

# Print logs from pods matching a pattern across all clusters
kubectl multi logs transport-controller*

# Print logs from a specific container in matching pods across all clusters
kubectl multi logs nginx-pod* -c nginx-container

# Follow logs from matching pods across all clusters
kubectl multi logs app-* -f

# Print logs with timestamps across all clusters
kubectl multi logs nginx-pod --timestamps`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("pod name or pattern must be specified")
			}

			kubeconfig, remoteCtx, _, namespace, allNamespaces := GetGlobalFlags()
			return handleLogsCommand(args[0], follow, previous, container, since, sinceTime, timestamps, tail, limitBytes, kubeconfig, remoteCtx, namespace, allNamespaces)
		},
	}

	// Add logs-specific flags
	cmd.Flags().BoolVarP(&follow, "follow", "f", false, "specify if the logs should be streamed")
	cmd.Flags().BoolVarP(&previous, "previous", "p", false, "if true, print the logs for the previous instance of the container in a pod if it exists")
	cmd.Flags().StringVarP(&container, "container", "c", "", "print the logs of this container")
	cmd.Flags().StringVar(&since, "since", "", "only return logs newer than a relative duration like 5s, 2m, or 3h")
	cmd.Flags().StringVar(&sinceTime, "since-time", "", "only return logs after a specific date (RFC3339)")
	cmd.Flags().BoolVar(&timestamps, "timestamps", false, "include timestamps on each line in the log output")
	cmd.Flags().Int64Var(&tail, "tail", -1, "lines of recent log file to display. Defaults to -1 with no selector, showing all log lines otherwise 10, if a selector is provided")
	cmd.Flags().Int64Var(&limitBytes, "limit-bytes", 0, "maximum bytes of logs to return. Defaults to no limit")

	cmd.SetHelpFunc(logsHelpFunc)

	return cmd
}

func handleLogsCommand(podPattern string, follow, previous bool, container, since, sinceTime string, timestamps bool, tail, limitBytes int64, kubeconfig, remoteCtx, namespace string, allNamespaces bool) error {
	clusters, err := cluster.DiscoverClusters(kubeconfig, remoteCtx)
	if err != nil {
		return fmt.Errorf("failed to discover clusters: %v", err)
	}

	if len(clusters) == 0 {
		return fmt.Errorf("no clusters discovered")
	}

	if follow {
		fmt.Println("Warning: Follow mode (-f) across multiple clusters can be overwhelming.")
		fmt.Println("Consider using this command on a specific cluster for follow mode.")
		fmt.Println("Example: kubectl logs pod-name -f --context=specific-cluster")
		fmt.Println()
	}

	fmt.Printf("Getting logs for pod pattern '%s' across %d clusters...\n\n", podPattern, len(clusters))

	foundAnyPod := false

	for _, clusterInfo := range clusters {
		if clusterInfo.Client == nil {
			fmt.Printf("Warning: skipping cluster %s (no client available)\n", clusterInfo.Name)
			continue
		}

		fmt.Printf("=== Cluster: %s (Context: %s) ===\n", clusterInfo.Name, clusterInfo.Context)

		// Get matching pods from this cluster
		matchingPods, err := getMatchingPods(clusterInfo, podPattern, namespace, allNamespaces)
		if err != nil {
			fmt.Printf("Error listing pods in cluster %s: %v\n", clusterInfo.Name, err)
			fmt.Printf("\n")
			continue
		}

		if len(matchingPods) == 0 {
			fmt.Printf("No pods matching pattern '%s' found in cluster %s\n", podPattern, clusterInfo.Name)
			fmt.Printf("\n")
			continue
		}

		for _, podName := range matchingPods {
			fmt.Printf("--- Pod: %s ---\n", podName)

			kubectlArgs := buildLogsArgs(podName, follow, previous, container, since, sinceTime, timestamps, tail, limitBytes, namespace, allNamespaces, clusterInfo.Context)

			output, err := executeKubectlLogs(kubectlArgs, kubeconfig, clusterInfo.Name)
			if err != nil {
				fmt.Printf("Error getting logs for pod '%s' in cluster %s: %v\n", podName, clusterInfo.Name, err)
			} else if strings.TrimSpace(output) != "" {
				fmt.Print(output)
				foundAnyPod = true
			} else {
				fmt.Printf("No logs available for pod '%s'\n", podName)
			}
			fmt.Printf("\n")
		}
	}

	if !foundAnyPod {
		fmt.Printf("No pods matching pattern '%s' found in any cluster\n", podPattern)
	}

	return nil
}

func buildLogsArgs(podName string, follow, previous bool, container, since, sinceTime string, timestamps bool, tail, limitBytes int64, namespace string, allNamespaces bool, clusterContext string) []string {
	var kubectlArgs []string

	kubectlArgs = append(kubectlArgs, "logs", podName)

	if container != "" {
		kubectlArgs = append(kubectlArgs, "-c", container)
	}

	if follow {
		kubectlArgs = append(kubectlArgs, "-f")
	}

	if previous {
		kubectlArgs = append(kubectlArgs, "-p")
	}

	if since != "" {
		kubectlArgs = append(kubectlArgs, "--since", since)
	}

	if sinceTime != "" {
		kubectlArgs = append(kubectlArgs, "--since-time", sinceTime)
	}

	if timestamps {
		kubectlArgs = append(kubectlArgs, "--timestamps")
	}

	if tail >= 0 {
		kubectlArgs = append(kubectlArgs, "--tail", fmt.Sprintf("%d", tail))
	}

	if limitBytes > 0 {
		kubectlArgs = append(kubectlArgs, "--limit-bytes", fmt.Sprintf("%d", limitBytes))
	}

	if !allNamespaces && namespace != "" {
		kubectlArgs = append(kubectlArgs, "-n", namespace)
	}

	kubectlArgs = append(kubectlArgs, "--context", clusterContext)

	return kubectlArgs
}

func executeKubectlLogs(args []string, kubeconfig, clusterName string) (string, error) {

	cmd := exec.Command("kubectl", args...)

	cmd.Env = os.Environ()
	if kubeconfig != "" {
		cmd.Env = append(cmd.Env, "KUBECONFIG="+kubeconfig)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	output := stdout.String()
	stderrOutput := stderr.String()

	if err != nil {
		if strings.Contains(stderrOutput, "not found") || strings.Contains(stderrOutput, "NotFound") {
			return "", fmt.Errorf("not found")
		}

		return "", fmt.Errorf("kubectl logs failed: %v\nStderr: %s", err, stderrOutput)
	}

	if stderrOutput != "" && !strings.Contains(stderrOutput, "not found") {
		output = "# Warning: " + strings.TrimSpace(stderrOutput) + "\n" + output
	}

	return output, nil
}

func getMatchingPods(clusterInfo cluster.ClusterInfo, pattern, namespace string, allNamespaces bool) ([]string, error) {
	var matchingPods []string

	targetNS := ""
	if allNamespaces {
		targetNS = ""
	} else if namespace != "" {
		targetNS = namespace
	} else {
		targetNS = "default"
	}

	pods, err := clusterInfo.Client.CoreV1().Pods(targetNS).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	hasWildcard := strings.Contains(pattern, "*")

	for _, pod := range pods.Items {
		if hasWildcard {

			matched, err := filepath.Match(pattern, pod.Name)
			if err != nil {
				continue
			}
			if matched {
				matchingPods = append(matchingPods, pod.Name)
			}
		} else {

			if pod.Name == pattern {
				matchingPods = append(matchingPods, pod.Name)
			}
		}
	}

	return matchingPods, nil
}
