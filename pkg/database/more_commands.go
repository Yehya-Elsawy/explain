package database

func init() {
	extraCommands := map[string]CommandDef{
		"cd": {
			Name:        "cd",
			Category:    "Navigation",
			Summary:     "Changes the current working directory in the shell",
			Description: "Shell builtin that updates the current working directory context ($PWD).",
			DefaultRisk: RiskSafe,
			Flags: map[string]FlagDef{
				"L": {Short: "-L", Description: "Follow symbolic links (default behavior)"},
				"P": {Short: "-P", Description: "Use physical directory structure without following symbolic links"},
			},
		},

		"pwd": {
			Name:        "pwd",
			Category:    "Navigation",
			Summary:     "Prints the absolute path of the current working directory",
			Description: "Displays the full directory path you are currently located in.",
			DefaultRisk: RiskSafe,
			Flags: map[string]FlagDef{
				"L": {Short: "-L", Long: "--logical", Description: "Print the logical path from environment (including symlinks)"},
				"P": {Short: "-P", Long: "--physical", Description: "Print physical directory path resolving all symbolic links"},
			},
		},

		"echo": {
			Name:        "echo",
			Category:    "Text",
			Summary:     "Prints lines of text or variable values to standard output",
			Description: "Writes arguments to standard output separated by spaces followed by a newline.",
			DefaultRisk: RiskSafe,
			Flags: map[string]FlagDef{
				"n": {Short: "-n", Description: "Do not print the trailing newline character"},
				"e": {Short: "-e", Description: "Enable interpretation of backslash escapes (e.g. \\n, \\t, \\e)"},
				"E": {Short: "-E", Description: "Disable interpretation of backslash escapes (default)"},
			},
		},

		"clear": {
			Name:        "clear",
			Category:    "System",
			Summary:     "Clears the terminal screen and scrollback buffer",
			Description: "Clears your terminal screen if possible.",
			DefaultRisk: RiskSafe,
		},

		"export": {
			Name:        "export",
			Category:    "Shell",
			Summary:     "Sets environment variables for the current shell and all child processes",
			Description: "Marks variables to be automatically passed to the environment of subsequently executed commands.",
			DefaultRisk: RiskLow,
			Flags: map[string]FlagDef{
				"p": {Short: "-p", Description: "Display list of all exported variables in reusable shell format"},
				"n": {Short: "-n", Description: "Remove export property from listed variables"},
			},
		},

		"alias": {
			Name:        "alias",
			Category:    "Shell",
			Summary:     "Defines or displays command aliases (shortcuts) in the shell",
			Description: "Creates short custom command aliases (e.g. alias ll='ls -la').",
			DefaultRisk: RiskSafe,
			Flags: map[string]FlagDef{
				"p": {Short: "-p", Description: "Print all defined aliases in reusable format"},
			},
		},

		"which": {
			Name:        "which",
			Category:    "System",
			Summary:     "Locates and prints the full executable path of a command in $PATH",
			Description: "Searches directories listed in PATH environment variable for executable matching name.",
			DefaultRisk: RiskSafe,
			Flags: map[string]FlagDef{
				"a": {Short: "-a", Description: "Print all matching pathnames in each directory of PATH (not just first)"},
			},
		},

		"source": {
			Name:        "source",
			Category:    "Shell",
			Summary:     "Executes commands from a file in the current shell environment",
			Description: "Reads and executes shell commands from the specified file directly in current shell context.",
			DefaultRisk: RiskMedium,
		},

		"cat": {
			Name:        "cat",
			Category:    "FileOps",
			Summary:     "Concatenates files and prints their contents to standard output",
			Description: "Reads files sequentially and outputs their content to stdout.",
			DefaultRisk: RiskSafe,
			Flags: map[string]FlagDef{
				"n": {Short: "-n", Long: "--number", Description: "Number all output lines starting from 1"},
				"b": {Short: "-b", Long: "--number-nonblank", Description: "Number non-empty output lines only"},
				"s": {Short: "-s", Long: "--squeeze-blank", Description: "Suppress repeated empty output lines"},
				"A": {Short: "-A", Long: "--show-all", Description: "Show non-printing characters, end-of-line markers ($), and tabs"},
			},
		},

		"head": {
			Name:        "head",
			Category:    "FileOps",
			Summary:     "Outputs the first part (default 10 lines) of files",
			Description: "Prints the initial lines or bytes of specified files.",
			DefaultRisk: RiskSafe,
			Flags: map[string]FlagDef{
				"n": {Short: "-n", Long: "--lines", TakesValue: true, ValueName: "K", Description: "Print the first K lines instead of default 10"},
				"c": {Short: "-c", Long: "--bytes", TakesValue: true, ValueName: "K", Description: "Print the first K bytes of each file"},
				"q": {Short: "-q", Long: "--quiet", Description: "Never print headers giving file names"},
				"v": {Short: "-v", Long: "--verbose", Description: "Always print headers giving file names"},
			},
		},

		"tail": {
			Name:        "tail",
			Category:    "FileOps",
			Summary:     "Outputs the last part of files (or follows logs in real time)",
			Description: "Prints the trailing lines of files, commonly used to follow active logs.",
			DefaultRisk: RiskSafe,
			Flags: map[string]FlagDef{
				"f": {Short: "-f", Long: "--follow", Description: "Follow mode: output appended data as the file grows in real time"},
				"F": {Short: "-F", Description: "Follow mode and retry if file is rotated or temporarily inaccessible"},
				"n": {Short: "-n", Long: "--lines", TakesValue: true, ValueName: "K", Description: "Output the last K lines instead of default 10 (e.g. -n 100 or -n +1)"},
				"c": {Short: "-c", Long: "--bytes", TakesValue: true, ValueName: "K", Description: "Output the last K bytes"},
			},
		},

		"awk": {
			Name:        "awk",
			Category:    "Text",
			Summary:     "Pattern scanning and text column processing language",
			Description: "Processes structured tabular text files, filters records, and extracts columns (e.g. $1, $2).",
			DefaultRisk: RiskSafe,
			Flags: map[string]FlagDef{
				"F": {Short: "-F", TakesValue: true, ValueName: "SEPARATOR", Description: "Set input field separator (e.g. -F ':' or -F ',')"},
				"v": {Short: "-v", TakesValue: true, ValueName: "VAR=VAL", Description: "Assign value to variable before execution begins"},
				"f": {Short: "-f", TakesValue: true, ValueName: "FILE", Description: "Read AWK program script from specified file"},
			},
		},

		"sed": {
			Name:        "sed",
			Category:    "Text",
			Summary:     "Stream editor for filtering and transforming text",
			Description: "Performs search, replace, deletion, and text substitutions on input streams.",
			DefaultRisk: RiskSafe,
			Flags: map[string]FlagDef{
				"i": {Short: "-i", Long: "--in-place", Description: "Edit files in-place directly on disk (OVERWRITES ORIGINAL FILE)", Risk: RiskHigh},
				"e": {Short: "-e", TakesValue: true, ValueName: "SCRIPT", Description: "Add the script/commands to the execution list"},
				"n": {Short: "-n", Long: "--quiet", Description: "Suppress automatic printing of pattern space (print only matching lines with 'p')"},
				"E": {Short: "-E", Long: "-r", Description: "Use extended regular expressions (ERE) in script"},
			},
		},

		"sort": {
			Name:        "sort",
			Category:    "Text",
			Summary:     "Sorts lines of text files alphabetically or numerically",
			Description: "Orders lines of text based on sorting keys.",
			DefaultRisk: RiskSafe,
			Flags: map[string]FlagDef{
				"n": {Short: "-n", Long: "--numeric-sort", Description: "Compare according to string numerical value"},
				"r": {Short: "-r", Long: "--reverse", Description: "Reverse the result of comparisons"},
				"u": {Short: "-u", Long: "--unique", Description: "Output only unique lines (remove duplicate lines)"},
				"k": {Short: "-k", Long: "--key", TakesValue: true, ValueName: "POS", Description: "Start a sorting key at column position POS"},
				"h": {Short: "-h", Long: "--human-numeric-sort", Description: "Compare human readable numbers (e.g. 2K 1G)"},
			},
		},

		"uniq": {
			Name:        "uniq",
			Category:    "Text",
			Summary:     "Reports or omits repeated adjacent lines",
			Description: "Filters out consecutive identical lines from input (usually paired with sort).",
			DefaultRisk: RiskSafe,
			Flags: map[string]FlagDef{
				"c": {Short: "-c", Long: "--count", Description: "Prefix lines by the number of occurrences"},
				"d": {Short: "-d", Long: "--repeated", Description: "Only print duplicate lines"},
				"u": {Short: "-u", Long: "--unique", Description: "Only print unique lines (lines that appear exactly once)"},
				"i": {Short: "-i", Long: "--ignore-case", Description: "Ignore differences in case when comparing"},
			},
		},

		"xargs": {
			Name:        "xargs",
			Category:    "System",
			Summary:     "Builds and executes command lines from standard input",
			Description: "Reads items from stdin and executes a command passing those items as arguments.",
			DefaultRisk: RiskMedium,
			Flags: map[string]FlagDef{
				"I": {Short: "-I", TakesValue: true, ValueName: "REPLACE_STR", Description: "Replace occurrences of REPLACE_STR in command with names read from stdin"},
				"n": {Short: "-n", TakesValue: true, ValueName: "MAX_ARGS", Description: "Use at most MAX_ARGS arguments per command line"},
				"P": {Short: "-P", TakesValue: true, ValueName: "MAX_PROCS", Description: "Run up to MAX_PROCS processes in parallel"},
				"0": {Short: "-0", Long: "--null", Description: "Input items are terminated by a null character instead of whitespace (safe for filenames with spaces)"},
				"r": {Short: "-r", Long: "--no-run-if-empty", Description: "If standard input does not contain any nonblanks, do not run the command"},
			},
		},

		"ssh": {
			Name:        "ssh",
			Category:    "Network",
			Summary:     "OpenSSH remote login client for secure remote terminal access",
			Description: "Connects securely to a remote Linux machine over encrypted SSH protocol.",
			DefaultRisk: RiskSafe,
			Flags: map[string]FlagDef{
				"i": {Short: "-i", TakesValue: true, ValueName: "KEY_FILE", Description: "Selects file from which the identity (private key) for public key authentication is read"},
				"p": {Short: "-p", TakesValue: true, ValueName: "PORT", Description: "Port to connect to on the remote host (default: 22)"},
				"v": {Short: "-v", Description: "Verbose mode (prints debugging messages about the connection progress)"},
				"N": {Short: "-N", Description: "Do not execute a remote command (useful for just forwarding ports)"},
				"L": {Short: "-L", TakesValue: true, ValueName: "[LOCAL_IP:]LOCAL_PORT:DEST_HOST:DEST_PORT", Description: "Local port forwarding: forward connections from local port to remote destination"},
				"R": {Short: "-R", TakesValue: true, ValueName: "[REMOTE_IP:]REMOTE_PORT:DEST_HOST:DEST_PORT", Description: "Remote port forwarding: forward connections from remote port to local host"},
				"D": {Short: "-D", TakesValue: true, ValueName: "[LOCAL_IP:]LOCAL_PORT", Description: "Dynamic application-level port forwarding (SOCKS proxy)"},
			},
		},

		"df": {
			Name:        "df",
			Category:    "System",
			Summary:     "Reports filesystem disk space usage",
			Description: "Displays available and used disk space on mounted filesystems.",
			DefaultRisk: RiskSafe,
			Flags: map[string]FlagDef{
				"h": {Short: "-h", Long: "--human-readable", Description: "Print sizes in human readable format (e.g., 1K 234M 2G)"},
				"T": {Short: "-T", Long: "--print-type", Description: "Print filesystem type (ext4, btrfs, tmpfs, etc.)"},
				"i": {Short: "-i", Long: "--inodes", Description: "List inode information instead of block usage"},
			},
		},

		"du": {
			Name:        "du",
			Category:    "System",
			Summary:     "Estimates file and directory space usage",
			Description: "Calculates space occupied by directories and their contents recursively.",
			DefaultRisk: RiskSafe,
			Flags: map[string]FlagDef{
				"h": {Short: "-h", Long: "--human-readable", Description: "Print sizes in human readable format (e.g., 1K 234M 2G)"},
				"s": {Short: "-s", Long: "--summarize", Description: "Display only a total for each argument (do not list subfolders)"},
				"d": {Short: "-d", Long: "--max-depth", TakesValue: true, ValueName: "N", Description: "Print total for directory only if it is N or fewer levels below command argument"},
				"c": {Short: "-c", Long: "--total", Description: "Produce a grand total at the end of output"},
			},
		},

		"free": {
			Name:        "free",
			Category:    "System",
			Summary:     "Displays amount of free and used memory (RAM) in the system",
			Description: "Shows physical RAM and swap memory usage statistics.",
			DefaultRisk: RiskSafe,
			Flags: map[string]FlagDef{
				"h": {Short: "-h", Long: "--human", Description: "Show all output fields automatically scaled to shortest three digit unit and display the units"},
				"m": {Short: "-m", Long: "--mega", Description: "Display the amount of memory in megabytes (MB)"},
				"g": {Short: "-g", Long: "--giga", Description: "Display the amount of memory in gigabytes (GB)"},
				"t": {Short: "-t", Long: "--total", Description: "Display a line showing the column totals"},
			},
		},

		"uptime": {
			Name:        "uptime",
			Category:    "System",
			Summary:     "Tells how long the system has been running and load averages",
			Description: "Shows current time, system uptime duration, number of logged-in users, and 1, 5, 15 min load averages.",
			DefaultRisk: RiskSafe,
			Flags: map[string]FlagDef{
				"p": {Short: "-p", Long: "--pretty", Description: "Show uptime in pretty, human-readable format"},
				"s": {Short: "-s", Long: "--since", Description: "Show system up since (boot timestamp)"},
			},
		},

		"uname": {
			Name:        "uname",
			Category:    "System",
			Summary:     "Prints system and Linux kernel information",
			Description: "Outputs hardware architecture, kernel version, hostname, and OS name.",
			DefaultRisk: RiskSafe,
			Flags: map[string]FlagDef{
				"a": {Short: "-a", Long: "--all", Description: "Print all information in standard order"},
				"r": {Short: "-r", Long: "--kernel-release", Description: "Print the operating system kernel release version"},
				"m": {Short: "-m", Long: "--machine", Description: "Print the machine hardware architecture (e.g. x86_64, aarch64)"},
				"s": {Short: "-s", Long: "--kernel-name", Description: "Print the kernel name (e.g. Linux)"},
			},
		},
	}

	for k, v := range extraCommands {
		BuiltinCommands[k] = v
	}
}
