package cmd

import (
	"fmt"

	"kubectl-multi/pkg/cluster"
	"kubectl-multi/pkg/util"

	"github.com/spf13/cobra"
)

// Custom help function for delete command
func deleteHelpFunc(cmd *cobra.Command, args []string) {
	// Get original kubectl help using the new implementation
	cmdInfo, err := util.GetKubectlCommandInfo("delete")
	if err != nil {
		// Fallback to default help if kubectl help is not available
		cmd.Help()
		return
	}

	// Multi-cluster plugin information
	multiClusterInfo := `Delete resources across all managed clusters.
This command deletes resources from all KubeStellar managed clusters.`

	// Multi-cluster examples
	multiClusterExamples := `# Delete a deployment from all managed clusters
kubectl multi delete deployment nginx

# Delete pods with a specific label from all clusters
kubectl multi delete pods -l app=nginx

# Delete resources from a file across all clusters
kubectl multi delete -f deployment.yaml

# Delete all pods in all clusters
kubectl multi delete pods --all

# Delete with force flag across all clusters
kubectl multi delete pod nginx --force`

	// Multi-cluster usage
	multiClusterUsage := `kubectl multi delete [TYPE[.VERSION][.GROUP] [NAME | -l label] | TYPE[.VERSION][.GROUP]/NAME ...] [flags]`

	// Format combined help using the new CommandInfo structure
	combinedHelp := util.FormatMultiClusterHelp(cmdInfo, multiClusterInfo, multiClusterExamples, multiClusterUsage)
	fmt.Fprintln(cmd.OutOrStdout(), combinedHelp)
}

func newDeleteCommand() *cobra.Command {
	var filename string
	var recursive bool
	var selector string
	var all bool
	var force bool
	var grace int
	var timeout string
	var wait bool

	cmd := &cobra.Command{
		Use:   "delete [TYPE[.VERSION][.GROUP] [NAME | -l label] | TYPE[.VERSION][.GROUP]/NAME ...]",
		Short: "Delete resources across all managed clusters",
		Long: `Delete resources across all managed clusters.
This command deletes resources from all KubeStellar managed clusters.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if filename == "" && len(args) == 0 {
				return fmt.Errorf("resource type and name, or filename must be specified")
			}
			kubeconfig, remoteCtx, _, namespace, allNamespaces := GetGlobalFlags()
			return handleDeleteCommand(args, filename, recursive, selector, all, force, grace, timeout, wait, kubeconfig, remoteCtx, namespace, allNamespaces)
		},
	}

	cmd.Flags().StringVarP(&filename, "filename", "f", "", "Filename, directory, or URL to files to delete")
	cmd.Flags().BoolVarP(&recursive, "recursive", "R", false, "Process the directory used in -f, --filename recursively")
	cmd.Flags().StringVarP(&selector, "selector", "l", "", "Selector (label query) to filter on")
	cmd.Flags().BoolVar(&all, "all", false, "Delete all resources in the namespace of the specified resource types")
	cmd.Flags().BoolVar(&force, "force", false, "If true, immediately remove resources from API and bypass graceful deletion")
	cmd.Flags().IntVar(&grace, "grace-period", -1, "Period of time in seconds given to the resource to terminate gracefully")
	cmd.Flags().StringVar(&timeout, "timeout", "0s", "The length of time to wait before giving up on a delete")
	cmd.Flags().BoolVar(&wait, "wait", true, "If true, wait for resources to be gone before returning")

	// Set custom help function
	cmd.SetHelpFunc(deleteHelpFunc)

	return cmd
}

func handleDeleteCommand(args []string, filename string, recursive bool, selector string, all, force bool, grace int, timeout string, wait bool, kubeconfig, remoteCtx, namespace string, allNamespaces bool) error {
	clusters, err := cluster.DiscoverClusters(kubeconfig, remoteCtx)
	if err != nil {
		return fmt.Errorf("failed to discover clusters: %v", err)
	}

	if len(clusters) == 0 {
		return fmt.Errorf("no clusters discovered")
	}

	// Build kubectl args
	cmdArgs := []string{"delete"}

	// Add filename or resource args
	if filename != "" {
		cmdArgs = append(cmdArgs, "-f", filename)
		if recursive {
			cmdArgs = append(cmdArgs, "-R")
		}
	} else {
		cmdArgs = append(cmdArgs, args...)
	}

	// Add flags
	if selector != "" {
		cmdArgs = append(cmdArgs, "-l", selector)
	}
	if all {
		cmdArgs = append(cmdArgs, "--all")
	}
	if force {
		cmdArgs = append(cmdArgs, "--force")
	}
	if grace >= 0 {
		cmdArgs = append(cmdArgs, fmt.Sprintf("--grace-period=%d", grace))
	}
	if timeout != "0s" {
		cmdArgs = append(cmdArgs, "--timeout="+timeout)
	}
	if !wait {
		cmdArgs = append(cmdArgs, "--wait=false")
	}
	if namespace != "" {
		cmdArgs = append(cmdArgs, "-n", namespace)
	}
	if allNamespaces {
		cmdArgs = append(cmdArgs, "-A")
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
			fmt.Printf("Error: %v\n", err)
		} else {
			fmt.Print(output)
		}
		fmt.Println()
	}

	return nil
}

func newExecCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exec POD [-c CONTAINER] -- COMMAND [args...]",
		Short: "Execute a command in a container across managed clusters",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("exec command not yet implemented")
		},
	}
	return cmd
}

func newEditCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit [TYPE[.VERSION][.GROUP]/]NAME",
		Short: "Edit a resource on the server across managed clusters",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("edit command not yet implemented")
		},
	}
	return cmd
}

func newPatchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "patch [TYPE[.VERSION][.GROUP]/]NAME --patch PATCH",
		Short: "Update field(s) of a resource across managed clusters",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("patch command not yet implemented")
		},
	}
	return cmd
}

func newScaleCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scale [TYPE[.VERSION][.GROUP]/]NAME --replicas=COUNT",
		Short: "Set a new size for a deployment, replica set, or stateful set across managed clusters",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("scale command not yet implemented")
		},
	}
	return cmd
}

func newPortForwardCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "port-forward POD [LOCAL_PORT:]REMOTE_PORT",
		Short: "Forward one or more local ports to a pod across managed clusters",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("port-forward command not yet implemented")
		},
	}
	return cmd
}

func newTopCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "top [TYPE]",
		Short: "Display resource (CPU/memory/storage) usage across managed clusters",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("top command not yet implemented")
		},
	}
	return cmd
}
