package tests

import (
	"testing"

	"github.com/Yehya-Elsawy/explain/pkg/analyzer"
	"github.com/Yehya-Elsawy/explain/pkg/ast"
)

func TestPipelineSummary(t *testing.T) {
	pipe, err := ast.Parse("ps aux | grep nginx | awk '{print $2}' | xargs kill -9")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	analysis := analyzer.AnalyzePipeline(pipe)
	if analysis.PipelineSummary == "" {
		t.Errorf("expected non-empty pipeline summary for multi-stage command")
	}

	if analysis.SmartTip == "" {
		t.Errorf("expected smart tip suggestion for ps | grep | kill pipeline")
	}
}

func TestSmartTipUselessCat(t *testing.T) {
	pipe, err := ast.Parse("cat /var/log/syslog | grep error")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	analysis := analyzer.AnalyzePipeline(pipe)
	if analysis.SmartTip == "" {
		t.Errorf("expected smart tip for useless cat command")
	}
}

func TestKubectlAndNpm(t *testing.T) {
	pipe1, _ := ast.Parse("kubectl get pods -n kube-system")
	analysis1 := analyzer.AnalyzePipeline(pipe1)
	if len(analysis1.Commands) != 1 || analysis1.Commands[0].Subcommand != "get" {
		t.Errorf("expected kubectl subcommand get, got %v", analysis1.Commands[0].Subcommand)
	}

	pipe2, _ := ast.Parse("npm install -D typescript")
	analysis2 := analyzer.AnalyzePipeline(pipe2)
	if len(analysis2.Commands) != 1 || analysis2.Commands[0].Subcommand != "install" {
		t.Errorf("expected npm subcommand install, got %v", analysis2.Commands[0].Subcommand)
	}
}

func TestChainedCommandSummaries(t *testing.T) {
	pipeAnd, _ := ast.Parse("apt update && apt upgrade")
	analysisAnd := analyzer.AnalyzePipeline(pipeAnd)
	if analysisAnd.PipelineSummary != "A conditional 2-stage execution chain (apt && apt) where each command runs only if the preceding command succeeds (exit code 0)." {
		t.Errorf("unexpected summary for AND chain: %q", analysisAnd.PipelineSummary)
	}

	pipeOr, _ := ast.Parse("make build || echo 'failed'")
	analysisOr := analyzer.AnalyzePipeline(pipeOr)
	if analysisOr.PipelineSummary != "A fallback 2-stage execution chain (make || echo) where each command runs only if the preceding command fails (non-zero exit code)." {
		t.Errorf("unexpected summary for OR chain: %q", analysisOr.PipelineSummary)
	}

	pipeSeq, _ := ast.Parse("git pull ; git status")
	analysisSeq := analyzer.AnalyzePipeline(pipeSeq)
	if analysisSeq.PipelineSummary != "A sequential 2-stage execution (git ; git) where commands run one after another, regardless of exit status." {
		t.Errorf("unexpected summary for sequential chain: %q", analysisSeq.PipelineSummary)
	}

	pipeCompound, _ := ast.Parse("cat file | grep err && echo done")
	analysisCompound := analyzer.AnalyzePipeline(pipeCompound)
	if analysisCompound.PipelineSummary != "A compound 3-stage execution combining pipelines and conditional operators." {
		t.Errorf("unexpected summary for compound execution: %q", analysisCompound.PipelineSummary)
	}
}
