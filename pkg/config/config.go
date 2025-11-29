package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

// Config holds the application configuration
type Config struct {
	APIUrl          string        `mapstructure:"api_url"`
	TokenID         string        `mapstructure:"token_id"`
	TokenSecret     string        `mapstructure:"token_secret"`
	RefreshInterval time.Duration `mapstructure:"refresh_interval"`
	SkipTLSVerify   bool          `mapstructure:"skip_tls_verify"`
}

// Loader is the interface for loading configuration
type Loader interface {
	Load() (*Config, error)
}

// ViperLoader loads configuration using Viper
type ViperLoader struct {
	configPath string
}

// NewLoader creates a new configuration loader
// If configPath is empty, it defaults to ~/.pvecrc
func NewLoader(configPath string) Loader {
	if configPath == "" {
		home, err := os.UserHomeDir()
		if err == nil {
			configPath = filepath.Join(home, ".pvecrc")
		}
	}
	return &ViperLoader{configPath: configPath}
}

// Load reads and parses the configuration file
func (l *ViperLoader) Load() (*Config, error) {
	v := viper.New()

	// Set defaults
	v.SetDefault("refresh_interval", "5s")
	v.SetDefault("skip_tls_verify", true)

	// Set config file path
	if l.configPath != "" {
		v.SetConfigFile(l.configPath)
		v.SetConfigType("json") // Explicitly set JSON type for .pvecrc files
	} else {
		return nil, fmt.Errorf("config path not set")
	}

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse into struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Validate required fields
	if cfg.APIUrl == "" {
		return nil, fmt.Errorf("api_url is required")
	}
	if cfg.TokenID == "" {
		return nil, fmt.Errorf("token_id is required")
	}
	if cfg.TokenSecret == "" {
		return nil, fmt.Errorf("token_secret is required")
	}

	return &cfg, nil
}

// Save writes the configuration back to file
func (l *ViperLoader) Save(cfg *Config) error {
	v := viper.New()
	v.SetConfigFile(l.configPath)
	v.SetConfigType("json")

	v.Set("api_url", cfg.APIUrl)
	v.Set("token_id", cfg.TokenID)
	v.Set("token_secret", cfg.TokenSecret)
	v.Set("refresh_interval", cfg.RefreshInterval.String())
	v.Set("skip_tls_verify", cfg.SkipTLSVerify)

	return v.WriteConfig()
}

// GetAuthToken returns the formatted authentication token
func (c *Config) GetAuthToken() string {
	return fmt.Sprintf("PVEAPIToken=%s=%s", c.TokenID, c.TokenSecret)
}
