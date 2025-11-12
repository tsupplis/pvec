package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_GetAuthToken(t *testing.T) {
	cfg := &Config{
		TokenID:     "user@pam!token",
		TokenSecret: "secret-uuid-here",
	}

	token := cfg.GetAuthToken()
	assert.Equal(t, "PVEAPIToken=user@pam!token=secret-uuid-here", token)
}

func TestViperLoader_Load_Success(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.json")

	configContent := `{
  "api_url": "https://proxmox.example.com:8006",
  "token_id": "user@pam!token",
  "token_secret": "secret-uuid",
  "refresh_interval": "10s",
  "skip_tls_verify": true
}`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Load config
	loader := NewLoader(configPath)
	cfg, err := loader.Load()

	require.NoError(t, err)
	assert.Equal(t, "https://proxmox.example.com:8006", cfg.APIUrl)
	assert.Equal(t, "user@pam!token", cfg.TokenID)
	assert.Equal(t, "secret-uuid", cfg.TokenSecret)
	assert.Equal(t, 10*time.Second, cfg.RefreshInterval)
	assert.True(t, cfg.SkipTLSVerify)
}

func TestViperLoader_Load_Defaults(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.json")

	// Minimal config (only required fields)
	configContent := `{
  "api_url": "https://proxmox.example.com:8006",
  "token_id": "user@pam!token",
  "token_secret": "secret-uuid"
}`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	loader := NewLoader(configPath)
	cfg, err := loader.Load()

	require.NoError(t, err)
	assert.Equal(t, 5*time.Second, cfg.RefreshInterval) // Default value
	assert.True(t, cfg.SkipTLSVerify)                   // Default value (changed to true)
}

func TestViperLoader_Load_MissingAPIUrl(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.json")

	configContent := `{
  "token_id": "user@pam!token",
  "token_secret": "secret-uuid"
}`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	loader := NewLoader(configPath)
	_, err = loader.Load()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "api_url is required")
}

func TestViperLoader_Load_MissingTokenID(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.json")

	configContent := `{
  "api_url": "https://proxmox.example.com:8006",
  "token_secret": "secret-uuid"
}`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	loader := NewLoader(configPath)
	_, err = loader.Load()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "token_id is required")
}

func TestViperLoader_Load_MissingTokenSecret(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.json")

	configContent := `{
  "api_url": "https://proxmox.example.com:8006",
  "token_id": "user@pam!token"
}`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	loader := NewLoader(configPath)
	_, err = loader.Load()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "token_secret is required")
}

func TestViperLoader_Load_FileNotFound(t *testing.T) {
	loader := NewLoader("/nonexistent/path/config.yaml")
	_, err := loader.Load()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read config file")
}

func TestViperLoader_Save(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.json")

	// Create initial config
	initialContent := `{
  "api_url": "https://proxmox.example.com:8006",
  "token_id": "user@pam!token",
  "token_secret": "secret-uuid"
}`
	err := os.WriteFile(configPath, []byte(initialContent), 0644)
	require.NoError(t, err)

	// Load and modify
	loader := NewLoader(configPath).(*ViperLoader)
	cfg := &Config{
		APIUrl:          "https://new-proxmox.com:8006",
		TokenID:         "newuser@pam!newtoken",
		TokenSecret:     "new-secret",
		RefreshInterval: 15 * time.Second,
		SkipTLSVerify:   true,
	}

	err = loader.Save(cfg)
	require.NoError(t, err)

	// Reload and verify
	cfg2, err := loader.Load()
	require.NoError(t, err)
	assert.Equal(t, "https://new-proxmox.com:8006", cfg2.APIUrl)
	assert.Equal(t, "newuser@pam!newtoken", cfg2.TokenID)
	assert.Equal(t, "new-secret", cfg2.TokenSecret)
	assert.Equal(t, 15*time.Second, cfg2.RefreshInterval)
	assert.True(t, cfg2.SkipTLSVerify)
}

func TestNewLoader_DefaultPath(t *testing.T) {
	loader := NewLoader("")
	viperLoader, ok := loader.(*ViperLoader)
	require.True(t, ok)

	// Should use home directory
	home, _ := os.UserHomeDir()
	expectedPath := filepath.Join(home, ".pvecrc")
	assert.Equal(t, expectedPath, viperLoader.configPath)
}

func TestNewLoader_CustomPath(t *testing.T) {
	customPath := "/custom/path/config.yaml"
	loader := NewLoader(customPath)
	viperLoader, ok := loader.(*ViperLoader)
	require.True(t, ok)

	assert.Equal(t, customPath, viperLoader.configPath)
}
