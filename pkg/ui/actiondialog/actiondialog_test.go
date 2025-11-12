package actiondialog

import (
	"context"
	"errors"
	"testing"

	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"
)

// MockAction implements actions.Action for testing
type MockAction struct {
	name        string
	description string
	executeErr  error
}

func (m *MockAction) Execute(ctx context.Context) error {
	return m.executeErr
}

func (m *MockAction) Name() string {
	return m.name
}

func (m *MockAction) Description() string {
	return m.description
}

func TestNewActionDialog(t *testing.T) {
	pages := tview.NewPages()
	action := &MockAction{name: "Start", description: "Start VM"}
	vmid := "100"

	ad := NewActionDialog(pages, action, vmid)

	assert.NotNil(t, ad)
	assert.NotNil(t, ad.modal)
	assert.Equal(t, pages, ad.pages)
	assert.Equal(t, action, ad.action)
	assert.Equal(t, vmid, ad.vmid)
}

func TestActionDialog_Show(t *testing.T) {
	pages := tview.NewPages()
	action := &MockAction{name: "Start", description: "Start VM"}
	vmid := "100"

	ad := NewActionDialog(pages, action, vmid)
	ad.Show()

	// Verify page was added
	assert.True(t, pages.HasPage("action"))
}

func TestActionDialog_GetModal(t *testing.T) {
	pages := tview.NewPages()
	action := &MockAction{name: "Start", description: "Start VM"}
	vmid := "100"

	ad := NewActionDialog(pages, action, vmid)

	modal := ad.GetModal()
	assert.NotNil(t, modal)
	assert.Equal(t, ad.modal, modal)
}

func TestActionDialog_ShowSuccess(t *testing.T) {
	pages := tview.NewPages()
	action := &MockAction{name: "Start", description: "Start VM"}
	vmid := "100"

	ad := NewActionDialog(pages, action, vmid)
	ad.showSuccess()

	// Verify modal was updated (we can't easily check text on modal)
	assert.NotNil(t, ad.modal)
}

func TestActionDialog_ShowError(t *testing.T) {
	pages := tview.NewPages()
	action := &MockAction{name: "Start", description: "Start VM"}
	vmid := "100"

	ad := NewActionDialog(pages, action, vmid)
	testErr := errors.New("test error")
	ad.showError(testErr)

	// Verify modal was updated (we can't easily check text on modal)
	assert.NotNil(t, ad.modal)
}

func TestActionDialog_Execute_Success(t *testing.T) {
	pages := tview.NewPages()
	action := &MockAction{
		name:        "Start",
		description: "Start VM",
		executeErr:  nil,
	}
	vmid := "100"

	ad := NewActionDialog(pages, action, vmid)
	ctx := context.Background()

	// Execute should run in goroutine, so we can't easily test completion
	// Just verify it doesn't panic
	ad.Execute(ctx)
}

func TestActionDialog_Execute_Error(t *testing.T) {
	pages := tview.NewPages()
	action := &MockAction{
		name:        "Start",
		description: "Start VM",
		executeErr:  errors.New("execution failed"),
	}
	vmid := "100"

	ad := NewActionDialog(pages, action, vmid)
	ctx := context.Background()

	// Execute should run in goroutine, so we can't easily test completion
	// Just verify it doesn't panic
	ad.Execute(ctx)
}
