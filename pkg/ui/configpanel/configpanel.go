package configpanel

import (
	"fmt"
	"time"

	"github.com/rivo/tview"
	"github.com/tsupplis/pvec/pkg/config"
	"github.com/tsupplis/pvec/pkg/ui/colors"
)

// ConfigPanel displays and edits configuration
type ConfigPanel struct {
	form   *tview.Form
	pages  *tview.Pages
	cfg    *config.Config
	loader config.Loader
}

// NewConfigPanel creates a new configuration panel
func NewConfigPanel(pages *tview.Pages, cfg *config.Config, loader config.Loader) *ConfigPanel {
	cp := &ConfigPanel{
		form:   tview.NewForm(),
		pages:  pages,
		cfg:    cfg,
		loader: loader,
	}

	cp.buildForm()
	return cp
}

// buildForm constructs the form fields
func (cp *ConfigPanel) buildForm() {
	cp.form.Clear(true)

	cp.form.AddInputField("API URL", cp.cfg.APIUrl, 50, nil, func(text string) {
		cp.cfg.APIUrl = text
	})

	cp.form.AddInputField("Token ID", cp.cfg.TokenID, 50, nil, func(text string) {
		cp.cfg.TokenID = text
	})

	cp.form.AddPasswordField("Token Secret", cp.cfg.TokenSecret, 50, '*', func(text string) {
		cp.cfg.TokenSecret = text
	})

	intervalStr := cp.cfg.RefreshInterval.String()
	cp.form.AddInputField("Refresh Interval", intervalStr, 20, nil, func(text string) {
		// Validation happens on save
	})

	cp.form.AddCheckbox("Skip TLS Verify", cp.cfg.SkipTLSVerify, func(checked bool) {
		cp.cfg.SkipTLSVerify = checked
	})

	cp.form.AddButton("Save", cp.handleSave)
	cp.form.AddButton("Cancel", cp.handleCancel)

	cp.form.SetBorder(true).
		SetTitle(" Configuration ").
		SetTitleAlign(tview.AlignLeft)

	cp.form.SetBackgroundColor(colors.Current.Background).
		SetBorderColor(colors.Current.Foreground).
		SetTitleColor(colors.Current.Foreground)

	cp.form.SetLabelColor(colors.Current.Foreground).
		SetFieldBackgroundColor(colors.Current.ActiveBackground).
		SetFieldTextColor(colors.Current.ActiveForeground).
		SetButtonBackgroundColor(colors.Current.ActiveBackground).
		SetButtonTextColor(colors.Current.ActiveForeground)

	cp.form.SetCancelFunc(func() {
		cp.handleCancel()
	})
}

// handleSave saves the configuration
func (cp *ConfigPanel) handleSave() {
	// Validate refresh interval
	intervalField := cp.form.GetFormItemByLabel("Refresh Interval").(*tview.InputField)
	intervalStr := intervalField.GetText()

	interval, err := time.ParseDuration(intervalStr)
	if err != nil {
		cp.showError(fmt.Sprintf("Invalid refresh interval: %v", err))
		return
	}
	cp.cfg.RefreshInterval = interval

	// Validate required fields
	if cp.cfg.APIUrl == "" {
		cp.showError("API URL is required")
		return
	}
	if cp.cfg.TokenID == "" {
		cp.showError("Token ID is required")
		return
	}
	if cp.cfg.TokenSecret == "" {
		cp.showError("Token Secret is required")
		return
	}

	// Save configuration
	if vl, ok := cp.loader.(*config.ViperLoader); ok {
		if err := vl.Save(cp.cfg); err != nil {
			cp.showError(fmt.Sprintf("Failed to save: %v", err))
			return
		}
	}

	cp.showSuccess("Configuration saved successfully!")
}

// handleCancel closes the panel without saving
func (cp *ConfigPanel) handleCancel() {
	cp.pages.RemovePage("config")
}

// showError displays an error message
func (cp *ConfigPanel) showError(message string) {
	modal := tview.NewModal().
		SetText(fmt.Sprintf("[%s]Error:[%s] %s",
			colors.Current.AlertColor.Name(),
			colors.Current.Foreground.Name(),
			message)).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			cp.pages.RemovePage("error")
		})

	modal.SetBackgroundColor(colors.Current.Background)
	modal.Box.SetBackgroundColor(colors.Current.Background).
		SetBorderColor(colors.Current.Foreground)

	// Wrap in flex for consistent background
	flex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(modal, 0, 1, true).
		AddItem(nil, 0, 1, false)
	flex.SetBackgroundColor(colors.Current.Background)

	cp.pages.AddPage("error", flex, true, true)
}

// showSuccess displays a success message
func (cp *ConfigPanel) showSuccess(message string) {
	modal := tview.NewModal().
		SetText(fmt.Sprintf("[%s]Success:[%s] %s\n\nRestart required for changes to take effect.",
			colors.Current.OkColor.Name(),
			colors.Current.Foreground.Name(),
			message)).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			cp.pages.RemovePage("success")
			cp.pages.RemovePage("config")
		})

	modal.SetBackgroundColor(colors.Current.Background)
	modal.Box.SetBackgroundColor(colors.Current.Background).
		SetBorderColor(colors.Current.Foreground)

	// Wrap in flex for consistent background
	flex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(modal, 0, 1, true).
		AddItem(nil, 0, 1, false)
	flex.SetBackgroundColor(colors.Current.Background)

	cp.pages.AddPage("success", flex, true, true)
}

// Show displays the configuration panel
func (cp *ConfigPanel) Show() {
	// Create a centered container with proper background
	innerFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(cp.form, 15, 1, true).
		AddItem(nil, 0, 1, false)
	innerFlex.SetBackgroundColor(colors.Current.Background)

	flex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(innerFlex, 80, 1, true).
		AddItem(nil, 0, 1, false)
	flex.SetBackgroundColor(colors.Current.Background)

	cp.pages.AddPage("config", flex, true, true)
}

// GetForm returns the underlying form
func (cp *ConfigPanel) GetForm() *tview.Form {
	return cp.form
}
