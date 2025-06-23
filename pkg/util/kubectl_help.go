package util

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// GetKubectlHelp retrieves the original kubectl help output for a given command
func GetKubectlHelp(command string) (string, error) {
	// Construct the kubectl command
	args := []string{command, "--help"}

	// Execute kubectl command
	cmd := exec.Command("kubectl", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		// If kubectl is not found or fails, return a fallback message
		return fmt.Sprintf("Original kubectl %s help not available: %v", command, err), nil
	}

	// Return the help output
	helpOutput := stdout.String()
	if helpOutput == "" {
		helpOutput = stderr.String()
	}

	return strings.TrimSpace(helpOutput), nil
}

// GetKubectlRootHelp retrieves the original kubectl root help output
func GetKubectlRootHelp() (string, error) {
	// Execute kubectl command
	cmd := exec.Command("kubectl", "--help")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		// If kubectl is not found or fails, return a fallback message
		return fmt.Sprintf("Original kubectl help not available: %v", err), nil
	}

	// Return the help output
	helpOutput := stdout.String()
	if helpOutput == "" {
		helpOutput = stderr.String()
	}

	return strings.TrimSpace(helpOutput), nil
}

// FormatMultiClusterHelp combines the original kubectl help with multi-cluster plugin information
func FormatMultiClusterHelp(originalHelp, multiClusterInfo, multiClusterExamples, multiClusterUsage string) string {
	if originalHelp == "" {
		return multiClusterInfo
	}

	// Split the original help into sections
	sections := splitHelpSections(originalHelp)

	// Build the combined help output
	var result strings.Builder

	// Add multi-cluster information at the top
	if multiClusterInfo != "" {
		result.WriteString(multiClusterInfo)
		result.WriteString("\n\n")
	}

	// Add original kubectl help sections in order: Description, Examples, Options, Usage
	if sections["description"] != "" {
		result.WriteString("Original kubectl description:\n")
		result.WriteString(sections["description"])
		result.WriteString("\n\n")
	}

	// Add kubectl-multi examples first, then original kubectl examples
	if multiClusterExamples != "" {
		result.WriteString("Examples:\n")
		result.WriteString(multiClusterExamples)
		result.WriteString("\n\n")
	}

	if sections["examples"] != "" {
		result.WriteString("Original kubectl examples:\n")
		result.WriteString(sections["examples"])
		result.WriteString("\n\n")
	}

	if sections["options"] != "" {
		result.WriteString("Options:\n")
		// Mark flags that may not be implemented yet
		markedOptions := markUnimplementedFlags(sections["options"])
		result.WriteString(markedOptions)
		result.WriteString("\n\n")
	}

	// Add kubectl-multi usage first, then original kubectl usage
	if multiClusterUsage != "" {
		result.WriteString("Usage:\n")
		result.WriteString(multiClusterUsage)
		result.WriteString("\n\n")
	}

	if sections["usage"] != "" {
		result.WriteString("Original kubectl usage:\n")
		result.WriteString(sections["usage"])
		result.WriteString("\n\n")
	}

	return strings.TrimSpace(result.String())
}

// FormatMultiClusterRootHelp formats the root command help
func FormatMultiClusterRootHelp(originalHelp, multiClusterInfo, multiClusterExamples, multiClusterUsage string) string {
	if originalHelp == "" {
		return multiClusterInfo
	}

	// Split the original help into sections
	sections := splitHelpSections(originalHelp)

	// Build the combined help output
	var result strings.Builder

	// Add multi-cluster information at the top
	if multiClusterInfo != "" {
		result.WriteString(multiClusterInfo)
		result.WriteString("\n\n")
	}

	// Add original kubectl help sections in order: Description, Examples, Options, Usage
	if sections["description"] != "" {
		result.WriteString("Original kubectl description:\n")
		result.WriteString(sections["description"])
		result.WriteString("\n\n")
	}

	// Add kubectl-multi examples first, then original kubectl examples
	if multiClusterExamples != "" {
		result.WriteString("Examples:\n")
		result.WriteString(multiClusterExamples)
		result.WriteString("\n\n")
	}

	if sections["examples"] != "" {
		result.WriteString("Original kubectl examples:\n")
		result.WriteString(sections["examples"])
		result.WriteString("\n\n")
	}

	if sections["options"] != "" {
		result.WriteString("Options:\n")
		// Mark flags that may not be implemented yet
		markedOptions := markUnimplementedFlags(sections["options"])
		result.WriteString(markedOptions)
		result.WriteString("\n\n")
	}

	// Add kubectl-multi usage first, then original kubectl usage
	if multiClusterUsage != "" {
		result.WriteString("Usage:\n")
		result.WriteString(multiClusterUsage)
		result.WriteString("\n\n")
	}

	if sections["usage"] != "" {
		result.WriteString("Original kubectl usage:\n")
		result.WriteString(sections["usage"])
		result.WriteString("\n\n")
	}

	return strings.TrimSpace(result.String())
}

// markUnimplementedFlags adds a note about flags that may not be implemented yet
func markUnimplementedFlags(options string) string {
	// Add a note at the beginning about flag forwarding
	note := "Note: Some flags may not be fully implemented in multi-cluster mode yet.\n\n"
	return note + options
}

// splitHelpSections parses the kubectl help output into sections
func splitHelpSections(helpOutput string) map[string]string {
	sections := make(map[string]string)
	lines := strings.Split(helpOutput, "\n")

	var currentSection string
	var currentContent strings.Builder

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Detect section headers
		if strings.HasPrefix(trimmedLine, "Examples:") {
			if currentSection != "" {
				// Don't trim the content to preserve indentation
				sections[currentSection] = currentContent.String()
			}
			currentSection = "examples"
			currentContent.Reset()
			continue
		}

		if strings.HasPrefix(trimmedLine, "Options:") {
			if currentSection != "" {
				// Don't trim the content to preserve indentation
				sections[currentSection] = currentContent.String()
			}
			currentSection = "options"
			currentContent.Reset()
			continue
		}

		if strings.HasPrefix(trimmedLine, "Usage:") {
			if currentSection != "" {
				// Don't trim the content to preserve indentation
				sections[currentSection] = currentContent.String()
			}
			currentSection = "usage"
			currentContent.Reset()
			continue
		}

		// If we haven't found a section header yet, this is the description
		if currentSection == "" && trimmedLine != "" {
			currentSection = "description"
		}

		if currentSection != "" {
			currentContent.WriteString(line)
			currentContent.WriteString("\n")
		}
	}

	// Add the last section
	if currentSection != "" {
		// Don't trim the content to preserve indentation
		sections[currentSection] = currentContent.String()
	}

	return sections
}
