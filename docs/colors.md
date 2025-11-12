# Color Theme System

## Overview

The pvec application now uses a centralized color theme system defined in `pkg/ui/colors/colors.go`. This makes it easy to customize the application's appearance by modifying the color definitions in one place.

## Theme Structure

```go
type Theme struct {
    Background       tcell.Color  // Main background color
    Foreground       tcell.Color  // Main text color
    ActiveBackground tcell.Color  // Background for active/selected items
    ActiveForeground tcell.Color  // Text color for active/selected items
    AccentForeground tcell.Color  // Color for highlighting keys/shortcuts
    AlertColor       tcell.Color  // Color for critical states (high usage, errors)
    WarningColor     tcell.Color  // Color for warning states (medium usage)
    OkColor          tcell.Color  // Color for normal/running states
    DisabledColor    tcell.Color  // Color for disabled/stopped states
}
```

## Default Colors

- **Background**: Black
- **Foreground**: White
- **ActiveBackground**: Dark Green
- **ActiveForeground**: White
- **AccentForeground**: Green (used for function key shortcuts)
- **AlertColor**: Red (used for CPU/Memory >80%)
- **WarningColor**: Orange (used for CPU/Memory >50%)
- **OkColor**: Green (used for running VMs/CTs)
- **DisabledColor**: Gray (used for stopped VMs/CTs)

Additional fixed colors:
- **VMColor**: Light Blue (for VM type indicator)
- **CTColor**: Light Cyan (for CT type indicator)

## Usage

All UI components now reference `colors.Current` instead of hardcoded colors:

```go
import "github.com/tsupplis/pvec/pkg/ui/colors"

// Example: Setting text color
cell.SetTextColor(colors.Current.Foreground)

// Example: Conditional colors for status
if status == "running" {
    cell.SetTextColor(colors.Current.OkColor)
} else if status == "stopped" {
    cell.SetTextColor(colors.Current.DisabledColor)
}

// Example: Conditional colors for resource usage
if cpuPercent > 80 {
    cell.SetTextColor(colors.Current.AlertColor)
} else if cpuPercent > 50 {
    cell.SetTextColor(colors.Current.WarningColor)
}
```

## Customization

To customize the color scheme, modify the `DefaultTheme` variable in `pkg/ui/colors/colors.go`:

```go
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
```

Future enhancement: Allow theme customization via configuration file.

## Components Using Color Theme

1. **Main List** (`pkg/ui/mainlist/mainlist.go`):
   - Headers use `AccentForeground`
   - Status indicators use `OkColor`, `AlertColor`, `WarningColor`, `DisabledColor`
   - Type indicators use `VMColor` and `CTColor`
   - Resource usage uses conditional `AlertColor`/`WarningColor`
   - All text uses `Foreground`

2. **Status Bar** (`main.go`):
   - Function keys use `AccentForeground`
   - Descriptions use `Foreground`

## Testing

The color theme system includes comprehensive tests in `pkg/ui/colors/colors_test.go` that verify:
- All default theme colors are set correctly
- VM and CT colors are correct
- Current theme is initialized properly
- Theme can be modified at runtime
