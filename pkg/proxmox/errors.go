package proxmox

import "errors"

var (
	// ErrNodeNotFound is returned when a VM/CT is not found in the cache
	ErrNodeNotFound = errors.New("node not found in cache")
)
