package projectgraph

import (
	"reflect"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

func TestSchemaEdgesLinkTheOwnerToItsConsumers(t *testing.T) {
	ps := []*Project{
		proj("api", "services/api", nil, nil),
		proj("go-client", "clients/go", nil, nil),
		proj("web", "apps/web", nil, nil),
		proj("events", "services/events", nil, nil),
		proj("other", "apps/other", nil, nil),
	}
	files := []string{
		"services/api/openapi.yaml", "services/events/asyncapi.yaml", "contracts/user.proto", "contracts/readme.md",
		"clients/go/go.mod", "apps/web/package.json", "apps/other/package.json",
	}
	cfg := workspace.Config{Projects: map[string]workspace.ProjectOverride{
		"events":    {Schemas: []string{"contracts/*.proto"}},
		"go-client": {ConsumesSchemas: []string{"services/api/openapi.yaml", "contracts/*.proto"}},
		"web":       {ConsumesSchemas: []string{"services/api/openapi.yaml", "apps/web/package.json"}},
	}}
	ResolveEdges(ps)
	ResolveSchemaEdges(ps, files, cfg)
	g, err := New(ps)
	if err != nil {
		t.Fatal(err)
	}
	if got := g.Get("api").Schemas; !reflect.DeepEqual(got, []string{"services/api/openapi.yaml"}) {
		t.Fatalf("api schemas %v", got)
	}
	// events owns its asyncapi file by being its root, and the proto by glob.
	if got := g.Get("events").Schemas; !reflect.DeepEqual(got, []string{"contracts/user.proto", "services/events/asyncapi.yaml"}) {
		t.Fatalf("events schemas %v", got)
	}
	if got := g.Get("go-client").Deps; !reflect.DeepEqual(got, []string{"api", "events"}) {
		t.Fatalf("go-client deps %v", got)
	}
	if got := g.Get("web").Deps; !reflect.DeepEqual(got, []string{"api"}) {
		t.Fatalf("web deps %v (a file of its own root is no dependency)", got)
	}
	if len(g.Get("other").Deps) != 0 || !reflect.DeepEqual(g.TransitiveDependents("api"), []string{"go-client", "web"}) {
		t.Fatal("unlinked project reached")
	}
	if g.SchemaOwner("contracts/user.proto").Name != "events" || g.SchemaOwner("contracts/readme.md") != nil {
		t.Fatal("schema owner")
	}
	if _, err := g.Order(); err != nil {
		t.Fatal(err)
	}
}
