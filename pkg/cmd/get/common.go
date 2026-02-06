package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

// buildKubectlGetArgs builds kubectl get command arguments
func buildKubectlGetArgs(resourceType, resourceName, outputFormat, selector, namespace string, allNamespaces bool, context string) []string {
	args := []string{"get", resourceType}

	if resourceName != "" {
		args = append(args, resourceName)
	}

	if outputFormat != "" {
		args = append(args, "-o", outputFormat)
	}

	if selector != "" {
		args = append(args, "-l", selector)
	}

	if allNamespaces {
		args = append(args, "-A")
	} else if namespace != "" {
		args = append(args, "-n", namespace)
	}

	args = append(args, "--context", context)

	return args
}

// runKubectlGet runs a kubectl command with the given args and kubeconfig, returns output and error
func runKubectlGet(args []string, kubeconfig string) (string, error) {
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

// formatLabels formats labels map to string representation
func formatLabels(labels map[string]string) string {
	if len(labels) == 0 {
		return "<none>"
	}
	result := ""
	for k, v := range labels {
		if result != "" {
			result += ","
		}
		result += fmt.Sprintf("%s=%s", k, v)
	}
	return result
}
