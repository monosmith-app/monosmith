package internallink

import (
	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// RegistryState is what the registry answered about a package name.
type RegistryState int

// What a lookup can find.
const (
	// RegistryMissing means the registry answered and has no such package.
	RegistryMissing RegistryState = iota
	// RegistryHas means the registry has a package with that name.
	RegistryHas
	// RegistryUnknown means the registry could not be asked (offline, no answer
	// in time, or the plugin cannot look names up).
	RegistryUnknown
)

// Lookup asks the registry of a project's language whether a package of that
// name exists. Tests replace it with a fake.
type Lookup func(from *projectgraph.Project, name string) RegistryState

// SessionLookup is the Lookup that asks the project's language plugin with the
// latest-version method, which answers within protocol.RegistryLookupTimeout and
// reports an unreachable registry instead of failing.
func SessionLookup(s *session.Session) Lookup {
	return func(from *projectgraph.Project, name string) RegistryState {
		c := s.Plugins.Get(from.Plugin)
		if c == nil || !c.Supports(protocol.MethodLatest) {
			return RegistryUnknown
		}
		var res protocol.LatestResult
		if err := c.Call(protocol.MethodLatest, protocol.LatestParams{Names: []string{name}}, &res, protocol.RegistryLookupTimeout*2); err != nil {
			return RegistryUnknown
		}
		if _, bad := res.Unreachable[name]; bad {
			return RegistryUnknown
		}
		if _, ok := res.Latest[name]; ok {
			return RegistryHas
		}
		return RegistryMissing
	}
}
