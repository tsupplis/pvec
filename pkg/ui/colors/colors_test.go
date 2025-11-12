package colors

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestDefaultTheme(t *testing.T) {
	tests := []struct {
		name     string
		color    tcell.Color
		expected tcell.Color
	}{
		{"Background", DefaultTheme.Background, tcell.ColorBlack},
		{"Foreground", DefaultTheme.Foreground, tcell.ColorWhite},
		{"ActiveBackground", DefaultTheme.ActiveBackground, tcell.ColorDarkGreen},
		{"ActiveForeground", DefaultTheme.ActiveForeground, tcell.ColorWhite},
		{"AccentForeground", DefaultTheme.AccentForeground, tcell.ColorGreen},
		{"AlertColor", DefaultTheme.AlertColor, tcell.ColorRed},
		{"WarningColor", DefaultTheme.WarningColor, tcell.ColorOrange},
		{"OkColor", DefaultTheme.OkColor, tcell.ColorGreen},
		{"DisabledColor", DefaultTheme.DisabledColor, tcell.ColorGray},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.color != tt.expected {
				t.Errorf("%s = %v, want %v", tt.name, tt.color, tt.expected)
			}
		})
	}
}

func TestVMColor(t *testing.T) {
	if VMColor != tcell.ColorLightBlue {
		t.Errorf("VMColor = %v, want %v", VMColor, tcell.ColorLightBlue)
	}
}

func TestCTColor(t *testing.T) {
	if CTColor != tcell.ColorLightCyan {
		t.Errorf("CTColor = %v, want %v", CTColor, tcell.ColorLightCyan)
	}
}

func TestCurrentTheme(t *testing.T) {
	// Current should be initialized to DefaultTheme
	if Current.Background != DefaultTheme.Background {
		t.Errorf("Current.Background = %v, want %v", Current.Background, DefaultTheme.Background)
	}
	if Current.Foreground != DefaultTheme.Foreground {
		t.Errorf("Current.Foreground = %v, want %v", Current.Foreground, DefaultTheme.Foreground)
	}
	if Current.ActiveBackground != DefaultTheme.ActiveBackground {
		t.Errorf("Current.ActiveBackground = %v, want %v", Current.ActiveBackground, DefaultTheme.ActiveBackground)
	}
	if Current.ActiveForeground != DefaultTheme.ActiveForeground {
		t.Errorf("Current.ActiveForeground = %v, want %v", Current.ActiveForeground, DefaultTheme.ActiveForeground)
	}
	if Current.AccentForeground != DefaultTheme.AccentForeground {
		t.Errorf("Current.AccentForeground = %v, want %v", Current.AccentForeground, DefaultTheme.AccentForeground)
	}
	if Current.AlertColor != DefaultTheme.AlertColor {
		t.Errorf("Current.AlertColor = %v, want %v", Current.AlertColor, DefaultTheme.AlertColor)
	}
	if Current.WarningColor != DefaultTheme.WarningColor {
		t.Errorf("Current.WarningColor = %v, want %v", Current.WarningColor, DefaultTheme.WarningColor)
	}
	if Current.OkColor != DefaultTheme.OkColor {
		t.Errorf("Current.OkColor = %v, want %v", Current.OkColor, DefaultTheme.OkColor)
	}
	if Current.DisabledColor != DefaultTheme.DisabledColor {
		t.Errorf("Current.DisabledColor = %v, want %v", Current.DisabledColor, DefaultTheme.DisabledColor)
	}
}

func TestThemeModification(t *testing.T) {
	// Save original
	original := Current

	// Modify theme
	Current = Theme{
		Background:       tcell.ColorBlue,
		Foreground:       tcell.ColorYellow,
		ActiveBackground: tcell.ColorRed,
		ActiveForeground: tcell.ColorBlack,
		AccentForeground: tcell.ColorPurple,
		AlertColor:       tcell.ColorOrange,
		WarningColor:     tcell.ColorYellow,
		OkColor:          tcell.ColorBlue,
		DisabledColor:    tcell.ColorDarkGray,
	}

	// Verify modification
	if Current.Background != tcell.ColorBlue {
		t.Errorf("Modified Current.Background = %v, want %v", Current.Background, tcell.ColorBlue)
	}
	if Current.AccentForeground != tcell.ColorPurple {
		t.Errorf("Modified Current.AccentForeground = %v, want %v", Current.AccentForeground, tcell.ColorPurple)
	}

	// Restore original
	Current = original
}
