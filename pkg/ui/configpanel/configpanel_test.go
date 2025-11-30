package configpanel

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tsupplis/pvec/pkg/config"
)

// MockLoader implements config.Loader for testing
type MockLoader struct {
	SaveError error
}

func (m *MockLoader) Load() (*config.Config, error) {
	return &config.Config{
		APIUrl:          "https://test.example.com:8006",
		TokenID:         "test@pam!test",
		TokenSecret:     "secret123",
		SkipTLSVerify:   false,
		RefreshInterval: 5 * time.Second,
	}, nil
}

func (m *MockLoader) Save(cfg *config.Config) error {
	return m.SaveError
}

func TestNewModel(t *testing.T) {
	cfg := &config.Config{
		APIUrl:          "https://proxmox.local:8006",
		TokenID:         "user@pam!token",
		TokenSecret:     "secretvalue",
		SkipTLSVerify:   true,
		RefreshInterval: 10 * time.Second,
	}
	loader := &MockLoader{}

	model := NewModel(cfg, loader)

	if model.cfg != cfg {
		t.Error("Config not set correctly")
	}
	if model.loader != loader {
		t.Error("Loader not set correctly")
	}
	if len(model.inputs) != 4 {
		t.Errorf("Expected 4 inputs, got %d", len(model.inputs))
	}
	if !model.skipTLSVerify {
		t.Error("SkipTLSVerify should be true")
	}
	if model.focusedField != 0 {
		t.Error("Initial focused field should be 0")
	}
}

func TestModel_Init(t *testing.T) {
	cfg := &config.Config{}
	loader := &MockLoader{}
	model := NewModel(cfg, loader)

	cmd := model.Init()
	if cmd == nil {
		t.Error("Init should return a command")
	}
}

func TestModel_NavigationKeys(t *testing.T) {
	cfg := &config.Config{}
	loader := &MockLoader{}
	model := NewModel(cfg, loader)

	tests := []struct {
		name          string
		key           tea.KeyType
		initialFocus  int
		expectedFocus int
	}{
		{"Tab from first field", tea.KeyTab, 0, 1},
		{"Tab from last field", tea.KeyTab, 6, 0},
		{"Shift+Tab from first", tea.KeyShiftTab, 0, 6},
		{"Down from first", tea.KeyDown, 0, 1},
		{"Up from last", tea.KeyUp, 6, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model.focusedField = tt.initialFocus
			msg := tea.KeyMsg{Type: tt.key}

			updatedModel, _ := model.Update(msg)
			m := updatedModel.(Model)
			if m.focusedField != tt.expectedFocus {
				t.Errorf("Expected focus %d, got %d", tt.expectedFocus, m.focusedField)
			}
		})
	}
}

func TestModel_EscapeKey(t *testing.T) {
	cfg := &config.Config{}
	loader := &MockLoader{}
	model := NewModel(cfg, loader)

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	updatedModel, cmd := model.Update(msg)

	if updatedModel == nil {
		t.Error("Model should not be nil")
	}
	if cmd == nil {
		t.Error("Command should not be nil (should return CloseMsg)")
	}

	// Execute the command to check if it returns CloseMsg
	result := cmd()
	if _, ok := result.(CloseMsg); !ok {
		t.Error("ESC should return CloseMsg")
	}
}

func TestModel_CheckboxToggle(t *testing.T) {
	cfg := &config.Config{SkipTLSVerify: false}
	loader := &MockLoader{}
	model := NewModel(cfg, loader)
	model.focusedField = 4 // Skip TLS checkbox

	// Press Enter to toggle
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	if !m.skipTLSVerify {
		t.Error("Checkbox should be toggled to true")
	}

	// Toggle again
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(Model)

	if m.skipTLSVerify {
		t.Error("Checkbox should be toggled back to false")
	}
}

func TestModel_CancelButton(t *testing.T) {
	cfg := &config.Config{}
	loader := &MockLoader{}
	model := NewModel(cfg, loader)
	model.focusedField = 6 // Cancel button

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	_, cmd := model.Update(msg)

	if cmd == nil {
		t.Error("Cancel should return a command")
	}

	result := cmd()
	if _, ok := result.(CloseMsg); !ok {
		t.Error("Cancel should return CloseMsg")
	}
}

func TestModel_View(t *testing.T) {
	cfg := &config.Config{
		APIUrl:          "https://test.local:8006",
		TokenID:         "test@pam!token",
		TokenSecret:     "secret",
		SkipTLSVerify:   false,
		RefreshInterval: 5 * time.Second,
	}
	loader := &MockLoader{}
	model := NewModel(cfg, loader)
	model.width = 80
	model.height = 24

	view := model.View()

	if view == "" {
		t.Error("View should not be empty")
	}

	// Check for expected elements
	expectedStrings := []string{
		"Configuration",
		"API URL:",
		"Token ID:",
		"Token Secret:",
		"Refresh Interval:",
		"Skip TLS Verify",
		"[Save]",
		"[Cancel]",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(view, expected) {
			t.Errorf("View should contain '%s'", expected)
		}
	}
}

func TestModel_WindowResize(t *testing.T) {
	cfg := &config.Config{}
	loader := &MockLoader{}
	model := NewModel(cfg, loader)

	msg := tea.WindowSizeMsg{Width: 100, Height: 30}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	if m.width != 100 {
		t.Errorf("Expected width 100, got %d", m.width)
	}
	if m.height != 30 {
		t.Errorf("Expected height 30, got %d", m.height)
	}
}

func TestModel_InputValues(t *testing.T) {
	cfg := &config.Config{
		APIUrl:          "https://original.local:8006",
		TokenID:         "original@pam!token",
		TokenSecret:     "originalsecret",
		RefreshInterval: 5 * time.Second,
	}
	loader := &MockLoader{}
	model := NewModel(cfg, loader)

	if model.inputs[0].Value() != cfg.APIUrl {
		t.Errorf("API URL input should be '%s', got '%s'", cfg.APIUrl, model.inputs[0].Value())
	}
	if model.inputs[1].Value() != cfg.TokenID {
		t.Errorf("Token ID input should be '%s', got '%s'", cfg.TokenID, model.inputs[1].Value())
	}
	if model.inputs[2].Value() != cfg.TokenSecret {
		t.Errorf("Token Secret input should be '%s', got '%s'", cfg.TokenSecret, model.inputs[2].Value())
	}
	if model.inputs[3].Value() != "5s" {
		t.Errorf("Refresh Interval input should be '5s', got '%s'", model.inputs[3].Value())
	}
}
