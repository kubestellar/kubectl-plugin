package cmd

import (
	"fmt"

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
	cmd := &cobra.Command{
		Use:   "delete [TYPE[.VERSION][.GROUP] [NAME | -l label] | TYPE[.VERSION][.GROUP]/NAME ...]",
		Short: "Delete resources across all managed clusters",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("delete command not yet implemented")
		},
	}

	// Set custom help function
	cmd.SetHelpFunc(deleteHelpFunc)

	return cmd
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

func newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create -f FILENAME",
		Short: "Create a resource from a file or from stdin across managed clusters",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("create command not yet implemented")
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
