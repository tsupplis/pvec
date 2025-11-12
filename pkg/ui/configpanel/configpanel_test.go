package configpanel

import (
	"testing"
	"time"

	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tsupplis/pvec/pkg/config"
)

func TestNewConfigPanel(t *testing.T) {
	pages := tview.NewPages()
	cfg := &config.Config{
		APIUrl:          "https://test.example.com:8006",
		TokenID:         "test@pam!token",
		TokenSecret:     "secret123",
		RefreshInterval: 10 * time.Second,
		SkipTLSVerify:   true,
	}
	loader := config.NewLoader("")

	cp := NewConfigPanel(pages, cfg, loader)

	assert.NotNil(t, cp)
	assert.NotNil(t, cp.form)
	assert.Equal(t, pages, cp.pages)
	assert.Equal(t, cfg, cp.cfg)
}

func TestConfigPanel_Show(t *testing.T) {
	pages := tview.NewPages()
	cfg := &config.Config{
		APIUrl:          "https://test.example.com:8006",
		TokenID:         "test@pam!token",
		TokenSecret:     "secret123",
		RefreshInterval: 5 * time.Second,
		SkipTLSVerify:   false,
	}
	loader := config.NewLoader("")

	cp := NewConfigPanel(pages, cfg, loader)
	cp.Show()

	// Verify page was added
	assert.True(t, pages.HasPage("config"))
}

func TestConfigPanel_GetForm(t *testing.T) {
	pages := tview.NewPages()
	cfg := &config.Config{
		APIUrl:          "https://test.example.com:8006",
		TokenID:         "test@pam!token",
		TokenSecret:     "secret123",
		RefreshInterval: 5 * time.Second,
		SkipTLSVerify:   true,
	}
	loader := config.NewLoader("")

	cp := NewConfigPanel(pages, cfg, loader)

	form := cp.GetForm()
	require.NotNil(t, form)

	// Check form has expected fields (5 fields: API URL, Token ID, Token Secret, Refresh Interval, Skip TLS Verify)
	assert.Equal(t, 5, cp.form.GetFormItemCount())
}

func TestConfigPanel_HandleCancel(t *testing.T) {
	pages := tview.NewPages()
	cfg := &config.Config{
		APIUrl:          "https://test.example.com:8006",
		TokenID:         "test@pam!token",
		TokenSecret:     "secret123",
		RefreshInterval: 5 * time.Second,
		SkipTLSVerify:   true,
	}
	loader := config.NewLoader("")

	cp := NewConfigPanel(pages, cfg, loader)
	cp.Show()

	// Simulate cancel
	cp.handleCancel()

	// Verify page was removed
	assert.False(t, pages.HasPage("config"))
}

func TestConfigPanel_ValidationErrors(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *config.Config
		wantErr string
	}{
		{
			name: "empty API URL",
			cfg: &config.Config{
				APIUrl:          "",
				TokenID:         "test@pam!token",
				TokenSecret:     "secret123",
				RefreshInterval: 5 * time.Second,
			},
			wantErr: "API URL is required",
		},
		{
			name: "empty Token ID",
			cfg: &config.Config{
				APIUrl:          "https://test.example.com:8006",
				TokenID:         "",
				TokenSecret:     "secret123",
				RefreshInterval: 5 * time.Second,
			},
			wantErr: "Token ID is required",
		},
		{
			name: "empty Token Secret",
			cfg: &config.Config{
				APIUrl:          "https://test.example.com:8006",
				TokenID:         "test@pam!token",
				TokenSecret:     "",
				RefreshInterval: 5 * time.Second,
			},
			wantErr: "Token Secret is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pages := tview.NewPages()
			loader := config.NewLoader("")
			cp := NewConfigPanel(pages, tt.cfg, loader)

			cp.handleSave()

			// Verify error page was shown
			assert.True(t, pages.HasPage("error"))
		})
	}
}
