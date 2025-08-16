package operations

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"kubectl-multi/pkg/cluster"
	"sigs.k8s.io/yaml"
)

// HelmOptions contains options for Helm operations
type HelmOptions struct {
	Kubeconfig      string
	Clusters        []cluster.ClusterInfo
	ReleaseName     string
	ChartName       string
	ChartVersion    string
	Namespace       string
	Values          map[string]interface{}
	ValuesFile      string
	ValuesFiles     []string // Multiple values files
	Wait            bool
	Timeout         string
	CreateNS        bool
	CreateNamespace bool // Alias for CreateNS
	DryRun          bool
	Atomic          bool
	Variables       map[string]string // Key-value pairs for --set
	Install         bool              // For upgrade --install
	KeepHistory     bool              // For uninstall
	AllNamespaces   bool              // For list
	Deployed        bool              // For list --deployed
	Failed          bool              // For list --failed
	Pending         bool              // For list --pending
	Revision        string            // For rollback and values
	AllValues       bool              // For values --all
}

// HelmInstall installs a Helm chart across all specified clusters
func HelmInstall(opts HelmOptions) error {
	if len(opts.Clusters) == 0 {
		return fmt.Errorf("no clusters specified for Helm installation")
	}

	successCount := 0
	failureCount := 0

	for _, clusterInfo := range opts.Clusters {
		fmt.Printf("\n--- Installing Helm chart to cluster: %s ---\n", clusterInfo.Name)

		args := buildHelmInstallArgs(opts, clusterInfo.Name)
		
		cmd := exec.Command("helm", args...)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err != nil {
			fmt.Printf("ERROR in cluster %s: %s\n", clusterInfo.Name, stderr.String())
			failureCount++
		} else {
			fmt.Printf("SUCCESS in cluster %s: %s\n", clusterInfo.Name, stdout.String())
			successCount++
		}
	}

	fmt.Printf("\n--- Helm Installation Summary ---\n")
	fmt.Printf("Successfully installed to %d cluster(s)\n", successCount)
	if failureCount > 0 {
		fmt.Printf("Failed to install to %d cluster(s)\n", failureCount)
		return fmt.Errorf("helm install failed on %d cluster(s)", failureCount)
	}

	return nil
}

// buildHelmInstallArgs builds the helm install command arguments
func buildHelmInstallArgs(opts HelmOptions, clusterContext string) []string {
	args := []string{"install", opts.ReleaseName, opts.ChartName}

	// Add kubeconfig and context
	if opts.Kubeconfig != "" {
		args = append(args, "--kubeconfig", opts.Kubeconfig)
	}
	args = append(args, "--kube-context", clusterContext)

	// Add namespace
	if opts.Namespace != "" {
		args = append(args, "--namespace", opts.Namespace)
	}

	// Add chart version if specified
	if opts.ChartVersion != "" {
		args = append(args, "--version", opts.ChartVersion)
	}

	// Add values file if specified
	if opts.ValuesFile != "" {
		args = append(args, "--values", opts.ValuesFile)
	}

	// Add individual variables
	for key, value := range opts.Variables {
		args = append(args, "--set", fmt.Sprintf("%s=%s", key, value))
	}

	// Add flags
	if opts.CreateNS {
		args = append(args, "--create-namespace")
	}

	if opts.Wait {
		args = append(args, "--wait")
	}

	if opts.Timeout != "" {
		args = append(args, "--timeout", opts.Timeout)
	}

	if opts.DryRun {
		args = append(args, "--dry-run")
	}

	if opts.Atomic {
		args = append(args, "--atomic")
	}

	return args
}

// HelmUpgrade upgrades a Helm release across all specified clusters
func HelmUpgrade(opts HelmOptions) error {
	if len(opts.Clusters) == 0 {
		return fmt.Errorf("no clusters specified for Helm upgrade")
	}

	successCount := 0
	failureCount := 0

	for _, clusterInfo := range opts.Clusters {
		fmt.Printf("\n--- Upgrading Helm release in cluster: %s ---\n", clusterInfo.Name)

		args := buildHelmUpgradeArgs(opts, clusterInfo.Name)
		
		cmd := exec.Command("helm", args...)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err != nil {
			fmt.Printf("ERROR in cluster %s: %s\n", clusterInfo.Name, stderr.String())
			failureCount++
		} else {
			fmt.Printf("SUCCESS in cluster %s: %s\n", clusterInfo.Name, stdout.String())
			successCount++
		}
	}

	fmt.Printf("\n--- Helm Upgrade Summary ---\n")
	fmt.Printf("Successfully upgraded in %d cluster(s)\n", successCount)
	if failureCount > 0 {
		fmt.Printf("Failed to upgrade in %d cluster(s)\n", failureCount)
		return fmt.Errorf("helm upgrade failed on %d cluster(s)", failureCount)
	}

	return nil
}

// buildHelmUpgradeArgs builds the helm upgrade command arguments
func buildHelmUpgradeArgs(opts HelmOptions, clusterContext string) []string {
	args := []string{"upgrade", opts.ReleaseName, opts.ChartName}

	// Add kubeconfig and context
	if opts.Kubeconfig != "" {
		args = append(args, "--kubeconfig", opts.Kubeconfig)
	}
	args = append(args, "--kube-context", clusterContext)

	// Add namespace
	if opts.Namespace != "" {
		args = append(args, "--namespace", opts.Namespace)
	}

	// Add chart version if specified
	if opts.ChartVersion != "" {
		args = append(args, "--version", opts.ChartVersion)
	}

	// Add values file if specified
	if opts.ValuesFile != "" {
		args = append(args, "--values", opts.ValuesFile)
	}

	// Add individual variables
	for key, value := range opts.Variables {
		args = append(args, "--set", fmt.Sprintf("%s=%s", key, value))
	}

	// Add flags
	if opts.Wait {
		args = append(args, "--wait")
	}

	if opts.Timeout != "" {
		args = append(args, "--timeout", opts.Timeout)
	}

	if opts.DryRun {
		args = append(args, "--dry-run")
	}

	if opts.Atomic {
		args = append(args, "--atomic")
	}

	return args
}

// HelmList lists Helm releases across all specified clusters
func HelmList(opts HelmOptions) error {
	if len(opts.Clusters) == 0 {
		return fmt.Errorf("no clusters specified")
	}

	for _, clusterInfo := range opts.Clusters {
		fmt.Printf("\n=== Cluster: %s ===\n", clusterInfo.Name)

		args := []string{"list"}

		// Add kubeconfig and context
		if opts.Kubeconfig != "" {
			args = append(args, "--kubeconfig", opts.Kubeconfig)
		}
		args = append(args, "--kube-context", clusterInfo.Name)

		// Add namespace or all namespaces
		if opts.AllNamespaces {
			args = append(args, "--all-namespaces")
		} else if opts.Namespace != "" {
			args = append(args, "--namespace", opts.Namespace)
		}
		
		// Add status filters
		if opts.Deployed {
			args = append(args, "--deployed")
		}
		if opts.Failed {
			args = append(args, "--failed")
		}
		if opts.Pending {
			args = append(args, "--pending")
		}

		cmd := exec.Command("helm", args...)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err != nil {
			fmt.Printf("ERROR: %s\n", stderr.String())
		} else {
			fmt.Print(stdout.String())
		}
	}

	return nil
}

// HelmUninstall uninstalls a Helm release from all specified clusters
func HelmUninstall(opts HelmOptions) error {
	if len(opts.Clusters) == 0 {
		return fmt.Errorf("no clusters specified")
	}

	successCount := 0
	failureCount := 0

	for _, clusterInfo := range opts.Clusters {
		fmt.Printf("\n--- Uninstalling Helm release from cluster: %s ---\n", clusterInfo.Name)

		args := []string{"uninstall", opts.ReleaseName}

		// Add kubeconfig and context
		if opts.Kubeconfig != "" {
			args = append(args, "--kubeconfig", opts.Kubeconfig)
		}
		args = append(args, "--kube-context", clusterInfo.Name)

		// Add namespace
		if opts.Namespace != "" {
			args = append(args, "--namespace", opts.Namespace)
		}
		
		// Add keep history
		if opts.KeepHistory {
			args = append(args, "--keep-history")
		}

		cmd := exec.Command("helm", args...)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err != nil {
			fmt.Printf("ERROR in cluster %s: %s\n", clusterInfo.Name, stderr.String())
			failureCount++
		} else {
			fmt.Printf("SUCCESS in cluster %s\n", clusterInfo.Name)
			fmt.Print(stdout.String())
			successCount++
		}
	}

	fmt.Printf("\n--- Summary ---\n")
	fmt.Printf("Successfully uninstalled from %d cluster(s)\n", successCount)
	if failureCount > 0 {
		fmt.Printf("Failed on %d cluster(s)\n", failureCount)
		return fmt.Errorf("helm uninstall failed on %d cluster(s)", failureCount)
	}

	return nil
}

// Legacy HelmUninstall function for backward compatibility
func HelmUninstallLegacy(kubeconfig string, clusters []cluster.ClusterInfo, releaseName, namespace string) error {
	opts := HelmOptions{
		Kubeconfig:  kubeconfig,
		Clusters:    clusters,
		ReleaseName: releaseName,
		Namespace:   namespace,
	}
	return HelmUninstall(opts)
}

// HelmRollback rolls back a Helm release to a previous revision
func HelmRollback(opts HelmOptions) error {
	if len(opts.Clusters) == 0 {
		return fmt.Errorf("no clusters specified")
	}

	successCount := 0
	failureCount := 0

	for _, clusterInfo := range opts.Clusters {
		fmt.Printf("\n--- Rolling back Helm release in cluster: %s ---\n", clusterInfo.Name)

		args := []string{"rollback", opts.ReleaseName}
		
		// Add revision if specified
		if opts.Revision != "" {
			args = append(args, opts.Revision)
		}

		// Add kubeconfig and context
		if opts.Kubeconfig != "" {
			args = append(args, "--kubeconfig", opts.Kubeconfig)
		}
		args = append(args, "--kube-context", clusterInfo.Name)

		// Add namespace
		if opts.Namespace != "" {
			args = append(args, "--namespace", opts.Namespace)
		}

		cmd := exec.Command("helm", args...)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err != nil {
			fmt.Printf("ERROR in cluster %s: %s\n", clusterInfo.Name, stderr.String())
			failureCount++
		} else {
			fmt.Printf("SUCCESS in cluster %s\n", clusterInfo.Name)
			fmt.Print(stdout.String())
			successCount++
		}
	}

	fmt.Printf("\n--- Summary ---\n")
	fmt.Printf("Successfully rolled back in %d cluster(s)\n", successCount)
	if failureCount > 0 {
		fmt.Printf("Failed on %d cluster(s)\n", failureCount)
		return fmt.Errorf("helm rollback failed on %d cluster(s)", failureCount)
	}

	return nil
}

// HelmStatus shows the status of a Helm release
func HelmStatus(opts HelmOptions) error {
	if len(opts.Clusters) == 0 {
		return fmt.Errorf("no clusters specified")
	}

	for _, clusterInfo := range opts.Clusters {
		fmt.Printf("\n=== Status for cluster: %s ===\n", clusterInfo.Name)

		args := []string{"status", opts.ReleaseName}

		// Add kubeconfig and context
		if opts.Kubeconfig != "" {
			args = append(args, "--kubeconfig", opts.Kubeconfig)
		}
		args = append(args, "--kube-context", clusterInfo.Name)

		// Add namespace
		if opts.Namespace != "" {
			args = append(args, "--namespace", opts.Namespace)
		}

		cmd := exec.Command("helm", args...)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err != nil {
			fmt.Printf("ERROR: %s\n", stderr.String())
		} else {
			fmt.Print(stdout.String())
		}
	}

	return nil
}

// HelmGetValues gets the values for a Helm release
func HelmGetValues(opts HelmOptions) error {
	if len(opts.Clusters) == 0 {
		return fmt.Errorf("no clusters specified")
	}

	for _, clusterInfo := range opts.Clusters {
		fmt.Printf("\n=== Values for cluster: %s ===\n", clusterInfo.Name)

		args := []string{"get", "values", opts.ReleaseName}

		// Add kubeconfig and context
		if opts.Kubeconfig != "" {
			args = append(args, "--kubeconfig", opts.Kubeconfig)
		}
		args = append(args, "--kube-context", clusterInfo.Name)

		// Add namespace
		if opts.Namespace != "" {
			args = append(args, "--namespace", opts.Namespace)
		}
		
		// Add all values flag
		if opts.AllValues {
			args = append(args, "--all")
		}
		
		// Add revision
		if opts.Revision != "" {
			args = append(args, "--revision", opts.Revision)
		}

		cmd := exec.Command("helm", args...)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err != nil {
			fmt.Printf("ERROR: %s\n", stderr.String())
		} else {
			fmt.Print(stdout.String())
		}
	}

	return nil
}


// LoadValuesFromYAML loads Helm values from a YAML configuration
func LoadValuesFromYAML(yamlContent string) (map[string]interface{}, error) {
	values := make(map[string]interface{})
	err := yaml.Unmarshal([]byte(yamlContent), &values)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML values: %v", err)
	}
	return values, nil
}

// GenerateValuesYAML generates a YAML string from values map
func GenerateValuesYAML(values map[string]interface{}) (string, error) {
	data, err := yaml.Marshal(values)
	if err != nil {
		return "", fmt.Errorf("failed to generate YAML: %v", err)
	}
	return string(data), nil
}

// ParseSetValues parses --set style key=value pairs
func ParseSetValues(setValues []string) map[string]string {
	values := make(map[string]string)
	for _, v := range setValues {
		parts := strings.SplitN(v, "=", 2)
		if len(parts) == 2 {
			values[parts[0]] = parts[1]
		}
	}
	return values
}

// HelmTemplatePreview generates and displays the rendered templates for a Helm chart
func HelmTemplatePreview(opts HelmOptions) error {
	fmt.Println("--- Helm Template Preview ---")

	args := []string{"template", opts.ReleaseName, opts.ChartName}

	// Add namespace
	if opts.Namespace != "" {
		args = append(args, "--namespace", opts.Namespace)
	}

	// Add chart version if specified
	if opts.ChartVersion != "" {
		args = append(args, "--version", opts.ChartVersion)
	}

	// Add values file if specified
	if opts.ValuesFile != "" {
		args = append(args, "--values", opts.ValuesFile)
	}

	// Add individual variables
	for key, value := range opts.Variables {
		args = append(args, "--set", fmt.Sprintf("%s=%s", key, value))
	}

	cmd := exec.Command("helm", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to generate template: %s", stderr.String())
	}

	fmt.Print(stdout.String())
	return nil
}

// HelmDependencyUpdate updates dependencies for a Helm chart
func HelmDependencyUpdate(chartPath string) error {
	fmt.Printf("Updating dependencies for chart: %s\n", chartPath)

	cmd := exec.Command("helm", "dependency", "update", chartPath)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to update dependencies: %s", stderr.String())
	}

	fmt.Print(stdout.String())
	return nil
}

// HelmRepoAdd adds a Helm repository
func HelmRepoAdd(name, url string) error {
	fmt.Printf("Adding Helm repository: %s (%s)\n", name, url)

	cmd := exec.Command("helm", "repo", "add", name, url)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to add repository: %s", stderr.String())
	}

	fmt.Print(stdout.String())

	// Update repository index
	return HelmRepoUpdate()
}

// HelmRepoUpdate updates Helm repository index
func HelmRepoUpdate() error {
	fmt.Println("Updating Helm repository index...")

	cmd := exec.Command("helm", "repo", "update")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to update repository: %s", stderr.String())
	}

	fmt.Print(stdout.String())
	return nil
}

// HelmRepoList lists configured Helm repositories
func HelmRepoList() error {
	fmt.Println("Configured Helm repositories:")

	cmd := exec.Command("helm", "repo", "list")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		// Check if it's because no repositories are configured
		if strings.Contains(stderr.String(), "no repositories") {
			fmt.Println("No repositories configured. Use 'helm repo add' to add repositories.")
			return nil
		}
		return fmt.Errorf("failed to list repositories: %s", stderr.String())
	}

	fmt.Print(stdout.String())
	return nil
}