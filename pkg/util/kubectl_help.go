package util

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/kubectl/pkg/cmd"
)

// CommandInfo holds information about a kubectl command
type CommandInfo struct {
	Description string
	Examples    string
	Usage       string
	Options     string
}

// GetKubectlCommandInfo retrieves command information by creating kubectl commands programmatically
func GetKubectlCommandInfo(command string) (*CommandInfo, error) {
	// Create kubectl command and extract help information
	helpOutput, err := executeKubectlHelp(command)
	if err != nil {
		return nil, fmt.Errorf("failed to get kubectl %s help: %v", command, err)
	}

	// Parse the help output into sections
	sections := splitHelpSections(helpOutput)

	return &CommandInfo{
		Description: strings.TrimSpace(sections["description"]),
		Examples:    strings.TrimSpace(sections["examples"]),
		Usage:       strings.TrimSpace(sections["usage"]),
		Options:     strings.TrimSpace(sections["options"]),
	}, nil
}

// GetKubectlRootInfo retrieves root command information
func GetKubectlRootInfo() (*CommandInfo, error) {
	// Create kubectl root command and extract help information
	helpOutput, err := executeKubectlHelp("")
	if err != nil {
		return nil, fmt.Errorf("failed to get kubectl root help: %v", err)
	}

	// Parse the help output into sections
	sections := splitHelpSections(helpOutput)

	return &CommandInfo{
		Description: strings.TrimSpace(sections["description"]),
		Examples:    strings.TrimSpace(sections["examples"]),
		Usage:       strings.TrimSpace(sections["usage"]),
		Options:     strings.TrimSpace(sections["options"]),
	}, nil
}

// executeKubectlHelp creates kubectl command objects and extracts help text programmatically
func executeKubectlHelp(command string) (string, error) {
	ioStreams := genericiooptions.IOStreams{In: os.Stdin, Out: &bytes.Buffer{}, ErrOut: &bytes.Buffer{}}

	// Create kubectl options
	kubectlOptions := cmd.KubectlOptions{
		IOStreams: ioStreams,
	}

	// Create the kubectl command
	kubectlCmd := cmd.NewKubectlCommand(kubectlOptions)

	// Find the specific subcommand
	var targetCmd *cobra.Command
	if command == "" {
		targetCmd = kubectlCmd
	} else {
		targetCmd = findSubcommand(kubectlCmd, command)
		if targetCmd == nil {
			return "", fmt.Errorf("command '%s' not found", command)
		}
	}

	// Capture the help output
	var helpBuffer bytes.Buffer
	targetCmd.SetOut(&helpBuffer)
	targetCmd.SetErr(&helpBuffer)

	// Generate help text
	targetCmd.Help()

	return helpBuffer.String(), nil
}

// findSubcommand recursively searches for a subcommand by name
func findSubcommand(cmd *cobra.Command, name string) *cobra.Command {
	if cmd.Name() == name {
		return cmd
	}

	for _, subcmd := range cmd.Commands() {
		if result := findSubcommand(subcmd, name); result != nil {
			return result
		}
	}

	return nil
}

// FormatMultiClusterHelp combines the kubectl command info with multi-cluster plugin information
func FormatMultiClusterHelp(cmdInfo *CommandInfo, multiClusterInfo, multiClusterExamples, multiClusterUsage string) string {
	if cmdInfo == nil {
		return multiClusterInfo
	}

	// Build the combined help output
	var result strings.Builder

	// Add multi-cluster information at the top
	if multiClusterInfo != "" {
		result.WriteString(multiClusterInfo)
		result.WriteString("\n\n")
	}

	// Add original kubectl help sections in order: Description, Examples, Options, Usage
	if cmdInfo.Description != "" {
		result.WriteString("Original kubectl description:\n")
		result.WriteString(cmdInfo.Description)
		result.WriteString("\n\n")
	}

	// Add kubectl-multi examples first, then original kubectl examples
	if multiClusterExamples != "" {
		result.WriteString("Examples:\n")
		result.WriteString(multiClusterExamples)
		result.WriteString("\n\n")
	}

	if cmdInfo.Examples != "" {
		result.WriteString("Original kubectl examples:\n")
		result.WriteString(cmdInfo.Examples)
		result.WriteString("\n\n")
	}

	if cmdInfo.Options != "" {
		result.WriteString("Options:\n")
		// Mark flags that may not be implemented yet
		markedOptions := markUnimplementedFlags(cmdInfo.Options)
		result.WriteString(markedOptions)
		result.WriteString("\n\n")
	}

	// Add kubectl-multi usage first, then original kubectl usage
	if multiClusterUsage != "" {
		result.WriteString("Usage:\n")
		result.WriteString(multiClusterUsage)
		result.WriteString("\n\n")
	}

	if cmdInfo.Usage != "" {
		result.WriteString("Original kubectl usage:\n")
		result.WriteString(cmdInfo.Usage)
		result.WriteString("\n\n")
	}

	return strings.TrimSpace(result.String())
}

// FormatMultiClusterRootHelp formats the root command help
func FormatMultiClusterRootHelp(cmdInfo *CommandInfo, multiClusterInfo, multiClusterExamples, multiClusterUsage string) string {
	return FormatMultiClusterHelp(cmdInfo, multiClusterInfo, multiClusterExamples, multiClusterUsage)
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

// Legacy functions for backward compatibility
func GetKubectlHelp(command string) (string, error) {
	cmdInfo, err := GetKubectlCommandInfo(command)
	if err != nil {
		return fmt.Sprintf("Original kubectl %s help not available: %v", command, err), nil
	}

	var result strings.Builder
	if cmdInfo.Description != "" {
		result.WriteString(cmdInfo.Description)
		result.WriteString("\n\n")
	}
	if cmdInfo.Examples != "" {
		result.WriteString(cmdInfo.Examples)
		result.WriteString("\n\n")
	}
	if cmdInfo.Options != "" {
		result.WriteString(cmdInfo.Options)
		result.WriteString("\n\n")
	}
	if cmdInfo.Usage != "" {
		result.WriteString(cmdInfo.Usage)
	}

	return strings.TrimSpace(result.String()), nil
}

func GetKubectlRootHelp() (string, error) {
	cmdInfo, err := GetKubectlRootInfo()
	if err != nil {
		return fmt.Sprintf("Original kubectl help not available: %v", err), nil
	}

	var result strings.Builder
	if cmdInfo.Description != "" {
		result.WriteString(cmdInfo.Description)
		result.WriteString("\n\n")
	}
	if cmdInfo.Examples != "" {
		result.WriteString(cmdInfo.Examples)
		result.WriteString("\n\n")
	}
	if cmdInfo.Options != "" {
		result.WriteString(cmdInfo.Options)
		result.WriteString("\n\n")
	}
	if cmdInfo.Usage != "" {
		result.WriteString(cmdInfo.Usage)
	}

	return strings.TrimSpace(result.String()), nil
}
