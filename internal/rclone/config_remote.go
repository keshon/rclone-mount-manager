package rclone

import (
	"fmt"
	"os/exec"
)

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
