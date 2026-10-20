package graphview

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// schemaWorkspace is a JVM service that owns an OpenAPI spec, a Go client and
// a TypeScript frontend that consume it, and a project nothing links to it.
func schemaWorkspace(t *testing.T) *projectgraph.Graph {
	t.Helper()
	mk := func(name, root, plugin string) *projectgraph.Project {
		return &projectgraph.Project{Project: protocol.Project{Name: name, Root: root}, Plugin: plugin}
	}
	projects := []*projectgraph.Project{
		mk("orders-api", "services/orders", "jvm"), mk("orders-go", "clients/go", "go"),
		mk("storefront", "apps/storefront", "node"), mk("docs", "apps/docs", "node"),
	}
	files := []string{"services/orders/openapi.yaml", "clients/go/main.go", "apps/storefront/package.json", "apps/docs/package.json"}
	cfg := workspace.Config{Projects: map[string]workspace.ProjectOverride{
		"orders-go":  {ConsumesSchemas: []string{"services/orders/openapi.yaml"}},
		"storefront": {ConsumesSchemas: []string{"services/**/openapi.yaml"}},
	}}
	projectgraph.ResolveEdges(projects)
	projectgraph.ResolveSchemaEdges(projects, files, cfg)
	g, err := projectgraph.New(projects)
	if err != nil {
		t.Fatal(err)
	}
	return g
}

func TestBlastRadiusOfASchemaFileReachesEveryClient(t *testing.T) {
	g := schemaWorkspace(t)
	for _, target := range []string{"services/orders/openapi.yaml", "orders-api"} {
		r, err := BlastRadius(g, target)
		if err != nil {
			t.Fatal(err)
		}
		if len(r.Roots) != 1 || r.Roots[0] != "orders-api" || len(r.Reached) != 2 {
			t.Fatalf("%s: %+v", target, r)
		}
		for _, x := range r.Reached {
			if x.Via != "schema" || x.From != "orders-api" || x.Name == "docs" {
				t.Fatalf("%s: %+v", target, x)
			}
		}
	}
	if text := BlastRadiusText(mustRadius(t, g, "orders-api")); !strings.Contains(text, "storefront") || !strings.Contains(text, "via schema of orders-api") {
		t.Fatal(text)
	}
}

func TestBlastRadiusOfALeafAndOfAnUnknownTarget(t *testing.T) {
	g := schemaWorkspace(t)
	if r := mustRadius(t, g, "docs"); len(r.Reached) != 0 || !strings.Contains(BlastRadiusText(r), "nothing else") {
		t.Fatalf("%+v", r)
	}
	if _, err := BlastRadius(g, "nowhere/file.txt"); err == nil {
		t.Fatal("an unknown target must be an error")
	}
}

func mustRadius(t *testing.T, g *projectgraph.Graph, target string) Radius {
	t.Helper()
	r, err := BlastRadius(g, target)
	if err != nil {
		t.Fatal(err)
	}
	return r
}
