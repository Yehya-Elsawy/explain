package ast

import (
	"bufio"
	"os"
	"strings"
)

// GetLastHistoryCommand reads the most recently executed command from shell history.
func GetLastHistoryCommand() string {
	histFile := os.Getenv("HISTFILE")
	if histFile == "" {
		home, err := os.UserHomeDir()
		if err == nil {
			// Check bash history
			if _, err := os.Stat(home + "/.bash_history"); err == nil {
				histFile = home + "/.bash_history"
			} else if _, err := os.Stat(home + "/.zsh_history"); err == nil {
				histFile = home + "/.zsh_history"
			}
		}
	}

	if histFile == "" {
		return ""
	}

	file, err := os.Open(histFile)
	if err != nil {
		return ""
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text != "" {
			// Zsh history format: ": 1620000000:0;command"
			if strings.HasPrefix(text, ": ") && strings.Contains(text, ";") {
				parts := strings.SplitN(text, ";", 2)
				if len(parts) == 2 {
					text = strings.TrimSpace(parts[1])
				}
			}
			// Skip explain command itself
			if !strings.HasPrefix(text, "explain") {
				lines = append(lines, text)
			}
		}
	}

	if len(lines) > 0 {
		return lines[len(lines)-1]
	}

	return ""
}
