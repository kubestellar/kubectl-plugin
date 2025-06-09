package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newApplyCommand() *cobra.Command {
	var filename string
	var recursive bool
	var dryRun string

	cmd := &cobra.Command{
		Use:   "apply (-f FILENAME | --filename=FILENAME)",
		Short: "Apply a configuration to resources across all managed clusters",
		Long: `Apply a configuration to resources across all managed clusters.
This command applies manifests to all KubeStellar managed clusters.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			kubeconfig, remoteCtx, _, namespace, allNamespaces := GetGlobalFlags()
			return handleApplyCommand(filename, recursive, dryRun, kubeconfig, remoteCtx, namespace, allNamespaces)
		},
	}

	cmd.Flags().StringVarP(&filename, "filename", "f", "", "filename, directory, or URL to files to use to apply the resource")
	cmd.Flags().BoolVarP(&recursive, "recursive", "R", false, "process the directory used in -f, --filename recursively")
	cmd.Flags().StringVar(&dryRun, "dry-run", "none", "must be \"none\", \"server\", or \"client\"")

	return cmd
}

func handleApplyCommand(filename string, recursive bool, dryRun, kubeconfig, remoteCtx, namespace string, allNamespaces bool) error {
	return fmt.Errorf("apply command not yet implemented")
}
