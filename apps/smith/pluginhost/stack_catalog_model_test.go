package pluginhost

import (
	"slices"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/host"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// choiceSet is a node language plugin and two plugins that offer choices.
func choiceSet(t *testing.T) *PluginSet {
	t.Helper()
	both := []protocol.Role{protocol.RoleScaffold, protocol.RoleToolchain}
	tc := []protocol.Role{protocol.RoleToolchain}
	web := plugin("web", protocol.KindConvention, both, []string{"node"}, "build")
	web.info.Requires = []protocol.PluginRef{protocol.RequiresLanguage("node")}
	web.info.Choices = []protocol.Choice{
		{Axis: "framework", Name: "react", Language: "node", Roles: both, Needs: []string{"test"}},
		{Axis: "style", Name: "bootstrap", Language: "node"},
	}
	web.info.Stacks = []protocol.Stack{{Name: "react", Choices: []string{"react"}}, {Name: "broken", Choices: []string{"bootstrap"}}}
	tests := plugin("vitest", protocol.KindConvention, tc, []string{"node"}, "test")
	tests.info.Requires = []protocol.PluginRef{protocol.RequiresLanguage("node")}
	tests.info.Choices = []protocol.Choice{{Axis: "test", Name: "vitest", Language: "node", Roles: tc, Default: true}}
	return &PluginSet{Clients: []*host.Client{
		connect(t, plugin("node", protocol.KindLanguage, nil, []string{"node"})),
		connect(t, scaffold{web}),
		connect(t, tests),
	}}
}

func TestACatalogExpandsSavedSelectionsAndKeepsAProblemWithAPublishedOne(t *testing.T) {
	cfg := workspace.Config{Stacks: map[string]workspace.StackConfig{"mine": {Choices: []string{"react", "bootstrap"}}}}
	c, err := NewCatalog(choiceSet(t), cfg)
	if err != nil {
		t.Fatal(err)
	}
	mine, err := c.Get("mine")
	if err != nil || mine.Scaffold != "web" || !slices.Contains(mine.Toolchain, "vitest") {
		t.Errorf("mine: %+v %v", mine, err)
	}
	if _, err := c.Get("react"); err != nil {
		t.Errorf("a published selection is usable: %v", err)
	}
	if _, err := c.Get("broken"); err == nil || !strings.Contains(err.Error(), "writes the project") {
		t.Errorf("a published selection that cannot be solved keeps its problem: %v", err)
	}
	cfg.Stacks["bad"] = workspace.StackConfig{Choices: []string{"reakt"}}
	if _, err := NewCatalog(choiceSet(t), cfg); err == nil || !strings.Contains(err.Error(), "[stacks.bad]") || !strings.Contains(err.Error(), "react") {
		t.Errorf("a declared selection that cannot be solved is an error that names it: %v", err)
	}
}
