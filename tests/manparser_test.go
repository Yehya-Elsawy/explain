package tests

import (
	"testing"

	"github.com/Yehya-Elsawy/explain/pkg/manparser"
)

func TestStripOverstrike(t *testing.T) {
	input := "_\bt_\be_\bs_\bt"
	expected := "test"
	actual := manparser.StripOverstrike(input)
	if actual != expected {
		t.Errorf("expected %s, got %s", expected, actual)
	}
}

func TestDynamicManExtraction(t *testing.T) {
	summary := manparser.ExtractCommandSummary("ls")
	if summary == "" {
		t.Skip("man ls not available in test environment")
	}
}
