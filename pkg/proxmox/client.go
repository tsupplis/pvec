package proxmox

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/tsupplis/pvec/pkg/models"
)

// Client is the interface for Proxmox API operations
type Client interface {
	// GetNodes retrieves all VMs and Containers
	GetNodes(ctx context.Context) ([]*models.VMStatus, error)
	// GetVMConfig retrieves detailed configuration for a VM or Container
	GetVMConfig(ctx context.Context, node, vmType, vmid string) (map[string]interface{}, error)
	// Start starts a VM or Container
	Start(ctx context.Context, node, vmType, vmid string) error
	// Shutdown gracefully shuts down a VM or Container
	Shutdown(ctx context.Context, node, vmType, vmid string) error
	// Reboot reboots a VM or Container
	Reboot(ctx context.Context, node, vmType, vmid string) error
	// Stop forcefully stops a VM or Container
	Stop(ctx context.Context, node, vmType, vmid string) error
}

// HTTPClient is the HTTP implementation of the Proxmox client
type HTTPClient struct {
	baseURL    string
	authToken  string
	httpClient *http.Client
}

// NewClient creates a new Proxmox HTTP client
func NewClient(baseURL, authToken string, skipTLSVerify bool) Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			// #nosec G402 - InsecureSkipVerify is intentional for Proxmox self-signed certificates
			// This is a user-configurable option and users are warned in documentation
			InsecureSkipVerify: skipTLSVerify,
		},
	}

	return &HTTPClient{
		baseURL:   baseURL,
		authToken: authToken,
		httpClient: &http.Client{
			Timeout:   30 * time.Second,
			Transport: transport,
		},
	}
}

// doRequest performs an HTTP request with authentication
func (c *HTTPClient) doRequest(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	url := fmt.Sprintf("%s/api2/json%s", c.baseURL, path)

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", c.authToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

// doRequestForm performs an HTTP request with form-encoded content type
func (c *HTTPClient) doRequestForm(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	url := fmt.Sprintf("%s/api2/json%s", c.baseURL, path)

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", c.authToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

// proxmoxResponse represents a generic Proxmox API response
type proxmoxResponse struct {
	Data json.RawMessage `json:"data"`
}

// nodeResource represents a node in the cluster
type nodeResource struct {
	Node string `json:"node"`
}

// vmResource represents a VM or Container resource
type vmResource struct {
	VMID   json.Number `json:"vmid"`
	Name   string      `json:"name"`
	Type   string      `json:"type"`
	Status string      `json:"status"`
	CPU    float64     `json:"cpu"`
	Mem    int64       `json:"mem"`
	MaxMem int64       `json:"maxmem"`
	MaxCPU int         `json:"maxcpu"`
	Uptime int64       `json:"uptime"`
}

// GetNodes retrieves all VMs and Containers from all nodes
func (c *HTTPClient) GetNodes(ctx context.Context) ([]*models.VMStatus, error) {
	// First, get list of nodes
	resp, err := c.doRequest(ctx, "GET", "/nodes", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get nodes: status %d", resp.StatusCode)
	}

	var nodesResp proxmoxResponse
	if err := json.NewDecoder(resp.Body).Decode(&nodesResp); err != nil {
		return nil, fmt.Errorf("failed to decode nodes response: %w", err)
	}

	var nodes []nodeResource
	if err := json.Unmarshal(nodesResp.Data, &nodes); err != nil {
		return nil, fmt.Errorf("failed to parse nodes data: %w", err)
	}

	// Collect VMs and Containers from all nodes
	var allVMs []*models.VMStatus
	for _, node := range nodes {
		vms, err := c.getNodeVMs(ctx, node.Node)
		if err != nil {
			// Log error but continue with other nodes
			continue
		}
		allVMs = append(allVMs, vms...)
	}

	return allVMs, nil
}

// GetVMConfig retrieves detailed configuration for a VM or Container
func (c *HTTPClient) GetVMConfig(ctx context.Context, node, vmType, vmid string) (map[string]interface{}, error) {
	path := fmt.Sprintf("/nodes/%s/%s/%s/config", node, vmType, vmid)
	resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get config for %s %s (status %d): %s", vmType, vmid, resp.StatusCode, string(body))
	}

	var apiResp proxmoxResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(apiResp.Data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return config, nil
}

// getNodeVMs retrieves VMs and Containers from a specific node
func (c *HTTPClient) getNodeVMs(ctx context.Context, nodeName string) ([]*models.VMStatus, error) {
	// Get QEMU VMs
	qemuVMs, _ := c.getResources(ctx, nodeName, "qemu")
	// Get LXC Containers
	lxcVMs, _ := c.getResources(ctx, nodeName, "lxc")

	return append(qemuVMs, lxcVMs...), nil
}

// getResources retrieves resources of a specific type (qemu or lxc)
func (c *HTTPClient) getResources(ctx context.Context, nodeName, resType string) ([]*models.VMStatus, error) {
	path := fmt.Sprintf("/nodes/%s/%s", nodeName, resType)
	resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get %s resources: status %d", resType, resp.StatusCode)
	}

	// Parse the API response
	resources, err := c.parseResourcesResponse(resp)
	if err != nil {
		return nil, err
	}

	// Convert to our model
	return c.convertResourcesToVMStatus(resources, resType, nodeName), nil
}

// parseResourcesResponse parses the API response into vmResource slice
func (c *HTTPClient) parseResourcesResponse(resp *http.Response) ([]vmResource, error) {
	var apiResp proxmoxResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var resources []vmResource
	if err := json.Unmarshal(apiResp.Data, &resources); err != nil {
		return nil, fmt.Errorf("failed to parse resources: %w", err)
	}

	return resources, nil
}

// convertResourcesToVMStatus converts vmResource slice to VMStatus slice
func (c *HTTPClient) convertResourcesToVMStatus(resources []vmResource, resType, nodeName string) []*models.VMStatus {
	result := make([]*models.VMStatus, 0, len(resources))
	for _, res := range resources {
		vmStatus := c.createVMStatusFromResource(res, resType, nodeName)
		result = append(result, vmStatus)
	}
	return result
}

// createVMStatusFromResource creates a VMStatus from a single vmResource
func (c *HTTPClient) createVMStatusFromResource(res vmResource, resType, nodeName string) *models.VMStatus {
	vmid := res.VMID.String()

	nodeType := models.TypeVM
	if resType == "lxc" {
		nodeType = models.TypeContainer
	}

	status := c.mapResourceStatus(res.Status)
	cpuPercent := res.CPU * 100
	memPercent := c.calculateMemoryPercentage(res.Mem, res.MaxMem)

	return &models.VMStatus{
		VMID:        vmid,
		Name:        res.Name,
		Type:        string(nodeType),
		Status:      string(status),
		Node:        nodeName,
		CPUUsage:    cpuPercent,
		MemoryUsage: memPercent,
		MaxMem:      res.MaxMem,
		MaxCPU:      res.MaxCPU,
		Uptime:      res.Uptime,
	}
}

// mapResourceStatus maps Proxmox status strings to our model states
func (c *HTTPClient) mapResourceStatus(status string) models.NodeState {
	switch status {
	case "running":
		return models.StateRunning
	case "stopped":
		return models.StateStopped
	case "paused":
		return models.StatePaused
	default:
		return models.StateUnknown
	}
}

// calculateMemoryPercentage calculates memory usage percentage
func (c *HTTPClient) calculateMemoryPercentage(mem, maxMem int64) float64 {
	if maxMem > 0 {
		return (float64(mem) / float64(maxMem)) * 100
	}
	return 0.0
}

// Start starts a VM or Container
func (c *HTTPClient) Start(ctx context.Context, node, vmType, vmid string) error {
	path := fmt.Sprintf("/nodes/%s/%s/%s/status/start", node, vmType, vmid)

	// Proxmox API expects form-encoded data for POST requests
	body := strings.NewReader("")
	resp, err := c.doRequestForm(ctx, "POST", path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to start %s %s (status %d): %s", vmType, vmid, resp.StatusCode, string(body))
	}

	return nil
}

// Shutdown gracefully shuts down a VM or Container
func (c *HTTPClient) Shutdown(ctx context.Context, node, vmType, vmid string) error {
	path := fmt.Sprintf("/nodes/%s/%s/%s/status/shutdown", node, vmType, vmid)

	body := strings.NewReader("")
	resp, err := c.doRequestForm(ctx, "POST", path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to shutdown %s %s (status %d): %s", vmType, vmid, resp.StatusCode, string(body))
	}

	return nil
}

// Reboot reboots a VM or Container
func (c *HTTPClient) Reboot(ctx context.Context, node, vmType, vmid string) error {
	path := fmt.Sprintf("/nodes/%s/%s/%s/status/reboot", node, vmType, vmid)

	body := strings.NewReader("")
	resp, err := c.doRequestForm(ctx, "POST", path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to reboot %s %s (status %d): %s", vmType, vmid, resp.StatusCode, string(body))
	}

	return nil
}

// Stop forcefully stops a VM or Container
func (c *HTTPClient) Stop(ctx context.Context, node, vmType, vmid string) error {
	path := fmt.Sprintf("/nodes/%s/%s/%s/status/stop", node, vmType, vmid)

	body := strings.NewReader("")
	resp, err := c.doRequestForm(ctx, "POST", path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to stop %s %s (status %d): %s", vmType, vmid, resp.StatusCode, string(body))
	}

	return nil
}
