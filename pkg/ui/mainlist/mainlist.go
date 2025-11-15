package mainlist

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rivo/tview"
	"github.com/tsupplis/pvec/pkg/models"
	"github.com/tsupplis/pvec/pkg/ui/colors"
)

// DataProvider is the interface for fetching node data
type DataProvider interface {
	GetNodes(ctx context.Context) ([]*models.VMStatus, error)
}

// MainList is the main scrolling list component
type MainList struct {
	table          *tview.Table
	nodes          []*models.VMStatus
	sortedNodes    []*models.VMStatus // Nodes in display order
	selectedRow    int
	provider       DataProvider
	refreshTicker  *time.Ticker
	stopRefresh    chan bool
	refreshMutex   sync.Mutex
	refreshEnabled bool
	app            *tview.Application
	onNodesUpdated func([]*models.VMStatus) // Callback when nodes are refreshed
}

// Config holds the configuration for the main list
type Config struct {
	RefreshInterval time.Duration
	Provider        DataProvider
	App             *tview.Application
	OnNodesUpdated  func([]*models.VMStatus) // Callback when nodes are refreshed
}

// NewMainList creates a new main list component
func NewMainList(cfg Config) *MainList {
	table := tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false).
		SetFixed(1, 0)

	ml := &MainList{
		table:          table,
		nodes:          make([]*models.VMStatus, 0),
		selectedRow:    1,
		provider:       cfg.Provider,
		stopRefresh:    make(chan bool),
		refreshEnabled: true,
		app:            cfg.App,
		onNodesUpdated: cfg.OnNodesUpdated,
	}

	// Set up table styling
	ml.table.SetBorder(true).
		SetTitle(" Proxmox VMs & Containers ").
		SetTitleAlign(tview.AlignLeft)

	// Set up selection change handler
	ml.table.SetSelectionChangedFunc(func(row, col int) {
		ml.selectedRow = row
	})

	// Start auto-refresh
	if cfg.RefreshInterval > 0 {
		ml.refreshTicker = time.NewTicker(cfg.RefreshInterval)
		go ml.autoRefresh()
	}

	return ml
}

// GetTable returns the underlying tview table
func (ml *MainList) GetTable() *tview.Table {
	return ml.table
}

// GetSelectedNode returns the currently selected VM/CT
func (ml *MainList) GetSelectedNode() *models.VMStatus {
	ml.refreshMutex.Lock()
	defer ml.refreshMutex.Unlock()

	// Get the current selection from the table directly
	row, _ := ml.table.GetSelection()

	// Row 0 is the header, so actual nodes start at row 1
	// Use sortedNodes since that's the order displayed in the table
	if row <= 0 || row > len(ml.sortedNodes) {
		return nil
	}
	return ml.sortedNodes[row-1]
}

// GetAllNodes returns all nodes
func (ml *MainList) GetAllNodes() []*models.VMStatus {
	ml.refreshMutex.Lock()
	defer ml.refreshMutex.Unlock()

	result := make([]*models.VMStatus, len(ml.nodes))
	copy(result, ml.nodes)
	return result
}

// Refresh fetches and updates the data
func (ml *MainList) Refresh(ctx context.Context) error {
	ml.refreshMutex.Lock()
	defer ml.refreshMutex.Unlock()

	nodes, err := ml.provider.GetNodes(ctx)
	if err != nil {
		return err
	}

	ml.nodes = nodes
	ml.updateTable()

	// Notify callback if set
	if ml.onNodesUpdated != nil {
		ml.onNodesUpdated(ml.nodes)
	}

	return nil
}

// updateTable rebuilds the table with current data
// sortNodes sorts nodes by type (CT first, then VM) and then by name
func sortNodes(nodes []*models.VMStatus) []*models.VMStatus {
	sorted := make([]*models.VMStatus, len(nodes))
	copy(sorted, nodes)

	// Bubble sort: CT < VM, then alphabetically by name
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if shouldSwap(sorted[i], sorted[j]) {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	return sorted
}

// shouldSwap determines if two nodes should be swapped in sort order
func shouldSwap(a, b *models.VMStatus) bool {
	// CT (container) should come before VM
	if a.Type == string(models.TypeVM) && b.Type == string(models.TypeContainer) {
		return true
	}
	// Within same type, sort alphabetically by name
	if a.Type == b.Type && a.Name > b.Name {
		return true
	}
	return false
}

// getStatusCell creates a status cell with appropriate color
func getStatusCell(status models.NodeState) *tview.TableCell {
	statusColor := colors.Current.DisabledColor
	switch status {
	case models.StateRunning:
		statusColor = colors.Current.OkColor
	case models.StateStopped:
		statusColor = colors.Current.AlertColor
	case models.StatePaused:
		statusColor = colors.Current.WarningColor
	}
	return tview.NewTableCell("●").
		SetTextColor(statusColor).
		SetAlign(tview.AlignCenter)
}

// getTypeCell creates a type cell with VM or CT label and color
func getTypeCell(nodeType models.NodeType) *tview.TableCell {
	typeText := "VM"
	typeColor := colors.VMColor
	if nodeType == models.TypeContainer {
		typeText = "CT"
		typeColor = colors.CTColor
	}
	return tview.NewTableCell(typeText).
		SetTextColor(typeColor).
		SetAlign(tview.AlignLeft)
}

// getResourceCell creates a cell for CPU or memory with threshold-based colors
func getResourceCell(value float64, text string, highThreshold, criticalThreshold float64) *tview.TableCell {
	cellColor := colors.Current.Foreground
	if value > criticalThreshold {
		cellColor = colors.Current.AlertColor
	} else if value > highThreshold {
		cellColor = colors.Current.WarningColor
	}
	return tview.NewTableCell(text).
		SetTextColor(cellColor).
		SetAlign(tview.AlignRight)
}

func (ml *MainList) updateTable() {
	ml.table.Clear()

	// Sort nodes
	ml.sortedNodes = sortNodes(ml.nodes)

	// Add header row
	headers := []string{"Status", "VMID", "Name", "Type", "Node", "CPU%", "Memory%", "Uptime"}
	for col, header := range headers {
		cell := tview.NewTableCell(header).
			SetTextColor(colors.Current.AccentForeground).
			SetAlign(tview.AlignLeft).
			SetSelectable(false)

		// Make the Name column (index 2) expandable
		if col == 2 {
			cell.SetExpansion(1)
		}

		ml.table.SetCell(0, col, cell)
	}

	// Add data rows
	for i, node := range ml.sortedNodes {
		row := i + 1

		// Status
		ml.table.SetCell(row, 0, getStatusCell(models.NodeState(node.Status)))

		// VMID
		ml.table.SetCell(row, 1, tview.NewTableCell(node.VMID).
			SetTextColor(colors.Current.Foreground).
			SetAlign(tview.AlignLeft))

		// Name (expandable to take remaining space)
		ml.table.SetCell(row, 2, tview.NewTableCell(node.Name).
			SetTextColor(colors.Current.Foreground).
			SetAlign(tview.AlignLeft).
			SetExpansion(1)) // This makes the column expand to fill available space

		// Type
		ml.table.SetCell(row, 3, getTypeCell(models.NodeType(node.Type)))

		// Node
		ml.table.SetCell(row, 4, tview.NewTableCell(node.Node).
			SetTextColor(colors.Current.Foreground).
			SetAlign(tview.AlignLeft))

		// CPU% (clamp negative values to 0)
		cpuUsage := node.CPUUsage
		if cpuUsage < 0 {
			cpuUsage = 0
		}
		cpuText := fmt.Sprintf("%.1f%%", cpuUsage)
		ml.table.SetCell(row, 5, getResourceCell(cpuUsage, cpuText, 50, 80))

		// Memory% (clamp negative values to 0)
		memUsage := node.MemoryUsage
		if memUsage < 0 {
			memUsage = 0
		}
		memText := fmt.Sprintf("%.1f%%", memUsage)
		ml.table.SetCell(row, 6, getResourceCell(memUsage, memText, 70, 90))

		// Uptime
		uptimeText := formatUptime(node.Uptime)
		ml.table.SetCell(row, 7, tview.NewTableCell(uptimeText).
			SetTextColor(colors.Current.Foreground).
			SetAlign(tview.AlignRight))
	}

	// Restore selection if valid
	if ml.selectedRow > 0 && ml.selectedRow <= len(ml.nodes) {
		ml.table.Select(ml.selectedRow, 0)
	} else if len(ml.nodes) > 0 {
		ml.table.Select(1, 0)
		ml.selectedRow = 1
	}
}

// autoRefresh periodically refreshes the data
func (ml *MainList) autoRefresh() {
	for {
		select {
		case <-ml.refreshTicker.C:
			if ml.refreshEnabled {
				ml.performRefresh()
			}
		case <-ml.stopRefresh:
			return
		}
	}
}

// performRefresh executes a single refresh cycle with timeout and error handling
func (ml *MainList) performRefresh() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := ml.Refresh(ctx)
	ml.updateTitleWithStatus(err)
}

// updateTitleWithStatus updates the table title based on refresh result
func (ml *MainList) updateTitleWithStatus(err error) {
	if ml.app == nil {
		return
	}

	var title string
	if err != nil {
		title = fmt.Sprintf(" Proxmox VMs & Containers [red](Error: %v)[-] ", err)
	} else {
		title = " Proxmox VMs & Containers "
	}

	ml.app.QueueUpdateDraw(func() {
		ml.table.SetTitle(title)
	})
}

// SetRefreshEnabled enables or disables auto-refresh
func (ml *MainList) SetRefreshEnabled(enabled bool) {
	ml.refreshEnabled = enabled
}

// Stop stops the auto-refresh
func (ml *MainList) Stop() {
	if ml.refreshTicker != nil {
		ml.refreshTicker.Stop()
	}
	close(ml.stopRefresh)
}

// formatUptime converts seconds to human-readable format
func formatUptime(seconds int64) string {
	if seconds == 0 {
		return "-"
	}

	days := seconds / 86400
	hours := (seconds % 86400) / 3600
	minutes := (seconds % 3600) / 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh", days, hours)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}
