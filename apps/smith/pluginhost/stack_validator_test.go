package pluginhost

import (
	"io"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/host"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/serve"
)

// fake describes itself with info.
type fake struct{ info protocol.InitializeResult }

func (f fake) Describe(protocol.InitializeParams) (protocol.InitializeResult, error) {
	return f.info, nil
}

// scaffold is a fake that can generate.
type scaffold struct{ fake }

func (scaffold) Generate(protocol.GenerateParams) (protocol.GenerateResult, error) {
	return protocol.GenerateResult{}, nil
}

func connect(t *testing.T, p serve.Plugin) *host.Client {
	t.Helper()
	toPlugin, fromHost := io.Pipe()
	toHost, fromPlugin := io.Pipe()
	go func() { _ = serve.Serve(p, toPlugin, fromPlugin) }()
	c, err := host.Connect(toHost, fromHost, host.Options{WorkspaceRoot: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	return c
}

func plugin(name string, kind protocol.PluginKind, roles []protocol.Role, languages []string, provides ...string) fake {
	return fake{protocol.InitializeResult{Plugin: protocol.Peer{Name: name}, Kind: kind, Roles: roles, Languages: languages, Provides: provides}}
}

func TestValidateStack(t *testing.T) {
	tc := []protocol.Role{protocol.RoleToolchain}
	set := &PluginSet{Clients: []*host.Client{
		connect(t, plugin("node", protocol.KindLanguage, nil, []string{"node", "ts"})),
		connect(t, plugin("go", protocol.KindLanguage, nil, []string{"go"})),
		connect(t, plugin("jest", protocol.KindConvention, tc, []string{"node"}, "test")),
		connect(t, plugin("vitest", protocol.KindConvention, tc, []string{"node"}, "test")),
		connect(t, plugin("slices", protocol.KindConvention, nil, []string{"go"})),
		connect(t, scaffold{fake{protocol.InitializeResult{Plugin: protocol.Peer{Name: "ts"}, Kind: protocol.KindConvention, Languages: []string{"node"}, Roles: []protocol.Role{protocol.RoleScaffold, protocol.RoleToolchain}, Provides: []string{"build"}}}}),
	}}
	for _, c := range []struct {
		stack protocol.Stack
		want  string
	}{
		{protocol.Stack{Language: "node", Toolchain: []string{"jest"}}, ""},
		{protocol.Stack{Language: "node", Scaffold: "ts", Toolchain: []string{"ts", "jest"}}, ""},
		{protocol.Stack{Language: "go", Structure: []string{"slices"}}, ""},
		{protocol.Stack{Language: "jest"}, "not a language plugin"},
		{protocol.Stack{Language: "rust"}, `"rust" is not loaded`},
		{protocol.Stack{Language: "node", Toolchain: []string{"jest", "vitest"}}, "jest and vitest both provide the test task"},
		{protocol.Stack{Language: "node", Toolchain: []string{"slices"}}, "listed as toolchain"},
		{protocol.Stack{Language: "node", Structure: []string{"slices"}}, "not to node projects"},
		{protocol.Stack{Language: "node", Scaffold: "jest"}, "listed as scaffold"},
		{protocol.Stack{Language: "node", Toolchain: []string{"jest", "jest"}}, "listed twice as toolchain"},
	} {
		err := ValidateStack(c.stack, set)
		if (c.want == "") != (err == nil) || (err != nil && !strings.Contains(err.Error(), c.want)) {
			t.Errorf("%+v: got %v, want %q", c.stack, err, c.want)
		}
	}

	_, err := NewCatalog(set, workspace.Config{Stacks: map[string]workspace.StackConfig{"bad": {Language: "node", Toolchain: []string{"jest", "vitest"}}}})
	if err == nil || !strings.Contains(err.Error(), "[stacks.bad]") {
		t.Fatalf("a bad smith.toml stack fails: %v", err)
	}
	_, err = NewCatalog(set, workspace.Config{DefaultStacks: map[string]string{"node": "missing"}})
	if err == nil || !strings.Contains(err.Error(), "[default_stacks]") {
		t.Fatalf("an unknown default fails: %v", err)
	}
}

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
