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

// clusterResource represents a resource from the cluster/resources endpoint
type clusterResource struct {
	ID        string      `json:"id"`
	VMID      json.Number `json:"vmid"`
	Name      string      `json:"name"`
	Type      string      `json:"type"`
	Status    string      `json:"status"`
	Node      string      `json:"node"`
	CPU       float64     `json:"cpu"`
	Mem       int64       `json:"mem"`
	MaxMem    int64       `json:"maxmem"`
	MaxCPU    int         `json:"maxcpu"`
	Uptime    int64       `json:"uptime"`
	DiskRead  int64       `json:"diskread"`
	DiskWrite int64       `json:"diskwrite"`
}

// GetNodes retrieves all VMs and Containers from all nodes using cluster resources
func (c *HTTPClient) GetNodes(ctx context.Context) ([]*models.VMStatus, error) {
	// Use cluster/resources endpoint for more accurate data
	resp, err := c.doRequest(ctx, "GET", "/cluster/resources", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get cluster resources: status %d", resp.StatusCode)
	}

	var resourcesResp proxmoxResponse
	if err := json.NewDecoder(resp.Body).Decode(&resourcesResp); err != nil {
		return nil, fmt.Errorf("failed to decode cluster resources response: %w", err)
	}

	var resources []clusterResource
	if err := json.Unmarshal(resourcesResp.Data, &resources); err != nil {
		return nil, fmt.Errorf("failed to parse cluster resources data: %w", err)
	}

	var allVMs []*models.VMStatus
	for _, resource := range resources {
		if resource.Type == "qemu" || resource.Type == "lxc" {
			vmStatus := c.createVMStatusFromClusterResource(resource)
			allVMs = append(allVMs, vmStatus)
		}
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

// createVMStatusFromClusterResource creates a VMStatus from a cluster resource
func (c *HTTPClient) createVMStatusFromClusterResource(res clusterResource) *models.VMStatus {
	vmid := res.VMID.String()

	nodeType := models.TypeVM
	if res.Type == "lxc" {
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
		Node:        res.Node,
		CPUUsage:    cpuPercent,
		MemoryUsage: memPercent,
		MaxMem:      res.MaxMem,
		MaxCPU:      res.MaxCPU,
		Uptime:      res.Uptime,
	}
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
