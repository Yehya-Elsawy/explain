package database

func init() {
	subcommandTools := map[string]CommandDef{
		"git": {
			Name:        "git",
			Category:    "VCS",
			Summary:     "Fast, scalable, distributed version control system",
			Description: "Tracks changes in source code during software development.",
			DefaultRisk: RiskLow,
			Subcommands: map[string]SubcommandDef{
				"commit": {
					Name:        "commit",
					Summary:     "Record staged changes to the repository history",
					Description: "Stores the current contents of the index in a new commit along with a log message.",
					DefaultRisk: RiskLow,
					Flags: map[string]FlagDef{
						"m":       {Short: "-m", Long: "--message", TakesValue: true, ValueName: "MSG", Description: "Use the given MSG as the commit message"},
						"a":       {Short: "-a", Long: "--all", Description: "Automatically stage modified and deleted files before committing"},
						"amend":   {Long: "--amend", Description: "Replace the tip of the current branch by creating a new commit (modifies last commit)"},
						"no-edit": {Long: "--no-edit", Description: "Use the selected commit message without launching an editor"},
					},
					ActionFmt: "Commits staged changes to the git repository.",
				},
				"push": {
					Name:        "push",
					Summary:     "Update remote refs along with associated objects",
					Description: "Uploads local branch commits to the remote repository.",
					DefaultRisk: RiskMedium,
					Flags: map[string]FlagDef{
						"u":     {Short: "-u", Long: "--set-upstream", Description: "Set upstream tracking branch for git pull and push"},
						"f":     {Short: "-f", Long: "--force", Description: "Force push (OVERWRITES remote branch history - DANGEROUS)", Risk: RiskCritical},
						"tags":  {Long: "--tags", Description: "Push all tags to remote"},
					},
					ActionFmt: "Uploads local branch commits to the remote repository.",
				},
				"pull": {
					Name:        "pull",
					Summary:     "Fetch from and integrate with another repository or a local branch",
					Description: "Fetches changes from remote and immediately merges them into current branch.",
					DefaultRisk: RiskMedium,
					Flags: map[string]FlagDef{
						"rebase": {Long: "--rebase", Description: "Rebase local commits on top of fetched branch instead of merging"},
					},
					ActionFmt: "Fetches and merges latest commits from remote.",
				},
				"checkout": {
					Name:        "checkout",
					Summary:     "Switch branches or restore working tree files",
					Description: "Updates files in working tree to match version in index or specified branch.",
					DefaultRisk: RiskLow,
					Flags: map[string]FlagDef{
						"b": {Short: "-b", TakesValue: true, ValueName: "NEW_BRANCH", Description: "Create and switch to a new branch"},
						"B": {Short: "-B", TakesValue: true, ValueName: "NEW_BRANCH", Description: "Create or reset and switch to branch"},
					},
					ActionFmt: "Switches branches or restores working tree files.",
				},
				"status": {
					Name:        "status",
					Summary:     "Show the working tree status",
					Description: "Displays paths that have differences between index and current HEAD commit.",
					DefaultRisk: RiskSafe,
					Flags: map[string]FlagDef{
						"s": {Short: "-s", Long: "--short", Description: "Give the output in short-format"},
					},
					ActionFmt: "Displays status of tracked, modified, and untracked files.",
				},
				"add": {
					Name:        "add",
					Summary:     "Add file contents to the index (staging area)",
					Description: "Stages changes for the next commit.",
					DefaultRisk: RiskLow,
					Flags: map[string]FlagDef{
						"A": {Short: "-A", Long: "--all", Description: "Stage all tracked and untracked file changes in the entire repo"},
						"p": {Short: "-p", Long: "--patch", Description: "Interactively choose hunks of patch to stage"},
						".": {Short: ".", Description: "Stage all changes in the current directory and subdirectories"},
					},
					ActionFmt: "Stages file changes ready to be committed.",
				},
				"reset": {
					Name:        "reset",
					Summary:     "Reset current HEAD to the specified state",
					Description: "Sets the current branch head (HEAD) to the specified commit.",
					DefaultRisk: RiskHigh,
					Flags: map[string]FlagDef{
						"hard": {Long: "--hard", Description: "Resets index and working tree. Any changes to tracked files are permanently discarded!", Risk: RiskCritical},
						"soft": {Long: "--soft", Description: "Resets HEAD only; leaves all changed files staged in index"},
					},
					ActionFmt: "Resets HEAD position and repository state.",
				},
			},
		},

		"docker": {
			Name:        "docker",
			Category:    "Container",
			Summary:     "Manage containers, images, volumes, and networks",
			Description: "Container virtualization runtime tool.",
			DefaultRisk: RiskMedium,
			Subcommands: map[string]SubcommandDef{
				"run": {
					Name:        "run",
					Summary:     "Create and run a new container from an image",
					Description: "Starts a process in an isolated container environment.",
					DefaultRisk: RiskMedium,
					Flags: map[string]FlagDef{
						"d":        {Short: "-d", Long: "--detach", Description: "Run container in background and print container ID"},
						"p":        {Short: "-p", Long: "--publish", TakesValue: true, ValueName: "HOST:CONTAINER", Description: "Publish a container's port(s) to the host (e.g. 8080:80)"},
						"v":        {Short: "-v", Long: "--volume", TakesValue: true, ValueName: "HOST_DIR:CONTAINER_DIR", Description: "Bind mount a volume from host into container"},
						"name":     {Long: "--name", TakesValue: true, ValueName: "NAME", Description: "Assign a custom name to the container"},
						"rm":       {Long: "--rm", Description: "Automatically remove container when it exits"},
						"it":       {Short: "-it", Description: "Allocate a pseudo-TTY connected to container's stdin (interactive mode)"},
						"e":        {Short: "-e", Long: "--env", TakesValue: true, ValueName: "KEY=VAL", Description: "Set environment variables inside container"},
						"restart":  {Long: "--restart", TakesValue: true, ValueName: "POLICY", Description: "Restart policy (no, on-failure, always, unless-stopped)"},
					},
					ActionFmt: "Creates and launches an isolated container instance.",
				},
				"ps": {
					Name:        "ps",
					Summary:     "List running Docker containers",
					Description: "Displays active containers, statuses, and port mappings.",
					DefaultRisk: RiskSafe,
					Flags: map[string]FlagDef{
						"a": {Short: "-a", Long: "--all", Description: "Show all containers (default shows just running)"},
						"q": {Short: "-q", Long: "--quiet", Description: "Only display numeric container IDs"},
					},
					ActionFmt: "Lists active Docker containers and their statuses.",
				},
				"exec": {
					Name:        "exec",
					Summary:     "Run a command in a running container",
					Description: "Executes an interactive or background command inside a container.",
					DefaultRisk: RiskMedium,
					Flags: map[string]FlagDef{
						"it": {Short: "-it", Description: "Run interactively with terminal attached"},
						"u":  {Short: "-u", Long: "--user", TakesValue: true, ValueName: "USER", Description: "Username or UID to execute command as"},
					},
					ActionFmt: "Runs a command inside a running Docker container.",
				},
				"stop": {
					Name:        "stop",
					Summary:     "Stop one or more running containers",
					Description: "Sends SIGTERM followed by SIGKILL after grace period to stop container.",
					DefaultRisk: RiskMedium,
					ActionFmt:   "Stops running Docker containers.",
				},
				"rm": {
					Name:        "rm",
					Summary:     "Remove one or more containers",
					Description: "Deletes stopped container instances from disk.",
					DefaultRisk: RiskHigh,
					Flags: map[string]FlagDef{
						"f": {Short: "-f", Long: "--force", Description: "Force the removal of a running container", Risk: RiskHigh},
						"v": {Short: "-v", Long: "--volumes", Description: "Remove anonymous volumes attached to container"},
					},
					ActionFmt: "Deletes container instances.",
				},
			},
		},

		"systemctl": {
			Name:        "systemctl",
			Category:    "System",
			Summary:     "Controls systemd system and service manager",
			Description: "Starts, stops, restarts, inspects, and manages system background daemons and units.",
			DefaultRisk: RiskMedium,
			Subcommands: map[string]SubcommandDef{
				"start": {
					Name:        "start",
					Summary:     "Start (activate) one or more units",
					DefaultRisk: RiskMedium,
					ActionFmt:   "Starts the specified system service/daemon.",
				},
				"stop": {
					Name:        "stop",
					Summary:     "Stop (deactivate) one or more units",
					DefaultRisk: RiskMedium,
					ActionFmt:   "Stops the specified system service/daemon.",
				},
				"restart": {
					Name:        "restart",
					Summary:     "Stop and then start one or more units",
					DefaultRisk: RiskMedium,
					ActionFmt:   "Restarts the specified system service/daemon.",
				},
				"status": {
					Name:        "status",
					Summary:     "Show runtime status about one or more units",
					DefaultRisk: RiskSafe,
					ActionFmt:   "Displays live status, PID, memory, and recent logs of the service.",
				},
				"enable": {
					Name:        "enable",
					Summary:     "Enable one or more units to start on boot",
					DefaultRisk: RiskMedium,
					Flags: map[string]FlagDef{
						"now": {Long: "--now", Description: "Start the unit immediately in addition to enabling on boot"},
					},
					ActionFmt: "Configures the service to launch automatically on system boot.",
				},
				"disable": {
					Name:        "disable",
					Summary:     "Disable one or more units from starting on boot",
					DefaultRisk: RiskMedium,
					Flags: map[string]FlagDef{
						"now": {Long: "--now", Description: "Stop the unit immediately in addition to disabling from boot"},
					},
					ActionFmt: "Prevents the service from automatically launching on boot.",
				},
				"daemon-reload": {
					Name:        "daemon-reload",
					Summary:     "Reload systemd manager configuration",
					DefaultRisk: RiskLow,
					ActionFmt:   "Rereads all service unit configuration files from disk.",
				},
			},
		},

		"apt": {
			Name:        "apt",
			Category:    "Package",
			Summary:     "Command-line package management interface for Debian/Ubuntu",
			Description: "Installs, updates, upgrades, and removes software packages.",
			DefaultRisk: RiskMedium,
			Subcommands: map[string]SubcommandDef{
				"update": {
					Name:        "update",
					Summary:     "Download latest package lists from repositories",
					DefaultRisk: RiskLow,
					ActionFmt:   "Fetches latest package list metadata from configured apt repositories.",
				},
				"upgrade": {
					Name:        "upgrade",
					Summary:     "Upgrade all installed packages to newest versions",
					DefaultRisk: RiskMedium,
					Flags: map[string]FlagDef{
						"y": {Short: "-y", Long: "--yes", Description: "Automatic yes to prompts; assume yes to all questions"},
					},
					ActionFmt: "Upgrades all installed system packages to their latest versions.",
				},
				"install": {
					Name:        "install",
					Summary:     "Install one or more new packages",
					DefaultRisk: RiskMedium,
					Flags: map[string]FlagDef{
						"y": {Short: "-y", Long: "--yes", Description: "Automatic yes to prompts"},
					},
					ActionFmt: "Downloads and installs specified package(s) and their dependencies.",
				},
				"remove": {
					Name:        "remove",
					Summary:     "Remove packages (leaving configuration files behind)",
					DefaultRisk: RiskHigh,
					ActionFmt:   "Uninstalls specified package binaries.",
				},
				"purge": {
					Name:        "purge",
					Summary:     "Remove packages and their global configuration files",
					DefaultRisk: RiskHigh,
					ActionFmt:   "Completely uninstalls package(s) and purges their configuration files.",
				},
			},
		},
	}

	for k, v := range subcommandTools {
		BuiltinCommands[k] = v
	}
}
