package actiondialog

import (
	"context"
	"fmt"

	"github.com/rivo/tview"
	"github.com/tsupplis/pvec/pkg/actions"
	"github.com/tsupplis/pvec/pkg/ui/colors"
)

// ActionDialog displays action execution progress
type ActionDialog struct {
	modal  *tview.Modal
	pages  *tview.Pages
	action actions.Action
	vmid   string
}

// NewActionDialog creates a new action dialog
func NewActionDialog(pages *tview.Pages, action actions.Action, vmid string) *ActionDialog {
	ad := &ActionDialog{
		pages:  pages,
		action: action,
		vmid:   vmid,
	}

	ad.modal = tview.NewModal().
		SetText(fmt.Sprintf("[%s]Executing %s on VM %s...",
			colors.Current.Foreground.Name(), action.Name(), vmid)).
		AddButtons([]string{"Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Cancel" {
				pages.RemovePage("action")
			}
		})

	ad.modal.SetBackgroundColor(colors.Current.Background)
	ad.modal.Box.SetBackgroundColor(colors.Current.Background).
		SetBorderColor(colors.Current.Foreground)

	return ad
}

// Execute runs the action and shows result
func (ad *ActionDialog) Execute(ctx context.Context) {
	// Update modal text
	ad.modal.SetText(fmt.Sprintf("[%s]Executing[%s] %s on VM %s...",
		colors.Current.WarningColor.Name(),
		colors.Current.Foreground.Name(),
		ad.action.Name(), ad.vmid))

	// Execute action in goroutine
	go func() {
		err := ad.action.Execute(ctx)

		// Update UI on main thread
		if err != nil {
			ad.showError(err)
		} else {
			ad.showSuccess()
		}
	}()
}

// showSuccess displays success message
func (ad *ActionDialog) showSuccess() {
	ad.modal.SetText(fmt.Sprintf("[%s]Success:[%s] %s completed on VM %s",
		colors.Current.OkColor.Name(),
		colors.Current.Foreground.Name(),
		ad.action.Name(), ad.vmid)).
		ClearButtons().
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			ad.pages.RemovePage("action")
		})
}

// showError displays error message
func (ad *ActionDialog) showError(err error) {
	ad.modal.SetText(fmt.Sprintf("[%s]Error:[%s] %s failed on VM %s\n\n%v",
		colors.Current.AlertColor.Name(),
		colors.Current.Foreground.Name(),
		ad.action.Name(), ad.vmid, err)).
		ClearButtons().
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			ad.pages.RemovePage("action")
		})
}

// Show displays the action dialog
func (ad *ActionDialog) Show() {
	// Wrap in flex for consistent background
	flex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(ad.modal, 0, 1, true).
		AddItem(nil, 0, 1, false)
	flex.SetBackgroundColor(colors.Current.Background)

	ad.pages.AddPage("action", flex, true, true)
}

// GetModal returns the underlying modal
func (ad *ActionDialog) GetModal() *tview.Modal {
	return ad.modal
}
