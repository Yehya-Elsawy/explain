package tests

import (
	"testing"

	"github.com/Yehya-Elsawy/explain-/pkg/ast"
)

func TestParseSimpleCommand(t *testing.T) {
	pipe, err := ast.Parse("tar -xzf backup.tar.gz")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pipe.Commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(pipe.Commands))
	}
	cmd := pipe.Commands[0]
	if cmd.Name != "tar" {
		t.Errorf("expected name 'tar', got '%s'", cmd.Name)
	}
	if len(cmd.Args) != 2 {
		t.Errorf("expected 2 args, got %d", len(cmd.Args))
	}
}

func TestParsePipeline(t *testing.T) {
	pipe, err := ast.Parse("ps aux | grep nginx | awk '{print $2}'")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pipe.Commands) != 3 {
		t.Fatalf("expected 3 commands in pipeline, got %d", len(pipe.Commands))
	}
	if pipe.Commands[0].Name != "ps" || !pipe.Commands[0].PipedToNext {
		t.Errorf("stage 1 failed: %v", pipe.Commands[0])
	}
	if pipe.Commands[1].Name != "grep" || !pipe.Commands[1].PipedToNext {
		t.Errorf("stage 2 failed: %v", pipe.Commands[1])
	}
	if pipe.Commands[2].Name != "awk" || pipe.Commands[2].PipedToNext {
		t.Errorf("stage 3 failed: %v", pipe.Commands[2])
	}
}

func TestParseRedirectsAndSudo(t *testing.T) {
	pipe, err := ast.Parse("sudo echo 'hello' > /tmp/output.txt 2>&1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pipe.Commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(pipe.Commands))
	}
	cmd := pipe.Commands[0]
	if len(cmd.Prefixes) != 1 || cmd.Prefixes[0] != "sudo" {
		t.Errorf("expected prefix 'sudo', got %v", cmd.Prefixes)
	}
	if cmd.Name != "echo" {
		t.Errorf("expected name 'echo', got '%s'", cmd.Name)
	}
	if len(cmd.Redirects) != 2 {
		t.Errorf("expected 2 redirects, got %d", len(cmd.Redirects))
	}
}
