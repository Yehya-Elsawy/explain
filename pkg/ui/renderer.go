package ui

import (
	"fmt"
	"strings"

	"github.com/elsawy/explain/pkg/analyzer"
	"github.com/elsawy/explain/pkg/database"
)

// RenderPipeline renders the complete formatted explanation to stdout.
func RenderPipeline(analysis *analyzer.PipelineAnalysis) {
	fmt.Println()

	numCmds := len(analysis.Commands)

	if numCmds > 1 {
		// Multi-stage pipeline overview
		renderPipelineHeader(analysis)
	}

	for idx, cmd := range analysis.Commands {
		if numCmds > 1 {
			stageNum := fmt.Sprintf("Stage %d of %d", idx+1, numCmds)
			fmt.Printf("  %s %s\n", Colorize(BoldCyan, "─── "+stageNum+" ───"), Colorize(Dim, "("+cmd.CommandName+")"))
		}

		renderSingleCommand(cmd)

		if idx < numCmds-1 {
			if cmd.PipedToNext {
				fmt.Printf("\n       %s %s\n\n", Colorize(BoldYellow, "│"), Colorize(Dim, "pipes stdout into next command"))
			} else if cmd.ChainOp != "" {
				fmt.Printf("\n       %s %s\n\n", Colorize(BoldYellow, "│"), Colorize(Dim, "then runs next ("+cmd.ChainOp+")"))
			}
		}
	}

	fmt.Println()
}

func renderPipelineHeader(analysis *analyzer.PipelineAnalysis) {
	var stages []string
	for _, c := range analysis.Commands {
		stages = append(stages, Colorize(BoldWhite, c.CommandName))
	}
	pipelineFlow := strings.Join(stages, Colorize(Yellow, "  ⟶  "))
	fmt.Printf("  %s %s\n\n", Colorize(BoldMagenta, "Pipeline:"), pipelineFlow)
}

func renderSingleCommand(cmd *analyzer.CommandAnalysis) {
	// 1. Header (Command Name — Summary)
	icon := getCommandIcon(cmd.CommandName)
	cmdTitle := cmd.CommandName
	if cmd.Subcommand != "" {
		cmdTitle = cmd.CommandName + " " + cmd.Subcommand
	}

	fmt.Printf("  %s %s %s %s\n\n",
		icon,
		Colorize(BoldCyan, cmdTitle),
		Colorize(Dim, "—"),
		Colorize(White, cmd.CommandSummary),
	)

	// 2. Breakdown Section
	if len(cmd.Items) > 0 {
		fmt.Printf("  %s\n", Colorize(BoldWhite, "Breakdown"))

		// Find maximum label width for neat column alignment
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

			fmt.Printf("    %s  %s %s\n",
				Colorize(labelColor, paddedLabel),
				Colorize(Dim, "→"),
				Colorize(White, item.Description),
			)
		}
		fmt.Println()
	}

	// 3. What this command does (Action)
	if cmd.ActionSummary != "" {
		fmt.Printf("  %s\n", Colorize(BoldGreen, "What this command does"))
		fmt.Printf("    %s\n\n", cmd.ActionSummary)
	}

	// 4. Redirections (if present)
	if len(cmd.Redirects) > 0 {
		fmt.Printf("  %s\n", Colorize(BoldYellow, "Redirection"))
		for _, r := range cmd.Redirects {
			target := r.Target
			if target == "" {
				target = "file"
			}
			desc := "Redirects standard output to " + target + " (overwrites file)"
			if r.Operator == ">>" {
				desc = "Appends output to " + target
			} else if r.Operator == "<" {
				desc = "Reads input from " + target
			} else if r.Operator == "2>&1" {
				desc = "Merges error stream (stderr) into standard output (stdout)"
			}
			fmt.Printf("    %s %s  %s %s\n", Colorize(BoldYellow, r.Operator), Colorize(Cyan, target), Colorize(Dim, "→"), desc)
		}
		fmt.Println()
	}

	// 5. Risk / Danger Assessment
	if cmd.Danger.Level != database.RiskSafe && cmd.Danger.Level != database.RiskLow {
		riskTitle := "Risk / things to know"
		riskColor := BoldYellow
		if cmd.Danger.Level == database.RiskCritical || cmd.Danger.Level == database.RiskHigh {
			riskColor = BoldRed
		}

		fmt.Printf("  %s %s\n", Colorize(riskColor, cmd.Danger.Badge), Colorize(BoldWhite, "· "+riskTitle))
		if cmd.Danger.Reason != "" {
			fmt.Printf("    %s\n", cmd.Danger.Reason)
		}
		if cmd.Danger.Warning != "" {
			fmt.Printf("    %s %s\n", Colorize(riskColor, "Warning:"), cmd.Danger.Warning)
		}
	}
}

func getCommandIcon(name string) string {
	switch name {
	case "tar", "zip", "unzip", "gzip":
		return "📦"
	case "rm", "shred":
		return "🗑️ "
	case "cp", "mv":
		return "📄"
	case "ls", "tree":
		return "📂"
	case "mkdir":
		return "📁"
	case "find", "locate", "grep", "rg":
		return "🔍"
	case "curl", "wget", "ssh", "scp", "ping":
		return "🌐"
	case "git":
		return "🌿"
	case "docker":
		return "🐳"
	case "kill", "pkill", "killall":
		return "⚡"
	case "ps", "top", "htop":
		return "📊"
	case "chmod", "chown", "sudo":
		return "🔒"
	case "systemctl", "service", "journalctl":
		return "⚙️ "
	case "dd", "mkfs", "fdisk":
		return "💾"
	case "cat", "head", "tail", "less", "awk", "sed":
		return "📝"
	default:
		return "💻"
	}
}
