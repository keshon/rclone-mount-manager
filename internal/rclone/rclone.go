package rclone

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"gopkg.in/ini.v1"
)

// -------------------- Manager --------------------

type Manager struct {
	ConfigPath   string
	Remotes      []Remote
	activeMounts map[string]*MountInstance
	rcProcess    *exec.Cmd
	rcCtx        context.CancelFunc
	rcPort       int
	mu           sync.Mutex

	progressLog []ProgressStats // history of snapshots
	logFile     string          // path to JSON log file
	maxLogItems int             // optional cap on log length (0 = unlimited)
}

// MountInstance represents an active mount
type MountInstance struct {
	Mount       Mount
	Status      string
	Progress    *ProgressStats
	LastUpdated time.Time
}

// -------------------- Progress Types --------------------

type ProgressStats struct {
	Transferring []TransferInfo `json:"transferring,omitempty"`
	Checking     []TransferInfo `json:"checking,omitempty"`

	Bytes     int64   `json:"bytes"`
	Checks    int64   `json:"checks"`
	Deletes   int64   `json:"deletes"`
	Errors    int64   `json:"errors"`
	Transfers int64   `json:"transfers"`
	Speed     float64 `json:"speed"`

	CheckSpeed    float64 `json:"checkSpeed"`
	TransferSpeed float64 `json:"transferSpeed"`

	ETA         *int64  `json:"eta,omitempty"`
	ElapsedTime float64 `json:"elapsedTime"`

	ServerSideOperations int64 `json:"serverSideOperations,omitempty"`

	Timestamp time.Time `json:"timestamp"`
}

type TransferInfo struct {
	Name       string  `json:"name"`
	Size       int64   `json:"size"`
	Bytes      int64   `json:"bytes"`
	Checked    bool    `json:"checked"`
	Percentage float64 `json:"percentage"`
	Speed      float64 `json:"speed"`
	ETA        *int64  `json:"eta,omitempty"`
}

// -------------------- Init --------------------

func NewManager(configPath, logFile string) (*Manager, error) {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, errors.New("rclone config not found at " + configPath)
	}

	m := &Manager{
		ConfigPath:   configPath,
		activeMounts: make(map[string]*MountInstance),
		rcPort:       5572,
		logFile:      logFile,
		maxLogItems:  0, // 0 = unlimited; set to e.g. 1000 to cap
	}

	if err := m.loadRemotes(); err != nil {
		return nil, err
	}

	if err := m.startRclone(); err != nil {
		return nil, err
	}

	return m, nil
}

func (m *Manager) loadRemotes() error {
	cfg, err := ini.Load(m.ConfigPath)
	if err != nil {
		return fmt.Errorf("parsing config failed: %w", err)
	}

	m.Remotes = nil
	for _, section := range cfg.Sections() {
		name := section.Name()
		if name == "DEFAULT" {
			continue
		}

		t := section.Key("type").String()
		if t == "" {
			continue
		}

		r := Remote{
			Name: name,
			Type: t,
		}

		if t == "union" {
			r.Upstreams = splitUpstreams(section.Key("upstreams").String())
		}

		m.Remotes = append(m.Remotes, r)
	}
	return nil
}

// -------------------- Rclone RCD --------------------

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
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start rclone rcd: %w", err)
	}

	m.rcProcess = cmd
	time.Sleep(1 * time.Second) // wait for RC to be ready
	return nil
}

func (m *Manager) StopRclone() {
	if m.rcCtx != nil {
		m.rcCtx()
	}
	if m.rcProcess != nil {
		_ = m.rcProcess.Wait()
	}
}

// -------------------- Progress --------------------

// UpdateProgress pulls /core/stats via POST and appends a snapshot to progressLog.
// Returns an error if HTTP/JSON/write steps fail.
func (m *Manager) UpdateProgress() error {
	url := fmt.Sprintf("http://127.0.0.1:%d/core/stats", m.rcPort)

	// rclone expects a POST (even if empty body)
	body := map[string]interface{}{}
	payload, _ := json.Marshal(body)

	resp, err := http.Post(url, "application/json", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("http post failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("stats returned %s", resp.Status)
	}

	var rawStats map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&rawStats); err != nil {
		return fmt.Errorf("decode stats failed: %w", err)
	}

	stats := m.parseStatsResponse(rawStats)
	stats.Timestamp = time.Now()

	m.mu.Lock()

	log.Println("[DEBUG] stats:", stats)

	// append and optionally cap
	m.progressLog = append(m.progressLog, *stats)
	if m.maxLogItems > 0 && len(m.progressLog) > m.maxLogItems {
		// keep only the tail
		m.progressLog = m.progressLog[len(m.progressLog)-m.maxLogItems:]
	}
	m.mu.Unlock()

	// write out snapshot log (overwrite)
	if m.logFile != "" {
		data, err := json.MarshalIndent(m.progressLog, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal log failed: %w", err)
		}
		if err := os.MkdirAll(filepath.Dir(m.logFile), 0755); err != nil {
			return fmt.Errorf("mkdir for log failed: %w", err)
		}
		if err := os.WriteFile(m.logFile, data, 0644); err != nil {
			return fmt.Errorf("write log failed: %w", err)
		}
	}

	return nil
}

// LastSnapshot returns the latest collected ProgressStats (nil,false if none)
func (m *Manager) LastSnapshot() (*ProgressStats, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.progressLog) == 0 {
		return nil, false
	}
	// return a copy
	last := m.progressLog[len(m.progressLog)-1]
	return &last, true
}

// GetRemoteProgress filters the last snapshot by "remoteName:" prefix and
// returns an aggregated ProgressStats for that remote (or false if nothing found).
func (m *Manager) GetRemoteProgress(remoteName string) (*ProgressStats, bool) {
	m.mu.Lock()
	if len(m.progressLog) == 0 {
		m.mu.Unlock()
		return nil, false
	}
	snap := m.progressLog[len(m.progressLog)-1]
	m.mu.Unlock()

	var res ProgressStats
	res.Timestamp = snap.Timestamp

	for _, t := range snap.Transferring {
		res.Transferring = append(res.Transferring, t)
		res.Bytes += t.Bytes
		res.Speed += t.Speed
		res.Transfers++

	}
	for _, c := range snap.Checking {
		res.Checking = append(res.Checking, c)
		res.Checks++

	}

	if len(res.Transferring) == 0 && len(res.Checking) == 0 {
		return nil, false
	}

	// copy some useful fields from global snapshot when relevant
	res.ElapsedTime = snap.ElapsedTime
	res.ETA = snap.ETA

	return &res, true
}

func (m *Manager) parseStatsResponse(rawStats map[string]interface{}) *ProgressStats {
	stats := &ProgressStats{}

	if val, ok := rawStats["bytes"].(float64); ok {
		stats.Bytes = int64(val)
	}
	if val, ok := rawStats["checks"].(float64); ok {
		stats.Checks = int64(val)
	}
	if val, ok := rawStats["deletes"].(float64); ok {
		stats.Deletes = int64(val)
	}
	if val, ok := rawStats["errors"].(float64); ok {
		stats.Errors = int64(val)
	}
	if val, ok := rawStats["transfers"].(float64); ok {
		stats.Transfers = int64(val)
	}
	if val, ok := rawStats["speed"].(float64); ok {
		stats.Speed = val
	}
	if val, ok := rawStats["elapsedTime"].(float64); ok {
		stats.ElapsedTime = val
	}
	if val, ok := rawStats["eta"].(float64); ok && val > 0 {
		eta := int64(val)
		stats.ETA = &eta
	}

	if transferring, ok := rawStats["transferring"].([]interface{}); ok {
		for _, t := range transferring {
			if tMap, ok := t.(map[string]interface{}); ok {
				stats.Transferring = append(stats.Transferring, m.parseTransferInfo(tMap))
			}
		}
	}
	if checking, ok := rawStats["checking"].([]interface{}); ok {
		for _, c := range checking {
			if cMap, ok := c.(map[string]interface{}); ok {
				stats.Checking = append(stats.Checking, m.parseTransferInfo(cMap))
			}
		}
	}

	return stats
}

func (m *Manager) parseTransferInfo(tMap map[string]interface{}) TransferInfo {
	transfer := TransferInfo{}

	if name, ok := tMap["name"].(string); ok {
		transfer.Name = name
	}
	if size, ok := tMap["size"].(float64); ok {
		transfer.Size = int64(size)
	}
	if bytes, ok := tMap["bytes"].(float64); ok {
		transfer.Bytes = int64(bytes)
	}
	if speed, ok := tMap["speed"].(float64); ok {
		transfer.Speed = speed
	}
	if eta, ok := tMap["eta"].(float64); ok && eta > 0 {
		etaInt := int64(eta)
		transfer.ETA = &etaInt
	}

	// compute percentage locally (rclone daemon often returns 0)
	if transfer.Size > 0 {
		transfer.Percentage = float64(transfer.Bytes) / float64(transfer.Size) * 100
	} else {
		transfer.Percentage = 0
	}

	return transfer
}

// -------------------- Mount API --------------------

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

func (m *Manager) ListActiveMounts() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	names := make([]string, 0, len(m.activeMounts))
	for name := range m.activeMounts {
		names = append(names, name)
	}
	return names
}

// -------------------- Config Remote --------------------

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

// -------------------- Helpers --------------------

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

func splitUpstreams(s string) []string {
	var out []string
	for _, u := range strings.Split(s, ":") {
		if u != "" {
			out = append(out, strings.TrimSpace(u))
		}
	}
	return out
}

// -------------------- Types --------------------

type Remote struct {
	Name      string
	Type      string
	Upstreams []string
	Config    map[string]string
}

type Mount struct {
	Name       string
	MountPoint string
	CacheMode  string
	CacheSize  string
	CacheAge   string
	CacheDir   string
	Status     string
}
