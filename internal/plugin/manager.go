package plugin

import "sync"

// Manager holds the registry of all active plugins.
type Manager struct {
	mu      sync.RWMutex
	plugins map[string]Plugin
}

// NewManager returns an empty plugin manager.
func NewManager() *Manager {
	return &Manager{plugins: make(map[string]Plugin)}
}

// Register adds a plugin. Silently replaces any existing plugin with the same name.
func (m *Manager) Register(p Plugin) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.plugins[p.Name()] = p
}

// Get returns a plugin by name.
func (m *Manager) Get(name string) (Plugin, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.plugins[name]
	return p, ok
}

// List returns all registered plugins in no guaranteed order.
func (m *Manager) List() []Plugin {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]Plugin, 0, len(m.plugins))
	for _, p := range m.plugins {
		out = append(out, p)
	}
	return out
}
