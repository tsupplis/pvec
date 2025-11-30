package configpanel

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tsupplis/pvec/pkg/config"
)

// Model represents the config panel state
type Model struct {
	cfg            *config.Config
	loader         config.Loader
	inputs         []textinput.Model
	focusedField   int
	skipTLSVerify  bool
	width          int
	height         int
	message        string
	messageIsError bool
}

// NewModel creates a new config panel model
func NewModel(cfg *config.Config, loader config.Loader) Model {
	inputs := make([]textinput.Model, 4)

	// API URL
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "https://proxmox.example.com:8006"
	inputs[0].SetValue(cfg.APIUrl)
	inputs[0].Width = 50
	inputs[0].Focus()

	// Token ID
	inputs[1] = textinput.New()
	inputs[1].Placeholder = "user@realm!tokenid"
	inputs[1].SetValue(cfg.TokenID)
	inputs[1].Width = 50

	// Token Secret
	inputs[2] = textinput.New()
	inputs[2].Placeholder = "secret-token-value"
	inputs[2].SetValue(cfg.TokenSecret)
	inputs[2].EchoMode = textinput.EchoPassword
	inputs[2].EchoCharacter = '*'
	inputs[2].Width = 50

	// Refresh Interval
	inputs[3] = textinput.New()
	inputs[3].Placeholder = "5s"
	inputs[3].SetValue(cfg.RefreshInterval.String())
	inputs[3].Width = 20

	return Model{
		cfg:           cfg,
		loader:        loader,
		inputs:        inputs,
		focusedField:  0,
		skipTLSVerify: cfg.SkipTLSVerify,
	}
}

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

// SaveResultMsg is sent when save completes
type SaveResultMsg struct {
	err error
}

// CloseMsg is sent when config panel wants to close
type CloseMsg struct{}

// Update implements tea.Model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case SaveResultMsg:
		return m.handleSaveResult(msg)

	case tea.KeyMsg:
		return m.handleKeyMsg(msg)
	}

	// Update the focused input field
	if m.focusedField < 4 {
		var cmd tea.Cmd
		m.inputs[m.focusedField], cmd = m.inputs[m.focusedField].Update(msg)
		return m, cmd
	}

	return m, nil
}

// handleSaveResult processes save operation results
func (m Model) handleSaveResult(msg SaveResultMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.message = fmt.Sprintf("Error: %v", msg.err)
		m.messageIsError = true
	} else {
		m.message = "Configuration saved successfully! Restart required."
		m.messageIsError = false
	}
	return m, nil
}

// handleKeyMsg processes keyboard input
func (m Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		return m, func() tea.Msg { return CloseMsg{} }
	case "tab", "down":
		return m.handleNavigationNext()
	case "shift+tab", "up":
		return m.handleNavigationPrev()
	case "enter", " ":
		return m.handleAction()
	}
	return m, nil
}

// handleNavigationNext moves focus to the next field
func (m Model) handleNavigationNext() (tea.Model, tea.Cmd) {
	if m.focusedField < 4 {
		m.inputs[m.focusedField].Blur()
	}
	m.focusedField++
	if m.focusedField > 6 {
		m.focusedField = 0
	}
	if m.focusedField < 4 {
		m.inputs[m.focusedField].Focus()
	}
	return m, nil
}

// handleNavigationPrev moves focus to the previous field
func (m Model) handleNavigationPrev() (tea.Model, tea.Cmd) {
	if m.focusedField < 4 {
		m.inputs[m.focusedField].Blur()
	}
	m.focusedField--
	if m.focusedField < 0 {
		m.focusedField = 6
	}
	if m.focusedField < 4 {
		m.inputs[m.focusedField].Focus()
	}
	return m, nil
}

// handleAction handles enter/space key on focused element
func (m Model) handleAction() (tea.Model, tea.Cmd) {
	switch m.focusedField {
	case 4: // Skip TLS checkbox
		m.skipTLSVerify = !m.skipTLSVerify
		return m, nil
	case 5: // Save button
		return m, m.save()
	case 6: // Cancel button
		return m, func() tea.Msg { return CloseMsg{} }
	}
	return m, nil
}

// save validates and saves the configuration
func (m *Model) save() tea.Cmd {
	return func() tea.Msg {
		// Validate and update config
		m.cfg.APIUrl = m.inputs[0].Value()
		m.cfg.TokenID = m.inputs[1].Value()
		m.cfg.TokenSecret = m.inputs[2].Value()
		m.cfg.SkipTLSVerify = m.skipTLSVerify

		// Parse refresh interval
		interval, err := time.ParseDuration(m.inputs[3].Value())
		if err != nil {
			return SaveResultMsg{err: fmt.Errorf("invalid refresh interval: %v", err)}
		}
		m.cfg.RefreshInterval = interval

		// Validate required fields
		if m.cfg.APIUrl == "" {
			return SaveResultMsg{err: fmt.Errorf("api URL is required")}
		}
		if m.cfg.TokenID == "" {
			return SaveResultMsg{err: fmt.Errorf("token ID is required")}
		}
		if m.cfg.TokenSecret == "" {
			return SaveResultMsg{err: fmt.Errorf("token secret is required")}
		}

		// Save configuration
		if vl, ok := m.loader.(*config.ViperLoader); ok {
			if err := vl.Save(m.cfg); err != nil {
				return SaveResultMsg{err: fmt.Errorf("failed to save: %v", err)}
			}
		}

		return SaveResultMsg{err: nil}
	}
}

// View implements tea.Model
func (m Model) View() string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#008000")).Bold(true)
	separatorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#008000"))
	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#008000")).Bold(true)

	// Title line
	title := "Configuration"
	titlePadding := (m.width - len(title)) / 2
	if titlePadding > 0 {
		b.WriteString(strings.Repeat(" ", titlePadding))
	}
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n")

	// Separator line
	b.WriteString(separatorStyle.Render(strings.Repeat("─", m.width)))
	b.WriteString("\n")

	// Blank line after separator
	b.WriteString("\n")

	// Calculate left padding for form
	formWidth := 60
	leftPadding := (m.width - formWidth) / 2
	if leftPadding < 2 {
		leftPadding = 2
	}
	padding := strings.Repeat(" ", leftPadding)

	// API URL field
	b.WriteString(padding)
	b.WriteString("API URL:")
	b.WriteString("\n")
	b.WriteString(padding)
	b.WriteString(m.inputs[0].View())
	b.WriteString("\n\n")

	// Token ID field
	b.WriteString(padding)
	b.WriteString("Token ID:")
	b.WriteString("\n")
	b.WriteString(padding)
	b.WriteString(m.inputs[1].View())
	b.WriteString("\n\n")

	// Token Secret field
	b.WriteString(padding)
	b.WriteString("Token Secret:")
	b.WriteString("\n")
	b.WriteString(padding)
	b.WriteString(m.inputs[2].View())
	b.WriteString("\n\n")

	// Refresh Interval field
	b.WriteString(padding)
	b.WriteString("Refresh Interval:")
	b.WriteString("\n")
	b.WriteString(padding)
	b.WriteString(m.inputs[3].View())
	b.WriteString("\n\n")

	// Skip TLS Verify checkbox
	b.WriteString(padding)
	checkbox := "[ ]"
	if m.skipTLSVerify {
		checkbox = "[X]"
	}
	if m.focusedField == 4 {
		b.WriteString("> " + checkbox + " Skip TLS Verify")
	} else {
		b.WriteString("  " + checkbox + " Skip TLS Verify")
	}
	b.WriteString("\n\n")

	// Buttons
	b.WriteString(padding)
	if m.focusedField == 5 {
		b.WriteString("> [Save] ")
	} else {
		b.WriteString("  [Save] ")
	}
	if m.focusedField == 6 {
		b.WriteString("> [Cancel]")
	} else {
		b.WriteString("  [Cancel]")
	}
	b.WriteString("\n")

	// Count lines written so far: title + separator + blank + fields
	// Title(1) + separator(1) + blank(1) +
	// API label+input+blank(3) + Token label+input+blank(3) +
	// Secret label+input+blank(3) + Interval label+input+blank(3) +
	// checkbox+blank(2) + buttons(1) = 18 total
	usedLines := 18

	// Add message line if present
	if m.message != "" {
		b.WriteString("\n")
		b.WriteString(padding)
		if m.messageIsError {
			b.WriteString(m.message + " - Press ESC to continue")
		} else {
			b.WriteString(m.message + " - Press ESC to close")
		}
		b.WriteString("\n")
		// Added: blank line(1) + message(1) + blank line(1) = 3 lines
		usedLines += 3
	}

	// Fill space to push status bar to last line
	// Total should be: usedLines + fillLines + statusBar(1) = height
	// So fillLines = height - usedLines - 1
	fillLines := m.height - usedLines - 1
	if fillLines > 0 {
		for i := 0; i < fillLines; i++ {
			b.WriteString("\n")
		}
	}

	// Status bar on the last line (no trailing newline)
	b.WriteString(statusStyle.Render(" Tab/↑↓: Navigate | Enter/Space: Select | ESC: Close"))

	return b.String()
}
