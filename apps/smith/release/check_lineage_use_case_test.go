package release

import (
	"io"
	"sync"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/pluginhost"
	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/host"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/serve"
)

// registryPlugin is a language plugin served in this process whose registry
// has every package at 9.9.9, and that records who was asked about.
type registryPlugin struct {
	mu    sync.Mutex
	asked []string
}

func (p *registryPlugin) Describe(protocol.InitializeParams) (protocol.InitializeResult, error) {
	return protocol.InitializeResult{Plugin: protocol.Peer{Name: "node", Version: "1.0.0"}, Kind: protocol.KindLanguage, Manifests: []string{"x.json"}}, nil
}

func (p *registryPlugin) Published(params protocol.PublishedParams) (protocol.PublishedResult, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.asked = append(p.asked, params.Project.Name)
	return protocol.PublishedResult{Version: "9.9.9", Registry: "registry.npmjs.org"}, nil
}

// lineageSession has one package per name, each tagged <name>@0.5.0 and so
// behind the fake registry's 9.9.9, with the registry setting the map gives.
func lineageSession(t *testing.T, modes map[string]*protocol.Registries) (*session.Session, *registryPlugin) {
	t.Helper()
	g := repo(t)
	var list []*projectgraph.Project
	for _, name := range []string{"a", "b", "c"} {
		mode, ok := modes[name]
		if !ok {
			continue
		}
		commit(t, g, "packages/"+name+"/index.js", "feat: "+name)
		if err := g.CreateTag(name+"@0.5.0", name+" 0.5.0"); err != nil {
			t.Fatal(err)
		}
		list = append(list, &projectgraph.Project{Project: protocol.Project{Name: name, Root: "packages/" + name, Kind: protocol.ProjectPackage, Version: "0.5.0", Exports: []string{name}, Registries: mode}, Plugin: "node"})
	}
	graph, err := projectgraph.New(list)
	if err != nil {
		t.Fatal(err)
	}
	fake := &registryPlugin{}
	hostR, pluginW := io.Pipe()
	pluginR, hostW := io.Pipe()
	go func() { _ = serve.Serve(fake, pluginR, pluginW); _ = pluginW.Close() }()
	c, err := host.Connect(hostR, hostW, host.Options{WorkspaceRoot: g.Dir})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	return &session.Session{Root: g.Dir, Config: workspace.Config{}, Git: g, Graph: graph, Plugins: &pluginhost.PluginSet{Clients: []*host.Client{c}}}, fake
}

func TestTheLineageCheckAsksNoRegistryForAProjectThatPublishesToNone(t *testing.T) {
	s, plugin := lineageSession(t, map[string]*protocol.Registries{"a": {NoPublish: true}, "b": {PackToRelease: true}})
	if findings := CheckLineage(s, false); len(findings) != 0 {
		t.Fatalf("tags behind a registry nobody publishes to are no finding, so planning proceeds: %+v", findings)
	}
	if len(plugin.asked) != 0 {
		t.Fatalf("no registry request is made: %v", plugin.asked)
	}
}

func TestTheLineageCheckStillRefusesAPublishingProjectBehindItsRegistry(t *testing.T) {
	s, plugin := lineageSession(t, map[string]*protocol.Registries{"a": {NoPublish: true}, "b": {PackToRelease: true}, "c": nil})
	findings := CheckLineage(s, false)
	if len(plugin.asked) != 1 || plugin.asked[0] != "c" {
		t.Fatalf("only the project that publishes is asked: %v", plugin.asked)
	}
	if len(findings) != 1 || findings[0].Project != "c" || !findings[0].Blocking {
		t.Fatalf("the check still works for it: %+v", findings)
	}
}
