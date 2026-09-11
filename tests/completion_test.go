package tests

import (
	"strings"
	"testing"

	"github.com/Yehya-Elsawy/explain/pkg/completion"
	"github.com/Yehya-Elsawy/explain/pkg/guard"
)

func TestCompletionGenerate(t *testing.T) {
	shells := []string{"bash", "zsh", "fish"}

	for _, sh := range shells {
		script, err := completion.Generate(sh)
		if err != nil {
			t.Fatalf("unexpected error generating completion for %s: %v", sh, err)
		}

		// Must contain primary subcommands
		for _, cmd := range []string{"guard", "update", "uninstall", "hook", "completion"} {
			if !strings.Contains(script, cmd) {
				t.Errorf("completion script for %s missing command %q", sh, cmd)
			}
		}

		// Must contain guard subcommands
		for _, sub := range []string{"enable", "disable", "status"} {
			if !strings.Contains(script, sub) {
				t.Errorf("completion script for %s missing guard subcommand %q", sh, sub)
			}
		}
	}

	_, err := completion.Generate("unsupported_shell")
	if err == nil {
		t.Error("expected error for unsupported shell")
	}
}

func TestHookIncludesCompletion(t *testing.T) {
	for _, sh := range []string{"bash", "zsh", "fish"} {
		hook, err := guard.GenerateHook(sh)
		if err != nil {
			t.Fatalf("unexpected error for %s: %v", sh, err)
		}

		// The hook must include completion directives
		if !strings.Contains(hook, "explain") {
			t.Errorf("hook for %s does not contain explain", sh)
		}
		if sh == "bash" && !strings.Contains(hook, "complete -F _explain_completions explain") {
			t.Errorf("bash hook does not include bash completion")
		}
		if sh == "zsh" && !strings.Contains(hook, "#compdef explain") {
			t.Errorf("zsh hook does not include zsh completion")
		}
		if sh == "fish" && !strings.Contains(hook, "complete -c explain") {
			t.Errorf("fish hook does not include fish completion")
		}
	}
}
