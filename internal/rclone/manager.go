// Package rclone provides management for rclone operations, including
// config parsing, mount management, progress tracking, daemon control,
// and filesystem operations.
package rclone

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sync"
)

// Manager manages rclone operations and state.
type Manager struct {
	ConfigPath   string                    // Path to rclone config
	Remotes      []Remote                  // Configured remotes
	activeMounts map[string]*MountInstance // Active mounts by name
	rcProcess    *exec.Cmd                 // rclone daemon process
	rcCtx        context.CancelFunc        // Daemon cancel func
	rcPort       int                       // rclone RC port
	mu           sync.Mutex                // Protects manager state

	progressLog []ProgressStats // Transfer stats history
	logFile     string          // Log file path
	maxLogItems int             // Max log items in memory
}

// NewManager creates a Manager with the given config and log paths.
// It validates the config file, loads remotes, and starts the rclone daemon.
func NewManager(configPath, logFile string) (*Manager, error) {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, errors.New("rclone config not found at " + configPath)
	}

	m := &Manager{
		ConfigPath:   configPath,
		activeMounts: make(map[string]*MountInstance),
		rcPort:       5572,
		logFile:      logFile,
	}

	if err := m.loadRemotes(); err != nil {
		return nil, err
	}
	if err := m.startRclone(); err != nil {
		return nil, err
	}

	return m, nil
}

// ConfigRemote opens rclone's interactive config tool in a Windows cmd window.
func (m *Manager) ConfigRemote() error {
	if m.ConfigPath == "" {
		return fmt.Errorf("ConfigPath is not set")
	}

	cmd := exec.Command(
		"cmd", "/C", "start", "cmd", "/K",
		".\\rclone\\rclone.exe", "config", "--config", m.ConfigPath,
	)

	return cmd.Start()
}
