package core

import "sync"

var (
	registryMu sync.RWMutex
	registry   = map[string]*Module{}
)

func Register(m *Module) {
	registryMu.Lock()
	defer registryMu.Unlock()
	registry[m.Name] = m
}

func AllModules() map[string]*Module {
	registryMu.RLock()
	defer registryMu.RUnlock()
	out := make(map[string]*Module, len(registry))
	for k, v := range registry {
		out[k] = v
	}
	return out
}

func FindCommand(name string) (CommandHandler, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	for _, m := range registry {
		if c, ok := m.Commands[name]; ok {
			return c.Handler, true
		}
	}
	return nil, false
}

func FindModule(name string) (*Module, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	m, ok := registry[name]
	return m, ok
}

func FindModuleByCommand(cmd string) (*Module, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	for _, m := range registry {
		if _, ok := m.Commands[cmd]; ok {
			return m, true
		}
	}
	return nil, false
}
