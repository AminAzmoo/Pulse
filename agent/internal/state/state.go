package state

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type AgentStatus string

const (
	StatusStarting       AgentStatus = "starting"
	StatusRunning        AgentStatus = "running"
	StatusOffline        AgentStatus = "offline"
	StatusNeedsAttention AgentStatus = "needs_attention"
)

type State struct {
	LastConnectionTime time.Time   `json:"last_connection_time"`
	LastHeartbeatTime  time.Time   `json:"last_heartbeat_time"`
	LastErrorCode      string      `json:"last_error_code"`
	AgentStatus        AgentStatus `json:"agent_status"`
}

type Manager struct {
	Path string
	mu   sync.Mutex
}

func (m *Manager) Load() (State, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var s State
	b, err := os.ReadFile(m.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return s, nil
		}
		return s, err
	}
	if len(b) == 0 {
		return s, errors.New("empty state file")
	}
	if err := json.Unmarshal(b, &s); err != nil {
		return s, err
	}
	return s, nil
}

func (m *Manager) Save(s State) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	dir := filepath.Dir(m.Path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	tmp := m.Path + ".tmp"
	b, _ := json.Marshal(s)
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	if err := os.Rename(tmp, m.Path); err != nil {
		return err
	}
	return nil
}

// Save writes simple state file with agent status and metadata.
func Save(path string, status string, metadata map[string]string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	payload := map[string]interface{}{
		"agent_status": status,
		"last_updated": time.Now().UTC(),
		"meta":         metadata,
	}
	b, _ := json.Marshal(payload)
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
