package rclone

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"rclone-gui/internal/utils"
	"strings"
	"syscall"
)

func (m *Manager) Mount(ctx context.Context, mount Mount, progress chan string) error {
	m.mu.Lock()
	if _, exists := m.activeMounts[mount.Name]; exists {
		m.mu.Unlock()
		return fmt.Errorf("mount %s is already active", mount.Name)
	}
	// create cancelable context
	ctx, cancel := context.WithCancel(ctx)
	m.activeMounts[mount.Name] = &MountInstance{
		Mount:  mount,
		Ctx:    cancel,
		Status: "mounting",
	}
	m.mu.Unlock()

	// Ensure logs folder exists
	logDir := filepath.Join("logs")
	_ = os.MkdirAll(logDir, 0755)

	logPath := filepath.Join(logDir, mount.Name+".log")
	logger := utils.FileLogger(logPath)

	args := []string{
		"--config", mount.LogPath,
		"mount", mount.Name + ":",
		mount.MountPoint,
		"--vfs-cache-mode", mount.CacheMode,
		"--vfs-cache-max-size", mount.CacheSize,
		"--vfs-cache-max-age", mount.CacheAge,
		"--cache-dir", mount.CacheDir,
		"--vfs-read-chunk-size", "64M",
		"--vfs-read-chunk-size-limit", "off",
		"--vfs-write-back", "0s",
		"--dir-cache-time", "1m",
		"--poll-interval", "15s",
		"--retries", "10",
		"--low-level-retries", "20",
		"--timeout", "1m",
		"--progress",
		"--stats", "10s",
		"-v",
	}

	cmd := exec.CommandContext(ctx, "rclone/rclone.exe", args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	capture := func(pipe *bufio.Scanner) {
		var buf bytes.Buffer
		for pipe.Scan() {
			line := pipe.Text()
			buf.WriteString(line + "\n")
			if strings.TrimSpace(line) == "" {
				msg := buf.String()
				progress <- msg
				logger.Print(msg)
				buf.Reset()
			}
		}
		if buf.Len() > 0 {
			msg := buf.String()
			progress <- msg
			logger.Print(msg)
		}
	}

	scannerOut := bufio.NewScanner(stdout)
	scannerErr := bufio.NewScanner(stderr)

	go capture(scannerOut)
	go capture(scannerErr)

	err = cmd.Wait()
	if err != nil {
		logger.Printf("rclone exited with error: %v\n", err)
		return fmt.Errorf("rclone exited: %w", err)
	}

	m.mu.Lock()
	delete(m.activeMounts, mount.Name)
	m.mu.Unlock()
	return nil
}

func (m *Manager) Unmount(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	inst, exists := m.activeMounts[name]
	if !exists {
		return fmt.Errorf("mount %s is not active", name)
	}
	inst.Ctx() // cancel context
	delete(m.activeMounts, name)
	return nil
}

func (m *Manager) ListActiveMounts() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	names := make([]string, 0, len(m.activeMounts))
	for name := range m.activeMounts {
		names = append(names, name)
	}
	return names
}
