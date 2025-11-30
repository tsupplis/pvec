package detailsdialog

import (
	"errors"
	"strings"
	"testing"

	"github.com/tsupplis/pvec/pkg/models"
)

func TestGetDetailsText(t *testing.T) {
	vm := &models.VMStatus{
		VMID:   "100",
		Name:   "test-vm",
		Type:   "qemu",
		Status: "running",
		Node:   "pve1",
	}

	config := map[string]interface{}{
		"cores":  2,
		"memory": 4096,
	}

	result := GetDetailsText(vm, config, 80, 24, 0)

	if result == "" {
		t.Error("GetDetailsText returned empty string")
	}

	if !strings.Contains(result, "test-vm") || !strings.Contains(result, "100") {
		t.Error("Missing VM info")
	}
}

func TestGetLoadingText(t *testing.T) {
	vm := &models.VMStatus{
		VMID: "100",
		Name: "test-vm",
	}

	result := GetLoadingText(vm, 80, 24)

	if result == "" {
		t.Error("GetLoadingText returned empty string")
	}

	if !strings.Contains(result, "Loading details...") {
		t.Error("Missing loading message")
	}
}

func TestGetErrorText(t *testing.T) {
	vm := &models.VMStatus{
		VMID:   "100",
		Name:   "test-vm",
		Type:   "qemu",
		Status: "running",
	}

	err := errors.New("connection timeout")
	result := GetErrorText(vm, err, 80, 24)

	if result == "" {
		t.Error("GetErrorText returned empty string")
	}

	if !strings.Contains(result, "connection timeout") {
		t.Error("Missing error message")
	}
}
