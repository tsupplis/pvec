package actiondialog

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// GetExecutingText shows action in progress
func GetExecutingText(actionName, vmName, vmid string, width, height int) string {
	var b strings.Builder

	// Dialog content
	title := "Action in Progress"
	message := fmt.Sprintf("Executing %s on %s (%s)...", actionName, vmName, vmid)
	footer := "Please wait..."

	// Calculate dialog dimensions
	dialogWidth := len(message) + 4
	if len(title)+4 > dialogWidth {
		dialogWidth = len(title) + 4
	}
	if len(footer)+4 > dialogWidth {
		dialogWidth = len(footer) + 4
	}
	if dialogWidth > width-4 {
		dialogWidth = width - 4
	}

	dialogHeight := 7
	topPadding := (height - dialogHeight) / 2
	leftPadding := (width - dialogWidth) / 2

	// Fill top padding
	for i := 0; i < topPadding; i++ {
		b.WriteString("\n")
	}

	// Top border
	b.WriteString(strings.Repeat(" ", leftPadding))
	b.WriteString("┌" + strings.Repeat("─", dialogWidth-2) + "┐\n")

	// Title
	titlePad := (dialogWidth - 4 - len(title)) / 2
	b.WriteString(strings.Repeat(" ", leftPadding))
	b.WriteString("│ " + strings.Repeat(" ", titlePad) + title + strings.Repeat(" ", dialogWidth-4-titlePad-len(title)) + " │\n")

	// Separator
	b.WriteString(strings.Repeat(" ", leftPadding))
	b.WriteString("├" + strings.Repeat("─", dialogWidth-2) + "┤\n")

	// Message
	msgPad := (dialogWidth - 4 - len(message)) / 2
	b.WriteString(strings.Repeat(" ", leftPadding))
	b.WriteString("│ " + strings.Repeat(" ", msgPad) + message + strings.Repeat(" ", dialogWidth-4-msgPad-len(message)) + " │\n")

	// Empty line
	b.WriteString(strings.Repeat(" ", leftPadding))
	b.WriteString("│" + strings.Repeat(" ", dialogWidth-2) + "│\n")

	// Footer
	footerPad := (dialogWidth - 4 - len(footer)) / 2
	b.WriteString(strings.Repeat(" ", leftPadding))
	b.WriteString("│ " + strings.Repeat(" ", footerPad) + footer + strings.Repeat(" ", dialogWidth-4-footerPad-len(footer)) + " │\n")

	// Bottom border
	b.WriteString(strings.Repeat(" ", leftPadding))
	b.WriteString("└" + strings.Repeat("─", dialogWidth-2) + "┘\n")

	return b.String()
}

// GetSuccessText shows action completed successfully
func GetSuccessText(actionName, vmName, vmid string, width, height int) string {
	var b strings.Builder

	borderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#008000"))
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#008000")).Bold(true)

	// Dialog content
	title := "Success"
	message := fmt.Sprintf("%s completed on %s (%s)", actionName, vmName, vmid)
	footer := "Press any key to continue"

	// Calculate dialog dimensions
	dialogWidth := len(message) + 4
	if len(title)+4 > dialogWidth {
		dialogWidth = len(title) + 4
	}
	if len(footer)+4 > dialogWidth {
		dialogWidth = len(footer) + 4
	}
	if dialogWidth > width-4 {
		dialogWidth = width - 4
	}

	dialogHeight := 7
	topPadding := (height - dialogHeight) / 2
	leftPadding := (width - dialogWidth) / 2

	// Fill top padding
	for i := 0; i < topPadding; i++ {
		b.WriteString("\n")
	}

	// Top border
	b.WriteString(strings.Repeat(" ", leftPadding))
	b.WriteString(borderStyle.Render("┌" + strings.Repeat("─", dialogWidth-2) + "┐"))
	b.WriteString("\n")

	// Title
	titlePad := (dialogWidth - 4 - len(title)) / 2
	b.WriteString(strings.Repeat(" ", leftPadding))
	b.WriteString(borderStyle.Render("│") + " " + titleStyle.Render(strings.Repeat(" ", titlePad)+title+strings.Repeat(" ", dialogWidth-4-titlePad-len(title))) + " " + borderStyle.Render("│"))
	b.WriteString("\n")

	// Separator
	b.WriteString(strings.Repeat(" ", leftPadding))
	b.WriteString(borderStyle.Render("├" + strings.Repeat("─", dialogWidth-2) + "┤"))
	b.WriteString("\n")

	// Message
	msgPad := (dialogWidth - 4 - len(message)) / 2
	b.WriteString(strings.Repeat(" ", leftPadding))
	b.WriteString(borderStyle.Render("│") + " " + strings.Repeat(" ", msgPad) + message + strings.Repeat(" ", dialogWidth-4-msgPad-len(message)) + " " + borderStyle.Render("│"))
	b.WriteString("\n")

	// Empty line
	b.WriteString(strings.Repeat(" ", leftPadding))
	b.WriteString(borderStyle.Render("│" + strings.Repeat(" ", dialogWidth-2) + "│"))
	b.WriteString("\n")

	// Footer
	footerPad := (dialogWidth - 4 - len(footer)) / 2
	b.WriteString(strings.Repeat(" ", leftPadding))
	b.WriteString(borderStyle.Render("│") + " " + strings.Repeat(" ", footerPad) + footer + strings.Repeat(" ", dialogWidth-4-footerPad-len(footer)) + " " + borderStyle.Render("│"))
	b.WriteString("\n")

	// Bottom border
	b.WriteString(strings.Repeat(" ", leftPadding))
	b.WriteString(borderStyle.Render("└" + strings.Repeat("─", dialogWidth-2) + "┘"))
	b.WriteString("\n")

	return b.String()
}

// GetErrorText shows action failed
func GetErrorText(actionName, vmName, vmid string, errorMsg string, width, height int) string {
	var b strings.Builder

	borderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#008000"))
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#008000")).Bold(true)

	// Dialog content
	title := "Error"
	message := fmt.Sprintf("%s failed on %s (%s)", actionName, vmName, vmid)
	footer := "Press any key to continue"

	// Calculate dialog dimensions
	dialogWidth := len(message) + 4
	if len(errorMsg)+4 > dialogWidth {
		dialogWidth = len(errorMsg) + 4
	}
	if len(title)+4 > dialogWidth {
		dialogWidth = len(title) + 4
	}
	if len(footer)+4 > dialogWidth {
		dialogWidth = len(footer) + 4
	}
	if dialogWidth > width-4 {
		dialogWidth = width - 4
	}

	dialogHeight := 8
	topPadding := (height - dialogHeight) / 2
	leftPadding := (width - dialogWidth) / 2

	// Fill top padding
	for i := 0; i < topPadding; i++ {
		b.WriteString("\n")
	}

	// Top border
	b.WriteString(strings.Repeat(" ", leftPadding))
	b.WriteString(borderStyle.Render("┌" + strings.Repeat("─", dialogWidth-2) + "┐"))
	b.WriteString("\n")

	// Title
	titlePad := (dialogWidth - 4 - len(title)) / 2
	b.WriteString(strings.Repeat(" ", leftPadding))
	b.WriteString(borderStyle.Render("│") + " " + titleStyle.Render(strings.Repeat(" ", titlePad)+title+strings.Repeat(" ", dialogWidth-4-titlePad-len(title))) + " " + borderStyle.Render("│"))
	b.WriteString("\n")

	// Separator
	b.WriteString(strings.Repeat(" ", leftPadding))
	b.WriteString(borderStyle.Render("├" + strings.Repeat("─", dialogWidth-2) + "┤"))
	b.WriteString("\n")

	// Message
	msgPad := (dialogWidth - 4 - len(message)) / 2
	b.WriteString(strings.Repeat(" ", leftPadding))
	b.WriteString(borderStyle.Render("│") + " " + strings.Repeat(" ", msgPad) + message + strings.Repeat(" ", dialogWidth-4-msgPad-len(message)) + " " + borderStyle.Render("│"))
	b.WriteString("\n")

	// Error message
	// Truncate error if too long
	maxErrLen := dialogWidth - 4
	if len(errorMsg) > maxErrLen {
		errorMsg = errorMsg[:maxErrLen-3] + "..."
	}
	errPad := (dialogWidth - 4 - len(errorMsg)) / 2
	b.WriteString(strings.Repeat(" ", leftPadding))
	b.WriteString(borderStyle.Render("│") + " " + strings.Repeat(" ", errPad) + errorMsg + strings.Repeat(" ", dialogWidth-4-errPad-len(errorMsg)) + " " + borderStyle.Render("│"))
	b.WriteString("\n")

	// Empty line
	b.WriteString(strings.Repeat(" ", leftPadding))
	b.WriteString(borderStyle.Render("│" + strings.Repeat(" ", dialogWidth-2) + "│"))
	b.WriteString("\n")

	// Footer
	footerPad := (dialogWidth - 4 - len(footer)) / 2
	b.WriteString(strings.Repeat(" ", leftPadding))
	b.WriteString(borderStyle.Render("│") + " " + strings.Repeat(" ", footerPad) + footer + strings.Repeat(" ", dialogWidth-4-footerPad-len(footer)) + " " + borderStyle.Render("│"))
	b.WriteString("\n")

	// Bottom border
	b.WriteString(strings.Repeat(" ", leftPadding))
	b.WriteString(borderStyle.Render("└" + strings.Repeat("─", dialogWidth-2) + "┘"))
	b.WriteString("\n")

	return b.String()
}
