package manparser

import (
	"bytes"
	"context"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// ExtractHelpFlagDescription runs `<cmd> --help` and parses the line containing the flag.
func ExtractHelpFlagDescription(cmdName, flag string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 800*time.Millisecond)
	defer cancel()

	cmd := exec.CommandContext(ctx, cmdName, "--help")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		// Try -h
		ctx2, cancel2 := context.WithTimeout(context.Background(), 800*time.Millisecond)
		defer cancel2()
		cmd2 := exec.CommandContext(ctx2, cmdName, "-h")
		cmd2.Stdout = &stdout
		if err2 := cmd2.Run(); err2 != nil {
			return ""
		}
	}

	lines := strings.Split(stdout.String(), "\n")
	escaped := regexp.QuoteMeta(flag)
	pattern := regexp.MustCompile(`(?:^|\s|,|\|)` + escaped + `(?:[=,\s\[\.\)]|$)`)

	for _, line := range lines {
		if pattern.MatchString(line) {
			trimmed := strings.TrimSpace(line)
			// Split by multiple spaces to separate flag from description
			parts := regexp.MustCompile(`\s{2,}`).Split(trimmed, -1)
			if len(parts) >= 2 {
				desc := strings.Join(parts[1:], " ")
				if len(desc) > 120 {
					return desc[:117] + "..."
				}
				return desc
			}
			return trimmed
		}
	}

	return ""
}
