package cmd

import (
	"fmt"

	"kubectl-multi/pkg/cluster"

	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
)

func newRunCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Create and run a particular image in a pod across all managed clusters",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check for interactive flags
			for _, arg := range args {
				if arg == "--attach" || arg == "-it" || arg == "-i" || arg == "-t" {
					fmt.Println("kubectl multi does not support interactive commands (attach/tty) yet.")
					return nil
				}
			}
			kubeconfig, remoteCtx, _, _, _ := GetGlobalFlags()
			return handleRunMulti(args, kubeconfig, remoteCtx)
		},
	}
	cmd.DisableFlagParsing = true
	return cmd
}

func handleRunMulti(args []string, kubeconfig, remoteCtx string) error {
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
		output, err := runKubectl(append([]string{"run"}, append(args, "--context", cinfo.Context)...), kubeconfig)
		fmt.Printf("=== Cluster: %s ===\n", cinfo.Context)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			fmt.Print(output)
		}
		fmt.Println()
	}

	// 2. Run for KubeStellar clusters (excluding ITS and current)
	for _, c := range clusters {
		if c.Context == currentContext || c.Context == itsContext {
			continue
		}
		output, err := runKubectl(append([]string{"run"}, append(args, "--context", c.Context)...), kubeconfig)
		fmt.Printf("=== Cluster: %s ===\n", c.Context)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			fmt.Print(output)
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
