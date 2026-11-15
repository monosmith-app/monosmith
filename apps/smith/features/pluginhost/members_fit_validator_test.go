package pluginhost

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestValidateMembers(t *testing.T) {
	info := func(name string, roles []protocol.Role, mod func(*protocol.InitializeResult)) protocol.InitializeResult {
		r := protocol.InitializeResult{Plugin: protocol.Peer{Name: name}, Kind: protocol.KindConvention, Roles: roles}
		if mod != nil {
			mod(&r)
		}
		return r
	}
	tc := []protocol.Role{protocol.RoleToolchain}
	node := Member{Role: protocol.RoleLanguage, Info: protocol.InitializeResult{Plugin: protocol.Peer{Name: "node"}, Kind: protocol.KindLanguage, Languages: []string{"node", "js"}}}
	python := Member{Role: protocol.RoleLanguage, Info: protocol.InitializeResult{Plugin: protocol.Peer{Name: "python"}, Kind: protocol.KindLanguage}}
	jest := Member{Role: protocol.RoleToolchain, Info: info("jest", tc, func(r *protocol.InitializeResult) {
		r.Provides = []string{"test"}
		r.Requires = []protocol.PluginRef{protocol.RequiresLanguage("js")}
	})}
	vitest := Member{Role: protocol.RoleToolchain, Info: info("vitest", tc, func(r *protocol.InitializeResult) {
		r.Provides = []string{"test"}
		r.Conflicts = []protocol.PluginRef{protocol.ConflictsWith("karma")}
	})}
	karma := Member{Role: protocol.RoleToolchain, Info: info("karma", tc, func(r *protocol.InitializeResult) { r.Provides = []string{"browser-test"} })}
	scaffold := func(name string) Member {
		return Member{Role: protocol.RoleScaffold, Info: info(name, []protocol.Role{protocol.RoleScaffold}, func(r *protocol.InitializeResult) { r.Capabilities.Generate = true })}
	}
	needsLint := Member{Role: protocol.RoleStructure, Info: info("slices", nil, func(r *protocol.InitializeResult) {
		r.Requires = []protocol.PluginRef{{Level: protocol.RoleToolchain, Slot: "lint"}}
	})}
	for _, c := range []struct {
		name    string
		members []Member
		want    string
	}{
		{"a requirement on a language alias is met", []Member{node, jest}, ""},
		{"no language", []Member{jest}, "no language plugin is chosen"},
		{"two languages", []Member{node, python}, "node and python are both chosen as the language plugin"},
		{"a requirement unmet", []Member{python, jest}, "jest requires a language plugin js, and none is chosen"},
		{"a slot requirement unmet", []Member{node, needsLint}, "slices requires a toolchain plugin providing lint"},
		{"a conflict", []Member{node, vitest, karma}, "vitest conflicts with karma"},
		{"two scaffolds", []Member{node, scaffold("ts"), scaffold("vite")}, "ts and vite are both chosen as the scaffold"},
		{"a shared slot", []Member{node, jest, vitest}, "jest and vitest both provide the test task"},
	} {
		err := ValidateMembers(c.members)
		if (c.want == "") != (err == nil) || (err != nil && !strings.Contains(err.Error(), c.want)) {
			t.Errorf("%s: got %v, want %q", c.name, err, c.want)
		}
	}
}

func TestFitsLeavesNonLanguageRequirementsForLater(t *testing.T) {
	node := Member{Role: protocol.RoleLanguage, Info: protocol.InitializeResult{Plugin: protocol.Peer{Name: "node"}, Roles: []protocol.Role{protocol.RoleLanguage}, Languages: []string{"node"}}}
	python := Member{Role: protocol.RoleLanguage, Info: protocol.InitializeResult{Plugin: protocol.Peer{Name: "python"}, Roles: []protocol.Role{protocol.RoleLanguage}}}
	lint := protocol.InitializeResult{Plugin: protocol.Peer{Name: "eslint"}, Roles: []protocol.Role{protocol.RoleToolchain}, Provides: []string{"lint"}, Requires: []protocol.PluginRef{protocol.RequiresLanguage("node")}}
	if err := Fits([]Member{node}, Member{Info: lint, Role: protocol.RoleToolchain}); err != nil {
		t.Errorf("a node linter does not fit node: %v", err)
	}
	if err := Fits([]Member{python}, Member{Info: lint, Role: protocol.RoleToolchain}); err == nil {
		t.Error("a node linter fits python")
	}
	needsPeer := protocol.InitializeResult{Plugin: protocol.Peer{Name: "x"}, Roles: []protocol.Role{protocol.RoleStructure}, Requires: []protocol.PluginRef{{Level: protocol.RoleToolchain, Slot: "test"}}}
	if err := Fits([]Member{node}, Member{Info: needsPeer, Role: protocol.RoleStructure}); err != nil {
		t.Errorf("a requirement a later plugin may meet failed: %v", err)
	}
	if err := ValidateMembers([]Member{node, {Info: needsPeer, Role: protocol.RoleStructure}}); err == nil {
		t.Error("ValidateMembers accepted an unmet requirement")
	}
}
