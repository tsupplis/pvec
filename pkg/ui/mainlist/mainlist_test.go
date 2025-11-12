package mainlist

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tsupplis/pvec/pkg/models"
)

// MockDataProvider for testing
type MockDataProvider struct {
	nodes []*models.VMStatus
	err   error
}

func (m *MockDataProvider) GetNodes(ctx context.Context) ([]*models.VMStatus, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.nodes, nil
}

func TestNewMainList(t *testing.T) {
	provider := &MockDataProvider{}
	app := tview.NewApplication()

	cfg := Config{
		RefreshInterval: 0, // No auto-refresh for test
		Provider:        provider,
		App:             app,
	}

	ml := NewMainList(cfg)
	require.NotNil(t, ml)
	require.NotNil(t, ml.GetTable())

	ml.Stop()
}

func TestMainList_Refresh(t *testing.T) {
	nodes := []*models.VMStatus{
		{VMID: "100", Name: "test-vm", Status: string(models.StateRunning), Type: string(models.TypeVM)},
		{VMID: "200", Name: "test-ct", Status: string(models.StateStopped), Type: string(models.TypeContainer)},
	}

	provider := &MockDataProvider{nodes: nodes}
	app := tview.NewApplication()

	cfg := Config{
		RefreshInterval: 0,
		Provider:        provider,
		App:             app,
	}

	ml := NewMainList(cfg)
	defer ml.Stop()

	err := ml.Refresh(context.Background())
	require.NoError(t, err)

	allNodes := ml.GetAllNodes()
	assert.Len(t, allNodes, 2)
	assert.Equal(t, "100", allNodes[0].VMID)
	assert.Equal(t, "200", allNodes[1].VMID)
}

func TestMainList_Refresh_Error(t *testing.T) {
	expectedErr := errors.New("fetch failed")
	provider := &MockDataProvider{err: expectedErr}
	app := tview.NewApplication()

	cfg := Config{
		RefreshInterval: 0,
		Provider:        provider,
		App:             app,
	}

	ml := NewMainList(cfg)
	defer ml.Stop()

	err := ml.Refresh(context.Background())
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestMainList_GetSelectedNode(t *testing.T) {
	nodes := []*models.VMStatus{
		{VMID: "100", Name: "test-ct", Type: string(models.TypeContainer)},
		{VMID: "200", Name: "test-vm", Type: string(models.TypeVM)},
	}

	provider := &MockDataProvider{nodes: nodes}
	app := tview.NewApplication()

	cfg := Config{
		RefreshInterval: 0,
		Provider:        provider,
		App:             app,
	}

	ml := NewMainList(cfg)
	defer ml.Stop()

	err := ml.Refresh(context.Background())
	require.NoError(t, err)

	// Initially should select first row
	ml.table.Select(1, 0)
	selected := ml.GetSelectedNode()
	require.NotNil(t, selected)
	assert.Equal(t, "100", selected.VMID)

	// Select second row
	ml.table.Select(2, 0)
	selected = ml.GetSelectedNode()
	require.NotNil(t, selected)
	assert.Equal(t, "200", selected.VMID)

	// Invalid row (beyond table range)
	ml.table.Select(99, 0)
	selected = ml.GetSelectedNode()
	assert.Nil(t, selected)
}

func TestMainList_SetRefreshEnabled(t *testing.T) {
	provider := &MockDataProvider{}
	app := tview.NewApplication()

	cfg := Config{
		RefreshInterval: 0,
		Provider:        provider,
		App:             app,
	}

	ml := NewMainList(cfg)
	defer ml.Stop()

	assert.True(t, ml.refreshEnabled)

	ml.SetRefreshEnabled(false)
	assert.False(t, ml.refreshEnabled)

	ml.SetRefreshEnabled(true)
	assert.True(t, ml.refreshEnabled)
}

func TestFormatUptime(t *testing.T) {
	tests := []struct {
		name     string
		seconds  int64
		expected string
	}{
		{"zero", 0, "-"},
		{"minutes", 300, "5m"},
		{"hours", 3600, "1h 0m"},
		{"hours and minutes", 7230, "2h 0m"},
		{"days", 86400, "1d 0h"},
		{"days and hours", 90000, "1d 1h"},
		{"multiple days", 518400, "6d 0h"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatUptime(tt.seconds)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMainList_UpdateTable(t *testing.T) {
	nodes := []*models.VMStatus{
		{
			VMID:        "100",
			Name:        "test-vm",
			Status:      string(models.StateRunning),
			Type:        string(models.TypeVM),
			Node:        "pve1",
			CPUUsage:    25.5,
			MemoryUsage: 60.2,
			Uptime:      3600,
		},
	}

	provider := &MockDataProvider{nodes: nodes}
	app := tview.NewApplication()

	cfg := Config{
		RefreshInterval: 0,
		Provider:        provider,
		App:             app,
	}

	ml := NewMainList(cfg)
	defer ml.Stop()

	err := ml.Refresh(context.Background())
	require.NoError(t, err)

	// Check table has header + 1 data row
	assert.Equal(t, 2, ml.table.GetRowCount())
}

func TestMainList_AutoRefresh(t *testing.T) {
	// This test verifies that auto-refresh can be started and stopped
	nodes := []*models.VMStatus{
		{VMID: "100", Name: "test-vm"},
	}

	provider := &MockDataProvider{nodes: nodes}
	app := tview.NewApplication()

	cfg := Config{
		RefreshInterval: 100 * time.Millisecond,
		Provider:        provider,
		App:             app,
	}

	ml := NewMainList(cfg)

	// Let it run for a bit
	time.Sleep(150 * time.Millisecond)

	// Stop should not panic
	ml.Stop()
}

func TestMainList_GetAllNodes(t *testing.T) {
	nodes := []*models.VMStatus{
		{VMID: "100", Name: "vm1"},
		{VMID: "200", Name: "vm2"},
		{VMID: "300", Name: "vm3"},
	}

	provider := &MockDataProvider{nodes: nodes}
	app := tview.NewApplication()

	cfg := Config{
		RefreshInterval: 0,
		Provider:        provider,
		App:             app,
	}

	ml := NewMainList(cfg)
	defer ml.Stop()

	err := ml.Refresh(context.Background())
	require.NoError(t, err)

	allNodes := ml.GetAllNodes()
	assert.Len(t, allNodes, 3)

	// Verify the data is correct
	assert.Equal(t, "100", allNodes[0].VMID)
	assert.Equal(t, "200", allNodes[1].VMID)
	assert.Equal(t, "300", allNodes[2].VMID)
}
