// Package rclone provides management for rclone operations, including
// config parsing, mount management, progress tracking, daemon control,
// and filesystem operations.
package rclone

import (
	"context"
	"fmt"
	"time"
)

// Mount holds the configuration for mounting a remote as a local file system.
type Mount struct {
	Name       string // Remote name (e.g., "dropbox", "gdrive")
	MountPoint string // Local mount point
	CacheMode  string // VFS cache mode
	CacheSize  string // Cache size limit
	CacheAge   string // Max cache age
	CacheDir   string // Cache directory
	Status     string // Mount status
}

// MountInstance represents an active mount and its runtime state.
type MountInstance struct {
	Mount       Mount          // Configuration
	Status      string         // Current status
	Progress    *ProgressStats // Transfer statistics
	LastUpdated time.Time      // Last update time
}

// Mount creates a new mount using the given configuration.
func (m *Manager) Mount(ctx context.Context, mount Mount) error {
	m.mu.Lock()
	if _, exists := m.activeMounts[mount.Name]; exists {
		m.mu.Unlock()
		return fmt.Errorf("mount %s is already active", mount.Name)
	}
	m.mu.Unlock()

	req := map[string]interface{}{
		"fs":         mount.Name + ":",
		"mountPoint": mount.MountPoint,
		"vfsOpt": map[string]string{
			"CacheMode": mount.CacheMode,
		},
	}

	if err := m.callRC("mount/mount", req); err != nil {
		return fmt.Errorf("mount failed: %w", err)
	}

	m.mu.Lock()
	m.activeMounts[mount.Name] = &MountInstance{
		Mount:       mount,
		Status:      "mounted",
		Progress:    &ProgressStats{},
		LastUpdated: time.Now(),
	}
	m.mu.Unlock()

	return nil
}

// Unmount removes an active mount by name.
func (m *Manager) Unmount(name string) error {
	m.mu.Lock()
	inst, exists := m.activeMounts[name]
	m.mu.Unlock()
	if !exists {
		return fmt.Errorf("mount %s is not active", name)
	}

	req := map[string]interface{}{
		"mountPoint": inst.Mount.MountPoint,
	}
	if err := m.callRC("mount/unmount", req); err != nil {
		return err
	}

	m.mu.Lock()
	delete(m.activeMounts, name)
	m.mu.Unlock()

	return nil
}

// ListActiveMounts returns the names of all active mounts.
func (m *Manager) ListActiveMounts() []string {
	m.mu.Lock()
	defer m.mu.Unlock()

	names := make([]string, 0, len(m.activeMounts))
	for name := range m.activeMounts {
		names = append(names, name)
	}
	return names
}
