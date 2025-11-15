package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/tsupplis/pvec/pkg/actions"
	"github.com/tsupplis/pvec/pkg/config"
	"github.com/tsupplis/pvec/pkg/models"
	"github.com/tsupplis/pvec/pkg/proxmox"
	"github.com/tsupplis/pvec/pkg/ui/colors"
	"github.com/tsupplis/pvec/pkg/ui/configpanel"
	"github.com/tsupplis/pvec/pkg/ui/detailsdialog"
	"github.com/tsupplis/pvec/pkg/ui/helpdialog"
	"github.com/tsupplis/pvec/pkg/ui/mainlist"
)

// initBorders sets single-line borders globally for all tview components
func initBorders() {
	tview.Borders.HorizontalFocus = '─'
	tview.Borders.VerticalFocus = '│'
	tview.Borders.TopLeftFocus = '┌'
	tview.Borders.TopRightFocus = '┐'
	tview.Borders.BottomLeftFocus = '└'
	tview.Borders.BottomRightFocus = '┘'
}

// parseFlags handles command-line flags and returns the config file path
func parseFlags() string {
	var showVersion bool
	flag.BoolVar(&showVersion, "v", false, "Show version information")
	flag.BoolVar(&showVersion, "version", false, "Show version information")

	configPath := flag.String("c", "", "Path to configuration file")
	flag.StringVar(configPath, "config", "", "Path to configuration file")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: pvec [options]\n")
		fmt.Fprintf(os.Stderr, "A terminal-based interface for managing Proxmox VMs and Containers\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fmt.Fprintf(os.Stderr, "  -c, --config   Path to configuration file (default: ~/.pvecrc)\n")
		fmt.Fprintf(os.Stderr, "  -v, --version  Show version information\n")
		fmt.Fprintf(os.Stderr, "  -h, --help     Print this help\n")
	}

	flag.Parse()

	if showVersion {
		fmt.Println("pvec version 1.0.0")
		os.Exit(0)
	}

	return getConfigPath(*configPath)
}

// getConfigPath returns the configuration file path, using default if not provided
func getConfigPath(configPath string) string {
	if configPath != "" {
		return configPath
	}

	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to get home directory: %v", err)
	}
	return filepath.Join(home, ".pvecrc")
}

// createStatusBar creates the status bar showing keyboard shortcuts
func createStatusBar() *tview.TextView {
	return tview.NewTextView().
		SetDynamicColors(true).
		SetText(fmt.Sprintf(" [%s]F1 H[%s]elp  [%s]F2 C[%s]onfig  [%s]F3 S[%s]tart  [%s]F4[%s] Shut[%s]d[%s]own  [%s]F5 R[%s]eboot  [%s]F6[%s] S[%s]t[%s]op  [%s]F10 Q[%s]uit",
			colors.Current.AccentForeground.Name(), colors.Current.Foreground.Name(),
			colors.Current.AccentForeground.Name(), colors.Current.Foreground.Name(),
			colors.Current.AccentForeground.Name(), colors.Current.Foreground.Name(),
			colors.Current.AccentForeground.Name(), colors.Current.Foreground.Name(),
			colors.Current.AccentForeground.Name(), colors.Current.Foreground.Name(),
			colors.Current.AccentForeground.Name(), colors.Current.Foreground.Name(),
			colors.Current.AccentForeground.Name(), colors.Current.Foreground.Name(),
			colors.Current.AccentForeground.Name(), colors.Current.Foreground.Name(),
			colors.Current.AccentForeground.Name(), colors.Current.Foreground.Name()))
}

// setupKeyHandlers sets up keyboard event handlers for the application
func setupKeyHandlers(app *tview.Application, pages *tview.Pages, ml *mainlist.MainList,
	executor actions.Executor, helpDlg *helpdialog.HelpDialog, cfgPanel *configpanel.ConfigPanel,
	detailsDlg *detailsdialog.DetailsDialog) {

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		return handleKeyEvent(event, app, pages, ml, executor, helpDlg, cfgPanel, detailsDlg)
	})
}

// handleKeyEvent processes individual key events and returns appropriate response
func handleKeyEvent(event *tcell.EventKey, app *tview.Application, pages *tview.Pages,
	ml *mainlist.MainList, executor actions.Executor, helpDlg *helpdialog.HelpDialog,
	cfgPanel *configpanel.ConfigPanel, detailsDlg *detailsdialog.DetailsDialog) *tcell.EventKey {

	// If a modal dialog is open, only allow Escape and Enter keys
	if isModalPageOpen(pages) {
		switch event.Key() {
		case tcell.KeyEscape, tcell.KeyEnter:
			return event // Allow these keys to be processed by the dialog
		default:
			return nil // Block all other keys
		}
	}

	switch event.Key() {
	case tcell.KeyF1:
		helpDlg.Show()
		return nil
	case tcell.KeyF2:
		cfgPanel.Show()
		return nil
	case tcell.KeyF3:
		executeAction(app, pages, ml, executor, "Start")
		return nil
	case tcell.KeyF4:
		executeAction(app, pages, ml, executor, "Shutdown")
		return nil
	case tcell.KeyF5:
		executeAction(app, pages, ml, executor, "Reboot")
		return nil
	case tcell.KeyF6:
		executeAction(app, pages, ml, executor, "Stop")
		return nil
	case tcell.KeyF10:
		app.Stop()
		return nil
	case tcell.KeyEnter:
		return handleEnterKey(pages, ml, detailsDlg, event)
	case tcell.KeyRune:
		return handleRuneKey(event, app, pages, ml, executor, helpDlg, cfgPanel)
	}
	return event
}

// handleEnterKey processes Enter key events with modal detection
func handleEnterKey(pages *tview.Pages, ml *mainlist.MainList, detailsDlg *detailsdialog.DetailsDialog, event *tcell.EventKey) *tcell.EventKey {
	// Only show details dialog if we're on the main page (no modals open)
	frontPage, _ := pages.GetFrontPage()
	if frontPage != "main" {
		// Let the modal dialog handle the Enter key
		return event
	}

	// Double check - also look for any modal pages that might be open
	if isModalPageOpen(pages) {
		return event
	}

	// Show details dialog for selected VM/CT
	if node := ml.GetSelectedNode(); node != nil {
		detailsDlg.Show(node)
	}
	return nil
}

// isModalPageOpen checks if any modal pages are currently open
func isModalPageOpen(pages *tview.Pages) bool {
	modalPages := []string{"progress", "modal", "help", "config", "details"}
	for _, page := range modalPages {
		if pages.HasPage(page) {
			return true
		}
	}
	return false
}

// handleRuneKey handles letter key events
func handleRuneKey(event *tcell.EventKey, app *tview.Application, pages *tview.Pages,
	ml *mainlist.MainList, executor actions.Executor, helpDlg *helpdialog.HelpDialog,
	cfgPanel *configpanel.ConfigPanel) *tcell.EventKey {

	switch event.Rune() {
	case 'h', 'H':
		helpDlg.Show()
		return nil
	case 'c', 'C':
		cfgPanel.Show()
		return nil
	case 's', 'S':
		executeAction(app, pages, ml, executor, "Start")
		return nil
	case 'd', 'D':
		executeAction(app, pages, ml, executor, "Shutdown")
		return nil
	case 'r', 'R':
		executeAction(app, pages, ml, executor, "Reboot")
		return nil
	case 't', 'T':
		executeAction(app, pages, ml, executor, "Stop")
		return nil
	case 'q', 'Q':
		app.Stop()
		return nil
	}
	return event
}

func main() {
	initBorders()
	cfgPath := parseFlags()

	// Load configuration
	loader := config.NewLoader(cfgPath)
	cfg, err := loader.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration from %s: %v\nPlease create a .pvecrc file in your home directory or specify one with -c flag.", cfgPath, err)
	}

	// Create Proxmox client
	client := proxmox.NewClient(cfg.APIUrl, cfg.GetAuthToken(), cfg.SkipTLSVerify)

	// Create action executor
	executor := proxmox.NewActionExecutor(client)

	// Create TUI application
	app := tview.NewApplication()
	pages := tview.NewPages()

	// Create main list with refresh interval from config
	listCfg := mainlist.Config{
		RefreshInterval: cfg.RefreshInterval,
		Provider:        client,
		App:             app,
		OnNodesUpdated: func(nodes []*models.VMStatus) {
			// Update executor cache when nodes are refreshed
			if ae, ok := executor.(*proxmox.ActionExecutor); ok {
				ae.UpdateNodes(nodes)
			}
		},
	}
	ml := mainlist.NewMainList(listCfg)
	defer ml.Stop()

	// Initial data load
	ctx := context.Background()
	if err := ml.Refresh(ctx); err != nil {
		log.Fatalf("Failed to load initial data: %v", err)
	}

	// Create dialogs
	helpDlg := helpdialog.NewHelpDialog(pages)
	cfgPanel := configpanel.NewConfigPanel(pages, cfg, loader)
	detailsDlg := detailsdialog.NewDetailsDialog(pages, app, client)

	// Create status bar and main layout
	statusBar := createStatusBar()
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(ml.GetTable(), 0, 1, true).
		AddItem(statusBar, 1, 0, false)

	// Add main page
	pages.AddPage("main", flex, true, true)

	// Set up keyboard handlers
	setupKeyHandlers(app, pages, ml, executor, helpDlg, cfgPanel, detailsDlg)

	// Start the application
	if err := app.SetRoot(pages, true).EnableMouse(true).Run(); err != nil {
		log.Fatalf("Error running application: %v", err)
	}
}

// executeAction executes an action on the selected node
func executeAction(app *tview.Application, pages *tview.Pages, ml *mainlist.MainList, executor actions.Executor, actionName string) {
	node := ml.GetSelectedNode()
	if node == nil {
		showModal(app, pages, "Error", "No node selected")
		return
	}

	// Create action based on name
	var action actions.Action
	switch actionName {
	case "Start":
		action = actions.NewStartAction(executor, node)
	case "Shutdown":
		action = actions.NewShutdownAction(executor, node)
	case "Reboot":
		action = actions.NewRebootAction(executor, node)
	case "Stop":
		action = actions.NewStopAction(executor, node)
	default:
		showModal(app, pages, "Error", fmt.Sprintf("Unknown action: %s", actionName))
		return
	}

	// Show progress modal
	progressModal := tview.NewModal().
		SetText(fmt.Sprintf("[%s]Executing %s on %s...",
			colors.Current.Foreground.Name(), action.Name(), node.Name)).
		AddButtons([]string{"Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			pages.RemovePage("progress")
		})

	progressModal.SetBackgroundColor(colors.Current.Background)
	progressModal.Box.SetBackgroundColor(colors.Current.Background).
		SetBorderColor(colors.Current.Foreground)

	// Wrap in flex for consistent background
	flex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(progressModal, 0, 1, true).
		AddItem(nil, 0, 1, false)
	flex.SetBackgroundColor(colors.Current.Background)

	pages.AddPage("progress", flex, true, true)

	// Disable refresh during action
	ml.SetRefreshEnabled(false)

	// Execute action in background
	go func() {
		ctx := context.Background()
		err := action.Execute(ctx)
		ml.SetRefreshEnabled(true)

		app.QueueUpdateDraw(func() {
			pages.RemovePage("progress")

			if err != nil {
				showModal(app, pages, "Error", fmt.Sprintf("Failed to %s %s: %v", actionName, node.Name, err))
			} else {
				showModal(app, pages, "Success", fmt.Sprintf("Successfully executed %s on %s", actionName, node.Name))
				// Force refresh after successful action
				_ = ml.Refresh(ctx)
			}
		})
	}()
}

// showModal shows a simple modal dialog
func showModal(_ *tview.Application, pages *tview.Pages, _, message string) {
	modal := tview.NewModal().
		SetText(fmt.Sprintf("[%s]%s", colors.Current.Foreground.Name(), message)).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			pages.RemovePage("modal")
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

	pages.AddPage("modal", flex, true, true)
}
