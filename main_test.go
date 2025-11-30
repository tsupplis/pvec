package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetConfigPath_WithFlagProvided(t *testing.T) {
	// Test that when a config path is provided via flag, it's used directly
	result := getConfigPath("/custom/path/config.json")
	expected := "/custom/path/config.json"

	if result != expected {
		t.Errorf("Expected config path %s, got %s", expected, result)
	}
}

func TestGetConfigPath_WithEmptyFlag(t *testing.T) {
	// Test that when no flag is provided, it defaults to ~/.pvecrc
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get user home directory: %v", err)
	}

	result := getConfigPath("")
	expected := filepath.Join(homeDir, ".pvecrc")

	if result != expected {
		t.Errorf("Expected config path %s, got %s", expected, result)
	}
}

func TestGetConfigPath_WithRelativePath(t *testing.T) {
	// Test that relative paths are accepted as-is
	result := getConfigPath("./config/test.json")
	expected := "./config/test.json"

	if result != expected {
		t.Errorf("Expected config path %s, got %s", expected, result)
	}
}
