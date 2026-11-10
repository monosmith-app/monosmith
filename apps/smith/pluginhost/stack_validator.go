package pluginhost

import (
	"fmt"
	"slices"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Member is one plugin chosen for a project, at the level it is chosen for,
// with what its handshake declared. A plugin chosen at two levels (a
// scaffold that is also a toolchain) is two members.
type Member struct {
	Info protocol.InitializeResult
	Role protocol.Role
}

// matches reports whether a chosen plugin is what a reference names: every
// field the reference sets must agree. A language plugin also answers to
// each of its languages.
func matches(ref protocol.PluginRef, m Member) bool {
	if ref.IsZero() {
		return false
	}
	if ref.Level != "" && ref.Level != m.Role {
		return false
	}
	if ref.Name != "" && ref.Name != m.Info.Plugin.Name {
		aliased := m.Role == protocol.RoleLanguage && slices.Contains(m.Info.Languages, ref.Name)
		if !aliased {
			return false
		}
	}
	return ref.Slot == "" || slices.Contains(m.Info.Provides, ref.Slot)
}

// ValidateStack checks a stack against the loaded plugins: every plugin it
// names is loaded, then the stack's make-up as ValidateMembers does.
func ValidateStack(st protocol.Stack, set *PluginSet) error {
	lang := set.Get(st.Language)
	if lang == nil {
		return fmt.Errorf("language plugin %q is not loaded", st.Language)
	}
	members := []Member{{Info: lang.Info, Role: protocol.RoleLanguage}}
	add := func(name string, role protocol.Role) error {
		c := set.Get(name)
		if c == nil {
			return fmt.Errorf("plugin %q is not loaded", name)
		}
		members = append(members, Member{Info: c.Info, Role: role})
		return nil
	}
	for _, list := range []struct {
		names []string
		role  protocol.Role
	}{{st.Toolchain, protocol.RoleToolchain}, {st.Structure, protocol.RoleStructure}, {[]string{st.Scaffold}, protocol.RoleScaffold}, {st.Deploy, protocol.RoleDeploy}} {
		for _, n := range list.names {
			if n == "" {
				continue
			}
			if err := add(n, list.role); err != nil {
				return err
			}
		}
	}
	return ValidateMembers(members)
}

func overlaps(a, b []string) bool {
	for _, x := range a {
		if slices.Contains(b, x) {
			return true
		}
	}
	return false
}
