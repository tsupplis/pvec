package detailsdialog

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/tsupplis/pvec/pkg/models"
	"github.com/tsupplis/pvec/pkg/proxmox"
	"github.com/tsupplis/pvec/pkg/ui/colors"
)

// DetailsDialog displays all attributes of a VM/CT in a scrollable key-value list
type DetailsDialog struct {
	pages  *tview.Pages
	app    *tview.Application
	client proxmox.Client
	table  *tview.Table
	flex   *tview.Flex
}

// NewDetailsDialog creates a new details dialog
func NewDetailsDialog(pages *tview.Pages, app *tview.Application, client proxmox.Client) *DetailsDialog {
	dd := &DetailsDialog{
		pages:  pages,
		app:    app,
		client: client,
	}

	// Create table with two columns
	dd.table = tview.NewTable()
	dd.table.SetBorders(false) // No internal borders between cells
	dd.table.SetBackgroundColor(colors.Current.Background)
	dd.table.SetSelectable(true, false)
	dd.table.SetFixed(1, 2) // Fix header row and 2 columns
	dd.table.SetBorderColor(colors.Current.Foreground)

	// Wrap in frame for borders (no title to maximize table space)
	frame := tview.NewFrame(dd.table).
		SetBorders(1, 1, 1, 1, 2, 2)
	frame.SetBorder(true).
		SetBackgroundColor(colors.Current.Background).
		SetBorderColor(colors.Current.Foreground)

	// Center the frame with responsive size (screen size - 4)
	dd.flex = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 2, 1, false). // Top padding of 2
		AddItem(tview.NewFlex().
						AddItem(nil, 2, 1, false).              // Left padding of 2
						AddItem(frame, 0, 1, true).             // Frame takes remaining space
						AddItem(nil, 2, 1, false), 0, 1, true). // Right padding of 2
		AddItem(nil, 2, 1, false) // Bottom padding of 2

	dd.flex.SetBackgroundColor(colors.Current.Background)

	// Set up input capture for closing
	dd.table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyESC || event.Key() == tcell.KeyEnter {
			dd.Hide()
			return nil
		}
		return event
	})

	return dd
}

// Show displays the details dialog for the given VM/CT
func (dd *DetailsDialog) Show(vm *models.VMStatus) {
	// Show loading state
	dd.showLoading()
	dd.pages.AddPage("details", dd.flex, true, true)
	dd.pages.ShowPage("details")

	// Fetch config in background
	go func() {
		ctx := context.Background()
		vmType := string(vm.Type)
		config, err := dd.client.GetVMConfig(ctx, vm.Node, vmType, vm.VMID)

		// Update UI on main thread
		dd.app.QueueUpdateDraw(func() {
			if err != nil {
				dd.showError(vm, err)
			} else {
				dd.populateTable(vm, config)
			}
		})
	}()
}

// Hide closes the details dialog
func (dd *DetailsDialog) Hide() {
	dd.pages.HidePage("details")
	dd.pages.RemovePage("details")
}

// showLoading displays loading message
func (dd *DetailsDialog) showLoading() {
	dd.table.Clear()
	dd.table.SetCell(0, 0, tview.NewTableCell("Loading details...").
		SetTextColor(colors.Current.Foreground).
		SetAlign(tview.AlignCenter))
}

// showError displays error message
func (dd *DetailsDialog) showError(vm *models.VMStatus, err error) {
	dd.table.Clear()
	dd.table.SetCell(0, 0, tview.NewTableCell("Error").
		SetTextColor(colors.Current.AlertColor).
		SetAlign(tview.AlignLeft))
	dd.table.SetCell(0, 1, tview.NewTableCell(err.Error()).
		SetTextColor(colors.Current.AlertColor).
		SetAlign(tview.AlignLeft))

	// Show basic info even on error
	dd.table.SetCell(2, 0, tview.NewTableCell("VMID").
		SetTextColor(colors.Current.Foreground))
	dd.table.SetCell(2, 1, tview.NewTableCell(vm.VMID).
		SetTextColor(colors.Current.Foreground))
	dd.table.SetCell(3, 0, tview.NewTableCell("Name").
		SetTextColor(colors.Current.Foreground))
	dd.table.SetCell(3, 1, tview.NewTableCell(vm.Name).
		SetTextColor(colors.Current.Foreground))
}

// populateTable fills the table with VM/CT details
func (dd *DetailsDialog) populateTable(vm *models.VMStatus, config map[string]interface{}) {
	dd.table.Clear()

	// Add header row
	dd.addHeaderRow()

	// Build and add detail rows
	details := dd.buildDetails(vm, config)
	dd.addDetailRows(details)

	// Select first data row and scroll to top
	if len(details) > 0 {
		dd.table.Select(1, 0)
		dd.table.ScrollToBeginning()
	}
}

// addHeaderRow creates the table header
func (dd *DetailsDialog) addHeaderRow() {
	dd.table.SetCell(0, 0, tview.NewTableCell("Property").
		SetTextColor(colors.Current.AccentForeground).
		SetAlign(tview.AlignLeft).
		SetSelectable(false))

	dd.table.SetCell(0, 1, tview.NewTableCell("Value").
		SetTextColor(colors.Current.AccentForeground).
		SetAlign(tview.AlignLeft).
		SetSelectable(false))
}

// addDetailRows adds all detail rows to the table
func (dd *DetailsDialog) addDetailRows(details []DetailItem) {
	for i, detail := range details {
		row := i + 1
		if dd.isSectionSeparator(detail) {
			dd.addSeparatorRow(row, detail)
		} else {
			dd.addPropertyRow(row, detail)
		}
	}
}

// isSectionSeparator checks if a detail item is a section separator
func (dd *DetailsDialog) isSectionSeparator(detail DetailItem) bool {
	return strings.HasPrefix(detail.Key, "--") && strings.HasSuffix(detail.Key, "--")
}

// addSeparatorRow adds a section separator row
func (dd *DetailsDialog) addSeparatorRow(row int, detail DetailItem) {
	dd.table.SetCell(row, 0, tview.NewTableCell(detail.Key).
		SetTextColor(colors.Current.AccentForeground).
		SetAlign(tview.AlignCenter).
		SetSelectable(false))
	dd.table.SetCell(row, 1, tview.NewTableCell("").
		SetTextColor(colors.Current.AccentForeground).
		SetSelectable(false))
}

// addPropertyRow adds a property-value row
func (dd *DetailsDialog) addPropertyRow(row int, detail DetailItem) {
	// Property name (fixed width - pad to 18 chars)
	keyText := dd.formatKeyText(detail.Key)
	dd.table.SetCell(row, 0, tview.NewTableCell(keyText).
		SetTextColor(colors.Current.Foreground).
		SetAlign(tview.AlignLeft))

	// Value (responsive - no fixed width, let it grow/shrink)
	valueText := detail.Value
	dd.table.SetCell(row, 1, tview.NewTableCell(valueText).
		SetTextColor(colors.Current.Foreground).
		SetAlign(tview.AlignLeft).
		SetExpansion(1)) // This makes the column expandable
}

// formatKeyText formats the key text with proper truncation and padding
func (dd *DetailsDialog) formatKeyText(key string) string {
	keyText := key
	if len(keyText) > 18 {
		keyText = keyText[:18]
	}
	return fmt.Sprintf("%-18s", keyText) // Left-align and pad to 18 chars
}

// DetailItem represents a key-value pair
type DetailItem struct {
	Key   string
	Value string
}

// buildDetails creates a list of key-value pairs from the VM status and config
func (dd *DetailsDialog) buildDetails(vm *models.VMStatus, config map[string]interface{}) []DetailItem {
	// Start with basic VM details
	details := dd.buildBasicDetails(vm)

	// Add VM-specific details (like guest agent)
	details = append(details, dd.buildVMSpecificDetails(vm, config)...)

	// Add categorized config details
	details = append(details, dd.buildConfigDetails(config)...)

	return details
}

// buildBasicDetails creates the core VM status details
func (dd *DetailsDialog) buildBasicDetails(vm *models.VMStatus) []DetailItem {
	return []DetailItem{
		{"VMID", vm.VMID},
		{"Name", vm.Name},
		{"Type", string(vm.Type)},
		{"Status", string(vm.Status)},
		{"Node", vm.Node},
		{"CPU Usage", fmt.Sprintf("%.2f%%", vm.CPUUsage)},
		{"Memory Usage", fmt.Sprintf("%.2f%%", vm.MemoryUsage)},
		{"Max Memory", formatBytes(vm.MaxMem)},
		{"Max CPU", fmt.Sprintf("%d cores", vm.MaxCPU)},
		{"Uptime", formatUptime(vm.Uptime)},
	}
}

// buildVMSpecificDetails adds VM-type specific details like guest agent
func (dd *DetailsDialog) buildVMSpecificDetails(vm *models.VMStatus, config map[string]interface{}) []DetailItem {
	var details []DetailItem

	// Add guest agent info for VMs (from config)
	if vm.Type == string(models.TypeVM) {
		details = append(details, DetailItem{"Guest Agent", dd.getGuestAgentStatus(config)})
	}

	return details
}

// getGuestAgentStatus determines the guest agent status from config
func (dd *DetailsDialog) getGuestAgentStatus(config map[string]interface{}) string {
	agentValue, exists := config["agent"]
	if !exists {
		return "Not Configured"
	}

	agentStr := fmt.Sprintf("%v", agentValue)
	// Parse agent configuration (can be "1", "1,type=isa", etc.)
	if strings.HasPrefix(agentStr, "1") {
		return "Enabled"
	}
	return "Disabled"
}

// buildConfigDetails organizes additional config fields by category
func (dd *DetailsDialog) buildConfigDetails(config map[string]interface{}) []DetailItem {
	if len(config) == 0 {
		return nil
	}

	// Categorize config keys
	resourceKeys, optionKeys := dd.categorizeConfigKeys(config)

	// Sort both categories
	sort.Strings(resourceKeys)
	sort.Strings(optionKeys)

	var details []DetailItem

	// Add resource fields first
	if len(resourceKeys) > 0 {
		details = append(details, DetailItem{"-- resources --", ""})
		details = append(details, dd.buildKeyValuePairs(resourceKeys, config)...)
	}

	// Add option fields
	if len(optionKeys) > 0 {
		details = append(details, DetailItem{"-- options --", ""})
		details = append(details, dd.buildKeyValuePairs(optionKeys, config)...)
	}

	return details
}

// categorizeConfigKeys separates config keys into resource and option categories
func (dd *DetailsDialog) categorizeConfigKeys(config map[string]interface{}) ([]string, []string) {
	var resourceKeys []string
	var optionKeys []string

	for k := range config {
		// Skip fields we already display
		if !dd.isDisplayedField(k) {
			if dd.isResourceField(k) {
				resourceKeys = append(resourceKeys, k)
			} else {
				optionKeys = append(optionKeys, k)
			}
		}
	}

	return resourceKeys, optionKeys
}

// buildKeyValuePairs creates DetailItems from a list of keys and their config values
func (dd *DetailsDialog) buildKeyValuePairs(keys []string, config map[string]interface{}) []DetailItem {
	var details []DetailItem
	for _, k := range keys {
		value := formatValue(config[k])
		details = append(details, DetailItem{strings.ToLower(k), value})
	}
	return details
}

// isDisplayedField checks if a field is already displayed in main section
func (dd *DetailsDialog) isDisplayedField(key string) bool {
	displayed := []string{
		"vmid", "name", "type", "status", "node",
		"cpu", "mem", "maxmem", "maxcpu", "uptime", "agent",
	}
	keyLower := strings.ToLower(key)
	for _, d := range displayed {
		if keyLower == d {
			return true
		}
	}
	return false
}

// isResourceField checks if a field is related to resources (hardware/networking)
func (dd *DetailsDialog) isResourceField(key string) bool {
	resourceFields := []string{
		"cores", "sockets", "cpu", "vcpus", "cpulimit", "cpuunits",
		"memory", "balloon", "shares",
		"net0", "net1", "net2", "net3", "net4", "net5",
		"scsi0", "scsi1", "scsi2", "scsi3", "ide0", "ide1", "ide2", "ide3",
		"sata0", "sata1", "sata2", "sata3", "virtio0", "virtio1", "virtio2", "virtio3",
		"mp0", "mp1", "mp2", "mp3", "mp4", "mp5", "mp6", "mp7", "mp8", "mp9",
		"rootfs", "swap",
		"numa", "hotplug", "localtime", "machine", "ostype",
		"bootdisk", "boot", "cdrom", "tablet", "usb0", "usb1", "usb2", "usb3",
		"serial0", "serial1", "serial2", "serial3",
		"vga", "args",
	}
	keyLower := strings.ToLower(key)
	for _, field := range resourceFields {
		if strings.HasPrefix(keyLower, field) {
			return true
		}
	}
	return false
}

// formatValue converts an interface{} to a readable string
func formatValue(v interface{}) string {
	if v == nil {
		return "null"
	}

	switch val := v.(type) {
	case string:
		return val
	case float64:
		// Check if it's a whole number
		if val == float64(int64(val)) {
			return fmt.Sprintf("%d", int64(val))
		}
		return fmt.Sprintf("%.2f", val)
	case bool:
		if val {
			return "true"
		}
		return "false"
	case []interface{}:
		// Format arrays
		var items []string
		for _, item := range val {
			items = append(items, formatValue(item))
		}
		return "[" + strings.Join(items, ", ") + "]"
	case map[string]interface{}:
		return "{...}" // Don't expand nested objects
	default:
		return fmt.Sprintf("%v", v)
	}
}

// formatBytes formats bytes into human-readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	units := []string{"KB", "MB", "GB", "TB", "PB", "EB"}
	if exp >= len(units) {
		exp = len(units) - 1 // Prevent index out of bounds
	}
	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), units[exp])
}

// formatUptime formats uptime seconds into human-readable format
func formatUptime(seconds int64) string {
	if seconds <= 0 {
		return "N/A"
	}

	days := seconds / 86400
	hours := (seconds % 86400) / 3600
	minutes := (seconds % 3600) / 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}
