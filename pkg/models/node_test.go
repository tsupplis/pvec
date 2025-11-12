package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVMStatus_String(t *testing.T) {
	vm := &VMStatus{
		VMID:        "100",
		Name:        "test-vm",
		Type:        string(TypeVM),
		Status:      string(StateRunning),
		CPUUsage:    25.5,
		MemoryUsage: 60.2,
	}

	result := vm.String()
	assert.Contains(t, result, "100")
	assert.Contains(t, result, "test-vm")
	assert.Contains(t, result, "running")
}

func TestVMStatus_IsRunning(t *testing.T) {
	tests := []struct {
		name     string
		status   NodeState
		expected bool
	}{
		{"running node", StateRunning, true},
		{"stopped node", StateStopped, false},
		{"paused node", StatePaused, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := &VMStatus{Status: string(tt.status)}
			assert.Equal(t, tt.expected, vm.IsRunning())
		})
	}
}

func TestVMStatus_CanStart(t *testing.T) {
	tests := []struct {
		name     string
		status   NodeState
		expected bool
	}{
		{"stopped node", StateStopped, true},
		{"running node", StateRunning, false},
		{"paused node", StatePaused, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := &VMStatus{Status: string(tt.status)}
			assert.Equal(t, tt.expected, vm.CanStart())
		})
	}
}

func TestVMStatus_CanStop(t *testing.T) {
	tests := []struct {
		name     string
		status   NodeState
		expected bool
	}{
		{"running node", StateRunning, true},
		{"paused node", StatePaused, true},
		{"stopped node", StateStopped, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := &VMStatus{Status: string(tt.status)}
			assert.Equal(t, tt.expected, vm.CanStop())
		})
	}
}

func TestNodeList_Add(t *testing.T) {
	list := NewNodeList()
	vm := &VMStatus{VMID: "100", Name: "test"}

	list.Add(vm)

	assert.Equal(t, 1, list.Count())
	retrieved, exists := list.Get("100")
	assert.True(t, exists)
	assert.Equal(t, "test", retrieved.Name)
}

func TestNodeList_Remove(t *testing.T) {
	list := NewNodeList()
	vm := &VMStatus{VMID: "100", Name: "test"}
	list.Add(vm)

	removed := list.Remove("100")
	assert.True(t, removed)
	assert.Equal(t, 0, list.Count())

	removed = list.Remove("999")
	assert.False(t, removed)
}

func TestNodeList_Get(t *testing.T) {
	list := NewNodeList()
	vm := &VMStatus{VMID: "100", Name: "test"}
	list.Add(vm)

	retrieved, exists := list.Get("100")
	assert.True(t, exists)
	assert.Equal(t, "test", retrieved.Name)

	_, exists = list.Get("999")
	assert.False(t, exists)
}

func TestNodeList_All(t *testing.T) {
	list := NewNodeList()
	vm1 := &VMStatus{VMID: "100", Name: "test1"}
	vm2 := &VMStatus{VMID: "101", Name: "test2"}

	list.Add(vm1)
	list.Add(vm2)

	all := list.All()
	assert.Equal(t, 2, len(all))
}

func TestNodeList_Clear(t *testing.T) {
	list := NewNodeList()
	vm := &VMStatus{VMID: "100", Name: "test"}
	list.Add(vm)

	list.Clear()
	assert.Equal(t, 0, list.Count())
}

func TestNodeList_AddNil(t *testing.T) {
	list := NewNodeList()
	list.Add(nil)
	assert.Equal(t, 0, list.Count())
}
