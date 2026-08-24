package ui

import (
	"fmt"
	"strings"

	"github.com/Yehya-Elsawy/explain/pkg/analyzer"
	"github.com/Yehya-Elsawy/explain/pkg/database"
)

// RenderPipeline renders the complete formatted explanation to stdout.
func RenderPipeline(analysis *analyzer.PipelineAnalysis) {
	fmt.Println()

	numCmds := len(analysis.Commands)

	// 1. Overall Pipeline Header (if multi-stage)
	if numCmds > 1 {
		renderPipelineOverview(analysis)
	}

	// 2. Render each command stage
	for idx, cmd := range analysis.Commands {
		if numCmds > 1 {
			stageNum := fmt.Sprintf("Stage %d of %d", idx+1, numCmds)
			fmt.Printf("  %s %s\n", Colorize(BoldCyan, "┌── "+stageNum+" ──────────────────────────────────"), Colorize(Dim, "("+cmd.CommandName+")"))
		}

		renderSingleCommand(cmd, numCmds > 1)

		if idx < numCmds-1 {
			if cmd.PipedToNext {
				fmt.Printf("       %s %s\n\n", Colorize(BoldYellow, "│"), Colorize(Dim, "pipes stdout into next command ──►"))
			} else if cmd.ChainOp != "" {
				fmt.Printf("       %s %s\n\n", Colorize(BoldYellow, "│"), Colorize(Dim, "then executes ("+cmd.ChainOp+") ──►"))
			}
		}
	}

	// 3. Render Smart Tip / Better Alternative (if available)
	if analysis.SmartTip != "" {
		renderSmartTip(analysis.SmartTip)
	}

	fmt.Println()
}

func renderPipelineOverview(analysis *analyzer.PipelineAnalysis) {
	var stages []string
	for _, c := range analysis.Commands {
		stages = append(stages, Colorize(BoldWhite, c.CommandName))
	}
	pipelineFlow := strings.Join(stages, Colorize(BoldYellow, " ──► "))

	fmt.Printf("  %s %s\n", Colorize(BoldMagenta, "Pipeline Overview:"), pipelineFlow)
	if analysis.PipelineSummary != "" {
		fmt.Printf("  %s %s\n", Colorize(Dim, "└─►"), Colorize(White, analysis.PipelineSummary))
	}
	fmt.Println()
}

func renderSingleCommand(cmd *analyzer.CommandAnalysis, isPipeline bool) {
	indent := "  "
	if isPipeline {
		indent = "  │ "
	}

	// 1. Command Header (Command Name + Summary)
	cmdTitle := cmd.CommandName
	if cmd.Subcommand != "" {
		cmdTitle = cmd.CommandName + " " + cmd.Subcommand
	}

	fmt.Printf("%s%s %s %s\n\n",
		indent,
		Colorize(BoldCyan, cmdTitle),
		Colorize(Dim, "—"),
		Colorize(White, cmd.CommandSummary),
	)

	// 2. What this command does (Human Summary)
	if cmd.ActionSummary != "" {
		fmt.Printf("%s%s\n", indent, Colorize(BoldGreen, "What this command does"))
		fmt.Printf("%s  %s %s\n\n", indent, Colorize(BoldGreen, "└─►"), Colorize(White, cmd.ActionSummary))
	}

	// 3. Breakdown Section (Flags & Arguments Table)
	if len(cmd.Items) > 0 {
		fmt.Printf("%s%s\n", indent, Colorize(BoldWhite, "Breakdown"))

		maxLabelLen := 0
		for _, item := range cmd.Items {
			if len(item.Label) > maxLabelLen {
				maxLabelLen = len(item.Label)
			}
		}
		if maxLabelLen < 12 {
			maxLabelLen = 12
		}
		if maxLabelLen > 24 {
			maxLabelLen = 24
		}

		for _, item := range cmd.Items {
			labelColor := BoldYellow
			if item.IsSubcmd {
				labelColor = BoldMagenta
			} else if item.IsPrefix {
				labelColor = BoldCyan
			} else if item.IsArg {
				labelColor = Cyan
			}

			paddedLabel := item.Label
			if len(paddedLabel) < maxLabelLen {
				paddedLabel += strings.Repeat(" ", maxLabelLen-len(paddedLabel))
			}

			fmt.Printf("%s  %s %s %s\n",
				indent,
				Colorize(labelColor, paddedLabel),
				Colorize(Dim, "→"),
				Colorize(White, item.Description),
			)
		}
		fmt.Println()
	}

	// 4. Redirections Section (if present)
	if len(cmd.Redirects) > 0 {
		fmt.Printf("%s%s\n", indent, Colorize(BoldYellow, "I/O Redirection"))
		for _, r := range cmd.Redirects {
			target := r.Target
			if target == "" {
				target = "file"
			}
			desc := "Redirects standard output to " + target + " (overwrites existing file)"
			if r.Operator == ">>" {
				desc = "Appends standard output to " + target
			} else if r.Operator == "<" {
				desc = "Reads standard input from " + target
			} else if r.Operator == "2>&1" {
				desc = "Merges error output (stderr) into standard output stream (stdout)"
			}
			fmt.Printf("%s  %s %s %s %s\n", indent, Colorize(BoldYellow, r.Operator), Colorize(Cyan, target), Colorize(Dim, "→"), Colorize(White, desc))
		}
		fmt.Println()
	}

	// 5. Safety & Danger Assessment (ONLY rendered if command is genuinely risky)
	renderDangerAssessment(cmd.Danger, indent)

	if isPipeline {
		fmt.Printf("  %s\n", Colorize(BoldCyan, "└───────────────────────────────────────────────"))
	}
}

func renderDangerAssessment(danger analyzer.DangerInfo, indent string) {
	// ONLY render risk box if the command is genuinely risky (Medium, High, Critical, or has warning)
	if danger.Level != database.RiskMedium && danger.Level != database.RiskHigh && danger.Level != database.RiskCritical && danger.Warning == "" {
		return
	}

	riskColor := BoldYellow
	if danger.Level == database.RiskHigh || danger.Level == database.RiskCritical {
		riskColor = BoldRed
	}

	badgeText := danger.Badge
	badgeText = strings.TrimSpace(badgeText)
	badgeText = strings.TrimPrefix(badgeText, "🚨 ")
	badgeText = strings.TrimPrefix(badgeText, "⚠️ ")
	badgeText = strings.TrimPrefix(badgeText, "🟡 ")
	badgeText = strings.TrimPrefix(badgeText, "🟢 ")

	fmt.Printf("%s%s %s\n", indent, Colorize(riskColor, "[ "+badgeText+" ]"), Colorize(Dim, "· Risk Assessment"))
	if danger.Reason != "" {
		fmt.Printf("%s  %s %s\n", indent, Colorize(Dim, "└─►"), Colorize(White, danger.Reason))
	}
	if danger.Warning != "" {
		fmt.Printf("%s  %s %s\n", indent, Colorize(riskColor, "WARNING:"), Colorize(BoldWhite, danger.Warning))
	}
	fmt.Println()
}

func renderSmartTip(tip string) {
	tipText := strings.TrimSpace(tip)
	tipText = strings.TrimPrefix(tipText, "💡 ")
	tipText = strings.TrimPrefix(tipText, "⚠️ ")
	tipText = strings.TrimPrefix(tipText, "🚨 ")
	
	if strings.HasPrefix(tipText, "Tip: ") || strings.HasPrefix(tipText, "Security Note: ") || strings.HasPrefix(tipText, "Critical Danger: ") {
		fmt.Printf("  %s\n", Colorize(BoldYellow, tipText))
	} else {
		fmt.Printf("  %s %s\n", Colorize(BoldYellow, "Tip:"), Colorize(White, tipText))
	}
}
