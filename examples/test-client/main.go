package main

import (
	"context"
	"fmt"
	"log"

	"github.com/tsupplis/pvec/pkg/config"
	"github.com/tsupplis/pvec/pkg/models"
	"github.com/tsupplis/pvec/pkg/proxmox"
)

func main() {
	// Create configuration loader and load config
	loader := config.NewLoader("")
	cfg, err := loader.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create Proxmox client
	authToken := fmt.Sprintf("%s=%s", cfg.TokenID, cfg.TokenSecret)
	client := proxmox.NewClient(cfg.APIUrl, authToken, cfg.SkipTLSVerify)

	// Test connection and get VM/CT statuses
	ctx := context.Background()
	vms, err := client.GetNodes(ctx)
	if err != nil {
		log.Fatalf("Failed to get VMs/CTs: %v", err)
	}

	// Print VM/CT information
	fmt.Printf("Connected to Proxmox at %s\n", cfg.APIUrl)
	fmt.Printf("Found %d VMs/Containers:\n\n", len(vms))

	if len(vms) == 0 {
		fmt.Println("No VMs or Containers found.")
		return
	}

	// Group by node
	nodeMap := make(map[string][]*models.VMStatus)
	for _, vm := range vms {
		nodeMap[vm.Node] = append(nodeMap[vm.Node], vm)
	}

	// Print organized by node
	for nodeName, nodeVMs := range nodeMap {
		fmt.Printf("Node: %s (%d VMs/CTs)\n", nodeName, len(nodeVMs))
		for _, vm := range nodeVMs {
			fmt.Printf("  %s (%s) - %s [%s] - CPU: %.2f%% Memory: %.2f%%\n",
				vm.Name, vm.VMID, vm.Status, vm.Type, vm.CPUUsage, vm.MemoryUsage)
		}
		fmt.Println()
	}

	fmt.Printf("Test completed successfully!\n")
}
