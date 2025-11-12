package colors

import "github.com/gdamore/tcell/v2"

// Theme defines the color scheme for the application
type Theme struct {
	Background       tcell.Color
	Foreground       tcell.Color
	ActiveBackground tcell.Color
	ActiveForeground tcell.Color
	AccentForeground tcell.Color
	AlertColor       tcell.Color
	WarningColor     tcell.Color
	OkColor          tcell.Color
	DisabledColor    tcell.Color
}

// Default theme with standard colors
var DefaultTheme = Theme{
	Background:       tcell.ColorBlack,
	Foreground:       tcell.ColorWhite,
	ActiveBackground: tcell.ColorDarkGreen,
	ActiveForeground: tcell.ColorWhite,
	AccentForeground: tcell.ColorGreen,
	AlertColor:       tcell.ColorRed,
	WarningColor:     tcell.ColorOrange,
	OkColor:          tcell.ColorGreen,
	DisabledColor:    tcell.ColorGray,
}

// Current active theme
var Current = DefaultTheme

// Type colors for VM and CT
var (
	VMColor = tcell.ColorLightBlue
	CTColor = tcell.ColorLightCyan
)
