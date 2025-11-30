package helpdialog

import (
	"strings"
	"testing"
)

func TestGetHelpText(t *testing.T) {
	result := GetHelpText(80, 24)

	if result == "" {
		t.Error("GetHelpText returned empty string")
	}

	if !strings.Contains(result, "Help - Keyboard Shortcuts") {
		t.Error("Help text missing title")
	}

	if !strings.Contains(result, "Press ESC or Enter to close") {
		t.Error("Help text missing status bar")
	}

	if !strings.Contains(result, "Navigation:") {
		t.Error("Help text missing Navigation section")
	}

	if !strings.Contains(result, "Actions:") {
		t.Error("Help text missing Actions section")
	}
}

func TestGetHelpText_ContainsKeyBindings(t *testing.T) {
	result := GetHelpText(80, 24)

	keys := []string{"F1", "F2", "F3", "F4", "F5", "F6", "F7", "F10"}
	for _, key := range keys {
		if !strings.Contains(result, key) {
			t.Errorf("Missing key: %s", key)
		}
	}
}

func TestGetHelpText_MinimalDimensions(t *testing.T) {
	result := GetHelpText(20, 10)
	if result == "" {
		t.Error("Failed with minimal dimensions")
	}
}
