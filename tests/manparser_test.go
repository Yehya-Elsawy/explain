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
	expected := "list directory contents"

	if summary == "" {
		t.Skip("Note: man ls not available or non-standard in test container, skipping assertion.")
	}

	if summary != expected {
		t.Errorf("Expected %s, got %s", expected, summary)
	}
}
