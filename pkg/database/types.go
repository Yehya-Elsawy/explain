package database

// RiskLevel defines the danger / safety level of a command.
type RiskLevel string

const (
	RiskSafe     RiskLevel = "SAFE"     // Read-only or completely benign (ls, cat, pwd, echo)
	RiskLow      RiskLevel = "LOW"      // Creates files or directories safely (mkdir, touch, cp)
	RiskMedium   RiskLevel = "MEDIUM"   // Modifies existing files or system state (tar, chmod, systemctl restart)
	RiskHigh     RiskLevel = "HIGH"     // Overwrites files, kills processes, network config (kill, rsync --delete, > file)
	RiskCritical RiskLevel = "CRITICAL" // Destructive, irreversible or severe system impact (rm -rf, dd of=/dev/..., mkfs, chmod 777)
)

// FlagDef represents the definition of a specific command-line flag or option.
type FlagDef struct {
	Short       string    `json:"short,omitempty"`       // e.g. "-x"
	Long        string    `json:"long,omitempty"`        // e.g. "--extract"
	TakesValue  bool      `json:"takes_value,omitempty"` // Whether this flag consumes the following argument (e.g. -f <file>)
	ValueName   string    `json:"value_name,omitempty"`  // e.g. "FILE", "DIR", "PORT"
	Description string    `json:"description"`           // Plain-English beginner friendly description
	Risk        RiskLevel `json:"risk,omitempty"`        // If this specific flag elevates risk (e.g. -f in rm)
}

// SubcommandDef represents a subcommand under a tool (e.g. "git commit", "docker run", "systemctl status").
type SubcommandDef struct {
	Name        string             `json:"name"`
	Summary     string             `json:"summary"`
	Description string             `json:"description"`
	Flags       map[string]FlagDef `json:"flags,omitempty"` // Map of flag name (short or long) to FlagDef
	DefaultRisk RiskLevel          `json:"default_risk"`
	ActionFmt   string             `json:"action_fmt,omitempty"` // Template for action summary
}

// CommandDef represents the complete knowledge of a Linux utility.
type CommandDef struct {
	Name          string                   `json:"name"`
	Category      string                   `json:"category"` // FileOps, System, Network, Process, Archive, etc.
	Summary       string                   `json:"summary"`  // 1-line short summary
	Description   string                   `json:"description"`
	Flags         map[string]FlagDef       `json:"flags,omitempty"`       // Map of flag without leading dash or with dash
	Subcommands   map[string]SubcommandDef `json:"subcommands,omitempty"` // Subcommands if applicable
	DefaultRisk   RiskLevel                `json:"default_risk"`
	DefaultAction string                   `json:"default_action,omitempty"`
}
