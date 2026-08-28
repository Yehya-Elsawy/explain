package analyzer

import (
	"strings"

	"github.com/Yehya-Elsawy/explain/pkg/ast"
	"github.com/Yehya-Elsawy/explain/pkg/database"
	"github.com/Yehya-Elsawy/explain/pkg/manparser"
)

// ExplainedItem represents a single token/flag/argument breakdown item.
type ExplainedItem struct {
	Token       string
	Label       string
	Description string
	IsFlag      bool
	IsSubcmd    bool
	IsArg       bool
	IsPrefix    bool
	Risk        database.RiskLevel
}

// CommandAnalysis contains the complete breakdown and explanation of a command.
type CommandAnalysis struct {
	RawCommand     string
	Prefixes       []string
	CommandName    string
	CommandSummary string
	Subcommand     string
	Items          []ExplainedItem
	PositionalArgs []string
	Redirects      []ast.Redirect
	Danger         DangerInfo
	ActionSummary  string
	PipedToNext    bool
	ChainOp        string
}

// PipelineAnalysis contains the analysis of all piped or chained commands.
type PipelineAnalysis struct {
	RawInput        string
	Commands        []*CommandAnalysis
	MaxRisk         database.RiskLevel
	PipelineSummary string
	SmartTip        string
}

// AnalyzePipeline analyzes a full parsed AST pipeline.
func AnalyzePipeline(pipe *ast.Pipeline) *PipelineAnalysis {
	result := &PipelineAnalysis{
		RawInput: pipe.RawInput,
		Commands: make([]*CommandAnalysis, 0, len(pipe.Commands)),
		MaxRisk:  database.RiskSafe,
	}

	for _, cmd := range pipe.Commands {
		analysis := AnalyzeSingleCommand(cmd)
		result.Commands = append(result.Commands, analysis)
		if isHigherRisk(analysis.Danger.Level, result.MaxRisk) {
			result.MaxRisk = analysis.Danger.Level
		}
	}

	CheckPipelineDangers(result)
	result.PipelineSummary = SynthesizePipelineSummary(result)
	result.SmartTip = SuggestAlternative(result)

	return result
}

// AnalyzeSingleCommand analyzes a single command node.
func AnalyzeSingleCommand(cmd *ast.SingleCommand) *CommandAnalysis {
	analysis := &CommandAnalysis{
		Prefixes:       cmd.Prefixes,
		CommandName:    cmd.Name,
		Redirects:      cmd.Redirects,
		PipedToNext:    cmd.PipedToNext,
		ChainOp:        cmd.ChainOp,
		Items:          []ExplainedItem{},
		PositionalArgs: []string{},
	}

	// 1. Add Prefixes (sudo, xargs, etc.)
	for _, p := range cmd.Prefixes {
		desc := "Runs the command with administrative (root) privileges"
		if p == "xargs" {
			desc = "Builds and executes this command using arguments read from standard input"
		} else if p == "nohup" {
			desc = "Runs command immune to hangups (continues running after terminal closes)"
		} else if p == "time" {
			desc = "Measures execution time and resource usage of this command"
		} else if p != "sudo" && p != "doas" {
			desc = "Command execution wrapper"
		}
		analysis.Items = append(analysis.Items, ExplainedItem{
			Token:       p,
			Label:       p,
			Description: desc,
			IsPrefix:    true,
			Risk:        database.RiskMedium,
		})
	}

	// 2. Fetch or lookup Command Definition
	cmdDef, hasDef := database.BuiltinCommands[cmd.Name]
	if hasDef {
		analysis.CommandSummary = cmdDef.Summary
	} else {
		// Fallback to dynamic man page extraction
		summary := manparser.ExtractCommandSummary(cmd.Name)
		if summary != "" {
			analysis.CommandSummary = summary
		} else {
			analysis.CommandSummary = "Linux executable command"
		}
	}

	// 3. Detect Subcommand (e.g. git commit, docker run, systemctl restart)
	args := cmd.Args
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		firstArg := args[0]
		if hasDef && cmdDef.Subcommands != nil {
			if subDef, isSub := cmdDef.Subcommands[firstArg]; isSub {
				analysis.Subcommand = firstArg
				analysis.Items = append(analysis.Items, ExplainedItem{
					Token:       firstArg,
					Label:       firstArg,
					Description: subDef.Summary,
					IsSubcmd:    true,
					Risk:        subDef.DefaultRisk,
				})
				args = args[1:]
			}
		}
	}

	// 4. Process Flags and Positional Arguments
	i := 0
	for i < len(args) {
		arg := args[i]

		// Special case: `tar xzf` or `tar czvf` (tar flags without leading dash)
		if cmd.Name == "tar" && i == 0 && !strings.HasPrefix(arg, "-") && isTarFlagCluster(arg) {
			arg = "-" + arg
		}

		// Special case: `ps aux` or `ps ax` (BSD-style flags without leading dash)
		if cmd.Name == "ps" && (arg == "aux" || arg == "ax" || arg == "lax") {
			analysis.Items = append(analysis.Items, ExplainedItem{
				Token:       arg,
				Label:       arg,
				Description: "Show all processes for all users in full BSD-style listing format",
				IsFlag:      true,
				Risk:        database.RiskSafe,
			})
			i++
			continue
		}

		// Long option: --flag or --flag=value
		if strings.HasPrefix(arg, "--") {
			flagKey := strings.TrimPrefix(arg, "--")
			flagVal := ""
			if strings.Contains(flagKey, "=") {
				parts := strings.SplitN(flagKey, "=", 2)
				flagKey = parts[0]
				flagVal = parts[1]
			}

			flagDef, found := findFlagDef(cmdDef, analysis.Subcommand, flagKey)
			if !found {
				flagDef = lookupDynamicFlagDef(cmd.Name, "--"+flagKey)
			}
			if flagVal == "" && flagDef.TakesValue && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				flagVal = args[i+1]
				i++
			}

			label := "--" + flagKey
			if flagDef.Long != "" {
				label = flagDef.Long
			}
			desc := flagDef.Description
			if desc == "" {
				desc = "Option flag --" + flagKey
			}

			if flagVal != "" {
				desc += " (" + flagVal + ")"
			}

			analysis.Items = append(analysis.Items, ExplainedItem{
				Token:       arg,
				Label:       label,
				Description: desc,
				IsFlag:      true,
				Risk:        flagDef.Risk,
			})
			i++
			continue
		}

		// Numeric signal flags (e.g. kill -9 or kill -15)
		if (cmd.Name == "kill" || cmd.Name == "pkill" || cmd.Name == "killall") && strings.HasPrefix(arg, "-") && isNumber(arg[1:]) {
			sigNum := arg[1:]
			sigDef, _ := findFlagDef(cmdDef, "", sigNum)
			desc := sigDef.Description
			if desc == "" {
				desc = "Send signal " + sigNum + " to process"
			}
			analysis.Items = append(analysis.Items, ExplainedItem{
				Token:       arg,
				Label:       arg,
				Description: desc,
				IsFlag:      true,
				Risk:        sigDef.Risk,
			})
			i++
			continue
		}

		// Short option(s): -x or -xzf or dd key=val style
		if strings.HasPrefix(arg, "-") && len(arg) > 1 && (!isNumber(arg[1:]) || cmd.Name == "find") {
			cluster := arg[1:]

			// Single named options with single dash (e.g. find -name, -type, -mtime, -maxdepth)
			if cmd.Name == "find" && len(cluster) > 1 {
				flagDef, found := findFlagDef(cmdDef, "", cluster)
				if !found {
					flagDef = lookupDynamicFlagDef(cmd.Name, arg)
				}
				desc := flagDef.Description
				if desc == "" {
					desc = "Option flag " + arg
				}
				if flagDef.TakesValue && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
					desc += " (" + args[i+1] + ")"
					i++
				}
				analysis.Items = append(analysis.Items, ExplainedItem{
					Token:       arg,
					Label:       arg,
					Description: desc,
					IsFlag:      true,
					Risk:        flagDef.Risk,
				})
				i++
				continue
			}

			// Clustered short flags (-xzvf)
			runes := []rune(cluster)
			for rIdx, r := range runes {
				charStr := string(r)
				flagDef, found := findFlagDef(cmdDef, analysis.Subcommand, charStr)
				if !found {
					flagDef = lookupDynamicFlagDef(cmd.Name, "-"+charStr)
				}
				label := "-" + charStr
				if flagDef.Long != "" {
					label = flagDef.Short + ", " + flagDef.Long
				} else if flagDef.Short != "" {
					label = flagDef.Short
				}

				desc := flagDef.Description
				if desc == "" {
					desc = "Option flag -" + charStr
				}

				consumedClusterRemainder := false
				if flagDef.TakesValue {
					val := ""
					if rIdx < len(runes)-1 {
						val = string(runes[rIdx+1:])
						consumedClusterRemainder = true
					} else if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
						val = args[i+1]
						i++ // consume the following value
					}
					if val != "" {
						desc += " (" + val + ")"
					}
				}

				analysis.Items = append(analysis.Items, ExplainedItem{
					Token:       "-" + charStr,
					Label:       label,
					Description: desc,
					IsFlag:      true,
					Risk:        flagDef.Risk,
				})
				if consumedClusterRemainder {
					break
				}
			}
			i++
			continue
		}

		// dd style args (if=file, of=file, bs=4M)
		if cmd.Name == "dd" && strings.Contains(arg, "=") {
			parts := strings.SplitN(arg, "=", 2)
			prefix := parts[0]
			val := parts[1]
			flagDef, found := cmdDef.Flags[prefix]
			desc := flagDef.Description
			if !found {
				desc = "Set " + prefix + " parameter"
			}
			desc += " (" + val + ")"
			analysis.Items = append(analysis.Items, ExplainedItem{
				Token:       arg,
				Label:       prefix + "=",
				Description: desc,
				IsFlag:      true,
				Risk:        flagDef.Risk,
			})
			i++
			continue
		}

		// Positional argument
		analysis.PositionalArgs = append(analysis.PositionalArgs, arg)
		analysis.Items = append(analysis.Items, ExplainedItem{
			Token:       arg,
			Label:       arg,
			Description: describePositionalArg(cmd.Name, analysis.Subcommand, arg, len(analysis.PositionalArgs)),
			IsArg:       true,
		})
		i++
	}

	// 5. Evaluate Danger
	analysis.Danger = EvaluateDanger(cmd, analysis)

	// 6. Synthesize Action Summary
	analysis.ActionSummary = SynthesizeAction(cmd, analysis)

	return analysis
}

func isTarFlagCluster(s string) bool {
	valid := "xczjJvftCp"
	if len(s) == 0 {
		return false
	}
	for _, r := range s {
		if !strings.ContainsRune(valid, r) {
			return false
		}
	}
	return true
}

func findFlagDef(cmdDef database.CommandDef, subcmd, key string) (database.FlagDef, bool) {
	if subcmd != "" && cmdDef.Subcommands != nil {
		if subDef, ok := cmdDef.Subcommands[subcmd]; ok && subDef.Flags != nil {
			if f, ok := subDef.Flags[key]; ok {
				return f, true
			}
		}
	}
	if cmdDef.Flags != nil {
		if f, ok := cmdDef.Flags[key]; ok {
			return f, true
		}
	}
	return database.FlagDef{}, false
}

func lookupDynamicFlagDef(cmdName, flag string) database.FlagDef {
	info := manparser.ExtractFlagInfo(cmdName, flag)
	if info.Description == "" && !info.TakesValue {
		info = manparser.ExtractHelpFlagInfo(cmdName, flag)
	}
	return database.FlagDef{
		Description: info.Description,
		TakesValue:  info.TakesValue,
		ValueName:   info.ValueName,
	}
}

func describePositionalArg(cmdName, subcmd, arg string, pos int) string {
	switch cmdName {
	case "cd":
		if arg == ".." {
			return "Parent directory (moves one level up)"
		} else if arg == "~" {
			return "Current user's home directory (/home/username)"
		} else if arg == "-" {
			return "Previous working directory ($OLDPWD)"
		} else if arg == "." {
			return "Current directory"
		}
		return "Target destination directory to switch to"
	case "pwd":
		return "Argument"
	case "echo":
		return "Text / string to print"
	case "source", ".":
		return "Script file path to execute"
	case "which":
		return "Command name to locate"
	case "export":
		return "Environment variable definition (KEY=VALUE)"
	case "tar":
		if strings.HasSuffix(arg, ".tar") || strings.HasSuffix(arg, ".tar.gz") || strings.HasSuffix(arg, ".tgz") || strings.HasSuffix(arg, ".tar.bz2") || strings.HasSuffix(arg, ".tar.xz") {
			return "Archive file"
		}
		return "Target file / directory"
	case "chmod":
		if pos == 1 {
			return "Permission mode (e.g. 755, 644, +x)"
		}
		return "Target file / directory"
	case "chown":
		if pos == 1 {
			return "Target User / Group ownership (e.g. root:root)"
		}
		return "Target file / directory"
	case "rm", "cp", "mv", "ls", "cat", "head", "tail":
		return "Target path / file"
	case "find":
		if pos == 1 && !strings.HasPrefix(arg, "-") {
			return "Starting directory path to search in"
		}
		return "Pattern / argument"
	case "grep":
		if pos == 1 {
			return "Search pattern / regex"
		}
		return "Target file / directory to search"
	case "curl", "wget":
		if strings.HasPrefix(arg, "http://") || strings.HasPrefix(arg, "https://") {
			return "Target URL"
		}
		return "Argument"
	case "ssh":
		if strings.Contains(arg, "@") {
			return "Remote host destination (user@hostname)"
		}
		return "Target hostname or remote command"
	case "docker":
		if subcmd == "run" {
			return "Docker image name / tag"
		}
	case "git":
		if subcmd == "checkout" || subcmd == "branch" {
			return "Branch name"
		}
	}
	return "Positional argument"
}

func isNumber(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func isHigherRisk(a, b database.RiskLevel) bool {
	order := map[database.RiskLevel]int{
		database.RiskSafe:     0,
		database.RiskLow:      1,
		database.RiskMedium:   2,
		database.RiskHigh:     3,
		database.RiskCritical: 4,
	}
	return order[a] > order[b]
}
