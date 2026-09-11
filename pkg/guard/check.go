package guard

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/Yehya-Elsawy/explain/pkg/analyzer"
	"github.com/Yehya-Elsawy/explain/pkg/ast"
	"github.com/Yehya-Elsawy/explain/pkg/database"
	"github.com/Yehya-Elsawy/explain/pkg/ui"
)

// isLethalSuicideCommand detects commands that have zero legitimate use cases
// and permanently/irreversibly destroy the operating system or user home.
func isLethalSuicideCommand(rawCmd string, pipeline *ast.Pipeline) bool {
	// 1. Fork bombs
	if strings.Contains(rawCmd, ":(){ :|:& };:") || strings.Contains(rawCmd, ":(){:|:&};:") {
		return true
	}

	for _, cmd := range pipeline.Commands {
		argsJoined := strings.Join(cmd.Args, " ")

		// 2. Recursive rm on system roots, home, or with --no-preserve-root
		if cmd.Name == "rm" {
			hasRec := false
			for _, arg := range cmd.Args {
				if arg == "-r" || arg == "-R" || arg == "--recursive" || (strings.HasPrefix(arg, "-") && (strings.Contains(arg, "r") || strings.Contains(arg, "R"))) {
					hasRec = true
					break
				}
			}

			if hasRec {
				for _, arg := range cmd.Args {
					if arg == "/" || arg == "/*" || arg == "~" || arg == "~/*" || arg == "$HOME" || arg == "$HOME/*" {
						return true
					}
				}
				if strings.Contains(argsJoined, "--no-preserve-root") {
					return true
				}
			}
		}

		// 3. Recursive chmod 777 on / or /*
		if cmd.Name == "chmod" {
			hasRec := strings.Contains(argsJoined, "-R") || strings.Contains(argsJoined, "--recursive")
			has777 := strings.Contains(argsJoined, "777") || strings.Contains(argsJoined, "a+rwx")
			if hasRec && has777 {
				for _, arg := range cmd.Args {
					if arg == "/" || arg == "/*" || arg == "~" || arg == "$HOME" {
						return true
					}
				}
			}
		}
	}

	return false
}

// Check inspects a raw command string. If the command poses a critical risk:
// - Suicidal commands (rm -rf /, fork bombs, chmod 777 /) are HARD BLOCKED unconditionally.
// - Other destructive operations (dd, mkfs, etc.) require explicitly typing 'CONFIRM'.
// Returns 0 if allowed/confirmed, or 1 if blocked/aborted.
func Check(rawCmd string) int {
	rawCmd = strings.TrimSpace(rawCmd)
	if rawCmd == "" {
		return 0
	}

	// 1. Immediate detection of fork bombs
	if strings.Contains(rawCmd, ":(){ :|:& };:") || strings.Contains(rawCmd, ":(){:|:&};:") {
		ui.InitColors(false)
		fmt.Println()
		fmt.Printf("  %s %s\n\n", ui.Colorize(ui.BoldRed, "[X]"), ui.Colorize(ui.BoldRed, "explain guard: execution permanently blocked"))
		fmt.Printf("      %s %s\n", ui.Colorize(ui.Dim, "Command :"), ui.Colorize(ui.BoldWhite, rawCmd))
		fmt.Printf("      %s %s\n", ui.Colorize(ui.Dim, "Target  :"), ui.Colorize(ui.Cyan, "fork bomb"))
		fmt.Printf("      %s %s\n", ui.Colorize(ui.Dim, "Risk    :"), ui.Colorize(ui.BoldRed, "[ SYSTEM FREEZE HAZARD ]"))
		fmt.Printf("      %s %s\n", ui.Colorize(ui.Dim, "Action  :"), ui.Colorize(ui.BoldYellow, "This command exhausts process slots and crashes the operating system."))
		fmt.Println()
		fmt.Printf("  %s %s\n\n", ui.Colorize(ui.BoldRed, "[!]"), ui.Colorize(ui.White, "explain guard will not execute this command under any circumstances."))
		return 1
	}

	pipeline, err := ast.Parse(rawCmd)
	if err != nil {
		// If command fails to parse, never block the user's terminal workflow
		return 0
	}

	analysis := analyzer.AnalyzePipeline(pipeline)
	if analysis.MaxRisk != database.RiskCritical {
		return 0
	}

	// Locate the specific critical command details
	var critDanger analyzer.DangerInfo
	var critCmdName string
	for _, cmd := range analysis.Commands {
		if cmd.Danger.Level == database.RiskCritical {
			critDanger = cmd.Danger
			critCmdName = cmd.CommandName
			break
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

	// CASE 1: LETHAL SUICIDE COMMANDS -> HARD BLOCK (NO CONFIRMATION POSSIBLE)
	if isLethalSuicideCommand(rawCmd, pipeline) {
		fmt.Println()
		fmt.Printf("  %s %s\n\n", ui.Colorize(ui.BoldRed, "[X]"), ui.Colorize(ui.BoldRed, "explain guard: execution permanently blocked"))
		fmt.Printf("      %s %s\n", ui.Colorize(ui.Dim, "Command :"), ui.Colorize(ui.BoldWhite, rawCmd))
		fmt.Printf("      %s %s\n", ui.Colorize(ui.Dim, "Target  :"), ui.Colorize(ui.Cyan, critCmdName))
		fmt.Printf("      %s %s\n", ui.Colorize(ui.Dim, "Risk    :"), ui.Colorize(ui.BoldRed, "[ LETHAL SYSTEM DESTRUCTION ]"))
		fmt.Printf("      %s %s\n", ui.Colorize(ui.Dim, "Action  :"), ui.Colorize(ui.BoldYellow, "This command destroys the operating system with zero legitimate use cases."))
		fmt.Println()
		fmt.Printf("  %s %s\n\n", ui.Colorize(ui.BoldRed, "[!]"), ui.Colorize(ui.White, "explain guard will not execute this command under any circumstances."))
		return 1
	}

	// CASE 2: HIGH-RISK OPERATIONS (dd, mkfs, format) -> REQUIRE TYPING 'CONFIRM'
	fmt.Println()
	fmt.Printf("  %s %s\n\n", ui.Colorize(ui.BoldRed, "[!]"), ui.Colorize(ui.BoldRed, "explain guard: critical risk operation detected"))
	fmt.Printf("      %s %s\n", ui.Colorize(ui.Dim, "Command :"), ui.Colorize(ui.BoldWhite, rawCmd))
	fmt.Printf("      %s %s\n", ui.Colorize(ui.Dim, "Target  :"), ui.Colorize(ui.Cyan, critCmdName))
	fmt.Printf("      %s %s\n", ui.Colorize(ui.Dim, "Risk    :"), ui.Colorize(ui.BoldRed, "[ "+badge+" ]"))

	if critDanger.Reason != "" {
		fmt.Printf("      %s %s\n", ui.Colorize(ui.Dim, "Impact  :"), ui.Colorize(ui.White, critDanger.Reason))
	}
	if critDanger.Warning != "" {
		fmt.Printf("      %s %s\n", ui.Colorize(ui.Dim, "Warning :"), ui.Colorize(ui.BoldYellow, critDanger.Warning))
	}

	fmt.Println()
	fmt.Printf("  %s %s", ui.Colorize(ui.BoldYellow, "[?]"), ui.Colorize(ui.BoldWhite, "To proceed, type 'CONFIRM' (or press Enter to cancel): "))

	// Read user response from /dev/tty to ensure direct interactive prompt
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
