package cmd

import (
	"fmt"
	"strings"

	"kubectl-multi/pkg/cluster"

	"github.com/spf13/cobra"
)

func newRolloutCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rollout",
		Short: "Manage the rollout of a resource across all managed clusters",
	}
	cmd.AddCommand(newRolloutHistoryCommand())
	cmd.AddCommand(newRolloutPauseCommand())
	cmd.AddCommand(newRolloutRestartCommand())
	cmd.AddCommand(newRolloutResumeCommand())
	cmd.AddCommand(newRolloutStatusCommand())
	cmd.AddCommand(newRolloutUndoCommand())
	return cmd
}

func newRolloutHistoryCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "history",
		Short: "View the rollout history of a resource across all managed clusters",
		RunE: func(cmd *cobra.Command, args []string) error {
			kubeconfig, remoteCtx, _, _, _ := GetGlobalFlags()
			return handleRolloutSubcommand("history", args, kubeconfig, remoteCtx)
		},
	}
	return cmd
}

func newRolloutPauseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pause",
		Short: "Pause a resource across all managed clusters",
		RunE: func(cmd *cobra.Command, args []string) error {
			kubeconfig, remoteCtx, _, _, _ := GetGlobalFlags()
			return handleRolloutSubcommand("pause", args, kubeconfig, remoteCtx)
		},
	}
	return cmd
}

func newRolloutRestartCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restart",
		Short: "Restart a resource across all managed clusters",
		RunE: func(cmd *cobra.Command, args []string) error {
			kubeconfig, remoteCtx, _, _, _ := GetGlobalFlags()
			return handleRolloutSubcommand("restart", args, kubeconfig, remoteCtx)
		},
	}
	return cmd
}

func newRolloutResumeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resume",
		Short: "Resume a resource across all managed clusters",
		RunE: func(cmd *cobra.Command, args []string) error {
			kubeconfig, remoteCtx, _, _, _ := GetGlobalFlags()
			return handleRolloutSubcommand("resume", args, kubeconfig, remoteCtx)
		},
	}
	return cmd
}

func newRolloutStatusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show the status of the rollout across all managed clusters",
		RunE: func(cmd *cobra.Command, args []string) error {
			kubeconfig, remoteCtx, _, _, _ := GetGlobalFlags()
			return handleRolloutSubcommand("status", args, kubeconfig, remoteCtx)
		},
	}
	return cmd
}

func newRolloutUndoCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "undo",
		Short: "Roll back to a previous rollout across all managed clusters",
		RunE: func(cmd *cobra.Command, args []string) error {
			kubeconfig, remoteCtx, _, _, _ := GetGlobalFlags()
			return handleRolloutSubcommand("undo", args, kubeconfig, remoteCtx)
		},
	}
	return cmd
}

func handleRolloutSubcommand(subcommand string, extraArgs []string, kubeconfig, remoteCtx string) error {
	clusters, err := cluster.DiscoverClusters(kubeconfig, remoteCtx)
	if err != nil {
		return fmt.Errorf("failed to discover clusters: %v", err)
	}
	if len(clusters) == 0 {
		return fmt.Errorf("no clusters discovered")
	}
	maxProc := 5
	sem := make(chan struct{}, maxProc)
	results := make(chan struct {
		Cluster string
		Output  string
		Err     error
	}, len(clusters))
	for _, c := range clusters {
		sem <- struct{}{}
		go func(clusterInfo cluster.ClusterInfo) {
			defer func() { <-sem }()
			args := []string{"rollout", subcommand}
			if len(extraArgs) > 0 {
				args = append(args, extraArgs...)
			}
			args = append(args, "--context", clusterInfo.Name)
			cmdOutput, err := runKubectl(args, kubeconfig)
			results <- struct {
				Cluster string
				Output  string
				Err     error
			}{Cluster: clusterInfo.Name, Output: cmdOutput, Err: err}
		}(c)
	}
	for i := 0; i < len(clusters); i++ {
		res := <-results
		fmt.Printf("=== Cluster: %s ===\n", res.Cluster)
		if res.Err != nil {
			if strings.Contains(res.Cluster, "its1") {
				fmt.Printf("Cannot perform this operation on ITS (control) cluster: %s\n", res.Cluster)
			} else {
				fmt.Printf("Error: %v\n", res.Err)
			}
		} else {
			fmt.Print(res.Output)
		}
		fmt.Println()
	}
	return nil
}
