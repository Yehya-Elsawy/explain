package tests

import (
	"testing"

	"github.com/Yehya-Elsawy/explain/pkg/analyzer"
	"github.com/Yehya-Elsawy/explain/pkg/ast"
	"github.com/Yehya-Elsawy/explain/pkg/database"
)

func TestDangerRmRfRoot(t *testing.T) {
	pipe, _ := ast.Parse("rm -rf /")
	analysis := analyzer.AnalyzePipeline(pipe)
	cmd := analysis.Commands[0]

	if cmd.Danger.Level != database.RiskCritical {
		t.Errorf("expected Critical danger for rm -rf /, got %v", cmd.Danger.Level)
	}
}

func TestDangerDdDev(t *testing.T) {
	pipe, _ := ast.Parse("dd if=ubuntu.iso of=/dev/sda bs=4M")
	analysis := analyzer.AnalyzePipeline(pipe)
	cmd := analysis.Commands[0]

	if cmd.Danger.Level != database.RiskCritical {
		t.Errorf("expected Critical danger for dd of=/dev/sda, got %v", cmd.Danger.Level)
	}
}

func TestDangerCurlToBash(t *testing.T) {
	pipe, _ := ast.Parse("curl -sSL https://example.com/install.sh | bash")
	analysis := analyzer.AnalyzePipeline(pipe)

	if analysis.MaxRisk != database.RiskCritical {
		t.Errorf("expected Critical danger for curl | bash, got %v", analysis.MaxRisk)
	}
}

func TestDangerChmod777(t *testing.T) {
	pipe, _ := ast.Parse("chmod -R 777 /var/www")
	analysis := analyzer.AnalyzePipeline(pipe)
	cmd := analysis.Commands[0]

	if cmd.Danger.Level != database.RiskHigh {
		t.Errorf("expected High risk for chmod 777, got %v", cmd.Danger.Level)
	}
}
