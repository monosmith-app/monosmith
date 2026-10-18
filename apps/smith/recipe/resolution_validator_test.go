package recipe

import (
	"fmt"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// fakePlugins describe a small set of plugins for validation, like the
// handshakes of the real ones.
func fakePlugins() PluginLookup {
	plugins := map[string]protocol.InitializeResult{
		"node":   {Plugin: protocol.Peer{Name: "node", Version: "2.1.0"}, Roles: []protocol.Role{protocol.RoleLanguage, protocol.RoleToolchain, protocol.RoleScaffold}, Languages: []string{"node", "javascript"}, Capabilities: protocol.Capabilities{Generate: true}},
		"python": {Plugin: protocol.Peer{Name: "python", Version: "2.1.0"}, Roles: []protocol.Role{protocol.RoleLanguage, protocol.RoleToolchain, protocol.RoleScaffold}, Languages: []string{"python"}, Capabilities: protocol.Capabilities{Generate: true}},
		"eslint": {Plugin: protocol.Peer{Name: "eslint", Version: "1.0.0"}, Roles: []protocol.Role{protocol.RoleToolchain}, Provides: []string{"lint"}, Requires: []protocol.PluginRef{protocol.RequiresLanguage("node")}},
		"jest":   {Plugin: protocol.Peer{Name: "jest", Version: "1.0.0"}, Roles: []protocol.Role{protocol.RoleToolchain}, Provides: []string{"test"}, Requires: []protocol.PluginRef{protocol.RequiresLanguage("node")}},
		"vitest": {Plugin: protocol.Peer{Name: "vitest", Version: "1.0.0"}, Roles: []protocol.Role{protocol.RoleToolchain}, Provides: []string{"test"}, Requires: []protocol.PluginRef{protocol.RequiresLanguage("node")}, Conflicts: []protocol.PluginRef{protocol.ConflictsWith("karma")}},
		"karma":  {Plugin: protocol.Peer{Name: "karma", Version: "1.0.0"}, Roles: []protocol.Role{protocol.RoleToolchain}, Provides: []string{"e2e"}},
		"docker": {Plugin: protocol.Peer{Name: "docker", Version: "1.0.0"}, Roles: []protocol.Role{protocol.RoleDeploy}},
		"slices": {Plugin: protocol.Peer{Name: "slices", Version: "1.0.0"}, Roles: []protocol.Role{protocol.RoleStructure}},
	}
	return func(name string) (protocol.InitializeResult, error) {
		if info, ok := plugins[name]; ok {
			return info, nil
		}
		return protocol.InitializeResult{}, fmt.Errorf("%w: %q", ErrUnknownPlugin, name)
	}
}

// resolved builds a resolution from "key=plugin,plugin" entries, each at its
// own made-up line.
func resolved(entries ...string) Resolution {
	r := Resolution{Name: "team"}
	for i, e := range entries {
		key, plugins, _ := strings.Cut(e, "=")
		r.setUse(key, strings.Split(plugins, ","), Origin{Recipe: "team", File: "team.toml", Line: 10 + i})
	}
	return r
}

func TestValidateAcceptsRecipesThatAddUp(t *testing.T) {
	for name, r := range map[string]Resolution{
		"a language alone":           resolved("language=python"),
		"node with its toolchain":    resolved("language=node", "lint=eslint", "test=jest"),
		"additive levels, a list":    resolved("language=python", "structure=slices", "deploy=docker"),
		"a version in range":         resolved("language=node", "lint=eslint@^1.0"),
		"the toolchain level":        resolved("language=node", "toolchain=eslint,jest"),
		"a plugin in two slots once": resolved("language=node", "lint=eslint", "toolchain=eslint"),
	} {
		t.Run(name, func(t *testing.T) {
			if err := Validate(r, fakePlugins()); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestValidateNamesBothSidesAndTheirLines(t *testing.T) {
	for name, tc := range map[string]struct {
		r    Resolution
		want []string
	}{
		"no language": {resolved("structure=slices"), []string{"chooses no language plugin"}},
		"two languages for one level": {resolved("language=node,python"),
			[]string{"team.toml:10 (recipe team): use.language takes one plugin, not node, python"}},
		"an unknown plugin": {resolved("language=python", "lint=ruff"),
			[]string{`use.lint names the plugin "ruff"`, "no such plugin"}},
		"a version out of range": {resolved("language=node", "lint=eslint@^2"),
			[]string{"use.lint wants eslint ^2, but the plugin is 1.0.0"}},
		"a bad constraint": {resolved("language=node", "lint=eslint@banana"),
			[]string{`use.lint = "eslint@banana"`}},
		"a plugin that does not provide the slot": {resolved("language=node", "lint=jest"),
			[]string{`use.lint names jest, which does not provide a "lint" slot (it provides: test)`}},
		"a typo of a level is a slot nobody provides": {resolved("langauge=node"),
			[]string{`which does not provide a "langauge" slot`}},
		"a requirement not met": {resolved("language=python", "lint=eslint"),
			[]string{"eslint requires a language plugin node, and none is chosen", "lint = \"eslint\"  team.toml:11 (recipe team)"}},
		"two plugins for one slot": {resolved("language=node", "lint=eslint", "test=jest", "toolchain=vitest"),
			[]string{"jest and vitest both provide the test task; keep one", "test = \"jest\"  team.toml:12 (recipe team)", "toolchain = \"vitest\"  team.toml:13 (recipe team)"}},
		"a conflict": {resolved("language=node", "test=vitest", "e2e=karma"),
			[]string{"vitest conflicts with karma", "e2e = \"karma\"", "test = \"vitest\""}},
	} {
		t.Run(name, func(t *testing.T) {
			err := Validate(tc.r, fakePlugins())
			if err == nil {
				t.Fatal("want an error")
			}
			for _, want := range tc.want {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("%q must contain %q", err, want)
				}
			}
		})
	}
}
