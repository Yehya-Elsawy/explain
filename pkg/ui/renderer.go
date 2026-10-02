package ui

import (
	"fmt"
	"strings"

	"github.com/Yehya-Elsawy/explain/pkg/analyzer"
	"github.com/Yehya-Elsawy/explain/pkg/database"
)

func RenderPipeline(analysis *analyzer.PipelineAnalysis) {
	fmt.Println()

	numCmds := len(analysis.Commands)

	if numCmds > 1 {
		renderPipelineOverview(analysis)
	}

	for idx, cmd := range analysis.Commands {
		if numCmds > 1 {
			stageNum := fmt.Sprintf("Stage %d of %d", idx+1, numCmds)
			fmt.Printf("  %s %s\n", Colorize(BoldCyan, "┌── "+stageNum+" ──────────────────────────────────"), Colorize(Dim, "("+cmd.CommandName+")"))
		}

		renderSingleCommand(cmd, numCmds > 1)

		if idx < numCmds-1 {
			if cmd.PipedToNext {
				fmt.Printf("       %s %s\n\n", Colorize(BoldYellow, "│"), Colorize(Dim, "pipes stdout into next command ──►"))
			} else if cmd.ChainOp == "&&" {
				fmt.Printf("       %s %s\n\n", Colorize(BoldYellow, "│"), Colorize(Dim, "runs next command only if this succeeds (exit code 0) ──►"))
			} else if cmd.ChainOp == "||" {
				fmt.Printf("       %s %s\n\n", Colorize(BoldYellow, "│"), Colorize(Dim, "runs next command only if this fails (non-zero exit) ──►"))
			} else if cmd.ChainOp == ";" {
				fmt.Printf("       %s %s\n\n", Colorize(BoldYellow, "│"), Colorize(Dim, "runs next command sequentially after completion ──►"))
			} else if cmd.ChainOp != "" {
				fmt.Printf("       %s %s\n\n", Colorize(BoldYellow, "│"), Colorize(Dim, "then executes ("+cmd.ChainOp+") ──►"))
			}
		}
	}

	if analysis.SmartTip != "" {
		renderSmartTip(analysis.SmartTip)
	}

	fmt.Println()
}

func getExecutionTitle(analysis *analyzer.PipelineAnalysis) string {
	hasPipe := false
	hasAnd := false
	hasOr := false
	hasSemi := false

	for i := 0; i < len(analysis.Commands)-1; i++ {
		cmd := analysis.Commands[i]
		if cmd.PipedToNext {
			hasPipe = true
		} else if cmd.ChainOp == "&&" {
			hasAnd = true
		} else if cmd.ChainOp == "||" {
			hasOr = true
		} else if cmd.ChainOp == ";" {
			hasSemi = true
		}
	}

	distinct := 0
	if hasPipe {
		distinct++
	}
	if hasAnd {
		distinct++
	}
	if hasOr {
		distinct++
	}
	if hasSemi {
		distinct++
	}

	if distinct > 1 {
		return "Compound Command Overview:"
	}
	if hasAnd {
		return "Chained Execution (AND):"
	}
	if hasOr {
		return "Fallback Execution (OR):"
	}
	if hasSemi {
		return "Sequential Execution:"
	}
	return "Pipeline Overview:"
}

func renderPipelineOverview(analysis *analyzer.PipelineAnalysis) {
	var sb strings.Builder
	for i, c := range analysis.Commands {
		sb.WriteString(Colorize(BoldWhite, c.CommandName))
		if i < len(analysis.Commands)-1 {
			opStr := " ──► "
			if c.PipedToNext {
				opStr = " ──[|]──► "
			} else if c.ChainOp == "&&" {
				opStr = " ──[&&]──► "
			} else if c.ChainOp == "||" {
				opStr = " ──[||]──► "
			} else if c.ChainOp == ";" {
				opStr = " ──[;]──► "
			}
			sb.WriteString(Colorize(BoldYellow, opStr))
		}
	}

	title := getExecutionTitle(analysis)
	fmt.Printf("  %s %s\n", Colorize(BoldMagenta, title), sb.String())
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

	if cmd.ActionSummary != "" {
		fmt.Printf("%s%s\n", indent, Colorize(BoldGreen, "What this command does"))
		fmt.Printf("%s  %s %s\n\n", indent, Colorize(BoldGreen, "└─►"), Colorize(White, cmd.ActionSummary))
	}

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

	renderDangerAssessment(cmd.Danger, indent)

	if isPipeline {
		fmt.Printf("  %s\n", Colorize(BoldCyan, "└───────────────────────────────────────────────"))
	}
}

func renderDangerAssessment(danger analyzer.DangerInfo, indent string) {
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
