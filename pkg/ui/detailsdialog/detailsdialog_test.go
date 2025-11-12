package detailsdialog

import (
	"context"
	"fmt"
	"testing"

	"github.com/rivo/tview"
	"github.com/tsupplis/pvec/pkg/models"
)

// MockClient is a simple mock for testing
type MockClient struct {
	vmConfig map[string]interface{}
	err      error
}

func (m *MockClient) GetNodes(ctx context.Context) ([]*models.VMStatus, error) {
	return nil, nil
}

func (m *MockClient) GetVMConfig(ctx context.Context, node, vmType, vmid string) (map[string]interface{}, error) {
	return m.vmConfig, m.err
}

func (m *MockClient) Start(ctx context.Context, node, vmType, vmid string) error {
	return m.err
}

func (m *MockClient) Shutdown(ctx context.Context, node, vmType, vmid string) error {
	return m.err
}

func (m *MockClient) Reboot(ctx context.Context, node, vmType, vmid string) error {
	return m.err
}

func (m *MockClient) Stop(ctx context.Context, node, vmType, vmid string) error {
	return m.err
}

func TestNewDetailsDialog(t *testing.T) {
	pages := tview.NewPages()
	app := tview.NewApplication()
	client := &MockClient{}

	dd := NewDetailsDialog(pages, app, client)

	if dd == nil {
		t.Fatal("NewDetailsDialog returned nil")
	}
	if dd.pages != pages {
		t.Error("pages not set correctly")
	}
	if dd.app != app {
		t.Error("app not set correctly")
	}
	// Note: client interface comparison skipped as it's not directly comparable
	if dd.table == nil {
		t.Error("table not created")
	}
	if dd.flex == nil {
		t.Error("flex not created")
	}
}

func TestDetailsDialog_buildDetails(t *testing.T) {
	pages := tview.NewPages()
	app := tview.NewApplication()
	client := &MockClient{}
	dd := NewDetailsDialog(pages, app, client)

	vm := &models.VMStatus{
		VMID:        "100",
		Name:        "test-vm",
		Type:        "qemu",
		Status:      "running",
		Node:        "node1",
		CPUUsage:    25.5,
		MemoryUsage: 45.2,
		MaxMem:      4294967296, // 4GB
		MaxCPU:      2,
		Uptime:      3600,
	}

	config := map[string]interface{}{
		"agent":  "1",
		"cores":  2.0,
		"memory": 4096.0,
		"net0":   "virtio=12:34:56:78:9a:bc,bridge=vmbr0",
	}

	details := dd.buildDetails(vm, config)

	if len(details) < 10 {
		t.Errorf("Expected at least 10 details, got %d", len(details))
	}

	// Check basic fields
	if details[0].Value != vm.VMID {
		t.Errorf("Expected VMID %s, got %s", vm.VMID, details[0].Value)
	}
	if details[1].Value != vm.Name {
		t.Errorf("Expected Name %s, got %s", vm.Name, details[1].Value)
	}
	if details[2].Value != vm.Type {
		t.Errorf("Expected Type %s, got %s", vm.Type, details[2].Value)
	}
}

func TestDetailsDialog_buildDetails_Container(t *testing.T) {
	pages := tview.NewPages()
	app := tview.NewApplication()
	client := &MockClient{}
	dd := NewDetailsDialog(pages, app, client)

	vm := &models.VMStatus{
		VMID:   "101",
		Name:   "test-ct",
		Type:   "lxc",
		Status: "stopped",
		Node:   "node2",
	}

	config := map[string]interface{}{
		"rootfs": "local:101/vm-101-disk-0.raw,size=8G",
		"swap":   512.0,
	}

	details := dd.buildDetails(vm, config)

	// Should not include guest agent for LXC containers
	guestAgentFound := false
	for _, detail := range details {
		if detail.Key == "Guest Agent" {
			guestAgentFound = true
			break
		}
	}

	if guestAgentFound {
		t.Error("Guest Agent should not be included for LXC containers")
	}
}

func TestDetailsDialog_isResourceField(t *testing.T) {
	pages := tview.NewPages()
	app := tview.NewApplication()
	client := &MockClient{}
	dd := NewDetailsDialog(pages, app, client)

	tests := []struct {
		field    string
		expected bool
	}{
		{"cores", true},
		{"memory", true},
		{"net0", true},
		{"scsi0", true},
		{"virtio0", true},
		{"mp0", true},
		{"rootfs", true},
		{"description", false},
		{"onboot", false},
		{"protection", false},
		{"unknown_field", false},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			result := dd.isResourceField(tt.field)
			if result != tt.expected {
				t.Errorf("isResourceField(%s) = %v, expected %v", tt.field, result, tt.expected)
			}
		})
	}
}

func TestDetailsDialog_isDisplayedField(t *testing.T) {
	pages := tview.NewPages()
	app := tview.NewApplication()
	client := &MockClient{}
	dd := NewDetailsDialog(pages, app, client)

	tests := []struct {
		field    string
		expected bool
	}{
		{"vmid", true},
		{"name", true},
		{"type", true},
		{"VMID", true}, // case insensitive
		{"cores", false},
		{"memory", false},
		{"description", false},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			result := dd.isDisplayedField(tt.field)
			if result != tt.expected {
				t.Errorf("isDisplayedField(%s) = %v, expected %v", tt.field, result, tt.expected)
			}
		})
	}
}

func TestFormatValue(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{"nil value", nil, "null"},
		{"string value", "test", "test"},
		{"integer float", 42.0, "42"},
		{"decimal float", 3.14, "3.14"},
		{"boolean true", true, "true"},
		{"boolean false", false, "false"},
		{"array", []interface{}{"a", "b", "c"}, "[a, b, c]"},
		{"nested object", map[string]interface{}{"key": "value"}, "{...}"},
		{"other type", 123, "123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatValue(tt.input)
			if result != tt.expected {
				t.Errorf("formatValue(%v) = %s, expected %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1048576, "1.0 MB"},
		{4294967296, "4.0 GB"},
	}

	for _, tt := range tests {
		result := formatBytes(tt.input)
		if result != tt.expected {
			t.Errorf("formatBytes(%d) = %s, expected %s", tt.input, result, tt.expected)
		}
	}
}

func TestFormatUptime(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{0, "N/A"},
		{-100, "N/A"},
		{45, "0m"},          // 45 seconds becomes 0 minutes
		{150, "2m"},         // 2 minutes 30 seconds becomes 2m
		{7200, "2h 0m"},     // exactly 2 hours
		{90000, "1d 1h 0m"}, // 1 day, 1 hour
	}

	for _, tt := range tests {
		result := formatUptime(tt.input)
		if result != tt.expected {
			t.Errorf("formatUptime(%d) = %s, expected %s", tt.input, result, tt.expected)
		}
	}
}

func TestDetailsDialog_Show(t *testing.T) {
	pages := tview.NewPages()
	app := tview.NewApplication()
	client := &MockClient{
		vmConfig: map[string]interface{}{
			"cores":  2.0,
			"memory": 4096.0,
		},
		err: nil,
	}
	dd := NewDetailsDialog(pages, app, client)

	vm := &models.VMStatus{
		VMID: "100",
		Name: "test-vm",
		Type: "qemu",
		Node: "node1",
	}

	// This should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Show() panicked: %v", r)
		}
	}()

	dd.Show(vm)
}

func TestDetailsDialog_Show_Error(t *testing.T) {
	pages := tview.NewPages()
	app := tview.NewApplication()
	client := &MockClient{
		vmConfig: nil,
		err:      fmt.Errorf("connection error"),
	}
	dd := NewDetailsDialog(pages, app, client)

	vm := &models.VMStatus{
		VMID: "100",
		Name: "test-vm",
		Type: "qemu",
		Node: "node1",
	}

	// This should not panic even with error
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Show() panicked on error: %v", r)
		}
	}()

	dd.Show(vm)
}

// Benchmark test for performance
func BenchmarkDetailsDialog_buildDetails(b *testing.B) {
	pages := tview.NewPages()
	app := tview.NewApplication()
	client := &MockClient{}
	dd := NewDetailsDialog(pages, app, client)

	vm := &models.VMStatus{
		VMID:        "100",
		Name:        "test-vm",
		Type:        "qemu",
		Status:      "running",
		Node:        "node1",
		CPUUsage:    25.5,
		MemoryUsage: 45.2,
		MaxMem:      4294967296,
		MaxCPU:      2,
		Uptime:      3600,
	}

	config := make(map[string]interface{})
	for i := 0; i < 50; i++ {
		config[fmt.Sprintf("key%d", i)] = fmt.Sprintf("value%d", i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = dd.buildDetails(vm, config)
	}
}
