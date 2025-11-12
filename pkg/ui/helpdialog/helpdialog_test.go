package helpdialog

import (
	"testing"

	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHelpDialog(t *testing.T) {
	pages := tview.NewPages()
	dialog := NewHelpDialog(pages)

	require.NotNil(t, dialog)
	require.NotNil(t, dialog.GetFrame())
	assert.NotNil(t, dialog.pages)
}

func TestHelpDialog_Show(t *testing.T) {
	pages := tview.NewPages()
	dialog := NewHelpDialog(pages)

	dialog.Show()

	// Check that the page was added
	assert.True(t, pages.HasPage("help"))
}

func TestHelpDialog_GetFrame(t *testing.T) {
	pages := tview.NewPages()
	dialog := NewHelpDialog(pages)

	frame := dialog.GetFrame()
	require.NotNil(t, frame)
}
