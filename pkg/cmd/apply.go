package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"kubectl-multi/pkg/cluster"
	"kubectl-multi/pkg/util"

	"github.com/spf13/cobra"
)

// Custom help function for apply command
func applyHelpFunc(cmd *cobra.Command, args []string) {
	// Get original kubectl help using the new implementation
	cmdInfo, err := util.GetKubectlCommandInfo("apply")
	if err != nil {
		// Fallback to default help if kubectl help is not available
		cmd.Help()
		return
	}

	// Multi-cluster plugin information
	multiClusterInfo := `Apply a configuration to resources across all managed clusters.
This command applies manifests to all KubeStellar managed clusters.`

	// Multi-cluster examples
	multiClusterExamples := `# Apply a deployment to all managed clusters
kubectl multi apply -f deployment.yaml

# Apply resources from a directory to all clusters
kubectl multi apply -k dir/

# Apply with dry-run to see what would be applied
kubectl multi apply -f deployment.yaml --dry-run=client

# Apply resources recursively from a directory
kubectl multi apply -f dir/ -R`

	// Multi-cluster usage
	multiClusterUsage := `kubectl multi apply (-f FILENAME | -k DIRECTORY) [flags]`

	// Format combined help using the new CommandInfo structure
	combinedHelp := util.FormatMultiClusterHelp(cmdInfo, multiClusterInfo, multiClusterExamples, multiClusterUsage)
	fmt.Fprintln(cmd.OutOrStdout(), combinedHelp)
}

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

	// Set custom help function
	cmd.SetHelpFunc(applyHelpFunc)

	// Add view-last-applied as a subcommand
	cmd.AddCommand(newViewLastAppliedCommand())
	cmd.AddCommand(newEditLastAppliedCommand())
	cmd.AddCommand(newSetLastAppliedCommand())

	return cmd
}

func handleApplyCommand(filename string, recursive bool, dryRun, kubeconfig, remoteCtx, namespace string, allNamespaces bool) error {
	clusters, err := cluster.DiscoverClusters(kubeconfig, remoteCtx)
	if err != nil {
		return fmt.Errorf("failed to discover clusters: %v", err)
	}
	if len(clusters) == 0 {
		return fmt.Errorf("no clusters discovered")
	}

	maxProc := 5 // fixed concurrency for now
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
			args := []string{"apply", "-f", filename, "--context", clusterInfo.Name}
			if recursive {
				args = append(args, "-R")
			}
			if dryRun != "none" && dryRun != "" {
				args = append(args, "--dry-run="+dryRun)
			}
			if namespace != "" {
				args = append(args, "-n", namespace)
			}
			output, err := runKubectl(args, kubeconfig)
			results <- struct {
				Cluster string
				Output  string
				Err     error
			}{Cluster: clusterInfo.Name, Output: output, Err: err}
		}(c)
	}

	// Wait for all results
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

func newViewLastAppliedCommand() *cobra.Command {
	var filename string
	var output string
	var recursive bool

	cmd := &cobra.Command{
		Use:   "view-last-applied",
		Short: "View the latest last-applied-configuration annotations across all managed clusters",
		Long:  `View the latest last-applied-configuration annotations by type/name or file across all KubeStellar managed clusters.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			kubeconfig, remoteCtx, _, namespace, allNamespaces := GetGlobalFlags()
			return handleViewLastAppliedCommand(filename, output, recursive, args, kubeconfig, remoteCtx, namespace, allNamespaces)
		},
	}

	cmd.Flags().StringVarP(&filename, "filename", "f", "", "Filename, directory, or URL to files that contains the last-applied-configuration annotations")
	cmd.Flags().StringVarP(&output, "output", "o", "yaml", "Output format. Must be one of yaml|json")
	cmd.Flags().BoolVarP(&recursive, "recursive", "R", false, "Process the directory used in -f, --filename recursively")

	return cmd
}

func handleViewLastAppliedCommand(filename, output string, recursive bool, extraArgs []string, kubeconfig, remoteCtx, namespace string, allNamespaces bool) error {
	clusters, err := cluster.DiscoverClusters(kubeconfig, remoteCtx)
	if err != nil {
		return fmt.Errorf("failed to discover clusters: %v", err)
	}
	if len(clusters) == 0 {
		return fmt.Errorf("no clusters discovered")
	}

	maxProc := 5 // fixed concurrency for now
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
			args := []string{"apply", "view-last-applied"}
			if filename != "" {
				args = append(args, "-f", filename)
			}
			if output != "" {
				args = append(args, "-o", output)
			}
			if recursive {
				args = append(args, "-R")
			}
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

	// Wait for all results
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

func newEditLastAppliedCommand() *cobra.Command {
	var filename string
	var output string
	var recursive bool

	cmd := &cobra.Command{
		Use:   "edit-last-applied",
		Short: "Edit the last-applied-configuration annotations across all managed clusters",
		Long:  `Edit the latest last-applied-configuration annotations by type/name or file across all KubeStellar managed clusters. Opens your default editor for each resource.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			kubeconfig, remoteCtx, _, namespace, allNamespaces := GetGlobalFlags()
			return handleEditLastAppliedCommand(filename, output, recursive, args, kubeconfig, remoteCtx, namespace, allNamespaces)
		},
	}

	cmd.Flags().StringVarP(&filename, "filename", "f", "", "Filename, directory, or URL to files to edit the last-applied-configuration annotations")
	cmd.Flags().StringVarP(&output, "output", "o", "yaml", "Output format. Must be one of yaml|json")
	cmd.Flags().BoolVarP(&recursive, "recursive", "R", false, "Process the directory used in -f, --filename recursively")

	return cmd
}

func handleEditLastAppliedCommand(filename, output string, recursive bool, extraArgs []string, kubeconfig, remoteCtx, namespace string, allNamespaces bool) error {
	fmt.Println("kubectl multi does not support interactive commands yet.")
	return nil
}

func newSetLastAppliedCommand() *cobra.Command {
	var filename string
	var output string
	var createAnnotation bool
	var dryRun string
	var recursive bool

	cmd := &cobra.Command{
		Use:   "set-last-applied",
		Short: "Set the last-applied-configuration annotations across all managed clusters",
		Long:  `Set the latest last-applied-configuration annotations by file across all KubeStellar managed clusters.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return handleSetLastAppliedCommand(filename, output, createAnnotation, dryRun, recursive)
		},
	}

	cmd.Flags().StringVarP(&filename, "filename", "f", "", "Filename, directory, or URL to files that contains the last-applied-configuration annotations")
	cmd.Flags().StringVarP(&output, "output", "o", "yaml", "Output format. Must be one of yaml|json")
	cmd.Flags().BoolVar(&createAnnotation, "create-annotation", false, "Create the annotation if it does not already exist")
	cmd.Flags().StringVar(&dryRun, "dry-run", "none", "Must be 'none', 'server', or 'client'")
	cmd.Flags().BoolVarP(&recursive, "recursive", "R", false, "Process the directory used in -f, --filename recursively")

	return cmd
}

func handleSetLastAppliedCommand(filename, output string, createAnnotation bool, dryRun string, recursive bool) error {
	fmt.Println("kubectl multi does not support interactive commands yet.")
	return nil
}

// runKubectl runs a kubectl command with the given args and kubeconfig, returns output and error
func runKubectl(args []string, kubeconfig string) (string, error) {
	cmd := exec.Command("kubectl", args...)
	if kubeconfig != "" {
		cmd.Env = append(os.Environ(), "KUBECONFIG="+kubeconfig)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return stdout.String() + stderr.String(), err
	}
	return stdout.String(), nil
}
