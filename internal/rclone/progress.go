// Package rclone provides management for rclone operations, including
// config parsing, mount management, progress tracking, daemon control,
// and filesystem operations.
package rclone

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// ProgressStats holds statistics for rclone operations.
type ProgressStats struct {
	Transferring  []TransferInfo `json:"transferring,omitempty"`
	Checking      []TransferInfo `json:"checking,omitempty"`
	Bytes         int64          `json:"bytes"`
	Checks        int64          `json:"checks"`
	Deletes       int64          `json:"deletes"`
	Errors        int64          `json:"errors"`
	Transfers     int64          `json:"transfers"`
	Speed         float64        `json:"speed"`
	CheckSpeed    float64        `json:"checkSpeed"`
	TransferSpeed float64        `json:"transferSpeed"`
	ETA           int64          `json:"eta,omitempty"`
	ElapsedTime   float64        `json:"elapsedTime"`
	Timestamp     time.Time      `json:"timestamp"`
}

// TransferInfo holds details about a single transfer or check.
type TransferInfo struct {
	Name       string  `json:"name"`
	Size       int64   `json:"size"`
	Bytes      int64   `json:"bytes"`
	Checked    bool    `json:"checked"`
	Percentage float64 `json:"percentage"`
	Speed      float64 `json:"speed"`
	ETA        int64   `json:"eta,omitempty"`
}

// UpdateProgress fetches progress stats from the rclone daemon
// and updates the in-memory and file logs.
func (m *Manager) UpdateProgress() error {
	url := fmt.Sprintf("http://127.0.0.1:%d/core/stats", m.rcPort)
	payload, _ := json.Marshal(map[string]interface{}{})

	resp, err := http.Post(url, "application/json", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("http post failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("stats returned %s", resp.Status)
	}

	var raw map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return fmt.Errorf("decode stats failed: %w", err)
	}

	stats := m.parseStatsResponse(raw)
	stats.Timestamp = time.Now()

	m.mu.Lock()
	log.Println("[DEBUG] stats:", stats)
	m.progressLog = append(m.progressLog, *stats)
	if m.maxLogItems > 0 && len(m.progressLog) > m.maxLogItems {
		m.progressLog = m.progressLog[len(m.progressLog)-m.maxLogItems:]
	}
	m.mu.Unlock()

	if m.logFile != "" {
		data, _ := json.MarshalIndent(m.progressLog, "", "  ")
		_ = os.MkdirAll(filepath.Dir(m.logFile), 0755)
		_ = os.WriteFile(m.logFile, data, 0644)
	}

	return nil
}

// LastSnapshot returns the most recent ProgressStats, if any.
func (m *Manager) LastSnapshot() (*ProgressStats, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.progressLog) == 0 {
		return nil, false
	}
	last := m.progressLog[len(m.progressLog)-1]
	return &last, true
}

// GetRemoteProgress returns stats for a specific remote.
func (m *Manager) GetRemoteProgress(remote string) (*ProgressStats, bool) {
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

	res.ElapsedTime = snap.ElapsedTime
	res.ETA = snap.ETA

	return &res, true
}

// parseStatsResponse converts raw JSON from rclone into ProgressStats.
func (m *Manager) parseStatsResponse(raw map[string]interface{}) *ProgressStats {
	stats := &ProgressStats{}

	if val, ok := raw["bytes"].(float64); ok {
		stats.Bytes = int64(val)
	}
	if val, ok := raw["checks"].(float64); ok {
		stats.Checks = int64(val)
	}
	if val, ok := raw["errors"].(float64); ok {
		stats.Errors = int64(val)
	}
	if val, ok := raw["transfers"].(float64); ok {
		stats.Transfers = int64(val)
	}
	if val, ok := raw["speed"].(float64); ok {
		stats.Speed = val
	}
	if val, ok := raw["elapsedTime"].(float64); ok {
		stats.ElapsedTime = val
	}
	if val, ok := raw["eta"].(float64); ok && val > 0 {
		stats.ETA = int64(val)
	}

	if tr, ok := raw["transferring"].([]interface{}); ok {
		for _, t := range tr {
			if tMap, ok := t.(map[string]interface{}); ok {
				stats.Transferring = append(stats.Transferring, m.parseTransferInfo(tMap))
			}
		}
	}
	if chk, ok := raw["checking"].([]interface{}); ok {
		for _, c := range chk {
			if cMap, ok := c.(map[string]interface{}); ok {
				stats.Checking = append(stats.Checking, m.parseTransferInfo(cMap))
			}
		}
	}

	return stats
}

// parseTransferInfo converts raw JSON into a TransferInfo.
func (m *Manager) parseTransferInfo(tMap map[string]interface{}) TransferInfo {
	var t TransferInfo

	if name, ok := tMap["name"].(string); ok {
		t.Name = name
	}
	if size, ok := tMap["size"].(float64); ok {
		t.Size = int64(size)
	}
	if b, ok := tMap["bytes"].(float64); ok {
		t.Bytes = int64(b)
	}
	if sp, ok := tMap["speed"].(float64); ok {
		t.Speed = sp
	}
	if eta, ok := tMap["eta"].(float64); ok && eta > 0 {
		t.ETA = int64(eta)
	}
	if t.Size > 0 {
		t.Percentage = float64(t.Bytes) / float64(t.Size) * 100
	}

	return t
}
