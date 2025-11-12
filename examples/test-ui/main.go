package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/tsupplis/pvec/pkg/config"
	"github.com/tsupplis/pvec/pkg/proxmox"
	"github.com/tsupplis/pvec/pkg/ui/mainlist"
)

func main() {
	// Parse command line flags
	configPath := flag.String("c", "", "Path to config file (default: ~/.pvecrc)")
	flag.StringVar(configPath, "config", "", "Path to config file (default: ~/.pvecrc)")
	flag.Parse()

	// Load configuration
	loader := config.NewLoader(*configPath)
	cfg, err := loader.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		fmt.Fprintf(os.Stderr, "Create ~/.pvecrc or use -c flag to specify config file\n")
		os.Exit(1)
	}

	// Create Proxmox client
	client := proxmox.NewClient(cfg.APIUrl, cfg.GetAuthToken(), cfg.SkipTLSVerify)

	// Create and setup UI
	app, pages, ml := setupUI(client, cfg)
	defer ml.Stop()

	// Initial data load
	ctx := context.Background()
	if err := ml.Refresh(ctx); err != nil {
		log.Fatalf("Failed to load initial data: %v", err)
	}

	// Setup key handlers
	setupKeyHandlers(app, pages, ml)

	// Run the application
	if err := app.SetRoot(pages, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}

// setupUI creates and configures the UI components
func setupUI(client proxmox.Client, cfg *config.Config) (*tview.Application, *tview.Pages, *mainlist.MainList) {
	// Create tview application
	app := tview.NewApplication()

	// Create main list
	listCfg := mainlist.Config{
		RefreshInterval: cfg.RefreshInterval,
		Provider:        client,
		App:             app,
	}
	ml := mainlist.NewMainList(listCfg)

	// Create status bar
	statusBar := tview.NewTextView().
		SetDynamicColors(true).
		SetText(" [yellow]F1/h[white]=Help  [yellow]F3/s[white]=Start/Stop  [yellow]F4/r[white]=Reboot  [yellow]F5/k[white]=Kill  [yellow]F10/q[white]=Quit  [yellow]↑↓[white]=Navigate")

	// Create layout
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(ml.GetTable(), 0, 1, true).
		AddItem(statusBar, 1, 0, false)

	// Create pages for modal support
	pages := tview.NewPages().
		AddPage("main", flex, true, true)

	return app, pages, ml
}

// setupKeyHandlers configures keyboard event handling
func setupKeyHandlers(app *tview.Application, pages *tview.Pages, ml *mainlist.MainList) {
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyF1:
			showHelpDialog(pages)
			return nil
		case tcell.KeyF3:
			showActionDialog(pages, ml, "toggle")
			return nil
		case tcell.KeyF4:
			showActionDialog(pages, ml, "reboot")
			return nil
		case tcell.KeyF5:
			showActionDialog(pages, ml, "force kill")
			return nil
		case tcell.KeyF10:
			app.Stop()
			return nil
		case tcell.KeyRune:
			return handleRuneKeys(event, app, pages, ml)
		}
		return event
	})
}

// handleRuneKeys handles letter key events
func handleRuneKeys(event *tcell.EventKey, app *tview.Application, pages *tview.Pages, ml *mainlist.MainList) *tcell.EventKey {
	switch event.Rune() {
	case 'h', 'H':
		showHelpDialog(pages)
		return nil
	case 'q', 'Q':
		app.Stop()
		return nil
	case 's', 'S':
		showActionDialog(pages, ml, "toggle")
		return nil
	case 'r', 'R':
		showActionDialog(pages, ml, "reboot")
		return nil
	case 'k', 'K':
		showActionDialog(pages, ml, "force kill")
		return nil
	}
	return event
}

// showHelpDialog displays the help information
func showHelpDialog(pages *tview.Pages) {
	showInfo(nil, pages, "Help", "F1/h: Help\nF3/s: Start/Stop\nF4/r: Reboot\nF5/k: Force Kill\nF10/q: Quit\n↑↓: Navigate")
}

// showActionDialog displays an action confirmation dialog
func showActionDialog(pages *tview.Pages, ml *mainlist.MainList, actionName string) {
	node := ml.GetSelectedNode()
	if node != nil {
		message := fmt.Sprintf("Would %s %s (%s)", actionName, node.Name, node.VMID)
		showInfo(nil, pages, "Action", message)
	}
}

func showInfo(_ *tview.Application, pages *tview.Pages, _, message string) {
	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			pages.RemovePage("modal")
		})

	pages.AddPage("modal", modal, true, true)
}
