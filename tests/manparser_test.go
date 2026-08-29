package tests

import (
	"strings"
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

func TestParseBSDManPage(t *testing.T) {
	manPage := `CAFFEINATE(8) System Manager's Manual CAFFEINATE(8)

NAME
     caffeinate – prevent the system from sleeping on behalf of a utility

SYNOPSIS
     caffeinate [-disu] [-t timeout] [-w pid] [utility arguments...]

DESCRIPTION
     Available options:

     -d      Create an assertion to prevent the display from sleeping.

     -s      Create an assertion to prevent the system from sleeping. This
             assertion is valid only when system is running on AC power.

     -t      Specifies the timeout value in seconds for which this assertion
             has to be valid. The assertion is dropped after the specified
             timeout.
`

	if got := manparser.ParseCommandSummary(manPage); got != "prevent the system from sleeping on behalf of a utility" {
		t.Fatalf("unexpected BSD-style summary: %q", got)
	}

	display := manparser.ParseFlagInfo(manPage, "-d")
	if display.Description != "Create an assertion to prevent the display from sleeping." {
		t.Errorf("unexpected same-line description: %q", display.Description)
	}
	if display.TakesValue {
		t.Error("expected -d not to consume a value")
	}

	system := manparser.ParseFlagInfo(manPage, "-s")
	if !strings.Contains(system.Description, "valid only when system is running on AC power") {
		t.Errorf("expected wrapped description, got %q", system.Description)
	}

	timeout := manparser.ParseFlagInfo(manPage, "-t")
	if !timeout.TakesValue || timeout.ValueName != "timeout" {
		t.Errorf("expected -t to consume timeout, got %+v", timeout)
	}
	if !strings.Contains(timeout.Description, "Specifies the timeout value in seconds") {
		t.Errorf("unexpected timeout description: %q", timeout.Description)
	}
}

func TestParseGNUStyleOptions(t *testing.T) {
	manPage := `NAME
    example - demonstrate GNU-style options

SYNOPSIS
    example [OPTION]

OPTIONS
    -o, --output=FILE
           write output to FILE instead of standard output

    --color[=WHEN]  colorize the output; WHEN can be always, auto, or never
`

	output := manparser.ParseFlagInfo(manPage, "-o")
	if output.Description != "write output to FILE instead of standard output" {
		t.Errorf("unexpected next-line description: %q", output.Description)
	}
	if !output.TakesValue || output.ValueName != "FILE" {
		t.Errorf("expected output flag value metadata, got %+v", output)
	}

	color := manparser.ParseFlagInfo(manPage, "--color")
	if !color.TakesValue || color.ValueName != "WHEN" {
		t.Errorf("expected optional color value metadata, got %+v", color)
	}
	if !strings.Contains(color.Description, "colorize the output") {
		t.Errorf("unexpected inline help description: %q", color.Description)
	}
}
