package cmd

import (
	"github.com/spf13/cobra"
)

func newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete ([-f FILENAME] | TYPE [(NAME | -l label | --all)])",
		Aliases: []string{"del", "rm"},
		Short:   "Delete resources across all managed clusters",
		Long: `Delete resources across all KubeStellar managed clusters by filenames, stdin, 
resources and names, or by resources and label selector.`,
		Example: `# Delete a pod across all clusters
kubestellar delete pod nginx-pod

# Delete pods with a label across all clusters
kubestellar delete pods -l app=nginx

# Delete resources from a file across all clusters
kubestellar delete -f deployment.yaml`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement delete operation using shared core
			cmd.Println("Delete command will be implemented with shared core operations")
			return nil
		},
	}

	return cmd
}