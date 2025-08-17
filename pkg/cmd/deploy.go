package cmd

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

type InstallOptions struct {
	genericclioptions.IOStreams

	ReleaseName string
	Namespace   string
	Version     string
	ChartPath   string

	// KubeFlex options
	InstallKubeFlex   bool
	InstallPostgreSQL bool
	IsOpenShift       bool
	Domain            string
	ExternalPort      int
	HostContainer     string
	ClusterName       string

	// Control Plane options
	ITSes []string
	WDSes []string

	// Installation options
	InstallPCHs bool
	DryRun      bool
	Wait        bool
	Timeout     string
	Verbosity   int
}

func NewInstallOptions(streams genericclioptions.IOStreams) *InstallOptions {
	return &InstallOptions{
		IOStreams:         streams,
		ReleaseName:       "ks-core",
		Namespace:         "default",
		Version:           "",
		InstallKubeFlex:   true,
		InstallPostgreSQL: true,
		IsOpenShift:       false,
		Domain:            "localtest.me",
		ExternalPort:      9443,
		HostContainer:     "kubeflex-control-plane",
		ClusterName:       "kubeflex",
		InstallPCHs:       true,
		DryRun:            false,
		Wait:              true,
		Timeout:           "10m",
		Verbosity:         2,
	}
}

func NewInstallCmd(streams genericclioptions.IOStreams) *cobra.Command {
	o := NewInstallOptions(streams)

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install KubeStellar core components using Helm chart",
		Long: `Install KubeStellar core components using the official Helm chart.

This command simplifies the installation of KubeStellar by providing a more
user-friendly interface to the underlying Helm chart. It can install KubeFlex,
create ITSes (Inventory and Transport Spaces), and WDSes (Workload Description Spaces).

Examples:
  # Basic installation with KubeFlex only
  kubectl ks install
  
  # Install with one ITS and one WDS
  kubectl ks install --its its1 --wds wds1
  
  # Install with custom cluster name
  kubectl ks install --cluster-name my-cluster --its its1 --wds wds1
  
  # Install for OpenShift
  kubectl ks install --openshift
  
  # Install with custom domain and port (for Kind clusters)
  kubectl ks install --domain example.com --port 8443 --cluster-name my-kind-cluster
  
  # Install specific version
  kubectl ks install --version v0.28.0
  
  # Dry run to see what would be installed
  kubectl ks install --dry-run --its its1 --wds wds1`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := o.Validate(); err != nil {
				return err
			}
			return o.Run(cmd.Context())
		},
	}

	// Core flags
	cmd.Flags().StringVar(&o.ReleaseName, "release-name", o.ReleaseName, "Helm release name")
	cmd.Flags().StringVarP(&o.Namespace, "namespace", "n", o.Namespace, "Kubernetes namespace for installation")
	cmd.Flags().StringVar(&o.Version, "version", o.Version, "KubeStellar version to install (defaults to latest)")
	cmd.Flags().StringVar(&o.ChartPath, "chart-path", o.ChartPath, "Local path to chart (for development)")

	// KubeFlex flags
	cmd.Flags().BoolVar(&o.InstallKubeFlex, "install-kubeflex", o.InstallKubeFlex, "Install KubeFlex operator")
	cmd.Flags().BoolVar(&o.InstallPostgreSQL, "install-postgresql", o.InstallPostgreSQL, "Install PostgreSQL dependency")
	cmd.Flags().BoolVar(&o.IsOpenShift, "openshift", o.IsOpenShift, "Set to true when installing on OpenShift")
	cmd.Flags().StringVar(&o.Domain, "domain", o.Domain, "DNS domain for accessing control planes")
	cmd.Flags().IntVar(&o.ExternalPort, "port", o.ExternalPort, "External port for accessing control planes")
	cmd.Flags().StringVar(&o.HostContainer, "host-container", o.HostContainer, "Name of the container running the hosting cluster")
	cmd.Flags().StringVar(&o.ClusterName, "cluster-name", o.ClusterName, "Name of the Kind/k3s cluster (auto-sets host-container)")

	// Control Plane flags
	cmd.Flags().StringSliceVar(&o.ITSes, "its", []string{}, "Create ITS control planes (can be specified multiple times)")
	cmd.Flags().StringSliceVar(&o.WDSes, "wds", []string{}, "Create WDS control planes (can be specified multiple times)")

	// Installation flags
	cmd.Flags().BoolVar(&o.InstallPCHs, "install-pchs", o.InstallPCHs, "Install Post Create Hooks")
	cmd.Flags().BoolVar(&o.DryRun, "dry-run", o.DryRun, "Show what would be installed without actually installing")
	cmd.Flags().BoolVar(&o.Wait, "wait", o.Wait, "Wait for installation to complete")
	cmd.Flags().StringVar(&o.Timeout, "timeout", o.Timeout, "Timeout for installation")

	// Verbosity
	cmd.Flags().IntVar(&o.Verbosity, "verbosity", o.Verbosity, "Controller log verbosity level")

	return cmd
}

func (o *InstallOptions) Validate() error {
	// Check if helm is available
	if _, err := exec.LookPath("helm"); err != nil {
		return fmt.Errorf("helm is not installed or not in PATH: %w", err)
	}

	// Validate port range
	if o.ExternalPort < 1 || o.ExternalPort > 65535 {
		return fmt.Errorf("external port must be between 1 and 65535, got %d", o.ExternalPort)
	}

	// Validate verbosity
	if o.Verbosity < 0 || o.Verbosity > 10 {
		return fmt.Errorf("verbosity must be between 0 and 10, got %d", o.Verbosity)
	}

	return nil
}

func (o *InstallOptions) Run(ctx context.Context) error {
	fmt.Fprintf(o.Out, "Installing KubeStellar core components...\n")

	// Auto-set host container if cluster name is provided
	if o.ClusterName != "" && o.HostContainer == "kubeflex-control-plane" {
		o.HostContainer = o.ClusterName + "-control-plane"
		fmt.Fprintf(o.Out, "Auto-setting host-container to: %s\n", o.HostContainer)
	}

	args := o.buildHelmArgs()

	if o.DryRun {
		fmt.Fprintf(o.Out, "Dry run - would execute: helm %s\n", strings.Join(args, " "))
		return nil
	}

	if o.ChartPath != "" {
		if err := o.updateHelmDependencies(ctx); err != nil {
			return fmt.Errorf("failed to update helm dependencies: %w", err)
		}
	}

	cmd := exec.CommandContext(ctx, "helm", args...)
	cmd.Stdout = o.Out
	cmd.Stderr = o.ErrOut
	cmd.Stdin = o.In

	fmt.Fprintf(o.Out, "Executing: helm %s\n", strings.Join(args, " "))

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("helm command failed: %w", err)
	}

	fmt.Fprintf(o.Out, "\n✅ KubeStellar core installation completed successfully!\n")

	o.printPostInstallInstructions()

	return nil
}

func (o *InstallOptions) buildHelmArgs() []string {
	args := []string{"upgrade", "--install"}

	// Add basic flags
	args = append(args, o.ReleaseName)

	// Chart source
	if o.ChartPath != "" {
		args = append(args, o.ChartPath)
	} else {
		chartURL := "oci://ghcr.io/kubestellar/kubestellar/core-chart"
		args = append(args, chartURL)
		if o.Version != "" {
			args = append(args, "--version", o.Version)
		}
	}

	if o.Namespace != "default" {
		args = append(args, "--namespace", o.Namespace, "--create-namespace")
	}

	if o.Wait && len(o.ITSes) == 0 && len(o.WDSes) == 0 {
		args = append(args, "--wait")
		if o.Timeout != "" {
			args = append(args, "--timeout", o.Timeout)
		}
	}

	values := o.buildHelmValues()
	for key, value := range values {
		args = append(args, "--set", fmt.Sprintf("%s=%s", key, value))
	}

	jsonValues := o.buildHelmJSONValues()
	for key, value := range jsonValues {
		args = append(args, "--set-json", fmt.Sprintf("%s=%s", key, value))
	}

	return args
}

func (o *InstallOptions) buildHelmValues() map[string]string {
	values := make(map[string]string)

	if !o.InstallKubeFlex {
		values["kubeflex-operator.install"] = "false"
	}

	if !o.InstallPostgreSQL {
		values["kubeflex-operator.installPostgreSQL"] = "false"
	}

	if o.IsOpenShift {
		values["kubeflex-operator.isOpenShift"] = "true"
	}

	if o.Domain != "localtest.me" {
		values["kubeflex-operator.domain"] = o.Domain
	}

	if o.ExternalPort != 9443 {
		values["kubeflex-operator.externalPort"] = fmt.Sprintf("%d", o.ExternalPort)
	}

	if o.ClusterName != "kubeflex" && o.HostContainer == o.ClusterName+"-control-plane" {
		values["kubeflex-operator.hostContainer"] = o.HostContainer
	} else if o.HostContainer != "kubeflex-control-plane" && o.HostContainer != o.ClusterName+"-control-plane" {
		values["kubeflex-operator.hostContainer"] = o.HostContainer
	}

	if !o.InstallPCHs {
		values["InstallPCHs"] = "false"
	}

	if o.Verbosity != 2 {
		values["verbosity.default"] = fmt.Sprintf("%d", o.Verbosity)
	}

	return values
}

func (o *InstallOptions) buildHelmJSONValues() map[string]string {
	jsonValues := make(map[string]string)

	// Build ITSes JSON
	if len(o.ITSes) > 0 {
		var itsesJSON strings.Builder
		itsesJSON.WriteString("[")
		for i, its := range o.ITSes {
			if i > 0 {
				itsesJSON.WriteString(",")
			}
			itsesJSON.WriteString(fmt.Sprintf(`{"name":"%s"}`, its))
		}
		itsesJSON.WriteString("]")
		jsonValues["ITSes"] = itsesJSON.String()
	}

	if len(o.WDSes) > 0 {
		var wdsesJSON strings.Builder
		wdsesJSON.WriteString("[")
		for i, wds := range o.WDSes {
			if i > 0 {
				wdsesJSON.WriteString(",")
			}
			itsName := ""
			if len(o.ITSes) > 0 {
				itsName = o.ITSes[0]
			}
			if itsName != "" {
				wdsesJSON.WriteString(fmt.Sprintf(`{"name":"%s","ITSName":"%s"}`, wds, itsName))
			} else {
				wdsesJSON.WriteString(fmt.Sprintf(`{"name":"%s"}`, wds))
			}
		}
		wdsesJSON.WriteString("]")
		jsonValues["WDSes"] = wdsesJSON.String()
	}

	return jsonValues
}

func (o *InstallOptions) updateHelmDependencies(ctx context.Context) error {
	fmt.Fprintf(o.Out, "Updating helm dependencies...\n")

	cmd := exec.CommandContext(ctx, "helm", "dependency", "update", o.ChartPath)
	cmd.Stdout = o.Out
	cmd.Stderr = o.ErrOut

	return cmd.Run()
}

func (o *InstallOptions) printPostInstallInstructions() {
	fmt.Fprintf(o.Out, "\n📋 Next Steps:\n")

	if len(o.ITSes) > 0 || len(o.WDSes) > 0 {
		fmt.Fprintf(o.Out, "\n1. Add kubeconfig contexts for your control planes:\n")

		// Show kflex method
		fmt.Fprintf(o.Out, "   Using kflex (recommended):\n")
		fmt.Fprintf(o.Out, "   kubectl config use-context <your-hosting-cluster-context>\n")
		fmt.Fprintf(o.Out, "   kflex ctx --set-current-for-hosting\n")

		for _, its := range o.ITSes {
			fmt.Fprintf(o.Out, "   kflex ctx --overwrite-existing-context %s\n", its)
		}
		for _, wds := range o.WDSes {
			fmt.Fprintf(o.Out, "   kflex ctx --overwrite-existing-context %s\n", wds)
		}

		fmt.Fprintf(o.Out, "\n   Or using the import script:\n")
		if o.Version != "" {
			fmt.Fprintf(o.Out, "   bash <(curl -s https://raw.githubusercontent.com/kubestellar/kubestellar/v%s/scripts/import-cp-contexts.sh) --merge\n", o.Version)
		} else {
			fmt.Fprintf(o.Out, "   bash <(curl -s https://raw.githubusercontent.com/kubestellar/kubestellar/main/scripts/import-cp-contexts.sh) --merge\n")
		}

		fmt.Fprintf(o.Out, "\n2. Verify your control planes are ready:\n")
		fmt.Fprintf(o.Out, "   kubectl get controlplane\n")

		fmt.Fprintf(o.Out, "\n3. Start using your control planes:\n")
		for _, its := range o.ITSes {
			fmt.Fprintf(o.Out, "   kubectl --context %s get managedclusters\n", its)
		}
		for _, wds := range o.WDSes {
			fmt.Fprintf(o.Out, "   kubectl --context %s get deployments\n", wds)
		}
	} else {
		fmt.Fprintf(o.Out, "\n1. Create control planes using additional helm commands or:\n")
		fmt.Fprintf(o.Out, "   kubectl ks install --its its1 --wds wds1\n")
	}

	fmt.Fprintf(o.Out, "\n📖 For more information, visit: https://docs.kubestellar.io/\n")
}
