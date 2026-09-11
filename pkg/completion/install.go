package completion

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GetCompletionPath returns the standard user completion file path for a shell.
func GetCompletionPath(shell string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("unable to determine user home directory: %w", err)
	}

	switch strings.ToLower(strings.TrimSpace(shell)) {
	case "fish":
		return filepath.Join(home, ".config", "fish", "completions", "explain.fish"), nil
	case "bash":
		return filepath.Join(home, ".local", "share", "bash-completion", "completions", "explain"), nil
	case "zsh":
		return filepath.Join(home, ".zsh", "completions", "_explain"), nil
	default:
		return "", fmt.Errorf("unsupported shell '%s'. Supported shells: bash, zsh, fish", shell)
	}
}

// Install writes the completion script for the target shell into its standard location.
func Install(shell string) (string, error) {
	targetPath, err := GetCompletionPath(shell)
	if err != nil {
		return "", err
	}

	script, err := Generate(shell)
	if err != nil {
		return "", err
	}

	dir := filepath.Dir(targetPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	if err := os.WriteFile(targetPath, []byte(script), 0644); err != nil {
		return "", fmt.Errorf("failed to write completion script to %s: %w", targetPath, err)
	}

	return targetPath, nil
}

// InstallAll installs completion scripts for all shells detected on the system.
func InstallAll() ([]string, error) {
	var installed []string
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	// 1. Fish: install if fish is in PATH or ~/.config/fish exists
	fishConfigDir := filepath.Join(home, ".config", "fish")
	if _, err := os.Stat(fishConfigDir); err == nil {
		if path, err := Install("fish"); err == nil {
			installed = append(installed, path)
		}
	} else if _, err := os.Stat("/usr/bin/fish"); err == nil {
		if path, err := Install("fish"); err == nil {
			installed = append(installed, path)
		}
	}

	// 2. Bash: always install into ~/.local/share/bash-completion/completions
	if path, err := Install("bash"); err == nil {
		installed = append(installed, path)
	}

	// 3. Zsh: install if ~/.zsh or ~/.zshrc exists
	zshrc := filepath.Join(home, ".zshrc")
	if _, err := os.Stat(zshrc); err == nil {
		if path, err := Install("zsh"); err == nil {
			installed = append(installed, path)
		}
	}

	return installed, nil
}

// AutoEnsure quickly checks and ensures completions exist in standard user locations.
// It is non-blocking and fails silently to never disturb normal CLI execution.
func AutoEnsure() {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}

	// If fish config exists, ensure fish completion is present
	fishDir := filepath.Join(home, ".config", "fish")
	if _, err := os.Stat(fishDir); err == nil {
		fishTarget := filepath.Join(fishDir, "completions", "explain.fish")
		if _, err := os.Stat(fishTarget); os.IsNotExist(err) {
			_ = os.MkdirAll(filepath.Dir(fishTarget), 0755)
			_ = os.WriteFile(fishTarget, []byte(FishScript), 0644)
		}
	}

	// If bash completion directory or ~/.bashrc exists, ensure bash completion is present
	bashDir := filepath.Join(home, ".local", "share", "bash-completion", "completions")
	bashTarget := filepath.Join(bashDir, "explain")
	if _, err := os.Stat(bashTarget); os.IsNotExist(err) {
		_ = os.MkdirAll(bashDir, 0755)
		_ = os.WriteFile(bashTarget, []byte(BashScript), 0644)
	}
}

// UninstallCompletions removes installed completion files across all shells.
func UninstallCompletions() {
	for _, sh := range []string{"fish", "bash", "zsh"} {
		if p, err := GetCompletionPath(sh); err == nil {
			_ = os.Remove(p)
		}
	}
}
