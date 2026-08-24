package analyzer

import (
	"fmt"
	"strings"

	"github.com/Yehya-Elsawy/explain/pkg/ast"
)

// SynthesizePipelineSummary creates a cohesive 1-2 sentence human explanation for multi-stage pipelines.
func SynthesizePipelineSummary(pipe *PipelineAnalysis) string {
	numCmds := len(pipe.Commands)
	if numCmds <= 1 {
		return ""
	}

	names := make([]string, len(pipe.Commands))
	for i, c := range pipe.Commands {
		names[i] = c.CommandName
	}
	joinedNames := strings.Join(names, " → ")

	// Pattern 1: ps ... | grep ... | awk ... | xargs kill
	if hasCmd(names, "ps") && (hasCmd(names, "kill") || hasCmd(names, "pkill")) {
		return "Searches for active process IDs matching a target process pattern and forcefully terminates them."
	}

	// Pattern 2: ps ... | grep ... | awk/cut
	if hasCmd(names, "ps") && hasCmd(names, "grep") {
		return "Lists running system processes, filters matching lines, and extracts specific process details."
	}

	// Pattern 3: Remote script download & execute (curl/wget | bash/sh)
	if (hasCmd(names, "curl") || hasCmd(names, "wget")) && (hasCmd(names, "bash") || hasCmd(names, "sh") || hasCmd(names, "python") || hasCmd(names, "python3")) {
		return "Downloads a remote script from the web and pipes it directly to a shell interpreter for immediate execution."
	}

	// Pattern 4: find ... | xargs rm / chmod / etc.
	if hasCmd(names, "find") && hasCmd(names, "xargs") {
		return "Finds files matching specified criteria and runs batch commands on every matched item."
	}

	// Pattern 5: cat/find | grep | wc/sort/uniq
	if (hasCmd(names, "cat") || hasCmd(names, "grep")) && (hasCmd(names, "wc") || hasCmd(names, "sort") || hasCmd(names, "uniq")) {
		return "Inspects and filters input data, then aggregates, counts, or sorts the results."
	}

	// General Pipeline Fallback
	return fmt.Sprintf("A %d-stage pipeline (%s) that feeds the output of each command directly into the next.", numCmds, joinedNames)
}

// SuggestAlternative checks for command anti-patterns and suggests cleaner/safer single-command alternatives.
func SuggestAlternative(pipe *PipelineAnalysis) string {
	if len(pipe.Commands) == 0 {
		return ""
	}

	// 1. Danger alerts take top priority for tips
	for _, cmd := range pipe.Commands {
		if cmd.CommandName == "rm" && strings.Contains(cmd.RawCommand, "-rf") && (strings.Contains(cmd.RawCommand, "/") || strings.Contains(cmd.RawCommand, "/*")) {
			return "Critical Danger: Running 'rm -rf /' will permanently erase your entire operating system!"
		}
		if cmd.CommandName == "chmod" && (strings.Contains(cmd.RawCommand, "777") || strings.Contains(cmd.RawCommand, "a+rwx")) {
			return "Security Tip: 'chmod 777' grants full permission to everyone. Use 'chmod 755' for directories/executables, or 'chmod 644' for regular files."
		}
	}

	// 2. Useless Use of Cat (cat file | grep pattern)
	if len(pipe.Commands) >= 2 && pipe.Commands[0].CommandName == "cat" && len(pipe.Commands[0].PositionalArgs) > 0 {
		file := pipe.Commands[0].PositionalArgs[0]
		second := pipe.Commands[1].CommandName
		if second == "grep" {
			pattern := ""
			if len(pipe.Commands[1].PositionalArgs) > 0 {
				pattern = pipe.Commands[1].PositionalArgs[0]
			}
			if pattern != "" {
				return fmt.Sprintf("Tip: Avoid unnecessary 'cat'. Search directly with: grep '%s' %s", pattern, file)
			}
			return fmt.Sprintf("Tip: Avoid unnecessary 'cat'. Pass the file directly to grep: grep <pattern> %s", file)
		}
		if second == "awk" || second == "sed" {
			return fmt.Sprintf("Tip: Avoid unnecessary 'cat'. You can pass the file directly to %s: %s '<script>' %s", second, second, file)
		}
	}

	// 3. Process killing pipeline (ps aux | grep name | awk ... | xargs kill)
	names := make([]string, len(pipe.Commands))
	for i, c := range pipe.Commands {
		names[i] = c.CommandName
	}
	if hasCmd(names, "ps") && hasCmd(names, "grep") && (hasCmd(names, "kill") || hasCmd(names, "xargs")) {
		target := "process_name"
		for _, c := range pipe.Commands {
			if c.CommandName == "grep" && len(c.PositionalArgs) > 0 {
				target = c.PositionalArgs[0]
				break
			}
		}
		return fmt.Sprintf("Tip: Instead of this multi-stage pipeline, simplify with a single command: pkill -9 %s", target)
	}

	// 4. find ... | xargs rm
	if hasCmd(names, "find") && (hasCmd(names, "xargs") || hasCmd(names, "rm")) {
		for _, c := range pipe.Commands {
			if c.CommandName == "rm" || (c.CommandName == "xargs" && strings.Contains(c.RawCommand, "rm")) {
				return "Tip: 'find' has a built-in delete flag! Simplify to: find . -name '<pattern>' -delete"
			}
		}
	}

	// 5. Remote script execution warning
	if (hasCmd(names, "curl") || hasCmd(names, "wget")) && (hasCmd(names, "bash") || hasCmd(names, "sh")) {
		return "Security Note: Executing unverified remote scripts directly can be dangerous. Consider downloading first to inspect: curl -O <url>"
	}

	return ""
}

// SynthesizeAction generates a concise plain-English sentence summarizing the exact command action.
func SynthesizeAction(cmd *ast.SingleCommand, analysis *CommandAnalysis) string {
	name := cmd.Name
	argsJoined := strings.Join(cmd.Args, " ")

	switch name {
	case "cd":
		if len(analysis.PositionalArgs) == 0 {
			return "Changes the current working directory to the user's home directory (~)."
		}
		target := analysis.PositionalArgs[0]
		if target == ".." {
			return "Changes the current working directory to the parent directory (moves one level up)."
		} else if target == "~" {
			return "Changes the current working directory to the user's home directory (~)."
		} else if target == "-" {
			return "Switches the current working directory back to the previous directory ($OLDPWD)."
		} else if target == "." {
			return "Keeps the current working directory unchanged."
		}
		return fmt.Sprintf("Changes the current working directory to '%s'.", target)

	case "pwd":
		return "Prints the full absolute path of your current working directory."

	case "echo":
		if len(analysis.PositionalArgs) > 0 {
			return fmt.Sprintf("Prints \"%s\" to standard output.", strings.Join(analysis.PositionalArgs, " "))
		}
		return "Prints an empty newline."

	case "clear":
		return "Clears all text from the terminal screen."

	case "cat":
		if len(analysis.PositionalArgs) > 0 {
			return fmt.Sprintf("Outputs the contents of '%s' to standard output.", strings.Join(analysis.PositionalArgs, ", "))
		}
		return "Reads text from standard input and prints it to standard output."

	case "head":
		lines := "10"
		if val := extractFlagVal(analysis, "-n", "--lines"); val != "" {
			lines = val
		}
		if len(analysis.PositionalArgs) > 0 {
			return fmt.Sprintf("Prints the first %s lines of '%s'.", lines, strings.Join(analysis.PositionalArgs, ", "))
		}
		return fmt.Sprintf("Prints the first %s lines of input.", lines)

	case "tail":
		lines := "10"
		if val := extractFlagVal(analysis, "-n", "--lines"); val != "" {
			lines = val
		}
		isFollow := hasAnyFlag(analysis, "-f", "--follow")
		if len(analysis.PositionalArgs) > 0 {
			if isFollow {
				return fmt.Sprintf("Monitors and streams new lines added to '%s' in real time.", strings.Join(analysis.PositionalArgs, ", "))
			}
			return fmt.Sprintf("Prints the last %s lines of '%s'.", lines, strings.Join(analysis.PositionalArgs, ", "))
		}
		return fmt.Sprintf("Prints the last %s lines of input.", lines)

	case "export":
		if len(analysis.PositionalArgs) > 0 {
			return fmt.Sprintf("Sets environment variable(s): %s.", strings.Join(analysis.PositionalArgs, ", "))
		}
		return "Lists all exported shell environment variables."

	case "which":
		if len(analysis.PositionalArgs) > 0 {
			return fmt.Sprintf("Locates executable path for command '%s'.", strings.Join(analysis.PositionalArgs, ", "))
		}
		return "Locates the executable file for a command."

	case "source", ".":
		if len(analysis.PositionalArgs) > 0 {
			return fmt.Sprintf("Executes script '%s' inside the current shell environment.", analysis.PositionalArgs[0])
		}
		return "Executes commands from a script file in current shell environment."

	case "alias":
		if len(analysis.PositionalArgs) > 0 {
			return fmt.Sprintf("Defines command shortcut: %s.", strings.Join(analysis.PositionalArgs, ", "))
		}
		return "Lists all currently active shell aliases."

	case "tar":
		hasExtract := hasAnyFlag(analysis, "-x", "--extract")
		hasCreate := hasAnyFlag(analysis, "-c", "--create")
		hasGzip := hasAnyFlag(analysis, "-z", "--gzip")
		hasBzip2 := hasAnyFlag(analysis, "-j", "--bzip2")
		hasXz := hasAnyFlag(analysis, "-J", "--xz")
		hasList := hasAnyFlag(analysis, "-t", "--list")

		archiveName := findArchiveArg(cmd.Args)
		compType := "archive"
		if hasGzip {
			compType = "gzip archive (.tar.gz)"
		} else if hasBzip2 {
			compType = "bzip2 archive (.tar.bz2)"
		} else if hasXz {
			compType = "xz archive (.tar.xz)"
		}

		if hasExtract {
			if archiveName != "" {
				return fmt.Sprintf("Extracts files from %s '%s' into current directory.", compType, archiveName)
			}
			return fmt.Sprintf("Extracts files from %s.", compType)
		}
		if hasCreate {
			targets := findNonArchiveArgs(cmd.Args)
			if archiveName != "" && len(targets) > 0 {
				return fmt.Sprintf("Bundles %s into new %s '%s'.", strings.Join(targets, ", "), compType, archiveName)
			}
			if archiveName != "" {
				return fmt.Sprintf("Creates new %s named '%s'.", compType, archiveName)
			}
			return fmt.Sprintf("Creates a new %s.", compType)
		}
		if hasList {
			if archiveName != "" {
				return fmt.Sprintf("Lists contents of archive '%s' without extracting.", archiveName)
			}
			return "Lists archive contents without extracting."
		}

	case "rm":
		isRecursive := hasAnyFlag(analysis, "-r", "-R", "--recursive")
		isForce := hasAnyFlag(analysis, "-f", "--force")
		target := strings.Join(analysis.PositionalArgs, ", ")
		if target == "" {
			target = "specified path"
		}

		if isRecursive && isForce {
			return fmt.Sprintf("Recursively and forcefully deletes '%s' without prompt.", target)
		}
		if isRecursive {
			return fmt.Sprintf("Recursively deletes '%s' and all contents inside.", target)
		}
		if isForce {
			return fmt.Sprintf("Permanently deletes '%s' without prompting.", target)
		}
		return fmt.Sprintf("Permanently deletes '%s'.", target)

	case "cp":
		isRecursive := hasAnyFlag(analysis, "-r", "-R", "--recursive", "-a", "--archive")
		if len(analysis.PositionalArgs) >= 2 {
			src := strings.Join(analysis.PositionalArgs[:len(analysis.PositionalArgs)-1], ", ")
			dst := analysis.PositionalArgs[len(analysis.PositionalArgs)-1]
			if isRecursive {
				return fmt.Sprintf("Recursively copies '%s' to destination '%s'.", src, dst)
			}
			return fmt.Sprintf("Copies '%s' to '%s'.", src, dst)
		}

	case "mv":
		if len(analysis.PositionalArgs) >= 2 {
			src := strings.Join(analysis.PositionalArgs[:len(analysis.PositionalArgs)-1], ", ")
			dst := analysis.PositionalArgs[len(analysis.PositionalArgs)-1]
			return fmt.Sprintf("Moves or renames '%s' to '%s'.", src, dst)
		}

	case "ls":
		details := []string{}
		if hasAnyFlag(analysis, "-a", "--all") {
			details = append(details, "hidden files included")
		}
		if hasAnyFlag(analysis, "-l") {
			details = append(details, "detailed list format")
		}
		if hasAnyFlag(analysis, "-h", "--human-readable") {
			details = append(details, "human-readable sizes")
		}
		if hasAnyFlag(analysis, "-t") {
			details = append(details, "sorted by date")
		}
		target := "current directory"
		if len(analysis.PositionalArgs) > 0 {
			target = "'" + strings.Join(analysis.PositionalArgs, ", ") + "'"
		}
		if len(details) > 0 {
			return fmt.Sprintf("Lists files in %s (%s).", target, strings.Join(details, ", "))
		}
		return fmt.Sprintf("Lists files and folders in %s.", target)

	case "chmod":
		isRecursive := hasAnyFlag(analysis, "-R", "--recursive")
		if len(analysis.PositionalArgs) >= 2 {
			mode := analysis.PositionalArgs[0]
			target := strings.Join(analysis.PositionalArgs[1:], ", ")
			if isRecursive {
				return fmt.Sprintf("Recursively changes permissions of '%s' to '%s'.", target, mode)
			}
			return fmt.Sprintf("Changes permissions of '%s' to '%s'.", target, mode)
		}

	case "chown":
		isRecursive := hasAnyFlag(analysis, "-R", "--recursive")
		if len(analysis.PositionalArgs) >= 2 {
			owner := analysis.PositionalArgs[0]
			target := strings.Join(analysis.PositionalArgs[1:], ", ")
			if isRecursive {
				return fmt.Sprintf("Recursively changes owner of '%s' to '%s'.", target, owner)
			}
			return fmt.Sprintf("Changes owner of '%s' to '%s'.", target, owner)
		}

	case "grep":
		isRecursive := hasAnyFlag(analysis, "-r", "-R", "--recursive")
		isIgnore := hasAnyFlag(analysis, "-i", "--ignore-case")
		pattern := ""
		target := "input text"
		if len(analysis.PositionalArgs) >= 1 {
			pattern = analysis.PositionalArgs[0]
		}
		if len(analysis.PositionalArgs) >= 2 {
			target = strings.Join(analysis.PositionalArgs[1:], ", ")
		}

		desc := "Searches for"
		if isIgnore {
			desc += " pattern (case-insensitive)"
		} else {
			desc += " pattern"
		}
		if pattern != "" {
			desc += fmt.Sprintf(" '%s'", pattern)
		}
		if isRecursive {
			desc += fmt.Sprintf(" recursively inside '%s'.", target)
		} else {
			desc += fmt.Sprintf(" in %s.", target)
		}
		return desc

	case "find":
		startPath := "."
		if len(analysis.PositionalArgs) > 0 && !strings.HasPrefix(analysis.PositionalArgs[0], "-") {
			startPath = analysis.PositionalArgs[0]
		}
		hasDelete := strings.Contains(argsJoined, "-delete")
		if hasDelete {
			return fmt.Sprintf("Searches '%s' and deletes matching files.", startPath)
		}
		return fmt.Sprintf("Searches '%s' and lists matching file/folder paths.", startPath)

	case "curl":
		urls := []string{}
		for _, arg := range analysis.PositionalArgs {
			if strings.HasPrefix(arg, "http://") || strings.HasPrefix(arg, "https://") {
				urls = append(urls, arg)
			}
		}
		if len(urls) > 0 {
			if hasAnyFlag(analysis, "-O", "--remote-name") {
				return fmt.Sprintf("Downloads file from '%s' and saves locally.", urls[0])
			}
			if hasAnyFlag(analysis, "-I", "--head") {
				return fmt.Sprintf("Fetches HTTP response headers from '%s'.", urls[0])
			}
			return fmt.Sprintf("Sends HTTP GET request to '%s' and outputs response.", urls[0])
		}

	case "awk":
		if len(analysis.PositionalArgs) > 0 {
			return fmt.Sprintf("Processes input text using AWK expression: %s", analysis.PositionalArgs[0])
		}
		return "Scans and processes structured column data from input text."

	case "sed":
		if len(analysis.PositionalArgs) > 0 {
			return fmt.Sprintf("Transforms input text using sed expression: %s", analysis.PositionalArgs[0])
		}
		return "Performs automated text transformations on input streams."

	case "xargs":
		return "Converts standard input lines into arguments and runs target command."

	case "systemctl":
		if analysis.Subcommand != "" {
			unit := ""
			if len(analysis.PositionalArgs) > 0 {
				unit = analysis.PositionalArgs[0]
			}
			if unit != "" {
				return fmt.Sprintf("%s systemd service unit '%s'.", strings.Title(analysis.Subcommand), unit)
			}
			return fmt.Sprintf("Runs systemctl %s command.", analysis.Subcommand)
		}

	case "journalctl":
		if hasAnyFlag(analysis, "-u", "--unit") {
			unit := extractFlagVal(analysis, "-u", "--unit")
			return fmt.Sprintf("Fetches systemd log entries for service unit '%s'.", unit)
		}
		return "Queries and views systemd log messages."

	case "rsync":
		if len(analysis.PositionalArgs) >= 2 {
			src := analysis.PositionalArgs[0]
			dst := analysis.PositionalArgs[1]
			return fmt.Sprintf("Synchronizes files efficiently between '%s' and '%s'.", src, dst)
		}
		return "Synchronizes file structures between locations."

	case "git":
		if analysis.Subcommand != "" {
			switch analysis.Subcommand {
			case "commit":
				msg := extractFlagVal(analysis, "-m", "--message")
				if msg != "" {
					return fmt.Sprintf("Commits staged changes with message: \"%s\".", msg)
				}
				return "Commits staged repository changes."
			case "push":
				return "Uploads local commits to remote git repository."
			case "pull":
				return "Fetches latest remote changes and merges into current branch."
			case "checkout":
				if len(analysis.PositionalArgs) > 0 {
					return fmt.Sprintf("Switches working branch/files to '%s'.", analysis.PositionalArgs[0])
				}
				return "Switches git branch or checks out files."
			case "status":
				return "Shows current working tree status and modified files."
			}
		}

	case "docker":
		if analysis.Subcommand == "run" {
			img := ""
			if len(analysis.PositionalArgs) > 0 {
				img = analysis.PositionalArgs[len(analysis.PositionalArgs)-1]
			}
			if img != "" {
				return fmt.Sprintf("Creates and launches container from image '%s'.", img)
			}
			return "Creates and runs a new Docker container."
		} else if analysis.Subcommand == "ps" {
			return "Lists running Docker containers."
		} else if analysis.Subcommand == "exec" {
			return "Executes command inside a running Docker container."
		}

	case "kubectl":
		if analysis.Subcommand != "" {
			resource := ""
			if len(analysis.PositionalArgs) > 0 {
				resource = analysis.PositionalArgs[0]
			}
			if resource != "" {
				return fmt.Sprintf("Executes kubectl %s on Kubernetes resource '%s'.", analysis.Subcommand, resource)
			}
			return fmt.Sprintf("Executes kubectl %s command.", analysis.Subcommand)
		}

	case "ps":
		if len(analysis.PositionalArgs) > 0 {
			return fmt.Sprintf("Reports snapshots of running processes matching '%s'.", strings.Join(analysis.PositionalArgs, ", "))
		}
		return "Reports a snapshot of active system processes."

	case "npm":
		if analysis.Subcommand != "" {
			return fmt.Sprintf("Runs Node package manager action: npm %s.", analysis.Subcommand)
		}
	}

	if analysis.CommandSummary != "" {
		return fmt.Sprintf("Executes '%s' (%s).", name, lowerFirstChar(analysis.CommandSummary))
	}

	return fmt.Sprintf("Executes '%s' with specified options.", name)
}

func hasCmd(list []string, target string) bool {
	for _, s := range list {
		if s == target {
			return true
		}
	}
	return false
}

func hasAnyFlag(analysis *CommandAnalysis, flags ...string) bool {
	for _, item := range analysis.Items {
		if !item.IsFlag {
			continue
		}
		for _, f := range flags {
			if item.Token == f || strings.HasPrefix(item.Label, f) {
				return true
			}
		}
	}
	return false
}

func extractFlagVal(analysis *CommandAnalysis, flags ...string) string {
	for _, item := range analysis.Items {
		if !item.IsFlag {
			continue
		}
		for _, f := range flags {
			if strings.HasPrefix(item.Label, f) && strings.Contains(item.Description, "(") {
				start := strings.LastIndex(item.Description, "(")
				end := strings.LastIndex(item.Description, ")")
				if start != -1 && end > start {
					return item.Description[start+1 : end]
				}
			}
		}
	}
	return ""
}

func findArchiveArg(args []string) string {
	for _, a := range args {
		if strings.HasSuffix(a, ".tar") || strings.HasSuffix(a, ".tar.gz") || strings.HasSuffix(a, ".tgz") || strings.HasSuffix(a, ".tar.bz2") || strings.HasSuffix(a, ".tar.xz") {
			return a
		}
	}
	return ""
}

func findNonArchiveArgs(args []string) []string {
	var res []string
	for _, a := range args {
		if strings.HasPrefix(a, "-") {
			continue
		}
		if strings.HasSuffix(a, ".tar") || strings.HasSuffix(a, ".tar.gz") || strings.HasSuffix(a, ".tgz") || strings.HasSuffix(a, ".tar.bz2") || strings.HasSuffix(a, ".tar.xz") {
			continue
		}
		res = append(res, a)
	}
	return res
}

func lowerFirstChar(s string) string {
	if len(s) == 0 {
		return s
	}
	runes := []rune(s)
	if runes[0] >= 'A' && runes[0] <= 'Z' {
		runes[0] = runes[0] + ('a' - 'A')
	}
	return string(runes)
}
