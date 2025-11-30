package mainlist

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tsupplis/pvec/pkg/actions"
	"github.com/tsupplis/pvec/pkg/config"
	"github.com/tsupplis/pvec/pkg/models"
	"github.com/tsupplis/pvec/pkg/proxmox"
	"github.com/tsupplis/pvec/pkg/ui/configpanel"
	"github.com/tsupplis/pvec/pkg/ui/detailsdialog"
	"github.com/tsupplis/pvec/pkg/ui/helpdialog"
)

// DataProvider is the interface for fetching node data
type DataProvider interface {
	GetNodes(ctx context.Context) ([]*models.VMStatus, error)
}

// MainList is the main scrolling list component
type MainList struct {
	program        *tea.Program
	model          *listModel
	nodes          []*models.VMStatus
	sortedNodes    []*models.VMStatus
	selectedIdx    int
	provider       DataProvider
	client         proxmox.Client
	refreshTicker  *time.Ticker
	stopRefresh    chan bool
	refreshMutex   sync.Mutex
	refreshEnabled bool
	onNodesUpdated func([]*models.VMStatus)
	lastError      error
	appConfig      *config.Config
	configLoader   config.Loader
}

type listModel struct {
	parent         *MainList
	width          int
	height         int
	scrollOffset   int
	cursorPosition int
	showHelp       bool
	showDetails    bool
	detailsVM      *models.VMStatus
	detailsConfig  map[string]interface{}
	detailsLoading bool
	detailsError   error
	detailsScroll  int
	showAction     bool
	actionVM       *models.VMStatus
	actionName     string
	actionDone     bool
	actionError    error
	showConfig     bool
	configModel    *configpanel.Model
}

type refreshMsg struct {
	nodes []*models.VMStatus
	err   error
}

type configLoadedMsg struct {
	config map[string]interface{}
	err    error
}

type actionResultMsg struct {
	err error
}

type tickMsg time.Time

// Config holds the configuration for the main list
type Config struct {
	RefreshInterval time.Duration
	Provider        DataProvider
	Client          proxmox.Client           // For fetching detailed config
	OnNodesUpdated  func([]*models.VMStatus) // Callback when nodes are refreshed
	AppConfig       *config.Config           // Application configuration
	ConfigLoader    config.Loader            // Configuration loader
}

// NewMainList creates a new main list component
func NewMainList(cfg Config) *MainList {
	ml := &MainList{
		nodes:          make([]*models.VMStatus, 0),
		selectedIdx:    0,
		provider:       cfg.Provider,
		client:         cfg.Client,
		stopRefresh:    make(chan bool),
		refreshEnabled: true,
		onNodesUpdated: cfg.OnNodesUpdated,
		appConfig:      cfg.AppConfig,
		configLoader:   cfg.ConfigLoader,
	}

	model := &listModel{
		parent:         ml,
		width:          80,
		height:         24,
		cursorPosition: 0,
		scrollOffset:   0,
	}

	ml.model = model
	ml.program = tea.NewProgram(model, tea.WithAltScreen())

	// Start auto-refresh
	if cfg.RefreshInterval > 0 {
		ml.refreshTicker = time.NewTicker(cfg.RefreshInterval)
		go ml.autoRefresh()
	}

	return ml
}

// Init implements tea.Model
func (m *listModel) Init() tea.Cmd {
	// Trigger initial refresh
	go m.parent.performRefresh()
	return tickCmd()
}

// Update implements tea.Model
func (m *listModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle config panel messages
	if handled, model, cmd := m.handleConfigPanelMsg(msg); handled {
		return model, cmd
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.handleWindowSize(msg)
	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	case refreshMsg:
		return m.handleRefresh(msg)
	case configLoadedMsg:
		return m.handleConfigLoaded(msg)
	case actionResultMsg:
		return m.handleActionResult(msg)
	case tickMsg:
		return m, tickCmd()
	}

	return m, nil
}

// handleConfigPanelMsg handles config panel related messages
func (m *listModel) handleConfigPanelMsg(msg tea.Msg) (bool, tea.Model, tea.Cmd) {
	// First, let the config panel handle the message if it's showing
	if m.showConfig && m.configModel != nil {
		updatedModel, cmd := m.configModel.Update(msg)
		if updated, ok := updatedModel.(configpanel.Model); ok {
			m.configModel = &updated
		}

		// Check if this generated a CloseMsg or SaveResultMsg
		if _, ok := msg.(configpanel.CloseMsg); ok {
			m.showConfig = false
			return true, m, nil
		}

		// Handle save result - reinitialize client with new config
		if saveMsg, ok := msg.(configpanel.SaveResultMsg); ok {
			// Clear the node list immediately (config might be invalid)
			m.parent.refreshMutex.Lock()
			m.parent.nodes = nil
			m.parent.sortedNodes = nil
			m.parent.selectedIdx = 0
			m.cursorPosition = 0
			m.scrollOffset = 0
			m.parent.refreshMutex.Unlock()

			if saveMsg.Err() == nil {
				// Close the config panel on successful save
				m.showConfig = false
				// Reinitialize the Proxmox client with updated configuration
				m.parent.reinitializeClient()
				// Trigger immediate refresh with new client
				return true, m, m.parent.refreshCmd()
			}
			// Keep panel open on error
			return true, m, cmd
		}

		return true, m, cmd
	}

	return false, m, nil
}

// handleWindowSize updates dimensions
func (m *listModel) handleWindowSize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height
	return m, nil
}

// handleRefresh processes node list refresh
func (m *listModel) handleRefresh(msg refreshMsg) (tea.Model, tea.Cmd) {
	m.parent.refreshMutex.Lock()
	m.parent.nodes = msg.nodes
	m.parent.lastError = msg.err
	if msg.nodes != nil {
		m.parent.sortedNodes = sortNodes(msg.nodes)
	}
	m.parent.refreshMutex.Unlock()

	if m.parent.onNodesUpdated != nil && msg.nodes != nil {
		m.parent.onNodesUpdated(msg.nodes)
	}
	return m, nil
}

// handleConfigLoaded processes loaded VM config
func (m *listModel) handleConfigLoaded(msg configLoadedMsg) (tea.Model, tea.Cmd) {
	m.detailsLoading = false
	m.detailsConfig = msg.config
	m.detailsError = msg.err
	return m, nil
}

// handleActionResult processes action execution result
func (m *listModel) handleActionResult(msg actionResultMsg) (tea.Model, tea.Cmd) {
	m.actionDone = true
	m.actionError = msg.err
	return m, nil
}

func (m *listModel) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Check if any dialog is open and handle accordingly
	if handled, model, cmd := m.handleDialogKeys(msg); handled {
		return model, cmd
	}

	// Handle function keys (F1-F10)
	if handled, model, cmd := m.handleFunctionKeys(msg); handled {
		return model, cmd
	}

	// Handle list navigation
	return m.handleNavigationKeys(msg)
}

// handleDialogKeys processes keys when a dialog is open
func (m *listModel) handleDialogKeys(msg tea.KeyMsg) (bool, tea.Model, tea.Cmd) {
	if m.showHelp {
		return m.handleHelpDialogKeys(msg)
	}
	if m.showDetails {
		return m.handleDetailsDialogKeys(msg)
	}
	if m.showAction {
		return m.handleActionDialogKeys()
	}
	return false, m, nil
}

// handleHelpDialogKeys handles keys when help dialog is open
func (m *listModel) handleHelpDialogKeys(msg tea.KeyMsg) (bool, tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "enter":
		m.showHelp = false
		return true, m, nil
	}
	return true, m, nil
}

// handleDetailsDialogKeys handles keys when details dialog is open
func (m *listModel) handleDetailsDialogKeys(msg tea.KeyMsg) (bool, tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "enter":
		m.showDetails = false
		m.detailsScroll = 0
		return true, m, nil
	case "up", "k":
		if m.detailsScroll > 0 {
			m.detailsScroll--
		}
		return true, m, nil
	case "down", "j":
		m.detailsScroll++
		return true, m, nil
	}
	return true, m, nil
}

// handleActionDialogKeys handles keys when action dialog is open
func (m *listModel) handleActionDialogKeys() (bool, tea.Model, tea.Cmd) {
	if m.actionDone {
		m.showAction = false
		m.actionDone = false
		m.actionError = nil
		return true, m, nil
	}
	return true, m, nil
}

// handleFunctionKeys processes function key presses
func (m *listModel) handleFunctionKeys(msg tea.KeyMsg) (bool, tea.Model, tea.Cmd) {
	switch msg.String() {
	case "f1", "h":
		return m.handleHelpKey()
	case "f2", "c":
		return m.handleConfigKey()
	case "f3", "enter":
		return m.handleDetailsKey()
	case "f4", "s":
		return m.handleActionKey("start")
	case "f5", "d":
		return m.handleActionKey("shutdown")
	case "f6", "r":
		return m.handleActionKey("reboot")
	case "f7", "t":
		return m.handleActionKey("stop")
	case "f10", "q", "ctrl+c":
		return true, m, tea.Quit
	}
	return false, m, nil
}

// handleHelpKey shows the help dialog
func (m *listModel) handleHelpKey() (bool, tea.Model, tea.Cmd) {
	m.showHelp = true
	return true, m, nil
}

// handleConfigKey shows the config panel
func (m *listModel) handleConfigKey() (bool, tea.Model, tea.Cmd) {
	m.showConfig = true
	// Always recreate the config model to reset any unsaved changes
	if m.parent.appConfig != nil && m.parent.configLoader != nil {
		model := configpanel.NewModel(m.parent.appConfig, m.parent.configLoader)
		updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
		if updated, ok := updatedModel.(configpanel.Model); ok {
			m.configModel = &updated
		}
	}
	return true, m, nil
}

// handleDetailsKey shows details for selected VM
func (m *listModel) handleDetailsKey() (bool, tea.Model, tea.Cmd) {
	m.parent.refreshMutex.Lock()
	if m.parent.selectedIdx >= 0 && m.parent.selectedIdx < len(m.parent.sortedNodes) {
		vm := m.parent.sortedNodes[m.parent.selectedIdx]
		m.parent.refreshMutex.Unlock()
		m.showDetails = true
		m.detailsVM = vm
		m.detailsLoading = true
		m.detailsConfig = nil
		m.detailsError = nil
		m.detailsScroll = 0
		return true, m, m.loadConfig(vm)
	}
	m.parent.refreshMutex.Unlock()
	return true, m, nil
}

// handleActionKey executes an action on selected VM
func (m *listModel) handleActionKey(action string) (bool, tea.Model, tea.Cmd) {
	model, cmd := m.executeAction(action)
	return true, model, cmd
}

// handleNavigationKeys processes list navigation keys
func (m *listModel) handleNavigationKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.parent.refreshMutex.Lock()
	defer m.parent.refreshMutex.Unlock()

	maxIdx := len(m.parent.sortedNodes) - 1
	if maxIdx < 0 {
		return m, nil
	}

	switch msg.String() {
	case "up", "k":
		m.moveCursorUp()
	case "down", "j":
		m.moveCursorDown(maxIdx)
	case "home", "g":
		m.moveCursorHome()
	case "end", "G":
		m.moveCursorEnd(maxIdx)
	case "pgup":
		m.moveCursorPageUp()
	case "pgdown":
		m.moveCursorPageDown(maxIdx)
	}

	return m, nil
}

// moveCursorUp moves cursor up one position
func (m *listModel) moveCursorUp() {
	if m.cursorPosition > 0 {
		m.cursorPosition--
		m.parent.selectedIdx = m.cursorPosition
		if m.cursorPosition < m.scrollOffset {
			m.scrollOffset = m.cursorPosition
		}
	}
}

// moveCursorDown moves cursor down one position
func (m *listModel) moveCursorDown(maxIdx int) {
	if m.cursorPosition < maxIdx {
		m.cursorPosition++
		m.parent.selectedIdx = m.cursorPosition
		visibleRows := m.height - 4
		if m.cursorPosition >= m.scrollOffset+visibleRows {
			m.scrollOffset = m.cursorPosition - visibleRows + 1
		}
	}
}

// moveCursorHome moves cursor to first position
func (m *listModel) moveCursorHome() {
	m.cursorPosition = 0
	m.parent.selectedIdx = 0
	m.scrollOffset = 0
}

// moveCursorEnd moves cursor to last position
func (m *listModel) moveCursorEnd(maxIdx int) {
	m.cursorPosition = maxIdx
	m.parent.selectedIdx = maxIdx
	visibleRows := m.height - 4
	if maxIdx >= visibleRows {
		m.scrollOffset = maxIdx - visibleRows + 1
	}
}

// moveCursorPageUp moves cursor up one page
func (m *listModel) moveCursorPageUp() {
	visibleRows := m.height - 4
	m.cursorPosition -= visibleRows
	if m.cursorPosition < 0 {
		m.cursorPosition = 0
	}
	m.parent.selectedIdx = m.cursorPosition
	m.scrollOffset = m.cursorPosition
}

// moveCursorPageDown moves cursor down one page
func (m *listModel) moveCursorPageDown(maxIdx int) {
	visibleRows := m.height - 4
	m.cursorPosition += visibleRows
	if m.cursorPosition > maxIdx {
		m.cursorPosition = maxIdx
	}
	m.parent.selectedIdx = m.cursorPosition
	if m.cursorPosition >= visibleRows {
		m.scrollOffset = m.cursorPosition - visibleRows + 1
	}
}

// loadConfig fetches VM config in background
func (m *listModel) loadConfig(vm *models.VMStatus) tea.Cmd {
	return func() tea.Msg {
		if m.parent.client == nil {
			return configLoadedMsg{config: nil, err: fmt.Errorf("client not available")}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		config, err := m.parent.client.GetVMConfig(ctx, vm.Node, vm.Type, vm.VMID)
		return configLoadedMsg{config: config, err: err}
	}
}

// executorAdapter wraps proxmox.Client to match actions.Executor interface
type executorAdapter struct {
	client proxmox.Client
	node   string
	vmType string
}

func (e *executorAdapter) Start(ctx context.Context, vmid string) error {
	return e.client.Start(ctx, e.node, e.vmType, vmid)
}

func (e *executorAdapter) Shutdown(ctx context.Context, vmid string) error {
	return e.client.Shutdown(ctx, e.node, e.vmType, vmid)
}

func (e *executorAdapter) Reboot(ctx context.Context, vmid string) error {
	return e.client.Reboot(ctx, e.node, e.vmType, vmid)
}

func (e *executorAdapter) Stop(ctx context.Context, vmid string) error {
	return e.client.Stop(ctx, e.node, e.vmType, vmid)
}

func (m *listModel) executeAction(actionName string) (tea.Model, tea.Cmd) {
	m.parent.refreshMutex.Lock()
	if m.parent.selectedIdx < 0 || m.parent.selectedIdx >= len(m.parent.sortedNodes) {
		m.parent.refreshMutex.Unlock()
		return m, nil
	}
	vm := m.parent.sortedNodes[m.parent.selectedIdx]
	m.parent.refreshMutex.Unlock()

	// Show action dialog in executing state
	m.showAction = true
	m.actionVM = vm
	m.actionName = actionName
	m.actionDone = false
	m.actionError = nil

	// Execute action asynchronously
	return m, func() tea.Msg {
		if m.parent.client == nil {
			return actionResultMsg{err: fmt.Errorf("client not available")}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		// Create adapter to match actions.Executor interface
		executor := &executorAdapter{
			client: m.parent.client,
			node:   vm.Node,
			vmType: vm.Type,
		}

		var action actions.Action
		switch actionName {
		case "start":
			action = actions.NewStartAction(executor, vm)
		case "shutdown":
			action = actions.NewShutdownAction(executor, vm)
		case "reboot":
			action = actions.NewRebootAction(executor, vm)
		case "stop":
			action = actions.NewStopAction(executor, vm)
		default:
			return actionResultMsg{err: fmt.Errorf("unknown action: %s", actionName)}
		}

		err := action.Execute(ctx)
		return actionResultMsg{err: err}
	}
}

// View implements tea.Model
func (m *listModel) View() string {
	// Show help dialog if requested (full screen)
	if m.showHelp {
		return helpdialog.GetHelpText(m.width, m.height)
	}

	// Show config panel if requested (full screen)
	if m.showConfig && m.configModel != nil {
		return m.configModel.View()
	}

	// Show details dialog if requested (full screen)
	if m.showDetails && m.detailsVM != nil {
		if m.detailsLoading {
			return detailsdialog.GetLoadingText(m.detailsVM, m.width, m.height)
		} else if m.detailsError != nil {
			return detailsdialog.GetErrorText(m.detailsVM, m.detailsError, m.width, m.height)
		} else {
			return detailsdialog.GetDetailsText(m.detailsVM, m.detailsConfig, m.width, m.height, m.detailsScroll)
		}
	}

	// Render main list
	return m.renderMainList()
}

// renderMainList renders the main VM/Container list
func (m *listModel) renderMainList() string {

	m.parent.refreshMutex.Lock()
	defer m.parent.refreshMutex.Unlock()

	var b strings.Builder

	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#008000")).Bold(true)

	// Title
	title := "Proxmox VMs & Containers "
	if m.parent.lastError != nil {
		title = "Proxmox VMs & Containers (Error Connecting) "
	}
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n")

	headerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#008000")).Bold(true)
	separatorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#008000"))

	// Header
	header := fmt.Sprintf("%-6s %-6s %-20s %-4s %-10s %8s %8s %10s",
		"Status", "VMID", "Name", "Type", "Node", "CPU%", "Memory%", "Uptime")
	b.WriteString(headerStyle.Render(header))
	b.WriteString("\n")

	// Separator
	separator := strings.Repeat("─", m.width)
	b.WriteString(separatorStyle.Render(separator))
	b.WriteString("\n")

	// Rows
	visibleRows := m.height - 4 // Title, header, separator, and status bar
	endIdx := m.scrollOffset + visibleRows
	if endIdx > len(m.parent.sortedNodes) {
		endIdx = len(m.parent.sortedNodes)
	}

	for i := m.scrollOffset; i < endIdx; i++ {
		node := m.parent.sortedNodes[i]
		b.WriteString(m.renderRow(node, i == m.cursorPosition))
		b.WriteString("\n")
	}

	// Fill remaining space
	for i := endIdx - m.scrollOffset; i < visibleRows; i++ {
		b.WriteString("\n")
	}

	// Status bar
	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#008000")).Bold(true)
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Bold(true)

	var statusText string
	if m.showAction && m.actionVM != nil {
		actionCap := strings.Title(m.actionName)
		if m.actionDone {
			if m.actionError != nil {
				statusText = errorStyle.Render(fmt.Sprintf("Failed to %s %s. - Press any key", m.actionName, m.actionVM.VMID))
			} else {
				statusText = errorStyle.Render(fmt.Sprintf("Succeeded in %s %s. - Press any key", m.actionName, m.actionVM.VMID))
			}
		} else {
			statusText = statusStyle.Render(fmt.Sprintf("%s on %s...", actionCap, m.actionVM.Name))
		}
	} else {
		statusText = "F1 Help  F2 Conf  F3 Info F4 Start  F5 shutDown  F6 Reboot  F7 sTop  F10 Quit"
	}
	b.WriteString(statusStyle.Render(statusText))
	return b.String()
}

func (m *listModel) renderRow(node *models.VMStatus, selected bool) string {
	// Status indicator
	statusSymbol := node.Status

	// Type
	typeText := "VM"
	if models.NodeType(node.Type) == models.TypeContainer {
		typeText = "CT"
	}

	// CPU and Memory
	cpuUsage := node.CPUUsage
	if cpuUsage < 0 {
		cpuUsage = 0
	}
	cpuText := fmt.Sprintf("%.1f%%", cpuUsage)

	memUsage := node.MemoryUsage
	if memUsage < 0 {
		memUsage = 0
	}
	memText := fmt.Sprintf("%.1f%%", memUsage)

	// Uptime
	uptimeText := formatUptime(node.Uptime)

	// Build row
	row := fmt.Sprintf("%-6s %-6s %-20s %-4s %-10s %8s %8s %10s",
		statusSymbol,
		node.VMID,
		truncate(node.Name, 20),
		typeText,
		truncate(node.Node, 10),
		cpuText,
		memText,
		uptimeText)

	// Apply selection style first
	if selected {
		rowStyle := lipgloss.NewStyle().
			Reverse(true).
			Width(m.width)
		return rowStyle.Render(row)
	}

	// Apply color to status symbol after selection (only for non-selected rows)
	runningStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#008000"))
	stoppedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))

	if node.Status == string(models.StateRunning) {
		// Replace the status symbol with colored version
		row = runningStyle.Render(statusSymbol) + row[len(statusSymbol):]
	} else if node.Status == string(models.StateStopped) {
		// Replace the status symbol with colored version
		row = stoppedStyle.Render(statusSymbol) + row[len(statusSymbol):]
	}

	return row
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// tickCmd generates tick messages for the event loop
func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
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

// performRefresh executes a single refresh cycle
func (ml *MainList) performRefresh() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	nodes, err := ml.provider.GetNodes(ctx)

	ml.program.Send(refreshMsg{nodes: nodes, err: err})
}

// GetSelectedNode returns the currently selected VM/CT
func (ml *MainList) GetSelectedNode() *models.VMStatus {
	ml.refreshMutex.Lock()
	defer ml.refreshMutex.Unlock()

	if ml.selectedIdx < 0 || ml.selectedIdx >= len(ml.sortedNodes) {
		return nil
	}
	return ml.sortedNodes[ml.selectedIdx]
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
	nodes, err := ml.provider.GetNodes(ctx)
	if err != nil {
		ml.program.Send(refreshMsg{nodes: nil, err: err})
		return err
	}

	ml.program.Send(refreshMsg{nodes: nodes, err: nil})
	return nil
}

// SetRefreshEnabled enables or disables auto-refresh
func (ml *MainList) SetRefreshEnabled(enabled bool) {
	ml.refreshEnabled = enabled
}

// reinitializeClient creates a new Proxmox client with updated configuration
func (ml *MainList) reinitializeClient() {
	// Create new client with updated config
	newClient := proxmox.NewClient(
		ml.appConfig.APIUrl,
		ml.appConfig.GetAuthToken(),
		ml.appConfig.SkipTLSVerify,
	)

	// Update the provider and client
	ml.client = newClient
	ml.provider = newClient

	// Update refresh interval if it changed
	if ml.refreshTicker != nil && ml.appConfig.RefreshInterval > 0 {
		ml.refreshTicker.Reset(ml.appConfig.RefreshInterval)
	}
}

// refreshCmd returns a command that triggers a refresh
func (ml *MainList) refreshCmd() tea.Cmd {
	return func() tea.Msg {
		go ml.performRefresh()
		return nil
	}
}

// Stop stops the auto-refresh and program
func (ml *MainList) Stop() {
	if ml.refreshTicker != nil {
		ml.refreshTicker.Stop()
	}
	close(ml.stopRefresh)
	if ml.program != nil {
		ml.program.Quit()
	}
}

// Run starts the program
func (ml *MainList) Run() error {
	_, err := ml.program.Run()
	return err
}

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
