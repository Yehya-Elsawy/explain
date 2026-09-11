package guard

import (
	"fmt"
	"strings"

	"github.com/Yehya-Elsawy/explain/pkg/completion"
)

// GenerateHook generates the shell integration script for the specified shell.
func GenerateHook(shell string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(shell)) {
	case "bash":
		return bashHookScript + "\n" + completion.BashScript, nil
	case "zsh":
		return zshHookScript + "\n" + completion.ZshScript, nil
	case "fish":
		return fishHookScript + "\n" + completion.FishScript, nil
	default:
		return "", fmt.Errorf("unsupported shell '%s'. Supported shells: bash, zsh, fish", shell)
	}
}

const bashHookScript = `# explain guard - Bash Integration
# Fast-filters safe commands in-shell to ensure 0.00ms latency.
# Only forwards potentially dangerous operations to explain guard check.

_explain_guard_bash() {
    # Ignore during tab completion, prompt rendering, or internal functions
    [[ -n "$COMP_LINE" ]] && return 0
    [[ ${#FUNCNAME[@]} -gt 1 ]] && return 0

    local cmd="$BASH_COMMAND"
    [[ -z "$cmd" ]] && return 0

    # Avoid self-recursion or internal shell constructs
    [[ "$cmd" == explain* || "$cmd" == _explain* ]] && return 0

    # Fast shell regex filter: only test commands matching hazardous patterns
    if [[ "$cmd" =~ (rm[[:space:]]|dd[[:space:]]|mkfs|chmod[[:space:]]|chown[[:space:]]|>[[:space:]]*/dev/|\|\s*(sudo\s*)?(ba)?sh) ]]; then
        explain guard check "$cmd"
        return $?
    fi
    return 0
}

shopt -s extdebug
trap '_explain_guard_bash' DEBUG
`

const zshHookScript = `# explain guard - Zsh Integration
# Intercepts commands on Enter using ZLE accept-line with zero latency for safe commands.

_explain_guard_zsh() {
    local cmd="$BUFFER"
    local trimmed="${cmd#"${cmd%%[![:space:]]*}"}"

    if [[ -n "$trimmed" && "$trimmed" != explain* && "$trimmed" != _explain* ]]; then
        if [[ "$trimmed" =~ (rm[[:space:]]|dd[[:space:]]|mkfs|chmod[[:space:]]|chown[[:space:]]|>[[:space:]]*/dev/|\|\s*(sudo\s*)?(ba)?sh) ]]; then
            explain guard check "$trimmed"
            if [[ $? -ne 0 ]]; then
                zle -M "  [!] Command cancelled by explain guard"
                return 1
            fi
        fi
    fi
    zle .accept-line
}

zle -N accept-line _explain_guard_zsh
`

const fishHookScript = `# explain guard - Fish Integration
# Intercepts commands on Enter key with zero latency for safe commands.

function _explain_guard_fish
    set -l cmd (commandline)
    set -l trimmed (string trim -- "$cmd")

    if test -n "$trimmed"; and not string match -rq '^(_?explain)' -- "$trimmed"
        if string match -rq '(rm\s|dd\s|mkfs|chmod\s|chown\s|>\s*/dev/|\|\s*(sudo\s*)?(ba)?sh)' -- "$trimmed"
            explain guard check "$trimmed"
            if test $status -ne 0
                echo "  [!] Command cancelled by explain guard"
                return 1
            end
        end
    end
    commandline -f execute
end

bind \r _explain_guard_fish
`
