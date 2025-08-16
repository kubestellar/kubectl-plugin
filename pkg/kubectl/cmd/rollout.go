package cmd

import (
	"fmt"

	"kubectl-multi/pkg/cluster"

	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
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

	// Find current context from kubeconfig
	currentContext := ""
	{
		loading := clientcmd.NewDefaultClientConfigLoadingRules()
		if kubeconfig != "" {
			loading.ExplicitPath = kubeconfig
		}
		cfg := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loading, &clientcmd.ConfigOverrides{})
		rawCfg, err := cfg.RawConfig()
		if err == nil {
			currentContext = rawCfg.CurrentContext
		}
	}

	// Identify ITS (control) cluster context
	itsContext := remoteCtx

	// Build maps for quick lookup
	contextToCluster := make(map[string]cluster.ClusterInfo)
	for _, c := range clusters {
		contextToCluster[c.Context] = c
	}

	// 1. Run for current context (if present)
	if cinfo, ok := contextToCluster[currentContext]; ok {
		args := []string{"rollout", subcommand}
		if len(extraArgs) > 0 {
			args = append(args, extraArgs...)
		}
		args = append(args, "--context", cinfo.Context)
		cmdOutput, err := runKubectl(args, kubeconfig)
		fmt.Printf("=== Cluster: %s ===\n", cinfo.Context)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			fmt.Print(cmdOutput)
		}
		fmt.Println()
	}

	// 2. Run for KubeStellar clusters (excluding ITS and current)
	for _, c := range clusters {
		if c.Context == currentContext || c.Context == itsContext {
			continue
		}
		args := []string{"rollout", subcommand}
		if len(extraArgs) > 0 {
			args = append(args, extraArgs...)
		}
		args = append(args, "--context", c.Context)
		cmdOutput, err := runKubectl(args, kubeconfig)
		fmt.Printf("=== Cluster: %s ===\n", c.Context)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			fmt.Print(cmdOutput)
		}
		fmt.Println()
	}

	// 3. Print warning for ITS (control) cluster
	if cinfo, ok := contextToCluster[itsContext]; ok {
		fmt.Printf("=== Cluster: %s ===\n", cinfo.Context)
		fmt.Printf("Cannot perform this operation on ITS (control) cluster: %s\n", cinfo.Context)
		fmt.Println()
	}

	return nil
}
