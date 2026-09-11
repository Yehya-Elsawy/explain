package guard

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/Yehya-Elsawy/explain/pkg/completion"
	"github.com/Yehya-Elsawy/explain/pkg/ui"
)

const (
	MarkerStart = "# >>> explain guard >>>"
	MarkerEnd   = "# <<< explain guard <<<"
)

// detectRunningShell inspects the process hierarchy to find the actual shell currently running.
func detectRunningShell() string {
	pid := os.Getppid()
	for i := 0; i < 6 && pid > 1; i++ {
		// Linux: read /proc/<pid>/comm
		commBytes, err := os.ReadFile(fmt.Sprintf("/proc/%d/comm", pid))
		if err == nil {
			comm := strings.ToLower(strings.TrimSpace(string(commBytes)))
			switch {
			case strings.Contains(comm, "bash"):
				return "bash"
			case strings.Contains(comm, "zsh"):
				return "zsh"
			case strings.Contains(comm, "fish"):
				return "fish"
			}
		}

		// macOS fallback
		if runtime.GOOS == "darwin" {
			out, err := exec.Command("ps", "-p", fmt.Sprintf("%d", pid), "-o", "comm=").Output()
			if err == nil {
				comm := strings.ToLower(strings.TrimSpace(string(out)))
				switch {
				case strings.Contains(comm, "bash"):
					return "bash"
				case strings.Contains(comm, "zsh"):
					return "zsh"
				case strings.Contains(comm, "fish"):
					return "fish"
				}
			}
		}

		// Read parent PID from /proc/<pid>/stat
		statBytes, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
		if err != nil {
			break
		}
		fields := strings.Fields(string(statBytes))
		if len(fields) < 4 {
			break
		}
		parentPid, err := strconv.Atoi(fields[3])
		if err != nil || parentPid <= 1 || parentPid == pid {
			break
		}
		pid = parentPid
	}
	return ""
}

// DetectShell identifies the user's shell and determines the corresponding config file.
// If an explicit shell name is provided, it configures that specific shell.
func DetectShell(requested ...string) (shellName string, rcPath string, err error) {
	if len(requested) > 0 && strings.TrimSpace(requested[0]) != "" {
		shellName = strings.ToLower(strings.TrimSpace(requested[0]))
	} else {
		shellName = detectRunningShell()
		if shellName == "" {
			rawShell := os.Getenv("SHELL")
			if rawShell == "" {
				rawShell = "/bin/bash"
			}
			base := strings.ToLower(filepath.Base(rawShell))
			switch {
			case strings.Contains(base, "zsh"):
				shellName = "zsh"
			case strings.Contains(base, "fish"):
				shellName = "fish"
			default:
				shellName = "bash"
			}
		}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", "", fmt.Errorf("unable to determine user home directory: %w", err)
	}

	switch shellName {
	case "zsh":
		rcPath = filepath.Join(home, ".zshrc")
	case "fish":
		rcPath = filepath.Join(home, ".config", "fish", "config.fish")
	case "bash":
		if runtime.GOOS == "darwin" {
			profilePath := filepath.Join(home, ".bash_profile")
			if _, statErr := os.Stat(profilePath); statErr == nil {
				rcPath = profilePath
				break
			}
		}
		rcPath = filepath.Join(home, ".bashrc")
	default:
		return "", "", fmt.Errorf("unsupported shell '%s'. Supported shells: bash, zsh, fish", shellName)
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

// Enable activates explain guard in the target shell configuration file.
func Enable(requested ...string) error {
	shellName, rcPath, err := DetectShell(requested...)
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

	// Ensure shell completions are installed
	_, _ = completion.InstallAll()

	return nil
}

// Disable deactivates explain guard by removing the snippet from the shell config.
func Disable(requested ...string) error {
	shellName, rcPath, err := DetectShell(requested...)
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
func Status(requested ...string) error {
	shellName, rcPath, err := DetectShell(requested...)
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
