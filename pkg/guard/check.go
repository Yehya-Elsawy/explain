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

// Check inspects a raw command string. If the command poses a critical risk,
// it halts, displays a risk warning, and prompts the user for confirmation.
// Returns 0 if allowed/safe/confirmed, or 1 if cancelled/aborted.
func Check(rawCmd string) int {
	rawCmd = strings.TrimSpace(rawCmd)
	if rawCmd == "" {
		return 0
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
	fmt.Printf("  %s %s", ui.Colorize(ui.BoldYellow, "[?]"), ui.Colorize(ui.BoldWhite, "Execute this dangerous command anyway? (y/N): "))

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

	resp = strings.ToLower(strings.TrimSpace(resp))
	if resp == "y" || resp == "yes" {
		fmt.Printf("  %s %s\n\n", ui.Colorize(ui.Dim, "[>]"), ui.Colorize(ui.Dim, "Proceeding with execution..."))
		return 0
	}

	fmt.Printf("\n  %s %s\n\n", ui.Colorize(ui.BoldRed, "[X]"), ui.Colorize(ui.BoldRed, "Command execution cancelled by explain guard."))
	return 1
}
