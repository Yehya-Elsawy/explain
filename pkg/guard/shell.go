package guard

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Yehya-Elsawy/explain/pkg/ui"
)

const (
	MarkerStart = "# >>> explain guard >>>"
	MarkerEnd   = "# <<< explain guard <<<"
)

// DetectShell identifies the current user's shell and determines the corresponding config file.
func DetectShell() (shellName string, rcPath string, err error) {
	rawShell := os.Getenv("SHELL")
	if rawShell == "" {
		rawShell = "/bin/bash"
	}

	base := strings.ToLower(filepath.Base(rawShell))
	home, err := os.UserHomeDir()
	if err != nil {
		return "", "", fmt.Errorf("unable to determine user home directory: %w", err)
	}

	switch {
	case strings.Contains(base, "zsh"):
		shellName = "zsh"
		rcPath = filepath.Join(home, ".zshrc")
	case strings.Contains(base, "fish"):
		shellName = "fish"
		rcPath = filepath.Join(home, ".config", "fish", "config.fish")
	default:
		shellName = "bash"
		// On macOS, bash uses ~/.bash_profile by default for login shells if present
		if runtime.GOOS == "darwin" {
			profilePath := filepath.Join(home, ".bash_profile")
			if _, statErr := os.Stat(profilePath); statErr == nil {
				rcPath = profilePath
				break
			}
		}
		rcPath = filepath.Join(home, ".bashrc")
	}

	return shellName, rcPath, nil
}

// GenerateRcSnippet creates the block placed inside the user's rc file.
func GenerateRcSnippet(shellName string) string {
	if shellName == "fish" {
		return fmt.Sprintf("%s\nexplain hook fish | source\n%s\n", MarkerStart, MarkerEnd)
	}
	return fmt.Sprintf("%s\neval \"$(explain hook %s)\"\n%s\n", MarkerStart, shellName, MarkerEnd)
}

// Enable activates explain guard in the detected shell configuration file.
func Enable() error {
	shellName, rcPath, err := DetectShell()
	if err != nil {
		return err
	}

	// Read existing rc file
	contentBytes, err := os.ReadFile(rcPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read %s: %w", rcPath, err)
	}
	content := string(contentBytes)

	// Check if already present
	if strings.Contains(content, MarkerStart) {
		fmt.Println()
		fmt.Printf("  %s %s\n", ui.Colorize(ui.BoldYellow, "[>]"), ui.Colorize(ui.BoldWhite, "explain guard is already enabled"))
		fmt.Printf("      %s %s (%s)\n\n", ui.Colorize(ui.Dim, "Config file:"), ui.Colorize(ui.White, rcPath), ui.Colorize(ui.Cyan, shellName))
		return nil
	}

	// Ensure parent directory exists (e.g. ~/.config/fish)
	if err := os.MkdirAll(filepath.Dir(rcPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory for %s: %w", rcPath, err)
	}

	// Append snippet cleanly
	snippet := GenerateRcSnippet(shellName)
	newContent := content
	if len(newContent) > 0 && !strings.HasSuffix(newContent, "\n") {
		newContent += "\n"
	}
	newContent += "\n" + snippet

	if err := os.WriteFile(rcPath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write to %s: %w", rcPath, err)
	}

	fmt.Println()
	fmt.Printf("  %s %s\n", ui.Colorize(ui.BoldGreen, "[✓]"), ui.Colorize(ui.BoldGreen, "explain guard enabled successfully"))
	fmt.Printf("      %s %s (%s)\n", ui.Colorize(ui.Dim, "Target file:"), ui.Colorize(ui.BoldWhite, rcPath), ui.Colorize(ui.Cyan, shellName))
	fmt.Println()
	fmt.Printf("  %s %s\n", ui.Colorize(ui.BoldYellow, "[>]"), ui.Colorize(ui.White, "To activate protection in your current terminal session, run:"))
	fmt.Printf("      %s\n\n", ui.Colorize(ui.BoldCyan, "source "+rcPath))

	return nil
}

// Disable deactivates explain guard by removing the snippet from the shell config.
func Disable() error {
	shellName, rcPath, err := DetectShell()
	if err != nil {
		return err
	}

	contentBytes, err := os.ReadFile(rcPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println()
			fmt.Printf("  %s %s\n\n", ui.Colorize(ui.BoldYellow, "[>]"), ui.Colorize(ui.White, "explain guard is not configured (config file does not exist)."))
			return nil
		}
		return fmt.Errorf("failed to read %s: %w", rcPath, err)
	}

	content := string(contentBytes)
	startIdx := strings.Index(content, MarkerStart)
	endIdx := strings.Index(content, MarkerEnd)

	if startIdx == -1 || endIdx == -1 || endIdx < startIdx {
		fmt.Println()
		fmt.Printf("  %s %s\n", ui.Colorize(ui.BoldYellow, "[>]"), ui.Colorize(ui.White, "explain guard is not currently active in "+rcPath))
		fmt.Println()
		return nil
	}

	// Slice out the block
	before := strings.TrimRight(content[:startIdx], "\r\n")
	after := strings.TrimLeft(content[endIdx+len(MarkerEnd):], "\r\n")

	var updated string
	if before != "" && after != "" {
		updated = before + "\n\n" + after
	} else if before != "" {
		updated = before + "\n"
	} else if after != "" {
		updated = after
	}

	if err := os.WriteFile(rcPath, []byte(updated), 0644); err != nil {
		return fmt.Errorf("failed to update %s: %w", rcPath, err)
	}

	fmt.Println()
	fmt.Printf("  %s %s\n", ui.Colorize(ui.BoldGreen, "[✓]"), ui.Colorize(ui.BoldGreen, "explain guard disabled successfully"))
	fmt.Printf("      %s %s (%s)\n", ui.Colorize(ui.Dim, "Target file:"), ui.Colorize(ui.BoldWhite, rcPath), ui.Colorize(ui.Cyan, shellName))
	fmt.Println()
	fmt.Printf("  %s %s\n", ui.Colorize(ui.BoldYellow, "[>]"), ui.Colorize(ui.White, "To apply changes to this session, restart your terminal or run:"))
	fmt.Printf("      %s\n\n", ui.Colorize(ui.BoldCyan, "source "+rcPath))

	return nil
}

// Status reports the current state of explain guard on the system.
func Status() error {
	shellName, rcPath, err := DetectShell()
	if err != nil {
		return err
	}

	contentBytes, err := os.ReadFile(rcPath)
	isEnabled := false
	if err == nil && strings.Contains(string(contentBytes), MarkerStart) {
		isEnabled = true
	}

	fmt.Println()
	fmt.Printf("  %s %s\n", ui.Colorize(ui.BoldCyan, "[>]"), ui.Colorize(ui.BoldWhite, "explain guard status"))
	fmt.Printf("      %s %s\n", ui.Colorize(ui.Dim, "Active shell :"), ui.Colorize(ui.BoldCyan, shellName))
	fmt.Printf("      %s %s\n", ui.Colorize(ui.Dim, "Config file  :"), ui.Colorize(ui.White, rcPath))

	if isEnabled {
		fmt.Printf("      %s %s\n\n", ui.Colorize(ui.Dim, "Guard status :"), ui.Colorize(ui.BoldGreen, "ACTIVE (Protection Enabled)"))
	} else {
		fmt.Printf("      %s %s\n\n", ui.Colorize(ui.Dim, "Guard status :"), ui.Colorize(ui.BoldYellow, "INACTIVE (Protection Disabled)"))
		fmt.Printf("  %s %s\n", ui.Colorize(ui.Dim, "To activate, run:"), ui.Colorize(ui.BoldCyan, "explain guard enable"))
		fmt.Println()
	}

	return nil
}
