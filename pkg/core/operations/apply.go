package operations

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	"kubectl-multi/pkg/cluster"
	"k8s.io/client-go/tools/clientcmd"
)

// ApplyOptions contains all the options for the apply operation
type ApplyOptions struct {
	Clusters      []cluster.ClusterInfo
	Filename      string
	Recursive     bool
	DryRun        string
	Namespace     string
	AllNamespaces bool
	Kubeconfig    string
}

// ExecuteApply performs the apply operation across all clusters
func ExecuteApply(opts ApplyOptions) error {
	if len(opts.Clusters) == 0 {
		return fmt.Errorf("no clusters discovered")
	}

	// Find current context from kubeconfig
	currentContext := ""
	{
		loading := clientcmd.NewDefaultClientConfigLoadingRules()
		if opts.Kubeconfig != "" {
			loading.ExplicitPath = opts.Kubeconfig
		}
		cfg := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loading, &clientcmd.ConfigOverrides{})
		rawCfg, err := cfg.RawConfig()
		if err == nil {
			currentContext = rawCfg.CurrentContext
		}
	}

	successCount := 0
	failureCount := 0

	for _, clusterInfo := range opts.Clusters {
		fmt.Printf("\n--- Applying to cluster: %s ---\n", clusterInfo.Name)

		args := []string{"apply", "-f", opts.Filename}
		args = append(args, "--context", clusterInfo.Name)

		if opts.Kubeconfig != "" {
			args = append(args, "--kubeconfig", opts.Kubeconfig)
		}

		if opts.Namespace != "" && opts.Namespace != "default" {
			args = append(args, "-n", opts.Namespace)
		}

		if opts.AllNamespaces {
			args = append(args, "--all-namespaces")
		}

		if opts.Recursive {
			args = append(args, "-R")
		}

		if opts.DryRun != "" && opts.DryRun != "none" {
			args = append(args, "--dry-run="+opts.DryRun)
		}

		cmd := exec.Command("kubectl", args...)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err != nil {
			fmt.Printf("ERROR: %s\n", stderr.String())
			failureCount++
		} else {
			fmt.Print(stdout.String())
			successCount++
		}
	}

	// Restore original context if it was changed
	if currentContext != "" && opts.Kubeconfig != "" {
		cmd := exec.Command("kubectl", "config", "use-context", currentContext, "--kubeconfig", opts.Kubeconfig)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		_ = cmd.Run()
	}

	fmt.Printf("\n--- Summary ---\n")
	fmt.Printf("Successfully applied to %d cluster(s)\n", successCount)
	if failureCount > 0 {
		fmt.Printf("Failed to apply to %d cluster(s)\n", failureCount)
		return fmt.Errorf("apply failed on %d cluster(s)", failureCount)
	}

	return nil
}