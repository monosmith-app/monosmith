package pluginhost

import (
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

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
