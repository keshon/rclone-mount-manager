// Package rclone provides management for rclone operations, including
// config parsing, mount management, progress tracking, daemon control,
// and filesystem operations.
package rclone

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"syscall"
	"time"
)

// startRclone starts the rclone daemon (rcd) with HTTP remote control.
// It binds to localhost on the configured port and runs with verbose
// logging. The process is cancellable via context.
func (m *Manager) startRclone() error {
	ctx, cancel := context.WithCancel(context.Background())
	m.rcCtx = cancel

	args := []string{
		"rcd",
		"--rc-no-auth",
		"--rc-addr", fmt.Sprintf("127.0.0.1:%d", m.rcPort),
		"--config", m.ConfigPath,
		"-v",
	}

	cmd := exec.CommandContext(ctx, "rclone/rclone.exe", args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start rclone rcd: %w", err)
	}

	m.rcProcess = cmd
	time.Sleep(time.Second) // allow startup
	return nil
}

// StopRclone stops the rclone daemon if running.
// It cancels the context and waits for the process.
func (m *Manager) StopRclone() {
	if m.rcCtx != nil {
		m.rcCtx()
	}
	if m.rcProcess != nil {
		_ = m.rcProcess.Wait()
	}
}

// callRC sends a command to the rclone daemon via its HTTP API.
// It posts a JSON body to the given endpoint and checks for errors.
func (m *Manager) callRC(command string, body map[string]interface{}) error {
	url := fmt.Sprintf("http://127.0.0.1:%d/%s", m.rcPort, command)
	data, _ := json.Marshal(body)

	resp, err := http.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("rc call failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("rc call %s failed: %s", command, resp.Status)
	}
	return nil
}
