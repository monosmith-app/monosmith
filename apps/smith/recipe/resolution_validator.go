package recipe

import (
	"fmt"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/pluginhost"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// PluginLookup describes a plugin by name from its handshake, or says why it
// cannot be found.
type PluginLookup func(name string) (protocol.InitializeResult, error)

// levelRoles are the keys of use that name a level; any other key names a slot,
// the task a toolchain plugin provides.
var levelRoles = map[string]protocol.Role{
	"language":  protocol.RoleLanguage,
	"scaffold":  protocol.RoleScaffold,
	"structure": protocol.RoleStructure,
	"deploy":    protocol.RoleDeploy,
	"toolchain": protocol.RoleToolchain,
}

// Validate checks a resolved recipe before anything is written:
//
//   - every plugin it names exists, and at a version its constraint (name@^1.2)
//     accepts
//   - a slot is filled by a plugin that provides it
//   - the language and the scaffold are one plugin each
//   - the whole set passes the rules a stack passes (pluginhost.ValidateMembers):
//     one plugin per exclusive level and slot, requirements met, no conflicts
//
// Every error names the plugins and the recipe lines that chose them.
func Validate(res Resolution, lookup PluginLookup) error {
	var members []pluginhost.Member
	seen := map[string]bool{}
	for _, u := range res.Uses {
		role, isLevel := levelRoles[u.Key]
		if !isLevel {
			role = protocol.RoleToolchain
		}
		if (role == protocol.RoleLanguage || role == protocol.RoleScaffold) && len(u.Plugins) > 1 {
			return &ValidationError{Message: fmt.Sprintf("%s: use.%s takes one plugin, not %s", u.Origin, u.Key, strings.Join(u.Plugins, ", "))}
		}
		for _, spec := range u.Plugins {
			name, constraint, _ := strings.Cut(spec, "@")
			info, err := lookup(name)
			if err != nil {
				return &ValidationError{Message: fmt.Sprintf("%s: use.%s names the plugin %q: %v", u.Origin, u.Key, name, err)}
			}
			if constraint != "" {
				ok, err := pluginhost.Satisfies(info.Plugin.Version, constraint)
				if err != nil {
					return &ValidationError{Message: fmt.Sprintf("%s: use.%s = %q: %v", u.Origin, u.Key, spec, err)}
				}
				if !ok {
					return &ValidationError{Message: fmt.Sprintf("%s: use.%s wants %s %s, but the plugin is %s", u.Origin, u.Key, name, constraint, info.Plugin.Version)}
				}
			}
			if !isLevel && !slices.Contains(info.Provides, u.Key) {
				return &ValidationError{Message: fmt.Sprintf("%s: use.%s names %s, which does not provide a %q slot (it provides: %s). A key of use is a role (language, scaffold, structure, deploy, toolchain) or the task a toolchain plugin provides", u.Origin, u.Key, name, u.Key, orNone(info.Provides))}
			}
			if id := name + "/" + string(role); !seen[id] {
				seen[id] = true
				members = append(members, pluginhost.Member{Info: info, Role: role})
			}
		}
	}
	if _, ok := res.UseFor("language"); !ok {
		return &ValidationError{Message: "the recipe chooses no language plugin; set language = \"...\" under [use], or in the when block of each language"}
	}
	if err := pluginhost.ValidateMembers(members); err != nil {
		return &ValidationError{Message: err.Error(), Chosen: chosenBy(res, err.Error())}
	}
	return nil
}

// chosenBy lists the recipe lines that chose each plugin a message names.
func chosenBy(res Resolution, message string) []string {
	var out []string
	for _, u := range res.Uses {
		for _, spec := range u.Plugins {
			name, _, _ := strings.Cut(spec, "@")
			if strings.Contains(message, name) {
				out = append(out, fmt.Sprintf("%s = %q  %s", u.Key, spec, u.Origin))
			}
		}
	}
	return out
}

func orNone(list []string) string {
	if len(list) == 0 {
		return "none"
	}
	return strings.Join(list, ", ")
}
