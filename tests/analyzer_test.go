package tests

import (
	"testing"

	"github.com/Yehya-Elsawy/explain-/pkg/analyzer"
	"github.com/Yehya-Elsawy/explain-/pkg/ast"
	"github.com/Yehya-Elsawy/explain-/pkg/database"
)

func TestAnalyzeCdCommand(t *testing.T) {
	pipe, _ := ast.Parse("cd /home/")
	analysis := analyzer.AnalyzePipeline(pipe)

	cmd := analysis.Commands[0]
	if cmd.CommandName != "cd" {
		t.Errorf("expected cd, got %s", cmd.CommandName)
	}
	if cmd.ActionSummary != "Changes the current working directory to '/home/'." {
		t.Errorf("unexpected action summary: %s", cmd.ActionSummary)
	}
	if len(cmd.Items) != 1 || cmd.Items[0].Description != "Target destination directory to switch to" {
		t.Errorf("unexpected item description: %v", cmd.Items)
	}
}

func TestAnalyzeTarCommand(t *testing.T) {
	pipe, _ := ast.Parse("tar -xzf backup.tar.gz")
	analysis := analyzer.AnalyzePipeline(pipe)

	if len(analysis.Commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(analysis.Commands))
	}
	cmd := analysis.Commands[0]
	if cmd.CommandName != "tar" {
		t.Errorf("expected tar, got %s", cmd.CommandName)
	}

	// Should decompose into -x, -z, -f
	hasX, hasZ, hasF := false, false, false
	for _, item := range cmd.Items {
		if item.Token == "-x" {
			hasX = true
		}
		if item.Token == "-z" {
			hasZ = true
		}
		if item.Token == "-f" {
			hasF = true
		}
	}

	if !hasX || !hasZ || !hasF {
		t.Errorf("missing decomposed flags: x=%v, z=%v, f=%v", hasX, hasZ, hasF)
	}
}

func TestAnalyzeGitCommit(t *testing.T) {
	pipe, _ := ast.Parse("git commit -m 'initial release'")
	analysis := analyzer.AnalyzePipeline(pipe)

	cmd := analysis.Commands[0]
	if cmd.Subcommand != "commit" {
		t.Errorf("expected subcommand commit, got %s", cmd.Subcommand)
	}
	if cmd.ActionSummary == "" {
		t.Errorf("expected action summary, got empty")
	}
}

func TestAnalyzeDockerRun(t *testing.T) {
	pipe, _ := ast.Parse("docker run -d -p 8080:80 --name web nginx:latest")
	analysis := analyzer.AnalyzePipeline(pipe)

	cmd := analysis.Commands[0]
	if cmd.Subcommand != "run" {
		t.Errorf("expected subcommand run, got %s", cmd.Subcommand)
	}
	if cmd.Danger.Level != database.RiskMedium {
		t.Errorf("expected RiskMedium, got %v", cmd.Danger.Level)
	}
}
