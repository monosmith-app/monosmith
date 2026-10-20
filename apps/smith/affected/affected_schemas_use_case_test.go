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
