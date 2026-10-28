package clicontract

import (
	"io"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/pluginhost"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/host"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/serve"
)

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

func TestKindsComeFromTheLanguagePlugins(t *testing.T) {
	set := &pluginhost.PluginSet{}
	for _, name := range []string{"node", "go", "jest"} {
		ctor, _ := pluginhost.Builtin(name)
		set.Clients = append(set.Clients, connect(t, ctor("1.0.0")))
	}
	kinds := Kinds(set)
	var names []string
	for _, k := range kinds {
		names = append(names, k.Kind)
	}
	if len(kinds) != 6 || names[0] != "go-app" || names[5] != "node-package" {
		t.Fatalf("kinds %v (a toolchain plugin is no kind)", names)
	}
	if node := kinds[3]; node.Kind != "node-app" || node.Language != "typescript" || node.Flags[0] != "stack" {
		t.Fatalf("node-app %+v", node)
	}
}

func TestKindsLeaveOutWhatAPluginDoesNotScaffold(t *testing.T) {
	ctor, _ := pluginhost.Builtin("bicep")
	set := &pluginhost.PluginSet{Clients: []*host.Client{connect(t, ctor("1.0.0"))}}
	var names []string
	for _, k := range Kinds(set) {
		names = append(names, k.Kind)
	}
	if len(names) != 2 || names[0] != "bicep-app" || names[1] != "bicep-lib" {
		t.Fatalf("kinds %v, want bicep-app and bicep-lib (a Bicep package is refused)", names)
	}
}
