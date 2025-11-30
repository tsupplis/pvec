package detailsdialog

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/tsupplis/pvec/pkg/models"
)

// GetDetailsText generates formatted text showing VM/CT details
func GetDetailsText(vm *models.VMStatus, config map[string]interface{}, width, height, scrollOffset int) string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#008000")).Bold(true)
	separatorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#008000"))
	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#008000")).Bold(true)

	// Title
	title := fmt.Sprintf(" Details: %s (%s) ", vm.Name, vm.VMID)
	padding := (width - len(title)) / 2
	if padding > 0 {
		b.WriteString(strings.Repeat(" ", padding))
	}
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n")
	b.WriteString(separatorStyle.Render(strings.Repeat("─", width)))
	b.WriteString("\n")

	// Build details
	details := buildDetails(vm, config)

	// Render visible rows with scrolling
	visibleRows := height - 3 // Title, separator, status bar
	endIdx := scrollOffset + visibleRows
	if endIdx > len(details) {
		endIdx = len(details)
	}

	for i := scrollOffset; i < endIdx; i++ {
		detail := details[i]
		if isSectionSeparator(detail) {
			// Section header - centered
			sepPadding := (width - len(detail.Key)) / 2
			if sepPadding > 0 {
				b.WriteString(strings.Repeat(" ", sepPadding))
			}
			b.WriteString(detail.Key)
		} else {
			// Property: Value
			b.WriteString(fmt.Sprintf("  %-18s : %s", detail.Key, detail.Value))
		}
		b.WriteString("\n")
	}

	// Fill remaining space
	for i := endIdx - scrollOffset; i < visibleRows; i++ {
		b.WriteString("\n")
	}

	// Status bar
	statusText := fmt.Sprintf(" ↑↓/jk=Scroll  ESC/Enter=Close  [%d/%d]", scrollOffset+1, len(details))
	b.WriteString(statusStyle.Render(statusText))

	return b.String()
}

// GetLoadingText shows a loading message
func GetLoadingText(vm *models.VMStatus, width, height int) string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#008000")).Bold(true)
	separatorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#008000"))
	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#008000")).Bold(true)

	title := fmt.Sprintf(" Details: %s (%s) ", vm.Name, vm.VMID)
	padding := (width - len(title)) / 2
	if padding > 0 {
		b.WriteString(strings.Repeat(" ", padding))
	}
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n")
	b.WriteString(separatorStyle.Render(strings.Repeat("─", width)))
	b.WriteString("\n")

	// Center loading message
	message := "Loading details..."
	for i := 0; i < (height-4)/2; i++ {
		b.WriteString("\n")
	}
	msgPadding := (width - len(message)) / 2
	if msgPadding > 0 {
		b.WriteString(strings.Repeat(" ", msgPadding))
	}
	b.WriteString(message)
	b.WriteString("\n")

	for i := 0; i < (height-4)/2; i++ {
		b.WriteString("\n")
	}

	b.WriteString(statusStyle.Render(" ESC/Enter Close"))
	return b.String()
}

// GetErrorText shows an error message
func GetErrorText(vm *models.VMStatus, err error, width, height int) string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#008000")).Bold(true)
	separatorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#008000"))
	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#008000")).Bold(true)

	title := fmt.Sprintf(" Details: %s (%s) ", vm.Name, vm.VMID)
	padding := (width - len(title)) / 2
	if padding > 0 {
		b.WriteString(strings.Repeat(" ", padding))
	}
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n")
	b.WriteString(separatorStyle.Render(strings.Repeat("─", width)))
	b.WriteString("\n\n")

	errorMsg := fmt.Sprintf("Error: %v", err)
	b.WriteString(errorMsg)
	b.WriteString("\n\n")

	// Show basic info
	b.WriteString(fmt.Sprintf("  %-18s : %s\n", "VMID", vm.VMID))
	b.WriteString(fmt.Sprintf("  %-18s : %s\n", "Name", vm.Name))
	b.WriteString(fmt.Sprintf("  %-18s : %s\n", "Type", vm.Type))
	b.WriteString(fmt.Sprintf("  %-18s : %s\n", "Status", vm.Status))

	// Fill remaining space
	currentLines := 7
	for i := currentLines; i < height-1; i++ {
		b.WriteString("\n")
	}

	b.WriteString(statusStyle.Render(" ESC/Enter Close"))
	return b.String()
}

// DetailItem represents a key-value pair
type DetailItem struct {
	Key   string
	Value string
}

// buildDetails creates a list of key-value pairs from the VM status and config
func buildDetails(vm *models.VMStatus, config map[string]interface{}) []DetailItem {
	// Start with basic VM details
	details := buildBasicDetails(vm)

	// Add VM-specific details (like guest agent)
	details = append(details, buildVMSpecificDetails(vm, config)...)

	// Add categorized config details
	details = append(details, buildConfigDetails(config)...)

	return details
}

// isSectionSeparator checks if a detail item is a section separator
func isSectionSeparator(detail DetailItem) bool {
	return strings.HasPrefix(detail.Key, "--") && strings.HasSuffix(detail.Key, "--")
}

// buildBasicDetails creates the core VM status details
func buildBasicDetails(vm *models.VMStatus) []DetailItem {
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
func buildVMSpecificDetails(vm *models.VMStatus, config map[string]interface{}) []DetailItem {
	var details []DetailItem

	// Add guest agent info for VMs (from config)
	if vm.Type == string(models.TypeVM) {
		details = append(details, DetailItem{"Guest Agent", getGuestAgentStatus(config)})
	}

	return details
}

// getGuestAgentStatus determines the guest agent status from config
func getGuestAgentStatus(config map[string]interface{}) string {
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
func buildConfigDetails(config map[string]interface{}) []DetailItem {
	if len(config) == 0 {
		return nil
	}

	// Categorize config keys
	resourceKeys, optionKeys := categorizeConfigKeys(config)

	// Sort both categories
	sort.Strings(resourceKeys)
	sort.Strings(optionKeys)

	var details []DetailItem

	// Add resource fields first
	if len(resourceKeys) > 0 {
		details = append(details, DetailItem{"-- resources --", ""})
		details = append(details, buildKeyValuePairs(resourceKeys, config)...)
	}

	// Add option fields
	if len(optionKeys) > 0 {
		details = append(details, DetailItem{"-- options --", ""})
		details = append(details, buildKeyValuePairs(optionKeys, config)...)
	}

	return details
}

// categorizeConfigKeys separates config keys into resource and option categories
func categorizeConfigKeys(config map[string]interface{}) ([]string, []string) {
	var resourceKeys []string
	var optionKeys []string

	for k := range config {
		// Skip fields we already display
		if !isDisplayedField(k) {
			if isResourceField(k) {
				resourceKeys = append(resourceKeys, k)
			} else {
				optionKeys = append(optionKeys, k)
			}
		}
	}

	return resourceKeys, optionKeys
}

// buildKeyValuePairs creates DetailItems from a list of keys and their config values
func buildKeyValuePairs(keys []string, config map[string]interface{}) []DetailItem {
	var details []DetailItem
	for _, k := range keys {
		value := formatValue(config[k])
		details = append(details, DetailItem{strings.ToLower(k), value})
	}
	return details
}

// isDisplayedField checks if a field is already displayed in main section
func isDisplayedField(key string) bool {
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
func isResourceField(key string) bool {
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
		return formatFloat(val)
	case bool:
		return formatBool(val)
	case []interface{}:
		return formatArray(val)
	case map[string]interface{}:
		return "{...}"
	default:
		return fmt.Sprintf("%v", v)
	}
}

// formatFloat formats a float64 value
func formatFloat(val float64) string {
	if val == float64(int64(val)) {
		return fmt.Sprintf("%d", int64(val))
	}
	return fmt.Sprintf("%.2f", val)
}

// formatBool formats a boolean value
func formatBool(val bool) string {
	if val {
		return "true"
	}
	return "false"
}

// formatArray formats an array of values
func formatArray(val []interface{}) string {
	var items []string
	for _, item := range val {
		items = append(items, formatValue(item))
	}
	return "[" + strings.Join(items, ", ") + "]"
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
