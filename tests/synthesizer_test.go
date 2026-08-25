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
