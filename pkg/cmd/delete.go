package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [TYPE[.VERSION][.GROUP] [NAME | -l label] | TYPE[.VERSION][.GROUP]/NAME ...]",
		Short: "Delete resources across all managed clusters",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("delete command not yet implemented")
		},
	}
	return cmd
}

func newLogsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logs [-f] [-p] POD [-c CONTAINER]",
		Short: "Print the logs for a container in a pod across managed clusters",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("logs command not yet implemented")
		},
	}
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

func newRolloutCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rollout SUBCOMMAND",
		Short: "Manage the rollout of a resource across managed clusters",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("rollout command not yet implemented")
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
