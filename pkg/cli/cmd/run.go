package cmd

import (
	"github.com/spf13/cobra"
)

func newRunCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run NAME --image=image [--env=\"key=value\"] [--port=port] [--dry-run=server|client] [--overrides=inline-json] [--command] -- [COMMAND] [args...]",
		Short: "Run a particular image across clusters",
		Long:  `Create and run a particular image across all KubeStellar managed clusters.`,
		Example: `# Start a nginx deployment
kubestellar run nginx --image=nginx

# Start a deployment with environment variables
kubestellar run nginx --image=nginx --env="ENV=production"

# Start a deployment exposing a port
kubestellar run nginx --image=nginx --port=80`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement run operation using shared core
			cmd.Println("Run command will be implemented with shared core operations")
			return nil
		},
	}

	return cmd
}