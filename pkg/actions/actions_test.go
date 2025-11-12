package actions

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tsupplis/pvec/pkg/models"
)

// MockExecutor is a mock implementation of Executor for testing
type MockExecutor struct {
	StartCalled    bool
	ShutdownCalled bool
	RebootCalled   bool
	StopCalled     bool
	LastVMID       string
	ReturnError    error
}

func (m *MockExecutor) Start(ctx context.Context, vmid string) error {
	m.StartCalled = true
	m.LastVMID = vmid
	return m.ReturnError
}

func (m *MockExecutor) Shutdown(ctx context.Context, vmid string) error {
	m.ShutdownCalled = true
	m.LastVMID = vmid
	return m.ReturnError
}

func (m *MockExecutor) Reboot(ctx context.Context, vmid string) error {
	m.RebootCalled = true
	m.LastVMID = vmid
	return m.ReturnError
}

func (m *MockExecutor) Stop(ctx context.Context, vmid string) error {
	m.StopCalled = true
	m.LastVMID = vmid
	return m.ReturnError
}

func TestStartAction(t *testing.T) {
	mock := &MockExecutor{}
	node := &models.VMStatus{VMID: "100", Name: "test-vm"}
	action := NewStartAction(mock, node)

	assert.Equal(t, "Start", action.Name())
	assert.Contains(t, action.Description(), "test-vm")
	assert.Contains(t, action.Description(), "100")

	err := action.Execute(context.Background())
	assert.NoError(t, err)
	assert.True(t, mock.StartCalled)
	assert.Equal(t, "100", mock.LastVMID)
}

func TestStartAction_Error(t *testing.T) {
	mock := &MockExecutor{ReturnError: errors.New("start failed")}
	node := &models.VMStatus{VMID: "100", Name: "test-vm"}
	action := NewStartAction(mock, node)

	err := action.Execute(context.Background())
	assert.Error(t, err)
	assert.Equal(t, "start failed", err.Error())
}

func TestShutdownAction(t *testing.T) {
	mock := &MockExecutor{}
	node := &models.VMStatus{VMID: "101", Name: "test-ct"}
	action := NewShutdownAction(mock, node)

	assert.Equal(t, "Shutdown", action.Name())
	assert.Contains(t, action.Description(), "test-ct")
	assert.Contains(t, action.Description(), "101")

	err := action.Execute(context.Background())
	assert.NoError(t, err)
	assert.True(t, mock.ShutdownCalled)
	assert.Equal(t, "101", mock.LastVMID)
}

func TestShutdownAction_Error(t *testing.T) {
	mock := &MockExecutor{ReturnError: errors.New("shutdown failed")}
	node := &models.VMStatus{VMID: "101", Name: "test-ct"}
	action := NewShutdownAction(mock, node)

	err := action.Execute(context.Background())
	assert.Error(t, err)
	assert.Equal(t, "shutdown failed", err.Error())
}

func TestRebootAction(t *testing.T) {
	mock := &MockExecutor{}
	node := &models.VMStatus{VMID: "102", Name: "test-vm2"}
	action := NewRebootAction(mock, node)

	assert.Equal(t, "Reboot", action.Name())
	assert.Contains(t, action.Description(), "test-vm2")
	assert.Contains(t, action.Description(), "102")

	err := action.Execute(context.Background())
	assert.NoError(t, err)
	assert.True(t, mock.RebootCalled)
	assert.Equal(t, "102", mock.LastVMID)
}

func TestRebootAction_Error(t *testing.T) {
	mock := &MockExecutor{ReturnError: errors.New("reboot failed")}
	node := &models.VMStatus{VMID: "102", Name: "test-vm2"}
	action := NewRebootAction(mock, node)

	err := action.Execute(context.Background())
	assert.Error(t, err)
	assert.Equal(t, "reboot failed", err.Error())
}

func TestStopAction(t *testing.T) {
	mock := &MockExecutor{}
	node := &models.VMStatus{VMID: "103", Name: "test-ct2"}
	action := NewStopAction(mock, node)

	assert.Equal(t, "Stop", action.Name())
	assert.Contains(t, action.Description(), "test-ct2")
	assert.Contains(t, action.Description(), "103")

	err := action.Execute(context.Background())
	assert.NoError(t, err)
	assert.True(t, mock.StopCalled)
	assert.Equal(t, "103", mock.LastVMID)
}

func TestStopAction_Error(t *testing.T) {
	mock := &MockExecutor{ReturnError: errors.New("stop failed")}
	node := &models.VMStatus{VMID: "103", Name: "test-ct2"}
	action := NewStopAction(mock, node)

	err := action.Execute(context.Background())
	assert.Error(t, err)
	assert.Equal(t, "stop failed", err.Error())
}
