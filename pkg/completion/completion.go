package completion

import (
	"fmt"
	"strings"
)

// Generate returns the autocompletion script for the specified shell.
func Generate(shell string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(shell)) {
	case "bash":
		return BashScript, nil
	case "zsh":
		return ZshScript, nil
	case "fish":
		return FishScript, nil
	default:
		return "", fmt.Errorf("unsupported shell '%s'. Supported shells: bash, zsh, fish", shell)
	}
}

const BashScript = `# explain - Bash Autocompletion
_explain_completions() {
    local cur prev cword
    _get_comp_words_by_ref -n : cur prev cword 2>/dev/null || {
        cur="${COMP_WORDS[COMP_CWORD]}"
        prev="${COMP_WORDS[COMP_CWORD-1]}"
        cword=$COMP_CWORD
    }

    local commands="guard update uninstall hook completion -i --interactive -r --run --json --no-color -v --version -h --help"
    local guard_cmds="enable disable status"
    local shells="bash zsh fish"

    if [[ $cword -eq 1 ]]; then
        COMPREPLY=($(compgen -W "$commands" -- "$cur"))
        return 0
    fi

    local cmd1="${COMP_WORDS[1]}"

    case "$cmd1" in
        guard)
            if [[ $cword -eq 2 ]]; then
                COMPREPLY=($(compgen -W "$guard_cmds" -- "$cur"))
                return 0
            elif [[ $cword -eq 3 ]]; then
                case "${COMP_WORDS[2]}" in
                    enable|disable|status)
                        COMPREPLY=($(compgen -W "$shells" -- "$cur"))
                        return 0
                        ;;
                esac
            fi
            ;;
        hook|completion)
            if [[ $cword -eq 2 ]]; then
                COMPREPLY=($(compgen -W "$shells" -- "$cur"))
                return 0
            fi
            ;;
    esac

    return 0
}

complete -F _explain_completions explain
`

const ZshScript = `#compdef explain

_explain() {
    local curcontext="$curcontext" state line
    typeset -A opt_args

    _arguments -C \
        '(-i --interactive)'{-i,--interactive}'[Launch interactive mode]' \
        '(-r --run)'{-r,--run}'[Ask to run command after explaining]' \
        '--json[Output analysis in JSON format]' \
        '--no-color[Disable colored output]' \
        '(-v --version)'{-v,--version}'[Show current explain version]' \
        '(-h --help)'{-h,--help}'[Show help message]' \
        '1: :->command' \
        '*:: :->args'

    case $state in
        command)
            local -a subcmds
            subcmds=(
                'guard:Terminal safety shield against destructive commands'
                'update:Check and upgrade explain to latest release'
                'uninstall:Remove explain CLI from system'
                'hook:Output shell hook script'
                'completion:Generate shell autocompletion script'
            )
            _describe -t subcmds 'explain commands' subcmds
            ;;
        args)
            case $line[1] in
                guard)
                    if (( CURRENT == 2 )); then
                        local -a guard_subcmds
                        guard_subcmds=(
                            'enable:Enable active terminal protection'
                            'disable:Disable active terminal protection'
                            'status:Check terminal protection status'
                        )
                        _describe -t guard_subcmds 'guard commands' guard_subcmds
                    elif (( CURRENT == 3 )); then
                        local -a shells
                        shells=('bash:Bash shell' 'zsh:Zsh shell' 'fish:Fish shell')
                        _describe -t shells 'target shell' shells
                    fi
                    ;;
                hook|completion)
                    if (( CURRENT == 2 )); then
                        local -a shells
                        shells=('bash:Bash shell' 'zsh:Zsh shell' 'fish:Fish shell')
                        _describe -t shells 'target shell' shells
                    fi
                    ;;
            esac
            ;;
    esac
}

_explain "$@"
`

const FishScript = `# Fish completion for explain
complete -c explain -f

# Top-level commands
complete -c explain -n "__fish_use_subcommand" -a "guard" -d "Terminal safety shield against destructive commands"
complete -c explain -n "__fish_use_subcommand" -a "update" -d "Upgrade explain to latest release"
complete -c explain -n "__fish_use_subcommand" -a "uninstall" -d "Remove explain CLI from system"
complete -c explain -n "__fish_use_subcommand" -a "hook" -d "Output shell integration hook"
complete -c explain -n "__fish_use_subcommand" -a "completion" -d "Generate shell completion script"

# Options & Flags
complete -c explain -s i -l interactive -d "Interactive mode (no quotes needed)"
complete -c explain -s r -l run -d "Execute command after explanation"
complete -c explain -l json -d "Output analysis in JSON format"
complete -c explain -l no-color -d "Disable colored output"
complete -c explain -s v -l version -d "Show current version"
complete -c explain -s h -l help -d "Show help message"

# Guard subcommands
complete -c explain -n "__fish_seen_subcommand_from guard; and not __fish_seen_subcommand_from enable disable status" -a "enable" -d "Enable terminal protection"
complete -c explain -n "__fish_seen_subcommand_from guard; and not __fish_seen_subcommand_from enable disable status" -a "disable" -d "Disable terminal protection"
complete -c explain -n "__fish_seen_subcommand_from guard; and not __fish_seen_subcommand_from enable disable status" -a "status" -d "Check terminal protection status"

# Target shells for hook, completion, and guard
complete -c explain -n "__fish_seen_subcommand_from hook completion" -a "bash zsh fish" -d "Target shell"
complete -c explain -n "__fish_seen_subcommand_from guard; and __fish_seen_subcommand_from enable disable status" -a "bash zsh fish" -d "Target shell"
`
