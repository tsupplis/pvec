package proxmox

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tsupplis/pvec/pkg/models"
)

// MockClient for testing executor
type MockClient struct {
	StartFunc    func(ctx context.Context, node, vmType, vmid string) error
	ShutdownFunc func(ctx context.Context, node, vmType, vmid string) error
	RebootFunc   func(ctx context.Context, node, vmType, vmid string) error
	StopFunc     func(ctx context.Context, node, vmType, vmid string) error
}

func (m *MockClient) GetNodes(ctx context.Context) ([]*models.VMStatus, error) {
	return nil, nil
}

func (m *MockClient) GetVMConfig(ctx context.Context, node, vmType, vmid string) (map[string]interface{}, error) {
	return nil, nil
}

func (m *MockClient) Start(ctx context.Context, node, vmType, vmid string) error {
	if m.StartFunc != nil {
		return m.StartFunc(ctx, node, vmType, vmid)
	}
	return nil
}

func (m *MockClient) Shutdown(ctx context.Context, node, vmType, vmid string) error {
	if m.ShutdownFunc != nil {
		return m.ShutdownFunc(ctx, node, vmType, vmid)
	}
	return nil
}

func (m *MockClient) Reboot(ctx context.Context, node, vmType, vmid string) error {
	if m.RebootFunc != nil {
		return m.RebootFunc(ctx, node, vmType, vmid)
	}
	return nil
}

func (m *MockClient) Stop(ctx context.Context, node, vmType, vmid string) error {
	if m.StopFunc != nil {
		return m.StopFunc(ctx, node, vmType, vmid)
	}
	return nil
}

func TestActionExecutor_Start(t *testing.T) {
	called := false
	mock := &MockClient{
		StartFunc: func(ctx context.Context, node, vmType, vmid string) error {
			called = true
			assert.Equal(t, "pve1", node)
			assert.Equal(t, "qemu", vmType)
			assert.Equal(t, "100", vmid)
			return nil
		},
	}

	executor := NewActionExecutor(mock).(*ActionExecutor)
	executor.UpdateNodes([]*models.VMStatus{
		{VMID: "100", Node: "pve1", Type: string(models.TypeVM)},
	})

	err := executor.Start(context.Background(), "100")
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestActionExecutor_Start_NotFound(t *testing.T) {
	mock := &MockClient{}
	executor := NewActionExecutor(mock).(*ActionExecutor)

	err := executor.Start(context.Background(), "999")
	assert.Error(t, err)
	assert.Equal(t, ErrNodeNotFound, err)
}

func TestActionExecutor_Shutdown(t *testing.T) {
	called := false
	mock := &MockClient{
		ShutdownFunc: func(ctx context.Context, node, vmType, vmid string) error {
			called = true
			assert.Equal(t, "pve1", node)
			assert.Equal(t, "lxc", vmType)
			assert.Equal(t, "200", vmid)
			return nil
		},
	}

	executor := NewActionExecutor(mock).(*ActionExecutor)
	executor.UpdateNodes([]*models.VMStatus{
		{VMID: "200", Node: "pve1", Type: string(models.TypeContainer)},
	})

	err := executor.Shutdown(context.Background(), "200")
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestActionExecutor_Reboot(t *testing.T) {
	called := false
	mock := &MockClient{
		RebootFunc: func(ctx context.Context, node, vmType, vmid string) error {
			called = true
			return nil
		},
	}

	executor := NewActionExecutor(mock).(*ActionExecutor)
	executor.UpdateNodes([]*models.VMStatus{
		{VMID: "100", Node: "pve1", Type: string(models.TypeVM)},
	})

	err := executor.Reboot(context.Background(), "100")
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestActionExecutor_Stop(t *testing.T) {
	called := false
	mock := &MockClient{
		StopFunc: func(ctx context.Context, node, vmType, vmid string) error {
			called = true
			return nil
		},
	}

	executor := NewActionExecutor(mock).(*ActionExecutor)
	executor.UpdateNodes([]*models.VMStatus{
		{VMID: "100", Node: "pve1", Type: string(models.TypeVM)},
	})

	err := executor.Stop(context.Background(), "100")
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestActionExecutor_ClientError(t *testing.T) {
	expectedErr := errors.New("client error")
	mock := &MockClient{
		StartFunc: func(ctx context.Context, node, vmType, vmid string) error {
			return expectedErr
		},
	}

	executor := NewActionExecutor(mock).(*ActionExecutor)
	executor.UpdateNodes([]*models.VMStatus{
		{VMID: "100", Node: "pve1", Type: string(models.TypeVM)},
	})

	err := executor.Start(context.Background(), "100")
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestActionExecutor_UpdateNodes(t *testing.T) {
	mock := &MockClient{}
	executor := NewActionExecutor(mock).(*ActionExecutor)

	nodes := []*models.VMStatus{
		{VMID: "100", Node: "pve1", Type: string(models.TypeVM)},
		{VMID: "200", Node: "pve1", Type: string(models.TypeContainer)},
		{VMID: "300", Node: "pve2", Type: string(models.TypeVM)},
	}

	executor.UpdateNodes(nodes)

	assert.Equal(t, 3, len(executor.nodes))

	node, vmType, found := executor.getNodeInfo("100")
	assert.True(t, found)
	assert.Equal(t, "pve1", node)
	assert.Equal(t, "qemu", vmType)

	node, vmType, found = executor.getNodeInfo("200")
	assert.True(t, found)
	assert.Equal(t, "pve1", node)
	assert.Equal(t, "lxc", vmType)

	_, _, found = executor.getNodeInfo("999")
	assert.False(t, found)
}
