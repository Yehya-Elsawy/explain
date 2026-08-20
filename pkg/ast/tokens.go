package ast

// TokenType identifies the category of shell token.
type TokenType int

const (
	TokenWord TokenType = iota
	TokenPipe           // |
	TokenAnd            // &&
	TokenOr             // ||
	TokenSemi           // ;
	TokenRedirectOut    // >
	TokenRedirectAppend // >>
	TokenRedirectIn     // <
	TokenRedirectErr    // 2> or 2>&1 or &>
)

// Token represents a single lexical token from shell command input.
type Token struct {
	Type  TokenType
	Value string
	Pos   int
}

// Redirect represents an I/O redirection in the command.
type Redirect struct {
	Operator string // >, >>, <, 2>, 2>&1
	Target   string // filename or file descriptor
}

// SingleCommand represents a single command execution unit (e.g. "tar -xzf backup.tar.gz").
type SingleCommand struct {
	Prefixes    []string   // sudo, nohup, time, env, etc.
	EnvVars     []string   // FOO=bar
	Name        string     // tar
	Subcommand  string     // commit, run, restart (if applicable)
	Args        []string   // Raw arguments excluding prefixes & redirects
	Redirects   []Redirect // Redirections
	PipedToNext bool       // Whether this command is piped into the next command
	ChainOp     string     // &&, ||, ;, or empty
}

// Pipeline represents a full shell pipeline or compound command.
type Pipeline struct {
	RawInput string
	Commands []*SingleCommand
}
