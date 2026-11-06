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
