package actions

import (
	"context"
	"fmt"

	"github.com/tsupplis/pvec/pkg/models"
)

// Action represents an executable action on a VM or Container
type Action interface {
	// Execute performs the action
	Execute(ctx context.Context) error
	// Name returns the action name
	Name() string
	// Description returns a human-readable description
	Description() string
}

// Executor is the interface for executing actions on Proxmox nodes
type Executor interface {
	Start(ctx context.Context, vmid string) error
	Shutdown(ctx context.Context, vmid string) error
	Reboot(ctx context.Context, vmid string) error
	Stop(ctx context.Context, vmid string) error
}

// BaseAction provides common functionality for all actions
type BaseAction struct {
	VMID     string
	VMName   string
	Executor Executor
}

// StartAction starts a stopped VM or container
type StartAction struct {
	BaseAction
}

func NewStartAction(executor Executor, node *models.VMStatus) *StartAction {
	return &StartAction{
		BaseAction: BaseAction{
			VMID:     node.VMID,
			VMName:   node.Name,
			Executor: executor,
		},
	}
}

func (a *StartAction) Execute(ctx context.Context) error {
	return a.Executor.Start(ctx, a.VMID)
}

func (a *StartAction) Name() string {
	return "Start"
}

func (a *StartAction) Description() string {
	return fmt.Sprintf("Starting %s (%s)", a.VMName, a.VMID)
}

// ShutdownAction gracefully shuts down a running VM or container
type ShutdownAction struct {
	BaseAction
}

func NewShutdownAction(executor Executor, node *models.VMStatus) *ShutdownAction {
	return &ShutdownAction{
		BaseAction: BaseAction{
			VMID:     node.VMID,
			VMName:   node.Name,
			Executor: executor,
		},
	}
}

func (a *ShutdownAction) Execute(ctx context.Context) error {
	return a.Executor.Shutdown(ctx, a.VMID)
}

func (a *ShutdownAction) Name() string {
	return "Shutdown"
}

func (a *ShutdownAction) Description() string {
	return fmt.Sprintf("Shutting down %s (%s)", a.VMName, a.VMID)
}

// RebootAction reboots a running VM or container
type RebootAction struct {
	BaseAction
}

func NewRebootAction(executor Executor, node *models.VMStatus) *RebootAction {
	return &RebootAction{
		BaseAction: BaseAction{
			VMID:     node.VMID,
			VMName:   node.Name,
			Executor: executor,
		},
	}
}

func (a *RebootAction) Execute(ctx context.Context) error {
	return a.Executor.Reboot(ctx, a.VMID)
}

func (a *RebootAction) Name() string {
	return "Reboot"
}

func (a *RebootAction) Description() string {
	return fmt.Sprintf("Rebooting %s (%s)", a.VMName, a.VMID)
}

// StopAction forcefully stops a running VM or container
type StopAction struct {
	BaseAction
}

func NewStopAction(executor Executor, node *models.VMStatus) *StopAction {
	return &StopAction{
		BaseAction: BaseAction{
			VMID:     node.VMID,
			VMName:   node.Name,
			Executor: executor,
		},
	}
}

func (a *StopAction) Execute(ctx context.Context) error {
	return a.Executor.Stop(ctx, a.VMID)
}

func (a *StopAction) Name() string {
	return "Stop"
}

func (a *StopAction) Description() string {
	return fmt.Sprintf("Force stopping %s (%s)", a.VMName, a.VMID)
}
