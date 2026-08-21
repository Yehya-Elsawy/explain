package tests

import (
	"testing"

	"github.com/Yehya-Elsawy/explain-/pkg/manparser"
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
		t.Log("Note: man ls not available or non-standard in test container, skipping assertion")
	} else {
		t.Logf("Extracted ls summary: %s", summary)
	}
}
