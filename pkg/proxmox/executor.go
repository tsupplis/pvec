package proxmox

import (
	"context"

	"github.com/tsupplis/pvec/pkg/actions"
	"github.com/tsupplis/pvec/pkg/models"
)

// ActionExecutor adapts the Proxmox Client to the actions.Executor interface
type ActionExecutor struct {
	client Client
	nodes  map[string]*models.VMStatus // Cache for node lookups
}

// NewActionExecutor creates a new action executor
func NewActionExecutor(client Client) actions.Executor {
	return &ActionExecutor{
		client: client,
		nodes:  make(map[string]*models.VMStatus),
	}
}

// UpdateNodes updates the internal cache of nodes
func (e *ActionExecutor) UpdateNodes(nodes []*models.VMStatus) {
	e.nodes = make(map[string]*models.VMStatus)
	for _, node := range nodes {
		e.nodes[node.VMID] = node
	}
}

// getNodeInfo retrieves node information from cache
func (e *ActionExecutor) getNodeInfo(vmid string) (node, vmType string, found bool) {
	vm, exists := e.nodes[vmid]
	if !exists {
		return "", "", false
	}

	typeStr := "qemu"
	if vm.Type == string(models.TypeContainer) {
		typeStr = "lxc"
	}

	return vm.Node, typeStr, true
}

// Start starts a VM or Container
func (e *ActionExecutor) Start(ctx context.Context, vmid string) error {
	node, vmType, found := e.getNodeInfo(vmid)
	if !found {
		return ErrNodeNotFound
	}
	return e.client.Start(ctx, node, vmType, vmid)
}

// Shutdown gracefully shuts down a VM or Container
func (e *ActionExecutor) Shutdown(ctx context.Context, vmid string) error {
	node, vmType, found := e.getNodeInfo(vmid)
	if !found {
		return ErrNodeNotFound
	}
	return e.client.Shutdown(ctx, node, vmType, vmid)
}

// Reboot reboots a VM or Container
func (e *ActionExecutor) Reboot(ctx context.Context, vmid string) error {
	node, vmType, found := e.getNodeInfo(vmid)
	if !found {
		return ErrNodeNotFound
	}
	return e.client.Reboot(ctx, node, vmType, vmid)
}

// Stop forcefully stops a VM or Container
func (e *ActionExecutor) Stop(ctx context.Context, vmid string) error {
	node, vmType, found := e.getNodeInfo(vmid)
	if !found {
		return ErrNodeNotFound
	}
	return e.client.Stop(ctx, node, vmType, vmid)
}
