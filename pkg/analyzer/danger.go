package analyzer

import (
	"strings"

	"github.com/Yehya-Elsawy/explain-/pkg/ast"
	"github.com/Yehya-Elsawy/explain-/pkg/database"
)

// DangerInfo contains the safety rating and warning messages for a command.
type DangerInfo struct {
	Level   database.RiskLevel
	Badge   string
	Reason  string
	Warning string
}

// EvaluateDanger inspects command structure, arguments, and flags for hazards.
func EvaluateDanger(cmd *ast.SingleCommand, analysis *CommandAnalysis) DangerInfo {
	name := cmd.Name
	argsJoined := strings.Join(cmd.Args, " ")

	// 1. Root / Critical RM checks
	if name == "rm" {
		hasRecursive := false
		hasForce := false
		hasRoot := false

		for _, item := range analysis.Items {
			if item.Token == "-r" || item.Token == "-R" || item.Token == "--recursive" {
				hasRecursive = true
			}
			if item.Token == "-f" || item.Token == "--force" {
				hasForce = true
			}
		}

		for _, arg := range cmd.Args {
			if arg == "/" || arg == "/*" || arg == "~" || arg == "*" || strings.HasPrefix(arg, "/etc") || strings.HasPrefix(arg, "/boot") {
				hasRoot = true
			}
		}

		if hasRecursive && (hasRoot || strings.Contains(argsJoined, "--no-preserve-root")) {
			return DangerInfo{
				Level:   database.RiskCritical,
				Badge:   "🚨 CRITICAL DANGER",
				Reason:  "Destroys entire filesystem or critical system directories irreversibly.",
				Warning: "DO NOT RUN THIS! This will destroy your operating system and all files.",
			}
		}

		if hasRecursive && hasForce {
			return DangerInfo{
				Level:   database.RiskCritical,
				Badge:   "🚨 HIGH RISK",
				Reason:  "Recursively and forcefully deletes directories and files without prompting.",
				Warning: "Deleted files bypass the Trash bin and are permanently lost.",
			}
		}

		if hasRecursive {
			return DangerInfo{
				Level:   database.RiskHigh,
				Badge:   "⚠️  HIGH RISK",
				Reason:  "Recursively removes directories and all their sub-contents.",
				Warning: "Ensure you double-check the target path before running.",
			}
		}

		return DangerInfo{
			Level:   database.RiskMedium,
			Badge:   "⚠️  MODERATE RISK",
			Reason:  "Permanently deletes specified file(s).",
			Warning: "Deleted files cannot be recovered from the terminal.",
		}
	}

	// 2. Low-level device writes / Disk formatting (dd, mkfs, fdisk)
	if name == "dd" {
		if strings.Contains(argsJoined, "of=/dev/") {
			return DangerInfo{
				Level:   database.RiskCritical,
				Badge:   "🚨 CRITICAL DANGER",
				Reason:  "Writes raw data directly to a disk device or partition.",
				Warning: "An incorrect 'of=' target will completely overwrite your disk and destroy data.",
			}
		}
		return DangerInfo{
			Level:   database.RiskHigh,
			Badge:   "⚠️  HIGH RISK",
			Reason:  "Direct low-level byte stream copy.",
			Warning: "Verify input (if=) and output (of=) file paths carefully.",
		}
	}

	if strings.HasPrefix(name, "mkfs") || name == "fdisk" || name == "parted" || name == "gdisk" {
		return DangerInfo{
			Level:   database.RiskCritical,
			Badge:   "🚨 CRITICAL DANGER",
			Reason:  "Modifies disk partition tables or creates new filesystems (formats drives).",
			Warning: "Formatting a partition permanently erases all existing data on it.",
		}
	}

	// 3. Permissive Chmod (chmod 777)
	if name == "chmod" {
		if strings.Contains(argsJoined, "777") || strings.Contains(argsJoined, "a+rwx") {
			return DangerInfo{
				Level:   database.RiskHigh,
				Badge:   "⚠️  SECURITY RISK",
				Reason:  "Grants full read, write, and execute permissions to all users on system.",
				Warning: "777 makes files vulnerable to unauthorized modification by any process or user.",
			}
		}
	}

	// 4. Force Git operations
	if name == "git" {
		if strings.Contains(argsJoined, "--force") || strings.Contains(argsJoined, "-f") {
			return DangerInfo{
				Level:   database.RiskHigh,
				Badge:   "⚠️  HIGH RISK",
				Reason:  "Force push overwrites remote branch history.",
				Warning: "May erase commits made by other team members or your own past work.",
			}
		}
		if strings.Contains(argsJoined, "--hard") {
			return DangerInfo{
				Level:   database.RiskHigh,
				Badge:   "⚠️  HIGH RISK",
				Reason:  "Hard reset permanently discards all uncommitted working directory changes.",
				Warning: "All uncommitted changes will be lost permanently.",
			}
		}
	}

	// 5. Mass Process Killing
	if name == "kill" || name == "killall" || name == "pkill" {
		if strings.Contains(argsJoined, "-9") || strings.Contains(argsJoined, "KILL") {
			return DangerInfo{
				Level:   database.RiskHigh,
				Badge:   "⚠️  FORCE TERMINATION",
				Reason:  "Sends uncatchable SIGKILL signal to immediately kill target process.",
				Warning: "The process will not have a chance to save open files or close gracefully.",
			}
		}
	}

	// 6. Output Redirects to Raw Devices or Files
	for _, r := range cmd.Redirects {
		if strings.HasPrefix(r.Target, "/dev/sd") || strings.HasPrefix(r.Target, "/dev/nvme") {
			return DangerInfo{
				Level:   database.RiskCritical,
				Badge:   "🚨 CRITICAL DANGER",
				Reason:  "Redirecting output directly to a raw disk device.",
				Warning: "Will corrupt partition table and destroy stored filesystem data.",
			}
		}
		if r.Operator == ">" {
			return DangerInfo{
				Level:   database.RiskMedium,
				Badge:   "⚠️  OVERWRITE",
				Reason:  "The '>' operator will truncate (overwrite) target file if it already exists.",
				Warning: "Use '>>' if you intended to append instead of overwriting.",
			}
		}
	}

	// 7. Rsync --delete
	if name == "rsync" && strings.Contains(argsJoined, "--delete") {
		return DangerInfo{
			Level:   database.RiskHigh,
			Badge:   "⚠️  DELETION RISK",
			Reason:  "Deletes files in destination folder that do not exist in source folder.",
			Warning: "Test with '-n' (--dry-run) first to avoid accidental deletions.",
		}
	}

	// 8. Find -delete / -exec rm
	if name == "find" && (strings.Contains(argsJoined, "-delete") || strings.Contains(argsJoined, "-exec rm")) {
		return DangerInfo{
			Level:   database.RiskHigh,
			Badge:   "⚠️  DELETION RISK",
			Reason:  "Automatically deletes every file matching the search conditions.",
			Warning: "Run without '-delete' first to review the list of matched files.",
		}
	}

	// 9. Default Command Risk from Database
	if def, ok := database.BuiltinCommands[name]; ok {
		switch def.DefaultRisk {
		case database.RiskSafe:
			return DangerInfo{Level: database.RiskSafe, Badge: "🟢 SAFE", Reason: "Read-only operation; does not modify system files."}
		case database.RiskLow:
			return DangerInfo{Level: database.RiskLow, Badge: "🟢 LOW RISK", Reason: "Creates or inspects files safely."}
		case database.RiskMedium:
			return DangerInfo{Level: database.RiskMedium, Badge: "🟡 MODERATE", Reason: "Modifies files or service runtime state."}
		case database.RiskHigh:
			return DangerInfo{Level: database.RiskHigh, Badge: "⚠️  HIGH RISK", Reason: "Modifies critical configurations or terminates processes."}
		case database.RiskCritical:
			return DangerInfo{Level: database.RiskCritical, Badge: "🚨 CRITICAL", Reason: "High potential for data loss or system impact."}
		}
	}

	return DangerInfo{Level: database.RiskSafe, Badge: "🟢 SAFE", Reason: "Standard command execution."}
}

// CheckPipelineDangers checks for dangerous combinations like `curl ... | bash` or `cat ... > ...`
func CheckPipelineDangers(pipe *PipelineAnalysis) {
	if len(pipe.Commands) < 2 {
		return
	}

	for i := 0; i < len(pipe.Commands)-1; i++ {
		first := pipe.Commands[i]
		second := pipe.Commands[i+1]

		// curl/wget piped to bash/sh/zsh/python/perl
		if (first.CommandName == "curl" || first.CommandName == "wget") && (second.CommandName == "bash" || second.CommandName == "sh" || second.CommandName == "zsh" || second.CommandName == "sudo" || second.CommandName == "python3" || second.CommandName == "python") {
			pipe.MaxRisk = database.RiskCritical
			second.Danger = DangerInfo{
				Level:   database.RiskCritical,
				Badge:   "🚨 REMOTE CODE EXECUTION",
				Reason:  "Piping remote web scripts directly into a shell interpreter.",
				Warning: "Executes unverified remote code immediately. Inspect the script content before running!",
			}
		}
	}
}
