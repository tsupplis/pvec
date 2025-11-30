package mainlist

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/tsupplis/pvec/pkg/models"
)

// MockDataProvider implements DataProvider for testing
type MockDataProvider struct {
	Nodes []*models.VMStatus
	Err   error
}

func (m *MockDataProvider) GetNodes(ctx context.Context) ([]*models.VMStatus, error) {
	return m.Nodes, m.Err
}

func TestSortNodes(t *testing.T) {
	nodes := []*models.VMStatus{
		{VMID: "101", Name: "vm-zebra", Type: "qemu"},
		{VMID: "200", Name: "ct-alpha", Type: "lxc"},
		{VMID: "102", Name: "vm-alpha", Type: "qemu"},
		{VMID: "201", Name: "ct-beta", Type: "lxc"},
	}

	sorted := sortNodes(nodes)

	// Containers should come first (lxc), then VMs (qemu)
	// Within each type, sorted alphabetically by name
	expected := []string{
		"ct-alpha", // lxc comes first
		"ct-beta",  // lxc, alphabetically
		"vm-alpha", // qemu after lxc
		"vm-zebra", // qemu, alphabetically
	}

	if len(sorted) != len(expected) {
		t.Fatalf("Expected %d nodes, got %d", len(expected), len(sorted))
	}

	for i, expectedName := range expected {
		if sorted[i].Name != expectedName {
			t.Errorf("Position %d: expected %s, got %s", i, expectedName, sorted[i].Name)
		}
	}
}

func TestShouldSwap(t *testing.T) {
	tests := []struct {
		name     string
		a        *models.VMStatus
		b        *models.VMStatus
		expected bool
	}{
		{
			name:     "VM before CT",
			a:        &models.VMStatus{Type: "qemu", Name: "vm-a"},
			b:        &models.VMStatus{Type: "lxc", Name: "ct-a"},
			expected: true, // Should swap (CT should come first)
		},
		{
			name:     "CT before VM",
			a:        &models.VMStatus{Type: "lxc", Name: "ct-a"},
			b:        &models.VMStatus{Type: "qemu", Name: "vm-a"},
			expected: false, // No swap needed
		},
		{
			name:     "Same type, alphabetical swap needed",
			a:        &models.VMStatus{Type: "qemu", Name: "vm-zebra"},
			b:        &models.VMStatus{Type: "qemu", Name: "vm-alpha"},
			expected: true, // Should swap (alphabetically)
		},
		{
			name:     "Same type, no swap needed",
			a:        &models.VMStatus{Type: "qemu", Name: "vm-alpha"},
			b:        &models.VMStatus{Type: "qemu", Name: "vm-zebra"},
			expected: false, // Already in order
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldSwap(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{"Short string", "hello", 10, "hello"},
		{"Exact length", "hello", 5, "hello"},
		{"Long string", "hello world", 8, "hello..."},
		{"Very short max", "hello", 3, "..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncate(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestFormatUptime(t *testing.T) {
	tests := []struct {
		name     string
		seconds  int64
		expected string
	}{
		{"Zero uptime", 0, "-"},
		{"Less than hour", 45 * 60, "45m"},
		{"Hours and minutes", 3*3600 + 30*60, "3h 30m"},
		{"Days and hours", 2*86400 + 5*3600, "2d 5h"},
		{"Only days", 3 * 86400, "3d 0h"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatUptime(tt.seconds)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestNewMainList(t *testing.T) {
	provider := &MockDataProvider{
		Nodes: []*models.VMStatus{
			{VMID: "100", Name: "test-vm", Type: "qemu", Status: "running"},
		},
	}

	cfg := Config{
		RefreshInterval: 5 * time.Second,
		Provider:        provider,
		Client:          nil,
		OnNodesUpdated:  nil,
	}

	ml := NewMainList(cfg)

	if ml == nil {
		t.Fatal("NewMainList returned nil")
	}
	if ml.provider != provider {
		t.Error("Provider not set correctly")
	}
	if ml.selectedIdx != 0 {
		t.Error("Initial selectedIdx should be 0")
	}
	if !ml.refreshEnabled {
		t.Error("Refresh should be enabled by default")
	}
	if ml.model == nil {
		t.Error("Internal model should be initialized")
	}
	if ml.program == nil {
		t.Error("Program should be initialized")
	}
}

func TestListModel_RenderRow(t *testing.T) {
	provider := &MockDataProvider{}
	ml := NewMainList(Config{Provider: provider})
	ml.model.width = 80

	node := &models.VMStatus{
		VMID:        "100",
		Name:        "test-vm",
		Type:        "qemu",
		Status:      "running",
		Node:        "node1",
		CPUUsage:    25.5,
		MemoryUsage: 50.0,
		Uptime:      3600, // 1 hour
	}

	// Test non-selected row
	row := ml.model.renderRow(node, false)
	if row == "" {
		t.Error("Row should not be empty")
	}
	if !strings.Contains(row, "100") {
		t.Error("Row should contain VMID")
	}
	if !strings.Contains(row, "test-vm") {
		t.Error("Row should contain name")
	}

	// Test selected row
	selectedRow := ml.model.renderRow(node, true)
	if selectedRow == "" {
		t.Error("Selected row should not be empty")
	}
	// Note: Selected row may or may not look different in test environment
	// depending on whether ANSI codes are rendered, so we just check it's not empty
}

func TestListModel_RenderRow_NegativeValues(t *testing.T) {
	provider := &MockDataProvider{}
	ml := NewMainList(Config{Provider: provider})
	ml.model.width = 80

	node := &models.VMStatus{
		VMID:        "100",
		Name:        "test-vm",
		Type:        "qemu",
		Status:      "stopped",
		Node:        "node1",
		CPUUsage:    -1.0, // Negative should be clamped to 0
		MemoryUsage: -1.0, // Negative should be clamped to 0
		Uptime:      0,
	}

	row := ml.model.renderRow(node, false)
	if !strings.Contains(row, "0.0%") {
		t.Error("Negative CPU/Memory should be rendered as 0.0%")
	}
}

func TestGetSelectedNode(t *testing.T) {
	nodes := []*models.VMStatus{
		{VMID: "100", Name: "vm1", Type: "qemu"},
		{VMID: "101", Name: "vm2", Type: "qemu"},
	}
	provider := &MockDataProvider{Nodes: nodes}
	ml := NewMainList(Config{Provider: provider})

	// Simulate having nodes
	ml.nodes = nodes
	ml.sortedNodes = sortNodes(nodes)
	ml.selectedIdx = 1

	selected := ml.GetSelectedNode()
	if selected == nil {
		t.Fatal("Selected node should not be nil")
	}
	if selected.VMID != "101" {
		t.Errorf("Expected VMID 101, got %s", selected.VMID)
	}
}

func TestGetSelectedNode_InvalidIndex(t *testing.T) {
	provider := &MockDataProvider{}
	ml := NewMainList(Config{Provider: provider})
	ml.selectedIdx = 10 // Out of bounds

	selected := ml.GetSelectedNode()
	if selected != nil {
		t.Error("Selected node should be nil for invalid index")
	}
}

func TestGetAllNodes(t *testing.T) {
	nodes := []*models.VMStatus{
		{VMID: "100", Name: "vm1", Type: "qemu"},
		{VMID: "101", Name: "vm2", Type: "qemu"},
	}
	provider := &MockDataProvider{Nodes: nodes}
	ml := NewMainList(Config{Provider: provider})
	ml.nodes = nodes

	allNodes := ml.GetAllNodes()
	if len(allNodes) != len(nodes) {
		t.Errorf("Expected %d nodes, got %d", len(nodes), len(allNodes))
	}

	// Verify it's a copy, not the same slice
	if len(allNodes) > 0 && len(nodes) > 0 && &allNodes[0] == &nodes[0] {
		t.Error("GetAllNodes should return a copy, not the same slice")
	}
}

func TestSetRefreshEnabled(t *testing.T) {
	provider := &MockDataProvider{}
	ml := NewMainList(Config{Provider: provider})

	if !ml.refreshEnabled {
		t.Error("Refresh should be enabled by default")
	}

	ml.SetRefreshEnabled(false)
	if ml.refreshEnabled {
		t.Error("Refresh should be disabled")
	}

	ml.SetRefreshEnabled(true)
	if !ml.refreshEnabled {
		t.Error("Refresh should be enabled")
	}
}
