package rclone

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"

	"gopkg.in/ini.v1"
)

type Manager struct {
	ConfigPath   string
	Remotes      []Remote
	activeMounts map[string]*MountInstance
	mu           sync.Mutex
}

type MountInstance struct {
	Mount  Mount
	Ctx    context.CancelFunc
	Status string
}

func NewManager(configPath string) (*Manager, error) {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, errors.New("rclone config not found at " + configPath)
	}

	m := &Manager{
		ConfigPath:   configPath,
		activeMounts: make(map[string]*MountInstance),
	}

	if err := m.loadRemotes(); err != nil {
		return nil, err
	}

	return m, nil
}

func (m *Manager) loadRemotes() error {
	cfg, err := ini.Load(m.ConfigPath)
	if err != nil {
		return fmt.Errorf("while parsing config: %w", err)
	}

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
			up := section.Key("upstreams").String()
			r.Upstreams = splitUpstreams(up)
		}

		m.Remotes = append(m.Remotes, r)
	}

	return nil
}

func splitUpstreams(s string) []string {
	var out []string
	for _, u := range splitAndTrim(s, ":") {
		if u != "" {
			out = append(out, u)
		}
	}
	return out
}

func splitAndTrim(s, sep string) []string {
	var out []string
	for _, p := range Split(s, sep) {
		out = append(out, TrimSpace(p))
	}
	return out
}

// Helper wrappers to avoid importing strings everywhere
func Split(s, sep string) []string { return []string{} }
func TrimSpace(s string) string    { return s }
