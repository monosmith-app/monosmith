package recipewizard

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/pluginhost"
	"github.com/monosmith-app/monosmith/apps/smith/recipe"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Scenario is one combination of answers the recipe being written can be
// used with, and the plugins it has chosen so far for it.
type Scenario struct {
	Answers map[string]string
	Members []pluginhost.Member
}

// Candidates are the plugins a recipe may name, described by their handshake.
type Candidates struct {
	infos []protocol.InitializeResult
}

// PluginNames lists every plugin a recipe can name: the built-in ones, the
// ones smith.toml declares and the ones installed on the search path.
func PluginNames(root string, external map[string]workspace.ExternalPlugin) []string {
	seen := map[string]bool{}
	for _, name := range pluginhost.BuiltinNames() {
		seen[name] = true
	}
	for name := range pluginhost.BuiltinAddons {
		seen[name] = true
	}
	for name := range external {
		seen[name] = true
	}
	if installed, err := pluginhost.FindInstalled(root); err == nil {
		for _, p := range installed {
			seen[p.Name] = true
		}
	}
	return keys(seen)
}

// NewCandidates describes each named plugin with lookup. A plugin that
// cannot be described is left out, and reported in skipped.
func NewCandidates(names []string, lookup recipe.PluginLookup) (c *Candidates, skipped []string) {
	c = &Candidates{}
	for _, name := range names {
		info, err := lookup(name)
		if err != nil {
			skipped = append(skipped, fmt.Sprintf("%s: %v", name, err))
			continue
		}
		c.infos = append(c.infos, info)
	}
	sort.Slice(c.infos, func(i, j int) bool { return c.infos[i].Plugin.Name < c.infos[j].Plugin.Name })
	return c, skipped
}

// Levels lists the levels some plugin can fill: the language, the scaffold,
// each task slot a toolchain provides, structure and deploy.
func (c *Candidates) Levels() []string {
	seen := map[string]bool{}
	for _, info := range c.infos {
		for _, role := range info.EffectiveRoles() {
			switch role {
			case protocol.RoleLanguage:
				seen[LevelLanguage] = true
			case protocol.RoleScaffold:
				seen[LevelScaffold] = true
			case protocol.RoleStructure:
				seen[LevelStructure] = true
			case protocol.RoleDeploy:
				seen[LevelDeploy] = true
			case protocol.RoleToolchain:
				for _, slot := range info.Provides {
					seen[slot] = true
				}
			}
		}
	}
	return OrderLevelKeys(keys(seen))
}

// roleOf is the role a level is filled in, and for a slot the task the plugin
// must provide.
func roleOf(level string) (role protocol.Role, slot string) {
	switch level {
	case LevelLanguage:
		return protocol.RoleLanguage, ""
	case LevelScaffold:
		return protocol.RoleScaffold, ""
	case LevelStructure:
		return protocol.RoleStructure, ""
	case LevelDeploy:
		return protocol.RoleDeploy, ""
	case LevelToolchain:
		return protocol.RoleToolchain, ""
	}
	return protocol.RoleToolchain, level
}

// For lists the plugins that can fill a level and fit the plugins every
// scenario has chosen so far. Before a language is chosen nothing can be
// judged beyond the level itself.
func (c *Candidates) For(level string, scenarios []Scenario) []protocol.InitializeResult {
	role, slot := roleOf(level)
	var out []protocol.InitializeResult
	for _, info := range c.infos {
		if !slices.Contains(info.EffectiveRoles(), role) || (slot != "" && !slices.Contains(info.Provides, slot)) {
			continue
		}
		fits := true
		for _, s := range scenarios {
			if hasLanguage(s.Members) || role == protocol.RoleLanguage {
				if pluginhost.Fits(s.Members, pluginhost.Member{Info: info, Role: role}) != nil {
					fits = false
					break
				}
			}
		}
		if fits {
			out = append(out, info)
		}
	}
	return out
}

func hasLanguage(members []pluginhost.Member) bool {
	return slices.ContainsFunc(members, func(m pluginhost.Member) bool { return m.Role == protocol.RoleLanguage })
}

// Info describes a plugin by name, whether or not it fits anywhere.
func (c *Candidates) Info(name string) (protocol.InitializeResult, bool) {
	for _, info := range c.infos {
		if info.Plugin.Name == name {
			return info, true
		}
	}
	return protocol.InitializeResult{}, false
}

// Describe is a line about a plugin: the levels it plays and what it covers.
func Describe(info protocol.InitializeResult) string {
	var parts []string
	for _, role := range info.EffectiveRoles() {
		parts = append(parts, string(role))
	}
	line := strings.Join(parts, ", ")
	if len(info.Provides) > 0 {
		line += ": " + strings.Join(info.Provides, ", ")
	}
	if len(info.Languages) > 0 {
		line += " (" + strings.Join(info.Languages, ", ") + ")"
	}
	return line
}
