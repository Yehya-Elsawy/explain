package analyzer

import (
	"fmt"
	"strings"

	"github.com/elsawy/explain/pkg/ast"
)

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
		return "Prints the full absolute path of the directory you are currently located in."

	case "echo":
		if len(analysis.PositionalArgs) > 0 {
			return fmt.Sprintf("Prints \"%s\" to standard output.", strings.Join(analysis.PositionalArgs, " "))
		}
		return "Prints an empty newline to standard output."

	case "clear":
		return "Clears all text and output from the current terminal window."

	case "export":
		if len(analysis.PositionalArgs) > 0 {
			return fmt.Sprintf("Sets and exports environment variable(s): %s.", strings.Join(analysis.PositionalArgs, ", "))
		}
		return "Lists all exported shell environment variables."

	case "which":
		if len(analysis.PositionalArgs) > 0 {
			return fmt.Sprintf("Locates and prints the executable binary path for '%s'.", strings.Join(analysis.PositionalArgs, ", "))
		}
		return "Locates the executable file associated with the given command."

	case "source", ".":
		if len(analysis.PositionalArgs) > 0 {
			return fmt.Sprintf("Reads and executes commands from '%s' directly in the current shell session.", analysis.PositionalArgs[0])
		}
		return "Executes shell commands from a script file in the current shell environment."

	case "alias":
		if len(analysis.PositionalArgs) > 0 {
			return fmt.Sprintf("Defines command shortcut(s): %s.", strings.Join(analysis.PositionalArgs, ", "))
		}
		return "Prints all currently defined shell aliases."

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
			compType = "gzip-compressed archive (.tar.gz)"
		} else if hasBzip2 {
			compType = "bzip2-compressed archive (.tar.bz2)"
		} else if hasXz {
			compType = "xz-compressed archive (.tar.xz)"
		}

		if hasExtract {
			if archiveName != "" {
				return fmt.Sprintf("Extracts files from %s '%s' into the current directory.", compType, archiveName)
			}
			return fmt.Sprintf("Extracts files from %s into the current directory.", compType)
		}
		if hasCreate {
			targets := findNonArchiveArgs(cmd.Args)
			if archiveName != "" && len(targets) > 0 {
				return fmt.Sprintf("Creates a new %s named '%s' bundling %s.", compType, archiveName, strings.Join(targets, ", "))
			}
			if archiveName != "" {
				return fmt.Sprintf("Creates a new %s named '%s'.", compType, archiveName)
			}
			return fmt.Sprintf("Creates a new %s.", compType)
		}
		if hasList {
			if archiveName != "" {
				return fmt.Sprintf("Lists the contents of archive '%s' without extracting.", archiveName)
			}
			return "Lists the contents of the archive without extracting."
		}

	case "rm":
		isRecursive := hasAnyFlag(analysis, "-r", "-R", "--recursive")
		isForce := hasAnyFlag(analysis, "-f", "--force")
		target := strings.Join(analysis.PositionalArgs, ", ")
		if target == "" {
			target = "specified path"
		}

		if isRecursive && isForce {
			return fmt.Sprintf("Recursively and permanently deletes '%s' without asking for confirmation.", target)
		}
		if isRecursive {
			return fmt.Sprintf("Recursively deletes '%s' and all files/directories inside it.", target)
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
			details = append(details, "including hidden files")
		}
		if hasAnyFlag(analysis, "-l") {
			details = append(details, "in detailed table format (permissions, size, date)")
		}
		if hasAnyFlag(analysis, "-h", "--human-readable") {
			details = append(details, "with human-readable sizes")
		}
		if hasAnyFlag(analysis, "-t") {
			details = append(details, "sorted by modification date")
		}
		target := "the current directory"
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
				return fmt.Sprintf("Recursively changes permissions of '%s' to mode '%s'.", target, mode)
			}
			return fmt.Sprintf("Changes permissions of '%s' to mode '%s'.", target, mode)
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
		target := "standard input"
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
			desc += fmt.Sprintf(" recursively across '%s'.", target)
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
			return fmt.Sprintf("Searches directory '%s' matching specified filters and deletes matching files.", startPath)
		}
		return fmt.Sprintf("Searches directory '%s' and prints matching file/folder paths.", startPath)

	case "curl":
		urls := []string{}
		for _, arg := range analysis.PositionalArgs {
			if strings.HasPrefix(arg, "http://") || strings.HasPrefix(arg, "https://") {
				urls = append(urls, arg)
			}
		}
		if len(urls) > 0 {
			if hasAnyFlag(analysis, "-O", "--remote-name") {
				return fmt.Sprintf("Downloads file from '%s' and saves it locally.", urls[0])
			}
			if hasAnyFlag(analysis, "-I", "--head") {
				return fmt.Sprintf("Fetches HTTP response headers from '%s'.", urls[0])
			}
			return fmt.Sprintf("Sends HTTP request to '%s' and prints response body to stdout.", urls[0])
		}

	case "dd":
		return "Performs low-level byte stream copy between specified input and output devices/files."

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
				return "Uploads committed changes to the remote git repository."
			case "pull":
				return "Fetches latest changes from remote and merges them into current branch."
			case "checkout":
				if len(analysis.PositionalArgs) > 0 {
					return fmt.Sprintf("Switches working branch or file state to '%s'.", analysis.PositionalArgs[0])
				}
				return "Switches git branch or checks out files."
			}
		}

	case "docker":
		if analysis.Subcommand == "run" {
			img := ""
			if len(analysis.PositionalArgs) > 0 {
				img = analysis.PositionalArgs[len(analysis.PositionalArgs)-1]
			}
			if img != "" {
				return fmt.Sprintf("Creates and runs a new Docker container from image '%s'.", img)
			}
			return "Creates and runs a new Docker container."
		}
	}

	if analysis.CommandSummary != "" {
		return fmt.Sprintf("Executes %s with the provided arguments.", name)
	}

	return fmt.Sprintf("Runs the '%s' command.", name)
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
			if strings.HasPrefix(item.Label, f) && strings.Contains(item.Description, "Value: ") {
				parts := strings.Split(item.Description, "Value: ")
				if len(parts) > 1 {
					return strings.TrimSuffix(parts[1], ")")
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
