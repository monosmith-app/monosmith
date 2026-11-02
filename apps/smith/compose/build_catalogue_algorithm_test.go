package compose

import (
	"testing"
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
