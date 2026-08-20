package manparser

import (
	"bytes"
	"context"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// StripOverstrike removes terminal bolding/underlining artifacts (e.g. "_\bc" or "c\bc").
func StripOverstrike(s string) string {
	var buf bytes.Buffer
	runes := []rune(s)
	n := len(runes)
	i := 0
	for i < n {
		if i+1 < n && runes[i+1] == '\b' {
			// Skip character and backspace, read the next one
			i += 2
			continue
		}
		buf.WriteRune(runes[i])
		i++
	}
	return buf.String()
}

// ExtractCommandSummary parses the NAME section of a man page.
func ExtractCommandSummary(cmdName string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "man", "-P", "cat", cmdName)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return ""
	}

	clean := StripOverstrike(stdout.String())
	lines := strings.Split(clean, "\n")

	inName := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "NAME" {
			inName = true
			continue
		}
		if inName {
			if strings.HasPrefix(line, "SYNOPSIS") || strings.HasPrefix(line, "DESCRIPTION") || (len(line) > 0 && line[0] != ' ' && line[0] != '\t') {
				break
			}
			if strings.Contains(trimmed, " - ") || strings.Contains(trimmed, " — ") {
				parts := strings.SplitN(trimmed, " - ", 2)
				if len(parts) < 2 {
					parts = strings.SplitN(trimmed, " — ", 2)
				}
				if len(parts) == 2 {
					return strings.TrimSpace(parts[1])
				}
			}
		}
	}
	return ""
}

// ExtractFlagDescription parses the man page to find description for a given flag (e.g. "-v" or "--version").
func ExtractFlagDescription(cmdName, flag string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "man", "-P", "cat", cmdName)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return ""
	}

	clean := StripOverstrike(stdout.String())
	lines := strings.Split(clean, "\n")

	escaped := regexp.QuoteMeta(flag)
	pattern := regexp.MustCompile(`^\s+.*(?:^|\s|,|/|\|)` + escaped + `(?:[=,\s\[\.\)]|$)`)

	found := false
	var descLines []string

	for _, line := range lines {
		if !found {
			if pattern.MatchString(line) {
				found = true
				continue
			}
		} else {
			if len(line) > 0 && line[0] != ' ' && line[0] != '\t' {
				break
			}
			trimmed := strings.TrimSpace(line)
			if trimmed == "" {
				if len(descLines) > 0 {
					break
				}
				continue
			}
			if regexp.MustCompile(`^\s+-[a-zA-Z0-9-]`).MatchString(line) {
				break
			}

			descLines = append(descLines, trimmed)
			if len(descLines) >= 3 {
				break
			}
		}
	}

	if len(descLines) > 0 {
		fullDesc := strings.Join(descLines, " ")
		if idx := strings.Index(fullDesc, ". "); idx != -1 && idx < 120 {
			return fullDesc[:idx+1]
		}
		if len(fullDesc) > 120 {
			return fullDesc[:117] + "..."
		}
		return fullDesc
	}

	return ""
}
