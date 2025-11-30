package helpdialog

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// GetHelpText returns the formatted help text
func GetHelpText(width, height int) string {
	var b strings.Builder

	// Help content - organized as key/action pairs
	type helpItem struct {
		keys   string
		action string
	}

	sections := []struct {
		title string
		items []helpItem
	}{
		{
			title: "Navigation:",
			items: []helpItem{
				{"↑ / k", "Move up"},
				{"↓ / j", "Move down"},
				{"PgUp", "Scroll page up"},
				{"PgDn", "Scroll page down"},
			},
		},
		{
			title: "Actions:",
			items: []helpItem{
				{"F1 / h", "Show this help"},
				{"F2 / c", "Configuration"},
				{"F3 / i", "Show VM/CT details"},
				{"F4 / s", "Start VM/CT"},
				{"F5 / d", "Shutdown VM/CT"},
				{"F6 / r", "Reboot VM/CT"},
				{"F7 / t", "Stop VM/CT"},
				{"F10 / q", "Quit application"},
				{"Ctrl+C", "Quit application"},
			},
		},
	}

	// Title and separator first
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#008000")).Bold(true)
	separatorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#008000"))
	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#008000")).Bold(true)

	title := "Help - Keyboard Shortcuts"
	titlePadding := (width - len(title)) / 2
	if titlePadding > 0 {
		b.WriteString(strings.Repeat(" ", titlePadding))
	}
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n")
	b.WriteString(separatorStyle.Render(strings.Repeat("─", width)))
	b.WriteString("\n\n")

	// Build help lines with proper column alignment
	var helpLines []string
	keyColWidth := 15 // Width for the key column

	for _, section := range sections {
		helpLines = append(helpLines, section.title)
		for _, item := range section.items {
			// Format: "  key" + padding + "action"
			line := "  " + item.keys
			// Pad to keyColWidth using rune count, then add action
			lineRunes := []rune(line)
			if len(lineRunes) < keyColWidth {
				line += strings.Repeat(" ", keyColWidth-len(lineRunes))
			}
			line += item.action
			helpLines = append(helpLines, line)
		}
		helpLines = append(helpLines, "") // Empty line between sections
	}

	// Center content vertically (accounting for title and separator already written)
	contentHeight := len(helpLines)
	topPadding := (height - contentHeight - 4) / 2 // 4 for title, separator, and status bar
	if topPadding > 2 {                            // Already have 2 lines for title and separator
		for i := 0; i < topPadding-2; i++ {
			b.WriteString("\n")
		}
	}

	// Calculate left padding to center the content block (not individual lines)
	// Find the longest line to determine content width
	maxLineLen := 0
	for _, line := range helpLines {
		if len(line) > maxLineLen {
			maxLineLen = len(line)
		}
	}
	leftPadding := (width - maxLineLen) / 2
	if leftPadding < 0 {
		leftPadding = 2 // Minimum padding
	}

	for _, line := range helpLines {
		// Apply same left padding to all lines (left-aligned columns)
		if leftPadding > 0 {
			b.WriteString(strings.Repeat(" ", leftPadding))
		}
		b.WriteString(line)
		b.WriteString("\n")
	}

	// Fill remaining space
	// Calculate total lines used: title(1) + separator(1) + blank line(1) + topPadding + contentHeight
	usedLines := 3 + (topPadding - 2) + contentHeight
	if usedLines < 0 {
		usedLines = 3 + contentHeight
	}

	for i := usedLines; i < height-1; i++ {
		b.WriteString("\n")
	}

	// Status bar
	b.WriteString(statusStyle.Render(" Press ESC or Enter to close"))

	return b.String()
}
