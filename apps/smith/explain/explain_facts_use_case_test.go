package explain

import (
	"bytes"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/factmerge"
	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func graphWithFacts(t *testing.T) *projectgraph.Graph {
	t.Helper()
	p := &projectgraph.Project{Project: protocol.Project{Name: "web", Root: "apps/web"}, Plugin: "node"}
	p.FactDetail = map[string]factmerge.Fact{
		"module.aliases": {Value: map[string]any{"@app": "src"}, Scope: factmerge.ScopeWorkspace, Source: "smith.toml [facts]"},
		"source.roots":   {Value: []any{"src"}, Scope: factmerge.ScopePluginDefault, Source: "plugin typescript (stack node-jest)"},
	}
	g, err := projectgraph.New([]*projectgraph.Project{p})
	if err != nil {
		t.Fatal(err)
	}
	return g
}

func TestExplainFactsShowsValueAndSource(t *testing.T) {
	g := graphWithFacts(t)
	facts, err := ExplainFacts(g, "web", "")
	if err != nil || len(facts) != 2 {
		t.Fatalf("%v %v", facts, err)
	}
	var out bytes.Buffer
	printFacts(&out, "web", facts)
	for _, want := range []string{`module.aliases  {"@app":"src"}`, "(workspace, from smith.toml [facts])", "(plugin default, from plugin typescript (stack node-jest))"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("output lacks %q:\n%s", want, out.String())
		}
	}
}

func TestExplainFactsOneKeyAndErrors(t *testing.T) {
	g := graphWithFacts(t)
	if facts, err := ExplainFacts(g, "web", "source.roots"); err != nil || len(facts) != 1 {
		t.Fatalf("%v %v", facts, err)
	}
	if _, err := ExplainFacts(g, "web", "test.patterns"); err == nil {
		t.Error("a fact no one set is an error")
	}
	if _, err := ExplainFacts(g, "nope", ""); err == nil {
		t.Error("unknown project")
	}
}
