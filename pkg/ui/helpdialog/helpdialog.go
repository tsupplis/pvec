package helpdialog

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/tsupplis/pvec/pkg/ui/colors"
)

// HelpDialog displays help information about key bindings
type HelpDialog struct {
	frame *tview.Frame
	pages *tview.Pages
}

// NewHelpDialog creates a new help dialog
func NewHelpDialog(pages *tview.Pages) *HelpDialog {
	helpText := fmt.Sprintf(`[%s]Keyboard Shortcuts:[%s]

[%s]F1 / h[%s]  - Show this help
[%s]F2 / c[%s]  - Open configuration editor
[%s]F3 / s[%s]  - Start selected VM/CT
[%s]F4 / d[%s]  - Shutdown selected VM/CT
[%s]F5 / r[%s]  - Reboot selected VM/CT
[%s]F6 / t[%s]  - Stop (force) selected VM/CT
[%s]F10 / q[%s] - Quit application

[%s]↑ / ↓[%s]   - Navigate up/down
[%s]PgUp/PgDn[%s] - Scroll page up/down
[%s]Home/End[%s] - Jump to first/last item
`,
		colors.Current.AccentForeground.Name(), colors.Current.Foreground.Name(),
		colors.Current.AccentForeground.Name(), colors.Current.Foreground.Name(),
		colors.Current.AccentForeground.Name(), colors.Current.Foreground.Name(),
		colors.Current.AccentForeground.Name(), colors.Current.Foreground.Name(),
		colors.Current.AccentForeground.Name(), colors.Current.Foreground.Name(),
		colors.Current.AccentForeground.Name(), colors.Current.Foreground.Name(),
		colors.Current.AccentForeground.Name(), colors.Current.Foreground.Name(),
		colors.Current.AccentForeground.Name(), colors.Current.Foreground.Name(),
		colors.Current.AccentForeground.Name(), colors.Current.Foreground.Name(),
		colors.Current.AccentForeground.Name(), colors.Current.Foreground.Name(),
		colors.Current.AccentForeground.Name(), colors.Current.Foreground.Name())

	textView := tview.NewTextView().
		SetText(helpText).
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft).
		SetWordWrap(true)
	textView.SetBackgroundColor(colors.Current.Background)

	frame := tview.NewFrame(textView).
		SetBorders(1, 1, 1, 1, 2, 2).
		AddText(" Help ", true, tview.AlignCenter, colors.Current.AccentForeground).
		AddText("Press ESC to close", false, tview.AlignCenter, colors.Current.Foreground)
	frame.SetBorder(true).
		SetBackgroundColor(colors.Current.Background)

	return &HelpDialog{
		frame: frame,
		pages: pages,
	}
}

// Show displays the help dialog
func (h *HelpDialog) Show() {
	// Create a responsive container with 2-char padding on all sides
	innerFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(nil, 2, 1, false).
		AddItem(h.frame, 0, 1, true).
		AddItem(nil, 2, 1, false)
	innerFlex.SetBackgroundColor(colors.Current.Background)

	flex := tview.NewFlex().
		AddItem(nil, 2, 1, false).
		AddItem(innerFlex, 0, 1, true).
		AddItem(nil, 2, 1, false)
	flex.SetBackgroundColor(colors.Current.Background)

	// Set up input handler to close on ESC
	flex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape || event.Key() == tcell.KeyEnter {
			h.pages.RemovePage("help")
			return nil
		}
		return event
	})

	h.pages.AddPage("help", flex, true, true)
}

// GetFrame returns the underlying frame
func (h *HelpDialog) GetFrame() *tview.Frame {
	return h.frame
}
