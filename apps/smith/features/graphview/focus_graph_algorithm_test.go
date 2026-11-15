package graphview

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/features/projectgraph"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// chain: web → sdk → ui, and a separate tool.
func chain(t *testing.T) *projectgraph.Graph {
	t.Helper()
	mk := func(name string, deps ...string) *projectgraph.Project {
		return &projectgraph.Project{Project: protocol.Project{Name: name, Root: "x/" + name, Kind: protocol.ProjectLib, Language: "node"}, Deps: deps, Plugin: "node"}
	}
	g, err := projectgraph.New([]*projectgraph.Project{mk("web", "sdk"), mk("sdk", "ui"), mk("ui"), mk("tool")})
	if err != nil {
		t.Fatal(err)
	}
	return g
}

func TestAroundKeepsWhatTouchesTheFocus(t *testing.T) {
	g := Around(chain(t), []string{"sdk"})
	if strings.Join(g.Names(), ",") != "sdk,ui,web" {
		t.Fatalf("names %v", g.Names())
	}
	if ui := Around(chain(t), []string{"ui"}); strings.Join(ui.Names(), ",") != "sdk,ui,web" {
		t.Fatalf("ui's dependents %v", ui.Names())
	}
	sub := Subgraph(chain(t), map[string]bool{"web": true, "ui": true})
	if len(sub.Get("web").Deps) != 0 {
		t.Fatal("an edge to a dropped project is dropped")
	}
}

func TestMarkedProjectsInEveryFormat(t *testing.T) {
	g := chain(t)
	marked := map[string]bool{"sdk": true}
	var parsed struct{ Projects []Node }
	_ = json.Unmarshal([]byte(JSON(g, marked)), &parsed)
	for _, n := range parsed.Projects {
		if n.Marked != (n.Name == "sdk") {
			t.Errorf("json %s marked=%v", n.Name, n.Marked)
		}
	}
	if !strings.Contains(Text(g, marked), "sdk *  lib") || !strings.Contains(DOT(g, marked), `"sdk" [label="sdk\nlib · node", style="rounded,bold"`) || !strings.Contains(HTML(g, "x", marked), `fill="#64748b33"`) {
		t.Fatal("a marked project stands out in text, dot and html")
	}
}
