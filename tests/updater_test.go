package tests

import (
	"testing"

	"github.com/Yehya-Elsawy/explain/pkg/updater"
)

func TestIsUpToDate(t *testing.T) {
	tests := []struct {
		current  string
		latest   string
		expected bool
	}{
		{"v2.0.0", "v2.0.0", true},
		{"v2.0.1", "v2.0.0", true},
		{"v2.1.0", "v2.0.0", true},
		{"v3.0.0", "v2.0.0", true},
		{"v1.10.0", "v1.2.0", true},
		{"v1.2.0", "v1.10.0", false},
		{"v1.9.9", "v2.0.0", false},
		{"v2.0.0", "v2.0.1", false},
		{"(devel)", "v2.0.0", true},
		{"", "v2.0.0", true},
	}

	for _, tt := range tests {
		got := updater.IsUpToDate(tt.current, tt.latest)
		if got != tt.expected {
			t.Errorf("IsUpToDate(%q, %q) = %v; want %v", tt.current, tt.latest, got, tt.expected)
		}
	}
}
