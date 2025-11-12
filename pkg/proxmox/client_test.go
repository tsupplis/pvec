package proxmox

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tsupplis/pvec/pkg/models"
)

func setupMockServer(_ *testing.T) (*httptest.Server, *HTTPClient) {
	mux := http.NewServeMux()

	// Mock nodes endpoint
	mux.HandleFunc("/api2/json/nodes", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		resp := map[string]interface{}{
			"data": []map[string]interface{}{
				{"node": "pve1"},
				{"node": "pve2"},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})

	// Mock QEMU VMs endpoint
	mux.HandleFunc("/api2/json/nodes/pve1/qemu", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"vmid":   100,
					"name":   "test-vm",
					"type":   "qemu",
					"status": "running",
					"cpu":    0.25,
					"mem":    2147483648,
					"maxmem": 4294967296,
					"maxcpu": 2,
					"uptime": 3600,
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("/api2/json/nodes/pve2/qemu", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"data": []map[string]interface{}{},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}) // Mock LXC containers endpoint
	mux.HandleFunc("/api2/json/nodes/pve1/lxc", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"vmid":   200,
					"name":   "test-ct",
					"type":   "lxc",
					"status": "stopped",
					"cpu":    0.0,
					"mem":    0,
					"maxmem": 1073741824,
					"maxcpu": 1,
					"uptime": 0,
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("/api2/json/nodes/pve2/lxc", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"data": []map[string]interface{}{},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})

	// Mock start endpoint
	mux.HandleFunc("/api2/json/nodes/pve1/qemu/100/status/start", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		resp := map[string]interface{}{
			"data": "OK",
		}
		_ = json.NewEncoder(w).Encode(resp)
	})

	// Mock shutdown endpoint
	mux.HandleFunc("/api2/json/nodes/pve1/qemu/100/status/shutdown", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		resp := map[string]interface{}{
			"data": "OK",
		}
		_ = json.NewEncoder(w).Encode(resp)
	})

	// Mock reboot endpoint
	mux.HandleFunc("/api2/json/nodes/pve1/qemu/100/status/reboot", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		resp := map[string]interface{}{
			"data": "OK",
		}
		_ = json.NewEncoder(w).Encode(resp)
	})

	// Mock stop endpoint
	mux.HandleFunc("/api2/json/nodes/pve1/qemu/100/status/stop", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		resp := map[string]interface{}{
			"data": "OK",
		}
		_ = json.NewEncoder(w).Encode(resp)
	})

	server := httptest.NewServer(mux)
	client := NewClient(server.URL, "test-token", true).(*HTTPClient)

	return server, client
}

func TestHTTPClient_GetNodes(t *testing.T) {
	server, client := setupMockServer(t)
	defer server.Close()

	nodes, err := client.GetNodes(context.Background())
	require.NoError(t, err)

	// Should have 2 resources (1 VM + 1 CT)
	assert.Len(t, nodes, 2)

	// Check VM
	vm := findNodeByID(nodes, "100")
	require.NotNil(t, vm)
	assert.Equal(t, "test-vm", vm.Name)
	assert.Equal(t, string(models.TypeVM), vm.Type)
	assert.Equal(t, string(models.StateRunning), vm.Status)
	assert.Equal(t, "pve1", vm.Node)
	assert.Equal(t, 25.0, vm.CPUUsage)
	assert.InDelta(t, 50.0, vm.MemoryUsage, 0.1)

	// Check Container
	ct := findNodeByID(nodes, "200")
	require.NotNil(t, ct)
	assert.Equal(t, "test-ct", ct.Name)
	assert.Equal(t, string(models.TypeContainer), ct.Type)
	assert.Equal(t, string(models.StateStopped), ct.Status)
	assert.Equal(t, "pve1", ct.Node)
}

func TestHTTPClient_Start(t *testing.T) {
	server, client := setupMockServer(t)
	defer server.Close()

	err := client.Start(context.Background(), "pve1", "qemu", "100")
	assert.NoError(t, err)
}

func TestHTTPClient_Shutdown(t *testing.T) {
	server, client := setupMockServer(t)
	defer server.Close()

	err := client.Shutdown(context.Background(), "pve1", "qemu", "100")
	assert.NoError(t, err)
}

func TestHTTPClient_Reboot(t *testing.T) {
	server, client := setupMockServer(t)
	defer server.Close()

	err := client.Reboot(context.Background(), "pve1", "qemu", "100")
	assert.NoError(t, err)
}

func TestHTTPClient_Stop(t *testing.T) {
	server, client := setupMockServer(t)
	defer server.Close()

	err := client.Stop(context.Background(), "pve1", "qemu", "100")
	assert.NoError(t, err)
}

func TestHTTPClient_GetNodes_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	client := NewClient(server.URL, "invalid-token", true)
	_, err := client.GetNodes(context.Background())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "401")
}

func TestHTTPClient_GetNodes_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", true)
	_, err := client.GetNodes(context.Background())

	assert.Error(t, err)
}

func TestHTTPClient_ContextCancellation(t *testing.T) {
	server, client := setupMockServer(t)
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := client.GetNodes(ctx)
	assert.Error(t, err)
}

func TestNewClient_TLSConfig(t *testing.T) {
	client := NewClient("https://example.com", "token", true)
	httpClient := client.(*HTTPClient)

	transport := httpClient.httpClient.Transport.(*http.Transport)
	assert.True(t, transport.TLSClientConfig.InsecureSkipVerify)

	client2 := NewClient("https://example.com", "token", false)
	httpClient2 := client2.(*HTTPClient)

	transport2 := httpClient2.httpClient.Transport.(*http.Transport)
	assert.False(t, transport2.TLSClientConfig.InsecureSkipVerify)
}

// Helper function to find a node by VMID
func findNodeByID(nodes []*models.VMStatus, vmid string) *models.VMStatus {
	for _, node := range nodes {
		if node.VMID == vmid {
			return node
		}
	}
	return nil
}
