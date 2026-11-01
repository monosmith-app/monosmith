package releasing

import (
	"sync"
	"time"
)

// registryTTL is how long the registry npm is configured with is remembered.
// One smith run asks for it once per project, so a short memory turns dozens
// of `npm config get registry` processes into one without going stale in a
// long-lived process such as the MCP server.
const registryTTL = 30 * time.Second

type registryEntry struct {
	registry string
	asked    time.Time
}

// registryStore remembers, per workspace root, the answer of `npm config get
// registry` for a short time.
type registryStore struct {
	mu      sync.Mutex
	now     func() time.Time
	entries map[string]registryEntry
}

// registries is the process-wide memory of configured registries.
var registries = &registryStore{now: time.Now, entries: map[string]registryEntry{}}

// configured returns the registry remembered for workspaceRoot, or the answer
// of ask, which it then remembers. Callers that arrive while ask runs wait
// for it, so many parallel lookups start one npm, not one each. A workspace
// has one root, so holding the lock while asking costs nothing.
func (s *registryStore) configured(workspaceRoot string, ask func() string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if e, ok := s.entries[workspaceRoot]; ok && s.now().Sub(e.asked) < registryTTL {
		return e.registry
	}
	registry := ask()
	s.entries[workspaceRoot] = registryEntry{registry: registry, asked: s.now()}
	return registry
}

// forget drops everything remembered, for tests.
func (s *registryStore) forget() {
	s.mu.Lock()
	s.entries = map[string]registryEntry{}
	s.mu.Unlock()
}
