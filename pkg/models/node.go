package models

import "fmt"

// NodeType represents the type of virtual resource
type NodeType string

const (
	TypeVM        NodeType = "qemu"
	TypeContainer NodeType = "lxc"
)

// NodeState represents the current state of the node
type NodeState string

const (
	StateRunning NodeState = "running"
	StateStopped NodeState = "stopped"
	StatePaused  NodeState = "paused"
	StateUnknown NodeState = "unknown"
)

// VMStatus represents the status of a VM or Container
type VMStatus struct {
	VMID        string  // Virtual Machine/Container ID
	Name        string  // Name of the VM/Container
	Type        string  // Type: qemu (VM) or lxc (Container)
	Status      string  // Status: running, stopped, etc.
	Node        string  // Proxmox node name
	CPUUsage    float64 // CPU usage percentage
	MemoryUsage float64 // Memory usage percentage
	MaxMem      int64   // Maximum memory in bytes
	MaxCPU      int     // Maximum CPU count
	Uptime      int64   // Uptime in seconds
}

// String returns a human-readable representation
func (v *VMStatus) String() string {
	return fmt.Sprintf("[%s] %s (%s) - %s - CPU: %.1f%% MEM: %.1f%%",
		v.VMID, v.Name, v.Type, v.Status, v.CPUUsage, v.MemoryUsage)
}

// IsRunning returns true if the node is currently running
func (v *VMStatus) IsRunning() bool {
	return v.Status == string(StateRunning)
}

// CanStart returns true if the node can be started
func (v *VMStatus) CanStart() bool {
	return v.Status == string(StateStopped)
}

// CanStop returns true if the node can be stopped
func (v *VMStatus) CanStop() bool {
	return v.Status == string(StateRunning) || v.Status == string(StatePaused)
}

// NodeList is an interface for managing a collection of nodes
type NodeList interface {
	// Add adds a node to the list
	Add(node *VMStatus)
	// Remove removes a node by VMID
	Remove(vmid string) bool
	// Get retrieves a node by VMID
	Get(vmid string) (*VMStatus, bool)
	// All returns all nodes
	All() []*VMStatus
	// Clear removes all nodes
	Clear()
	// Count returns the number of nodes
	Count() int
}

// InMemoryNodeList is an in-memory implementation of NodeList
type InMemoryNodeList struct {
	nodes map[string]*VMStatus
}

// NewNodeList creates a new in-memory node list
func NewNodeList() NodeList {
	return &InMemoryNodeList{
		nodes: make(map[string]*VMStatus),
	}
}

func (n *InMemoryNodeList) Add(node *VMStatus) {
	if node != nil {
		n.nodes[node.VMID] = node
	}
}

func (n *InMemoryNodeList) Remove(vmid string) bool {
	if _, exists := n.nodes[vmid]; exists {
		delete(n.nodes, vmid)
		return true
	}
	return false
}

func (n *InMemoryNodeList) Get(vmid string) (*VMStatus, bool) {
	node, exists := n.nodes[vmid]
	return node, exists
}

func (n *InMemoryNodeList) All() []*VMStatus {
	result := make([]*VMStatus, 0, len(n.nodes))
	for _, node := range n.nodes {
		result = append(result, node)
	}
	return result
}

func (n *InMemoryNodeList) Clear() {
	n.nodes = make(map[string]*VMStatus)
}

func (n *InMemoryNodeList) Count() int {
	return len(n.nodes)
}
