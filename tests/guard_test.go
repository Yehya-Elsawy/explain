package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Yehya-Elsawy/explain/pkg/analyzer"
	"github.com/Yehya-Elsawy/explain/pkg/ast"
	"github.com/Yehya-Elsawy/explain/pkg/database"
	"github.com/Yehya-Elsawy/explain/pkg/guard"
)

func TestGenerateHook(t *testing.T) {
	shells := []string{"bash", "zsh", "fish"}
	for _, sh := range shells {
		hook, err := guard.GenerateHook(sh)
		if err != nil {
			t.Fatalf("expected hook for %s, got error: %v", sh, err)
		}
		if !strings.Contains(hook, "explain guard check") {
			t.Errorf("hook for %s missing 'explain guard check'", sh)
		}
	}

	_, err := guard.GenerateHook("unknown_shell")
	if err == nil {
		t.Error("expected error for unsupported shell")
	}
}

func TestDetectShell(t *testing.T) {
	origShell := os.Getenv("SHELL")
	defer os.Setenv("SHELL", origShell)

	os.Setenv("SHELL", "/bin/zsh")
	sh, rc, err := guard.DetectShell()
	if err != nil || sh != "zsh" || !strings.HasSuffix(rc, ".zshrc") {
		t.Errorf("expected zsh and .zshrc, got %s, %s, err: %v", sh, rc, err)
	}

	os.Setenv("SHELL", "/usr/bin/fish")
	sh, rc, err = guard.DetectShell()
	if err != nil || sh != "fish" || !strings.HasSuffix(rc, "config.fish") {
		t.Errorf("expected fish and config.fish, got %s, %s, err: %v", sh, rc, err)
	}

	os.Setenv("SHELL", "/bin/bash")
	sh, _, err = guard.DetectShell()
	if err != nil || sh != "bash" {
		t.Errorf("expected bash, got %s, err: %v", sh, err)
	}
}

func TestRcSnippetManipulation(t *testing.T) {
	tempDir := t.TempDir()
	rcFile := filepath.Join(tempDir, ".test_bashrc")

	initialContent := "# User custom config\nexport PATH=$PATH:/usr/local/bin\nalias l='ls -la'\n"
	if err := os.WriteFile(rcFile, []byte(initialContent), 0644); err != nil {
		t.Fatalf("failed to create dummy rc file: %v", err)
	}

	snippet := guard.GenerateRcSnippet("bash")

	// 1. Simulate insertion
	updated := initialContent + "\n" + snippet
	if err := os.WriteFile(rcFile, []byte(updated), 0644); err != nil {
		t.Fatalf("failed to write updated rc: %v", err)
	}

	readBytes, _ := os.ReadFile(rcFile)
	readContent := string(readBytes)
	if !strings.Contains(readContent, guard.MarkerStart) || !strings.Contains(readContent, guard.MarkerEnd) {
		t.Fatalf("markers not found in rc file")
	}
	if !strings.Contains(readContent, "eval \"$(explain hook bash)\"") {
		t.Fatalf("hook snippet not found in rc file")
	}

	// 2. Simulate removal
	startIdx := strings.Index(readContent, guard.MarkerStart)
	endIdx := strings.Index(readContent, guard.MarkerEnd)
	before := strings.TrimRight(readContent[:startIdx], "\r\n")
	after := strings.TrimLeft(readContent[endIdx+len(guard.MarkerEnd):], "\r\n")

	var stripped string
	if before != "" && after != "" {
		stripped = before + "\n\n" + after
	} else if before != "" {
		stripped = before + "\n"
	}

	if err := os.WriteFile(rcFile, []byte(stripped), 0644); err != nil {
		t.Fatalf("failed to save stripped rc: %v", err)
	}

	finalBytes, _ := os.ReadFile(rcFile)
	finalContent := string(finalBytes)
	if strings.Contains(finalContent, guard.MarkerStart) {
		t.Fatalf("guard marker still present after removal")
	}
	if !strings.Contains(finalContent, "alias l='ls -la'") {
		t.Fatalf("original user configuration was damaged during removal")
	}
}

func TestGuardDangerClassification(t *testing.T) {
	// Routine commands that must NEVER be Critical (must bypass guard seamlessly)
	safeCommands := []string{
		"ls -la",
		"cd /home/user",
		"cat app.log | grep ERROR",
		"rm file.txt",
		"rm -f temp.log",
		"rm -r build/",
		"git commit -m 'feat: something'",
		"docker run -d -p 80:80 nginx",
		"kill 1234",
	}

	for _, cmdStr := range safeCommands {
		pipe, err := ast.Parse(cmdStr)
		if err != nil {
			t.Fatalf("parse error for %q: %v", cmdStr, err)
		}
		analysis := analyzer.AnalyzePipeline(pipe)
		if analysis.MaxRisk == database.RiskCritical {
			t.Errorf("routine command %q was falsely marked as RiskCritical!", cmdStr)
		}
	}

	// Catastrophic commands that MUST be intercepted as RiskCritical
	criticalCommands := []string{
		"rm -rf /",
		"rm -rf /*",
		"dd if=/dev/zero of=/dev/sda",
		"mkfs.ext4 /dev/sdb1",
		"chmod -R 777 /",
	}

	for _, cmdStr := range criticalCommands {
		pipe, err := ast.Parse(cmdStr)
		if err != nil {
			t.Fatalf("parse error for %q: %v", cmdStr, err)
		}
		analysis := analyzer.AnalyzePipeline(pipe)
		if analysis.MaxRisk != database.RiskCritical {
			t.Errorf("critical command %q was NOT flagged as RiskCritical! (got: %v)", cmdStr, analysis.MaxRisk)
		}
	}
}

func TestGuardHardBlock(t *testing.T) {
	lethal := []string{
		"rm -rf /",
		"rm -rf /*",
		"rm -rf ~",
		"rm -rf --no-preserve-root /",
		":(){ :|:& };:",
		"chmod -R 777 /",
	}

	for _, cmd := range lethal {
		exitCode := guard.Check(cmd)
		if exitCode != 1 {
			t.Errorf("expected lethal command %q to be hard-blocked with exit code 1, got %d", cmd, exitCode)
		}
	}
}
