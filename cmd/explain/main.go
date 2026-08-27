package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime/debug"
	"strings"

	"github.com/Yehya-Elsawy/explain/pkg/analyzer"
	"github.com/Yehya-Elsawy/explain/pkg/ast"
	"github.com/Yehya-Elsawy/explain/pkg/ui"
	"github.com/Yehya-Elsawy/explain/pkg/updater"
)

var Version = "v1.1.0"

func init() {
	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Version != "" && info.Main.Version != "(devel)" {
			Version = info.Main.Version
		}
	}
}

func printHelp() {
	ui.InitColors(false)
	fmt.Printf(`
%s — Linux command explainer for beginners.

%s
  explain <command with arguments>
  explain "<piped | or compound command>"
  explain !! (explain the last executed command)
  explain update (update explain to the latest version)
  explain -i (interactive mode - no quotes needed)

%s
  explain tar -xzf backup.tar.gz
  explain cd /home/
  explain !!
  explain "ps aux | grep nginx | awk '{print $2}' | xargs kill -9"
  explain "curl -fsSL https://get.docker.com | sh"
  explain "rm -rf /tmp/cache"

%s
  explain update    Check and upgrade explain to the latest release from GitHub
  explain uninstall Remove explain CLI from your system
  -i, --interactive Launch interactive mode (paste complex pipelines without quotes)
  -r, --run         Ask to run the command after explaining it
  --json            Output structured analysis in JSON format
  --no-color        Disable colored output
  -v, --version     Show current explain version
  -h, --help        Show this help message

%s
  Why quotes for pipes? The shell intercepts '|', '>', '&&' before passing them to programs.
  In interactive mode (explain -i), you can paste complex pipelines directly without any quotes!
`,
		ui.Colorize(ui.BoldCyan, "explain "+Version),
		ui.Colorize(ui.BoldWhite, "USAGE:"),
		ui.Colorize(ui.BoldWhite, "EXAMPLES:"),
		ui.Colorize(ui.BoldWhite, "COMMANDS & OPTIONS:"),
		ui.Colorize(ui.Dim, "PRO TIP:"),
	)
}

func runInteractive() {
	ui.InitColors(false)
	fmt.Println()
	fmt.Println(ui.Colorize(ui.BoldCyan, "  💡 explain interactive mode ("+Version+")"))
	fmt.Println(ui.Colorize(ui.Dim, "  Type or paste any command with pipes (|), redirects (>), or operators (no quotes needed)."))
	fmt.Println(ui.Colorize(ui.Dim, "  Type 'exit' or press Ctrl+C to quit.\n"))

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(ui.Colorize(ui.BoldGreen, "  explain > "))
		input, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}
		if input == "exit" || input == "quit" || input == "q" {
			break
		}

		if input == "!!" {
			lastCmd := ast.GetLastHistoryCommand()
			if lastCmd != "" {
				input = lastCmd
			}
		}

		pipeline, err := ast.Parse(input)
		if err != nil {
			fmt.Printf("  %s %v\n\n", ui.Colorize(ui.BoldRed, "Error parsing:"), err)
			continue
		}

		analysis := analyzer.AnalyzePipeline(pipeline)
		ui.RenderPipeline(analysis)
	}
}

func main() {
	args := os.Args[1:]

	// Check if data is piped via stdin (e.g. echo "ps aux | grep nginx" | explain)
	stat, _ := os.Stdin.Stat()
	isPiped := (stat.Mode() & os.ModeCharDevice) == 0

	if len(args) == 0 && isPiped {
		bytes, err := io.ReadAll(os.Stdin)
		if err == nil && len(strings.TrimSpace(string(bytes))) > 0 {
			input := strings.TrimSpace(string(bytes))
			pipeline, err := ast.Parse(input)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing command: %v\n", err)
				os.Exit(1)
			}
			analysis := analyzer.AnalyzePipeline(pipeline)
			ui.RenderPipeline(analysis)
			return
		}
	}

	if len(args) == 0 {
		printHelp()
		os.Exit(0)
	}

	disableColor := false
	outputJSON := false
	runAfter := false
	interactive := false
	doUpdate := false
	doUninstall := false
	var cmdTokens []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "-h" || arg == "--help" {
			printHelp()
			return
		}
		if arg == "-v" || arg == "--version" || arg == "version" {
			fmt.Printf("explain %s\n", Version)
			return
		}
		if arg == "update" || arg == "--update" || arg == "-u" {
			doUpdate = true
			break
		}
		if arg == "uninstall" || arg == "--uninstall" {
			doUninstall = true
			break
		}
		if arg == "-i" || arg == "--interactive" {
			interactive = true
			continue
		}
		if arg == "--no-color" {
			disableColor = true
			continue
		}
		if arg == "--json" {
			outputJSON = true
			continue
		}
		if arg == "-r" || arg == "--run" {
			runAfter = true
			continue
		}
		cmdTokens = append(cmdTokens, args[i:]...)
		break
	}

	ui.InitColors(disableColor)

	if doUpdate {
		if err := updater.SelfUpdate(Version); err != nil {
			fmt.Fprintf(os.Stderr, "%s %v\n", ui.Colorize(ui.BoldRed, "Error:"), err)
			os.Exit(1)
		}
		return
	}

	if doUninstall {
		runUninstall()
		return
	}

	if interactive {
		runInteractive()
		return
	}

	if len(cmdTokens) == 0 {
		printHelp()
		return
	}

	rawInput := strings.Join(cmdTokens, " ")

	// Handle !! (explain last history command)
	trimmed := strings.TrimSpace(rawInput)
	if trimmed == "!!" || trimmed == "\"!!\"" || trimmed == "'!!'" {
		lastCmd := ast.GetLastHistoryCommand()
		if lastCmd != "" {
			fmt.Printf("  %s %s\n", ui.Colorize(ui.Dim, "Explaining last command:"), ui.Colorize(ui.BoldYellow, lastCmd))
			rawInput = lastCmd
		}
	}

	pipeline, err := ast.Parse(rawInput)
	if err != nil {
		if strings.TrimSpace(rawInput) == "" {
			fmt.Fprintf(os.Stderr, "Error parsing command: %v\n", err)
			printHelp()
		} else {
			fmt.Fprintf(os.Stderr, "Error parsing command: %v\n", err)
		}
		os.Exit(1)
	}

	if len(pipeline.Commands) == 0 {
		fmt.Fprintln(os.Stderr, "No command provided to explain.")
		os.Exit(1)
	}

	analysis := analyzer.AnalyzePipeline(pipeline)

	if outputJSON {
		jsonData, err := json.MarshalIndent(analysis, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "JSON error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(jsonData))
		return
	}

	ui.RenderPipeline(analysis)

	if runAfter {
		fmt.Printf("  %s %s? (y/N): ", ui.Colorize(ui.BoldYellow, "Execute this command now"), ui.Colorize(ui.Dim, "['"+rawInput+"']"))
		reader := bufio.NewReader(os.Stdin)
		resp, _ := reader.ReadString('\n')
		resp = strings.TrimSpace(strings.ToLower(resp))
		if resp == "y" || resp == "yes" {
			fmt.Println()
			execCmd := exec.Command("bash", "-c", rawInput)
			execCmd.Stdout = os.Stdout
			execCmd.Stderr = os.Stderr
			execCmd.Stdin = os.Stdin
			if err := execCmd.Run(); err != nil {
				os.Exit(1)
			}
		}
	}
}

func runUninstall() {
	ui.InitColors(false)
	fmt.Println()

	execPath, err := os.Executable()
	if err != nil {
		execPath = ""
	}

	targets := []string{}
	if execPath != "" {
		targets = append(targets, execPath)
	}

	home, _ := os.UserHomeDir()
	if home != "" {
		targets = append(targets, home+"/.local/bin/explain")
	}
	targets = append(targets, "/usr/local/bin/explain")

	removedAny := false
	seen := make(map[string]bool)

	for _, path := range targets {
		if seen[path] {
			continue
		}
		seen[path] = true

		if _, err := os.Stat(path); err == nil {
			err := os.Remove(path)
			if err != nil {
				// Try removing with sudo if permission denied
				execCmd := exec.Command("sudo", "rm", "-f", path)
				if err2 := execCmd.Run(); err2 == nil {
					fmt.Printf("  %s Removed %s\n", ui.Colorize(ui.BoldGreen, "✓"), path)
					removedAny = true
				} else {
					fmt.Printf("  %s Could not remove %s (permission denied)\n", ui.Colorize(ui.BoldRed, "✗"), path)
				}
			} else {
				fmt.Printf("  %s Removed %s\n", ui.Colorize(ui.BoldGreen, "✓"), path)
				removedAny = true
			}
		}
	}

	if removedAny {
		fmt.Println(ui.Colorize(ui.BoldGreen, "\n  Successfully uninstalled explain CLI."))
		fmt.Println(ui.Colorize(ui.Dim, "  Thank you for using explain!\n"))
	} else {
		fmt.Println(ui.Colorize(ui.BoldYellow, "  No installed explain binary found to remove.\n"))
	}
}
