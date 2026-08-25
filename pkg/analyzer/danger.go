package analyzer

import (
	"strings"

	"github.com/Yehya-Elsawy/explain/pkg/ast"
	"github.com/Yehya-Elsawy/explain/pkg/database"
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
				Badge:   "CRITICAL DANGER",
				Reason:  "Destroys system or root directories irreversibly.",
				Warning: "DO NOT RUN THIS! This will permanently erase your operating system and files.",
			}
		}

		if hasRecursive && hasForce {
			return DangerInfo{
				Level:   database.RiskCritical,
				Badge:   "HIGH RISK DELETION",
				Reason:  "Recursively and forcefully deletes directories and files without prompting.",
				Warning: "Deleted items bypass the Trash bin and are permanently lost.",
			}
		}

		if hasRecursive {
			return DangerInfo{
				Level:   database.RiskHigh,
				Badge:   "HIGH RISK DELETION",
				Reason:  "Recursively removes directories and all sub-contents.",
				Warning: "Double-check the target path before running.",
			}
		}

		return DangerInfo{
			Level:   database.RiskMedium,
			Badge:   "MODERATE DELETION",
			Reason:  "Permanently deletes specified file(s).",
			Warning: "Deleted terminal files cannot be recovered.",
		}
	}

	// 2. Low-level device writes / Disk formatting (dd, mkfs, fdisk)
	if name == "dd" {
		if strings.Contains(argsJoined, "of=/dev/") {
			return DangerInfo{
				Level:   database.RiskCritical,
				Badge:   "CRITICAL DANGER",
				Reason:  "Writes raw data directly to a disk device or partition.",
				Warning: "Target 'of=/dev/...' will completely overwrite disk data and destroy partitions.",
			}
		}
		return DangerInfo{
			Level:   database.RiskHigh,
			Badge:   "HIGH RISK WRITE",
			Reason:  "Direct low-level byte stream copy to destination.",
			Warning: "Verify input (if=) and output (of=) file paths carefully.",
		}
	}

	if strings.HasPrefix(name, "mkfs") || name == "fdisk" || name == "parted" || name == "gdisk" {
		return DangerInfo{
			Level:   database.RiskCritical,
			Badge:   "CRITICAL DISK OPERATION",
			Reason:  "Modifies disk partition tables or formats filesystems.",
			Warning: "Formatting a partition permanently erases all existing data on it.",
		}
	}

	// 3. Permissive Chmod (chmod 777)
	if name == "chmod" {
		if strings.Contains(argsJoined, "777") || strings.Contains(argsJoined, "a+rwx") {
			return DangerInfo{
				Level:   database.RiskHigh,
				Badge:   "SECURITY RISK",
				Reason:  "Grants full read, write, and execute access to all users on the system.",
				Warning: "777 permissions leave files open to unauthorized modification by any process or user.",
			}
		}
	}

	// 4. Force Git operations
	if name == "git" {
		if strings.Contains(argsJoined, "--force") || strings.Contains(argsJoined, "-f") {
			return DangerInfo{
				Level:   database.RiskHigh,
				Badge:   "FORCE OVERWRITE",
				Reason:  "Force push overwrites remote branch commit history.",
				Warning: "May permanently erase commits made by teammates or your past work.",
			}
		}
		if strings.Contains(argsJoined, "--hard") {
			return DangerInfo{
				Level:   database.RiskHigh,
				Badge:   "HARD RESET",
				Reason:  "Hard reset discards all uncommitted working directory changes.",
				Warning: "All uncommitted code changes will be lost permanently.",
			}
		}
	}

	// 5. Mass Process Killing
	if name == "kill" || name == "killall" || name == "pkill" {
		if strings.Contains(argsJoined, "-9") || strings.Contains(argsJoined, "KILL") {
			return DangerInfo{
				Level:   database.RiskHigh,
				Badge:   "FORCE TERMINATION",
				Reason:  "Sends uncatchable SIGKILL signal to immediately stop target process.",
				Warning: "The process will terminate immediately without saving open files.",
			}
		}
	}

	// 6. Output Redirects to Raw Devices or Files
	for _, r := range cmd.Redirects {
		if strings.HasPrefix(r.Target, "/dev/sd") || strings.HasPrefix(r.Target, "/dev/nvme") {
			return DangerInfo{
				Level:   database.RiskCritical,
				Badge:   "CRITICAL DEVICE WRITE",
				Reason:  "Redirecting raw output directly into a storage device.",
				Warning: "Will corrupt partition tables and destroy filesystem data.",
			}
		}
		if r.Operator == ">" {
			return DangerInfo{
				Level:   database.RiskMedium,
				Badge:   "FILE OVERWRITE",
				Reason:  "The '>' operator truncates (overwrites) the target file if it exists.",
				Warning: "Use '>>' if you intended to append data instead of overwriting.",
			}
		}
	}

	// 7. Rsync --delete
	if name == "rsync" && strings.Contains(argsJoined, "--delete") {
		return DangerInfo{
			Level:   database.RiskHigh,
			Badge:   "DELETION RISK",
			Reason:  "Deletes files in destination folder that do not exist in source folder.",
			Warning: "Test with '-n' (--dry-run) first to avoid unexpected deletions.",
		}
	}

	// 8. Find -delete / -exec rm
	if name == "find" && (strings.Contains(argsJoined, "-delete") || strings.Contains(argsJoined, "-exec rm")) {
		return DangerInfo{
			Level:   database.RiskHigh,
			Badge:   "AUTOMATIC DELETION",
			Reason:  "Automatically deletes every file matching search conditions.",
			Warning: "Run find without '-delete' first to review matched files.",
		}
	}

	// 9. Specific handling for common benign commands (tar, ls, cd, pwd, cat)
	if name == "tar" && !strings.Contains(argsJoined, " /") {
		return DangerInfo{Level: database.RiskLow, Badge: "LOW RISK", Reason: "Bundles or extracts files locally."}
	}

	// 10. Default Command Risk from Database
	if def, ok := database.BuiltinCommands[name]; ok {
		switch def.DefaultRisk {
		case database.RiskSafe:
			return DangerInfo{Level: database.RiskSafe, Badge: "SAFE TO RUN", Reason: "Read-only operation; safe to run."}
		case database.RiskLow:
			return DangerInfo{Level: database.RiskLow, Badge: "LOW RISK", Reason: "Inspects or creates files safely."}
		case database.RiskMedium:
			return DangerInfo{Level: database.RiskMedium, Badge: "MODERATE RISK", Reason: "Modifies files or system runtime state."}
		case database.RiskHigh:
			return DangerInfo{Level: database.RiskHigh, Badge: "HIGH RISK", Reason: "Modifies system configurations or process state."}
		case database.RiskCritical:
			return DangerInfo{Level: database.RiskCritical, Badge: "CRITICAL DANGER", Reason: "High risk of data loss or system impact."}
		}
	}

	return DangerInfo{Level: database.RiskSafe, Badge: "SAFE TO RUN", Reason: "Standard command execution."}
}

// CheckPipelineDangers checks for dangerous combinations like `curl ... | bash`
func CheckPipelineDangers(pipe *PipelineAnalysis) {
	if len(pipe.Commands) < 2 {
		return
	}

	for i := 0; i < len(pipe.Commands)-1; i++ {
		first := pipe.Commands[i]
		second := pipe.Commands[i+1]

		if (first.CommandName == "curl" || first.CommandName == "wget") && (second.CommandName == "bash" || second.CommandName == "sh" || second.CommandName == "zsh" || second.CommandName == "sudo" || second.CommandName == "python3" || second.CommandName == "python") {
			pipe.MaxRisk = database.RiskCritical
			second.Danger = DangerInfo{
				Level:   database.RiskCritical,
				Badge:   "REMOTE CODE EXECUTION",
				Reason:  "Piping remote web scripts directly into a shell interpreter.",
				Warning: "Executes unverified remote code immediately. Always inspect script content before running!",
			}
		}
	}
}
