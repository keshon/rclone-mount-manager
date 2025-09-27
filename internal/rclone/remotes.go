// Package rclone provides management for rclone operations, including
// config parsing, mount management, progress tracking, daemon control,
// and filesystem operations.
package rclone

import (
	"fmt"
	"strings"

	"gopkg.in/ini.v1"
)

// Remote represents a configured remote storage provider.
type Remote struct {
	Name      string            // Remote name
	Type      string            // Remote type (e.g., "gdrive", "dropbox", "s3", "union")
	Upstreams []string          // Upstream remotes (for union remotes)
	Config    map[string]string // Additional configuration parameters
}

// loadRemotes parses the rclone configuration file and loads all remotes.
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

// splitUpstreams parses a colon-separated string of upstream remotes.
func splitUpstreams(s string) []string {
	var out []string
	for _, u := range strings.Split(s, ":") {
		if cleaned := strings.TrimSpace(u); cleaned != "" {
			out = append(out, cleaned)
		}
	}
	return out
}
