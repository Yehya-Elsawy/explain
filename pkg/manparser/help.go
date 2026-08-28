package manparser

import (
	"context"
	"os/exec"
	"time"
)

// ExtractHelpFlagDescription runs `<cmd> --help` and parses the line containing the flag.
func ExtractHelpFlagDescription(cmdName, flag string) string {
	return ExtractHelpFlagInfo(cmdName, flag).Description
}

// ExtractHelpFlagInfo runs a command's help mode and parses flag metadata.
// Some commands print help to stderr and return a non-zero exit status, so the
// parser intentionally accepts combined output whenever it contains useful data.
func ExtractHelpFlagInfo(cmdName, flag string) FlagInfo {
	for _, helpArg := range []string{"--help", "-h"} {
		ctx, cancel := context.WithTimeout(context.Background(), 800*time.Millisecond)
		output, _ := exec.CommandContext(ctx, cmdName, helpArg).CombinedOutput()
		cancel()
		if len(output) == 0 {
			continue
		}
		if info := ParseFlagInfo(string(output), flag); info.Description != "" || info.TakesValue {
			return info
		}
	}
	return FlagInfo{}
}
