package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/tsupplis/pvec/pkg/config"
	"github.com/tsupplis/pvec/pkg/models"
	"github.com/tsupplis/pvec/pkg/proxmox"
	"github.com/tsupplis/pvec/pkg/ui/mainlist"
)

// parseFlags handles command-line flags and returns the config file path
func parseFlags() string {
	var showVersion bool
	flag.BoolVar(&showVersion, "v", false, "Show version information")
	flag.BoolVar(&showVersion, "version", false, "Show version information")

	configPath := flag.String("c", "", "Path to configuration file")
	flag.StringVar(configPath, "config", "", "Path to configuration file")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: pvec [options]\n")
		fmt.Fprintf(os.Stderr, "A terminal-based interface for managing Proxmox VMs and Containers\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fmt.Fprintf(os.Stderr, "  -c, --config   Path to configuration file (default: ~/.pvecrc)\n")
		fmt.Fprintf(os.Stderr, "  -v, --version  Show version information\n")
		fmt.Fprintf(os.Stderr, "  -h, --help     Print this help\n")
	}

	flag.Parse()

	if showVersion {
		fmt.Println("pvec version 1.0.0")
		os.Exit(0)
	}

	return getConfigPath(*configPath)
}

// getConfigPath returns the configuration file path, using default if not provided
func getConfigPath(configPath string) string {
	if configPath != "" {
		return configPath
	}

	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to get home directory: %v", err)
	}
	return filepath.Join(home, ".pvecrc")
}

func main() {
	cfgPath := parseFlags()

	// Load configuration
	loader := config.NewLoader(cfgPath)
	cfg, err := loader.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration from %s: %v\nPlease create a .pvecrc file in your home directory or specify one with -c flag.", cfgPath, err)
	}

	// Create Proxmox client
	client := proxmox.NewClient(cfg.APIUrl, cfg.GetAuthToken(), cfg.SkipTLSVerify)

	// Create action executor
	executor := proxmox.NewActionExecutor(client)

	// Create main list with refresh interval from config
	listCfg := mainlist.Config{
		RefreshInterval: cfg.RefreshInterval,
		Provider:        client,
		Client:          client,
		AppConfig:       cfg,
		ConfigLoader:    loader,
		OnNodesUpdated: func(nodes []*models.VMStatus) {
			// Update executor cache when nodes are refreshed
			if ae, ok := executor.(*proxmox.ActionExecutor); ok {
				ae.UpdateNodes(nodes)
			}
		},
	}
	ml := mainlist.NewMainList(listCfg)
	defer ml.Stop()

	// Run Bubble Tea implementation (initial refresh happens in Init)
	if err := ml.Run(); err != nil {
		log.Fatalf("Error running application: %v", err)
	}
}
