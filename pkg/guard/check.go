package guard

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Yehya-Elsawy/explain/pkg/analyzer"
	"github.com/Yehya-Elsawy/explain/pkg/ast"
	"github.com/Yehya-Elsawy/explain/pkg/database"
	"github.com/Yehya-Elsawy/explain/pkg/ui"
)

func isLethalPath(arg string) bool {
	arg = strings.TrimSpace(arg)
	if arg == "" {
		return false
	}

	if arg == "/" || arg == "/*" || arg == "~" || arg == "~/*" || arg == "$HOME" || arg == "$HOME/*" {
		return true
	}

	if strings.HasSuffix(arg, "/*") {
		base := strings.TrimSuffix(arg, "/*")
		if isLethalDir(filepath.Clean(base)) {
			return true
		}
	}

	cleaned := filepath.Clean(arg)
	return isLethalDir(cleaned)
}

func isLethalDir(dir string) bool {
	if dir == "/" {
		return true
	}
	lethalDirs := []string{
		"/etc",
		"/boot",
		"/usr",
		"/bin",
		"/sbin",
		"/lib",
		"/lib64",
		"/var",
		"/dev",
		"/sys",
		"/proc",
		"/root",
	}
	for _, d := range lethalDirs {
		if dir == d {
			return true
		}
	}
	return false
}

func isLethalSuicideCommand(rawCmd string, pipeline *ast.Pipeline) bool {
	if strings.Contains(rawCmd, ":(){ :|:& };:") || strings.Contains(rawCmd, ":(){:|:&};:") {
		return true
	}

	for _, cmd := range pipeline.Commands {
		baseName := filepath.Base(cmd.Name)
		argsJoined := strings.Join(cmd.Args, " ")

		if baseName == "bash" || baseName == "sh" || baseName == "zsh" || baseName == "dash" || baseName == "ksh" {
			for i, arg := range cmd.Args {
				if arg == "-c" && i+1 < len(cmd.Args) {
					subPipe, err := ast.Parse(cmd.Args[i+1])
					if err == nil && isLethalSuicideCommand(cmd.Args[i+1], subPipe) {
						return true
					}
				}
			}
		}
		if baseName == "eval" && len(cmd.Args) > 0 {
			subPipe, err := ast.Parse(argsJoined)
			if err == nil && isLethalSuicideCommand(argsJoined, subPipe) {
				return true
			}
		}

		if baseName == "rm" {
			hasRec := false
			for _, arg := range cmd.Args {
				if arg == "-r" || arg == "-R" || arg == "--recursive" || (strings.HasPrefix(arg, "-") && (strings.Contains(arg, "r") || strings.Contains(arg, "R"))) {
					hasRec = true
					break
				}
			}

			if hasRec {
				for _, arg := range cmd.Args {
					if isLethalPath(arg) {
						return true
					}
				}
				if strings.Contains(argsJoined, "--no-preserve-root") {
					return true
				}
			}
		}

		if baseName == "chmod" {
			hasRec := strings.Contains(argsJoined, "-R") || strings.Contains(argsJoined, "--recursive")
			has777 := strings.Contains(argsJoined, "777") || strings.Contains(argsJoined, "a+rwx")
			if hasRec && has777 {
				for _, arg := range cmd.Args {
					if isLethalPath(arg) {
						return true
					}
				}
			}
		}
	}

	return false
}

func extractLethalTarget(pipeline *ast.Pipeline, rawCmd string) string {
	if strings.Contains(rawCmd, ":(){ :|:& };:") || strings.Contains(rawCmd, ":(){:|:&};:") {
		return "operating system process table"
	}
	for _, cmd := range pipeline.Commands {
		baseName := filepath.Base(cmd.Name)
		if baseName == "rm" || baseName == "chmod" {
			for _, arg := range cmd.Args {
				if isLethalPath(arg) {
					return arg
				}
			}
		}
	}
	return "root filesystem (/)"
}

func extractTargetResource(cmd *analyzer.CommandAnalysis, rawCmd string) string {
	if cmd == nil {
		return ""
	}

	baseName := filepath.Base(cmd.CommandName)

	for _, r := range cmd.Redirects {
		if strings.HasPrefix(r.Target, "/dev/") {
			return r.Target
		}
	}

	switch baseName {
	case "rm":
		if len(cmd.PositionalArgs) > 0 {
			return strings.Join(cmd.PositionalArgs, " ")
		}
	case "dd":
		for _, item := range cmd.Items {
			if strings.HasPrefix(item.Token, "of=") {
				return strings.TrimPrefix(item.Token, "of=")
			}
		}
		if len(cmd.PositionalArgs) > 0 {
			return strings.Join(cmd.PositionalArgs, " ")
		}
	case "fdisk", "parted", "gdisk", "sfdisk", "sgdisk", "wipefs", "shred":
		if len(cmd.PositionalArgs) > 0 {
			return strings.Join(cmd.PositionalArgs, " ")
		}
	case "chmod", "chown":
		if len(cmd.PositionalArgs) > 1 {
			return strings.Join(cmd.PositionalArgs[1:], " ")
		} else if len(cmd.PositionalArgs) == 1 {
			return cmd.PositionalArgs[0]
		}
	case "kill", "pkill", "killall":
		if len(cmd.PositionalArgs) > 0 {
			return fmt.Sprintf("PID %s", strings.Join(cmd.PositionalArgs, " "))
		}
	}

	if strings.HasPrefix(baseName, "mkfs") {
		if len(cmd.PositionalArgs) > 0 {
			return strings.Join(cmd.PositionalArgs, " ")
		}
	}

	if baseName == "bash" || baseName == "sh" || baseName == "zsh" {
		return "shell interpreter (remote script execution)"
	}

	if len(cmd.PositionalArgs) > 0 {
		return strings.Join(cmd.PositionalArgs, " ")
	}

	return cmd.CommandName
}

func Check(rawCmd string) int {
	rawCmd = strings.TrimSpace(rawCmd)
	if rawCmd == "" {
		return 0
	}

	if strings.Contains(rawCmd, ":(){ :|:& };:") || strings.Contains(rawCmd, ":(){:|:&};:") {
		ui.InitColors(false)
		fmt.Println()
		fmt.Printf("  %s %s\n\n", ui.Colorize(ui.BoldRed, "[X]"), ui.Colorize(ui.BoldRed, "explain guard: execution permanently blocked"))
		fmt.Printf("      %s %s\n", ui.Colorize(ui.Dim, "Command :"), ui.Colorize(ui.BoldWhite, rawCmd))
		fmt.Printf("      %s %s\n", ui.Colorize(ui.Dim, "Target  :"), ui.Colorize(ui.Cyan, "operating system process table"))
		fmt.Printf("      %s %s\n", ui.Colorize(ui.Dim, "Risk    :"), ui.Colorize(ui.BoldRed, "[ SYSTEM FREEZE HAZARD ]"))
		fmt.Printf("      %s %s\n", ui.Colorize(ui.Dim, "Action  :"), ui.Colorize(ui.BoldYellow, "This command exhausts process slots and crashes the operating system."))
		fmt.Println()
		fmt.Printf("  %s %s\n\n", ui.Colorize(ui.BoldRed, "[!]"), ui.Colorize(ui.White, "explain guard will not execute this command under any circumstances."))
		return 1
	}

	pipeline, err := ast.Parse(rawCmd)
	if err != nil {
		return 0
	}

	if isLethalSuicideCommand(rawCmd, pipeline) {
		ui.InitColors(false)
		lethalTarget := extractLethalTarget(pipeline, rawCmd)
		analysis := analyzer.AnalyzePipeline(pipeline)

		var flagLines []string
		for _, cmd := range analysis.Commands {
			for _, item := range cmd.Items {
				if item.IsFlag && item.Description != "" {
					flagLines = append(flagLines, fmt.Sprintf("%s → %s", ui.Colorize(ui.BoldWhite, item.Token), ui.Colorize(ui.White, item.Description)))
				}
			}
		}

		fmt.Println()
		fmt.Printf("  %s %s\n\n", ui.Colorize(ui.BoldRed, "[X]"), ui.Colorize(ui.BoldRed, "explain guard: execution permanently blocked"))
		fmt.Printf("      %s %s\n", ui.Colorize(ui.Dim, "Command :"), ui.Colorize(ui.BoldWhite, rawCmd))
		fmt.Printf("      %s %s\n", ui.Colorize(ui.Dim, "Target  :"), ui.Colorize(ui.Cyan, lethalTarget))
		if len(flagLines) > 0 {
			for i, line := range flagLines {
				if i == 0 {
					fmt.Printf("      %s %s\n", ui.Colorize(ui.Dim, "Options :"), line)
				} else {
					fmt.Printf("                %s\n", line)
				}
			}
		}
		fmt.Printf("      %s %s\n", ui.Colorize(ui.Dim, "Risk    :"), ui.Colorize(ui.BoldRed, "[ LETHAL SYSTEM DESTRUCTION ]"))
		fmt.Printf("      %s %s\n", ui.Colorize(ui.Dim, "Action  :"), ui.Colorize(ui.BoldYellow, "This command destroys the operating system with zero legitimate use cases."))
		fmt.Println()
		fmt.Printf("  %s %s\n\n", ui.Colorize(ui.BoldRed, "[!]"), ui.Colorize(ui.White, "explain guard will not execute this command under any circumstances."))
		return 1
	}

	for _, cmd := range pipeline.Commands {
		baseName := filepath.Base(cmd.Name)
		if baseName == "bash" || baseName == "sh" || baseName == "zsh" || baseName == "dash" || baseName == "ksh" {
			for i, arg := range cmd.Args {
				if arg == "-c" && i+1 < len(cmd.Args) {
					innerRes := Check(cmd.Args[i+1])
					if innerRes != 0 {
						return innerRes
					}
				}
			}
		} else if baseName == "eval" && len(cmd.Args) > 0 {
			innerRes := Check(strings.Join(cmd.Args, " "))
			if innerRes != 0 {
				return innerRes
			}
		}
	}

	analysis := analyzer.AnalyzePipeline(pipeline)
	if analysis.MaxRisk != database.RiskCritical {
		return 0
	}

	var critDanger analyzer.DangerInfo
	var critTarget string
	var critCmd *analyzer.CommandAnalysis
	for _, cmd := range analysis.Commands {
		if cmd.Danger.Level == database.RiskCritical {
			critDanger = cmd.Danger
			critTarget = extractTargetResource(cmd, rawCmd)
			critCmd = cmd
			break
		}
	}
	if critTarget == "" {
		critTarget = rawCmd
	}

	var flagLines []string
	if critCmd != nil {
		for _, item := range critCmd.Items {
			if item.IsFlag && item.Description != "" {
				flagLines = append(flagLines, fmt.Sprintf("%s → %s", ui.Colorize(ui.BoldWhite, item.Token), ui.Colorize(ui.White, item.Description)))
			}
		}
	}

	badge := critDanger.Badge
	badge = strings.TrimSpace(badge)
	badge = strings.TrimPrefix(badge, "🚨 ")
	badge = strings.TrimPrefix(badge, "⚠️ ")
	badge = strings.TrimPrefix(badge, "🟡 ")
	badge = strings.TrimPrefix(badge, "🟢 ")
	if badge == "" {
		badge = "CRITICAL DANGER"
	}

	ui.InitColors(false)

	fmt.Println()
	fmt.Printf("  %s %s\n\n", ui.Colorize(ui.BoldRed, "[!]"), ui.Colorize(ui.BoldRed, "explain guard: critical risk operation detected"))
	fmt.Printf("      %s %s\n", ui.Colorize(ui.Dim, "Command :"), ui.Colorize(ui.BoldWhite, rawCmd))
	fmt.Printf("      %s %s\n", ui.Colorize(ui.Dim, "Target  :"), ui.Colorize(ui.Cyan, critTarget))
	if len(flagLines) > 0 {
		for i, line := range flagLines {
			if i == 0 {
				fmt.Printf("      %s %s\n", ui.Colorize(ui.Dim, "Options :"), line)
			} else {
				fmt.Printf("                %s\n", line)
			}
		}
	}
	fmt.Printf("      %s %s\n", ui.Colorize(ui.Dim, "Risk    :"), ui.Colorize(ui.BoldRed, "[ "+badge+" ]"))

	if critDanger.Reason != "" {
		fmt.Printf("      %s %s\n", ui.Colorize(ui.Dim, "Impact  :"), ui.Colorize(ui.White, critDanger.Reason))
	}
	if critDanger.Warning != "" {
		fmt.Printf("      %s %s\n", ui.Colorize(ui.Dim, "Warning :"), ui.Colorize(ui.BoldYellow, critDanger.Warning))
	}

	fmt.Println()
	fmt.Printf("  %s %s", ui.Colorize(ui.BoldYellow, "[?]"), ui.Colorize(ui.BoldWhite, "To proceed, type 'CONFIRM' (or press Enter to cancel): "))

	var reader *bufio.Reader
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err == nil {
		defer tty.Close()
		reader = bufio.NewReader(tty)
	} else {
		reader = bufio.NewReader(os.Stdin)
	}

	resp, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println()
		return 1
	}

	resp = strings.TrimSpace(resp)
	if resp == "CONFIRM" {
		fmt.Printf("\n  %s %s\n\n", ui.Colorize(ui.BoldGreen, "[✓]"), ui.Colorize(ui.BoldGreen, "Confirmation accepted. Proceeding with execution..."))
		return 0
	}

	fmt.Printf("\n  %s %s\n\n", ui.Colorize(ui.BoldRed, "[X]"), ui.Colorize(ui.BoldRed, "Command execution cancelled by explain guard."))
	return 1
}
