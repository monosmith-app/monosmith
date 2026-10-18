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

// ValidateMembers checks a project's make-up, whatever chose it (a stack, a
// recipe): one language plugin and at most one scaffold; every plugin plays
// the level it is chosen for and applies to the language; each slot (a task
// a toolchain provides) has one plugin; every requirement is met and no two
// plugins conflict. Errors name both sides.
func ValidateMembers(members []Member) error { return validateMembers(members, false) }

// Fits says why a plugin cannot join the plugins already chosen for a project,
// or nil when it can. It applies the rules of ValidateMembers, except that a
// requirement on a plugin that is not a language may still be met by a
// plugin chosen later: only the language is known to be settled when the
// others are being picked. A recipe wizard uses it to offer only the plugins
// that fit.
func Fits(chosen []Member, candidate Member) error {
	return validateMembers(append(slices.Clone(chosen), candidate), true)
}

func validateMembers(members []Member, partial bool) error {
	var lang *Member
	for i, m := range members {
		if m.Role != protocol.RoleLanguage {
			continue
		}
		if lang != nil {
			return fmt.Errorf("%s and %s are both chosen as the language plugin; keep one", lang.Info.Plugin.Name, m.Info.Plugin.Name)
		}
		lang = &members[i]
	}
	if lang == nil {
		return fmt.Errorf("no language plugin is chosen")
	}
	if !slices.Contains(lang.Info.EffectiveRoles(), protocol.RoleLanguage) {
		return fmt.Errorf("%s is not a language plugin", lang.Info.Plugin.Name)
	}
	langNames := append([]string{lang.Info.Plugin.Name}, lang.Info.Languages...)
	type seat struct {
		name string
		role protocol.Role
	}
	seen := map[seat]bool{}
	checked := map[string]bool{}
	providedBy := map[string]string{}
	exclusive := map[protocol.Role]string{}
	for _, m := range members {
		name := m.Info.Plugin.Name
		if m.Role == protocol.RoleLanguage {
			continue
		}
		// One plugin may play several levels in a project (typescript is the
		// scaffold and a toolchain), but each level lists it once.
		if seen[seat{name, m.Role}] {
			return fmt.Errorf("%s is listed twice as %s", name, m.Role)
		}
		seen[seat{name, m.Role}] = true
		if !slices.Contains(m.Info.EffectiveRoles(), m.Role) {
			return fmt.Errorf("%s is listed as %s, but it plays %v", name, m.Role, m.Info.EffectiveRoles())
		}
		if m.Role.Exclusive() {
			if other, taken := exclusive[m.Role]; taken {
				return fmt.Errorf("%s and %s are both chosen as the %s; keep one", other, name, m.Role)
			}
			exclusive[m.Role] = name
		}
		if len(m.Info.Languages) > 0 && !overlaps(m.Info.Languages, langNames) {
			return fmt.Errorf("%s applies to %v, not to %s projects", name, m.Info.Languages, lang.Info.Plugin.Name)
		}
		if m.Role == protocol.RoleScaffold && !m.Info.Capabilities.Generate {
			return fmt.Errorf("scaffold %s cannot generate projects", name)
		}
		if checked[name] {
			continue
		}
		checked[name] = true
		for _, task := range m.Info.Provides {
			if other, dup := providedBy[task]; dup {
				return fmt.Errorf("%s and %s both provide the %s task; keep one", other, name, task)
			}
			providedBy[task] = name
		}
	}
	for _, m := range members {
		for _, ref := range m.Info.Requires {
			if partial && ref.Level != protocol.RoleLanguage {
				continue
			}
			if !slices.ContainsFunc(members, func(o Member) bool { return o.Info.Plugin.Name != m.Info.Plugin.Name && matches(ref, o) }) {
				return fmt.Errorf("%s requires a %s, and none is chosen", m.Info.Plugin.Name, ref)
			}
		}
		for _, ref := range m.Info.Conflicts {
			for _, o := range members {
				if o.Info.Plugin.Name != m.Info.Plugin.Name && matches(ref, o) {
					return fmt.Errorf("%s conflicts with %s (%s); keep one", m.Info.Plugin.Name, o.Info.Plugin.Name, ref)
				}
			}
		}
	}
	return nil
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
