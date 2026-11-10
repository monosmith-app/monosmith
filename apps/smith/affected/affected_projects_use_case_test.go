package affected

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/gitclient"
	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func setup(t *testing.T) (*gitclient.Git, *projectgraph.Graph) {
	t.Helper()
	g := gitclient.New(t.TempDir())
	if err := g.Init(); err != nil {
		t.Skip("git unavailable")
	}
	for _, kv := range [][2]string{{"user.email", "t@e.com"}, {"user.name", "T"}, {"commit.gpgsign", "false"}} {
		_, _ = g.Run("config", kv[0], kv[1])
	}
	write := func(rel string) {
		p := filepath.Join(g.Dir, filepath.FromSlash(rel))
		_ = os.MkdirAll(filepath.Dir(p), 0o755)
		_ = os.WriteFile(p, []byte(rel+"x"), 0o644)
	}
	for _, f := range []string{"apps/web/a", "libs/ui/a", "libs/core/a", "apps/api/a", "README.md", "smith.toml"} {
		write(f)
	}
	if err := g.CommitAll("init"); err != nil {
		t.Fatal(err)
	}
	_, _ = g.Run("branch", "base")
	mk := func(name, root string, deps ...string) *projectgraph.Project {
		return &projectgraph.Project{Project: protocol.Project{Name: name, Root: root}, Deps: deps}
	}
	pg, _ := projectgraph.New([]*projectgraph.Project{mk("web", "apps/web", "ui"), mk("ui", "libs/ui", "core"), mk("core", "libs/core"), mk("api", "apps/api")})
	return g, pg
}

func TestAffectedFollowsDependents(t *testing.T) {
	g, pg := setup(t)
	_ = os.WriteFile(filepath.Join(g.Dir, "libs", "ui", "a"), []byte("changed"), 0o644)
	_ = os.WriteFile(filepath.Join(g.Dir, "README.md"), []byte("docs only"), 0o644)
	r := Compute(g, pg, "base")
	if r.All || !reflect.DeepEqual(r.Projects, []string{"ui", "web"}) {
		t.Fatalf("result %+v", r)
	}
}

func TestAChangeInAnExcludedDirectoryAffectsNothing(t *testing.T) {
	g, pg := setup(t)
	pg.SetExcluded(workspace.Config{Workspace: workspace.WorkspaceSection{Exclude: []string{"examples/*"}}}.Excluded)
	p := filepath.Join(g.Dir, "examples", "movies", "run.mjs")
	_ = os.MkdirAll(filepath.Dir(p), 0o755)
	_ = os.WriteFile(p, []byte("changed"), 0o644)
	if r := Compute(g, pg, "base"); r.All || len(r.Projects) != 0 {
		t.Fatalf("an excluded directory's change affected %+v", r)
	}
}

func TestAffectedFallsBackToEverything(t *testing.T) {
	g, pg := setup(t)
	if r := Compute(g, pg, "no-such-branch"); !r.All {
		t.Fatalf("an unknown base must verify everything: %+v", r)
	}
	if r := Compute(g, pg, ""); !r.All {
		t.Fatal("no base must verify everything")
	}
	_ = os.WriteFile(filepath.Join(g.Dir, "smith.toml"), []byte("changed"), 0o644)
	if r := Compute(g, pg, "base"); !r.All || len(r.Projects) != 4 {
		t.Fatalf("a workspace file change must affect everything: %+v", r)
	}
}

func TestNoChanges(t *testing.T) {
	g, pg := setup(t)
	r := Compute(g, pg, "base")
	if r.All || len(r.Projects) != 0 {
		t.Fatalf("result %+v", r)
	}
}

func TestAChangedSchemaAffectsItsOwnerAndEveryConsumer(t *testing.T) {
	g := gitclient.New(t.TempDir())
	if err := g.Init(); err != nil {
		t.Skip("git unavailable")
	}
	for _, kv := range [][2]string{{"user.email", "t@e.com"}, {"user.name", "T"}, {"commit.gpgsign", "false"}} {
		_, _ = g.Run("config", kv[0], kv[1])
	}
	files := []string{"services/orders/openapi.yaml", "contracts/events.proto", "clients/go/main.go", "apps/web/app.ts", "apps/docs/a.md", "smith.toml"}
	for _, rel := range files {
		p := filepath.Join(g.Dir, filepath.FromSlash(rel))
		_ = os.MkdirAll(filepath.Dir(p), 0o755)
		_ = os.WriteFile(p, []byte(rel), 0o644)
	}
	if err := g.CommitAll("init"); err != nil {
		t.Fatal(err)
	}
	_, _ = g.Run("branch", "base")
	mk := func(name, root string) *projectgraph.Project {
		return &projectgraph.Project{Project: protocol.Project{Name: name, Root: root}}
	}
	projects := []*projectgraph.Project{mk("orders", "services/orders"), mk("go-client", "clients/go"), mk("web", "apps/web"), mk("docs", "apps/docs")}
	cfg := workspace.Config{Projects: map[string]workspace.ProjectOverride{
		"orders":    {Schemas: []string{"contracts/*.proto"}},
		"go-client": {ConsumesSchemas: []string{"services/orders/openapi.yaml"}},
		"web":       {ConsumesSchemas: []string{"services/orders/openapi.yaml", "contracts/events.proto"}},
	}}
	projectgraph.ResolveSchemaEdges(projects, files, cfg)
	pg, _ := projectgraph.New(projects)

	spec := filepath.Join(g.Dir, "services", "orders", "openapi.yaml")
	_ = os.WriteFile(spec, []byte("changed"), 0o644)
	if r := Compute(g, pg, "base"); r.All || !reflect.DeepEqual(r.Projects, []string{"go-client", "orders", "web"}) {
		t.Fatalf("an OpenAPI change affected %+v", r)
	}
	_ = os.WriteFile(spec, []byte("services/orders/openapi.yaml"), 0o644)
	// A schema outside every project's directory still names its owner, and
	// the owner's dependents follow (the Go client depends on orders).
	_ = os.WriteFile(filepath.Join(g.Dir, "contracts", "events.proto"), []byte("changed"), 0o644)
	if r := Compute(g, pg, "base"); r.All || !reflect.DeepEqual(r.Projects, []string{"go-client", "orders", "web"}) {
		t.Fatalf("a proto change affected %+v", r)
	}
}
