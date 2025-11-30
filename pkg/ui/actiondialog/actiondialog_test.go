package actiondialog

import (
"strings"
"testing"
)

func TestGetExecutingText(t *testing.T) {
	result := GetExecutingText("start", "test-vm", "100", 80, 24)

	if result == "" {
		t.Error("GetExecutingText returned empty string")
	}

	if !strings.Contains(result, "Action in Progress") {
		t.Error("Missing title")
	}

	if !strings.Contains(result, "start") || !strings.Contains(result, "test-vm") {
		t.Error("Missing action or VM info")
	}
}

func TestGetSuccessText(t *testing.T) {
	result := GetSuccessText("start", "test-vm", "100", 80, 24)

	if result == "" {
		t.Error("GetSuccessText returned empty string")
	}

	if !strings.Contains(result, "Success") {
		t.Error("Missing title")
	}

	if !strings.Contains(result, "completed") {
		t.Error("Missing completed message")
	}
}

func TestGetErrorText(t *testing.T) {
	result := GetErrorText("start", "test-vm", "100", "Connection failed", 80, 24)

	if result == "" {
		t.Error("GetErrorText returned empty string")
	}

	if !strings.Contains(result, "Error") {
		t.Error("Missing title")
	}

	if !strings.Contains(result, "failed") || !strings.Contains(result, "Connection failed") {
		t.Error("Missing error info")
	}
}
