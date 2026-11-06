package compose

import (
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/choiceselect"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestBuildPlacesEveryItemOnALayerInStackOrder(t *testing.T) {
	cat := Build(Source{
		Offers:    fakeOffers(),
		Languages: []Language{{Name: "node", Title: "TypeScript / JavaScript"}, {Name: "python", Title: "Python"}},
		Plugins:   []WorkspacePlugin{{Name: "slices", Layer: "architecture", Description: "structure plugin"}, {Name: "gh", Layer: "ci"}},
	})
	var ids []string
	for _, l := range cat.Layers {
		ids = append(ids, l.ID)
	}
	want := []string{"language", "framework", "shell", "host", "style", "test", "lint", "architecture", "ci"}
	if len(ids) != len(want) {
		t.Fatalf("layers = %v, want %v", ids, want)
	}
	for i := range want {
		if ids[i] != want[i] {
			t.Fatalf("layers = %v, want %v", ids, want)
		}
	}
	first := cat.Items[0]
	if first.Kind != KindLanguage || first.Layer != "language" {
		t.Fatalf("the language comes first: %+v", first)
	}
}

func TestBuildTakesTheLayerFromTheChoiceThenTheAxisAndTheTitleFromTheNameWhenNoneIsGiven(t *testing.T) {
	cat := Build(Source{Offers: fakeOffers(), Languages: []Language{{Name: "node"}, {Name: "python"}}})
	electron, ok := cat.Find("node/electron")
	if !ok || electron.Layer != "shell" || electron.Title != "Electron" {
		t.Fatalf("electron sits on the shell layer: %+v", electron)
	}
	bootstrap, _ := cat.Find("node/bootstrap")
	if bootstrap.Layer != "style" || bootstrap.Title != "bootstrap" {
		t.Fatalf("no layer means the axis, no title means the name: %+v", bootstrap)
	}
}

func TestBuildKeepsTheSameNameForTwoLanguagesAsTwoItems(t *testing.T) {
	cat := Build(Source{Offers: fakeOffers(), Languages: []Language{{Name: "node"}, {Name: "python"}}})
	if _, ok := cat.Find("azure-functions"); ok {
		t.Fatal("a bare name that two languages offer is not one item")
	}
	if _, ok := cat.Find("python/azure-functions"); !ok {
		t.Fatal("the id is language/name")
	}
}

func TestBuildMarksAChoiceWhoseLanguagePluginIsNotLoaded(t *testing.T) {
	cat := Build(Source{Offers: fakeOffers(), Languages: []Language{{Name: "node"}}})
	it, _ := cat.Find("python/azure-functions")
	if it.Problem == "" {
		t.Fatal("python is not loaded, so its items say why they cannot be used")
	}
	if ok, _ := cat.Find("node/react"); ok.Problem != "" {
		t.Fatalf("node items work: %q", ok.Problem)
	}
}

func TestAPluginsOwnLayerNeedsNoChangeInTheHost(t *testing.T) {
	offers := fakeOffers()
	offers[0].Layer = "database"
	cat := Build(Source{Offers: offers, Languages: []Language{{Name: "node"}}})
	last := cat.Layers[len(cat.Layers)-1]
	if last.ID != "database" {
		t.Fatalf("an unknown layer comes after the known ones: %+v", cat.Layers)
	}
}

func TestOffersOfBringsTheChoicesBackForTheSolver(t *testing.T) {
	cat := Build(Source{Offers: fakeOffers(), Languages: []Language{{Name: "node"}, {Name: "python"}}})
	got := OffersOf(cat.Items)
	if len(got) != len(fakeOffers()) {
		t.Fatalf("one offer for each choice, none for languages: %d", len(got))
	}
	if check := CheckPicks(got, "node", []string{"react"}); !check.Complete {
		t.Fatalf("the offers rebuilt from the catalogue solve like the originals: %+v", check)
	}
}

// offer is a choice of a fake plugin.
func offer(plugin string, c protocol.Choice) choiceselect.Offer {
	if c.Language == "" {
		c.Language = "node"
	}
	return choiceselect.Offer{Plugin: plugin, Choice: c}
}

var scaffoldRole = []protocol.Role{protocol.RoleScaffold}
var toolRole = []protocol.Role{protocol.RoleToolchain}

// fakeOffers are the choices of three fake plugins: a web starter, a desktop
// shell, a function host, and the tools they need.
func fakeOffers() []choiceselect.Offer {
	return []choiceselect.Offer{
		offer("vite", protocol.Choice{Axis: "framework", Name: "react", Title: "React", Aliases: []string{"web", "ui"}, Description: "a React app", Roles: append(append([]protocol.Role{}, scaffoldRole...), protocol.RoleToolchain), Needs: []string{"lint", "test"}}),
		offer("electron", protocol.Choice{Axis: "framework", Name: "electron", Title: "Electron", Layer: "shell", Aliases: []string{"desktop", "app"}, Description: "a desktop app", Roles: scaffoldRole, Needs: []string{"lint", "test"}, Conflicts: []string{"azure-functions"}}),
		offer("functions", protocol.Choice{Axis: "framework", Name: "azure-functions", Title: "Azure Functions", Layer: "host", Aliases: []string{"serverless", "cloud"}, Description: "an Azure function app", Roles: scaffoldRole, Needs: []string{"trigger", "lint", "test"}}),
		offer("functions", protocol.Choice{Axis: "trigger", Name: "http-trigger", Title: "HTTP trigger", Layer: "host", Suggested: true, Description: "answers a request"}),
		offer("functions", protocol.Choice{Axis: "trigger", Name: "timer-trigger", Title: "Timer trigger", Layer: "host", Description: "runs on a schedule"}),
		offer("eslint", protocol.Choice{Axis: "lint", Name: "eslint", Title: "ESLint", Roles: toolRole, Default: true, Description: "lint"}),
		offer("vitest", protocol.Choice{Axis: "test", Name: "vitest", Title: "Vitest", Roles: toolRole, Default: true, Description: "test"}),
		offer("vite", protocol.Choice{Axis: "style", Name: "bootstrap", Description: "css"}),
		offer("py", protocol.Choice{Axis: "framework", Name: "azure-functions", Language: "python", Title: "Azure Functions", Layer: "host", Aliases: []string{"serverless"}, Roles: scaffoldRole, Description: "python function app"}),
	}
}
