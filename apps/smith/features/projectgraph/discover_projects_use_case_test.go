package projectgraph

import (
	"path"
	"reflect"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/features/pluginhost"
	"github.com/monosmith-app/monosmith/apps/smith/features/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/host"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// manifestLanguage makes a project of every x.json it is asked about.
type manifestLanguage struct{}

func (manifestLanguage) Describe(protocol.InitializeParams) (protocol.InitializeResult, error) {
	return protocol.InitializeResult{Plugin: protocol.Peer{Name: "xlang", Version: "1.0.0"}, Kind: protocol.KindLanguage, Languages: []string{"x"}, Manifests: []string{"x.json"}}, nil
}

func (manifestLanguage) Discover(p protocol.DiscoverParams) (protocol.DiscoverResult, error) {
	var res protocol.DiscoverResult
	for _, m := range p.Manifests {
		root := path.Dir(m)
		res.Projects = append(res.Projects, protocol.Project{Name: path.Base(root), Root: root, Language: "x", Kind: protocol.ProjectApp})
	}
	return res, nil
}

func TestDiscoverLeavesOutTheExcludedDirectories(t *testing.T) {
	set := &pluginhost.PluginSet{Clients: []*host.Client{connect(t, manifestLanguage{})}}
	cfg := workspace.Config{Workspace: workspace.WorkspaceSection{Name: "w", Exclude: []string{"examples/*"}}}
	catalog, err := pluginhost.NewCatalog(set, cfg)
	if err != nil {
		t.Fatal(err)
	}
	files := []string{"apps/web/x.json", "examples/movies/x.json", "examples/shop/x.json", "examples/x.json"}
	g, err := Discover(files, set, catalog, cfg)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, p := range g.Projects {
		names = append(names, p.Name)
	}
	// examples/x.json is in examples itself, which examples/* does not match.
	if !reflect.DeepEqual(names, []string{"examples", "web"}) {
		t.Errorf("projects %v", names)
	}
	if !g.Excluded("examples/shop/run.mjs") || g.Excluded("apps/web/a.ts") {
		t.Error("the graph does not report the excluded directories")
	}
}
