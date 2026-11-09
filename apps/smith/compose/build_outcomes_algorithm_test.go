package compose

import (
	"reflect"
	"testing"
)

func outcomeItems() []Item {
	return []Item{
		{Kind: KindLanguage, Name: "node", Title: "TypeScript / JavaScript"},
		{Kind: KindLanguage, Name: "python", Title: "Python"},
		{Kind: KindLanguage, Name: "go", Title: "Go"},
		{Kind: KindChoice, Name: "react", Title: "React", Layer: "framework", Language: "node", Aliases: []string{"ui"}},
		{Kind: KindChoice, Name: "tauri", Title: "Tauri", Layer: "shell", Language: "node", Aliases: []string{"desktop"}},
		{Kind: KindChoice, Name: "azure-functions", Title: "Azure Functions", Layer: "host", Language: "node", Needs: []string{"trigger"}},
		{Kind: KindChoice, Name: "azure-functions", Title: "Azure Functions", Layer: "host", Language: "python", Needs: []string{"trigger"}},
		{Kind: KindChoice, Name: "http-trigger", Title: "HTTP trigger", Layer: "host", Language: "node"},
		{Kind: KindChoice, Name: "gcp-functions", Title: "Google Cloud Functions", Layer: "host", Language: "node", Needs: []string{"trigger"}, Problem: "off"},
	}
}

func byID(outcomes []Outcome, id string) (Outcome, bool) {
	for _, o := range outcomes {
		if o.ID == id {
			return o, true
		}
	}
	return Outcome{}, false
}

func TestOutcomesTakeTheirLanguagesAndWordsFromTheCatalogue(t *testing.T) {
	out := Outcomes(outcomeItems())
	web, _ := byID(out, "web")
	if !reflect.DeepEqual(web.Languages, []string{"node"}) || !reflect.DeepEqual(web.Typical, []string{"React"}) {
		t.Fatalf("a website is what the framework layer holds: %+v", web)
	}
	fn, _ := byID(out, "function")
	if !reflect.DeepEqual(fn.Languages, []string{"node", "python"}) || len(fn.Items) != 2 {
		t.Fatalf("a function app needs a host: no trigger, no host that cannot be used: %+v", fn)
	}
	desktop, _ := byID(out, "desktop")
	for _, want := range []string{"tauri", "Tauri", "desktop"} {
		found := false
		for _, w := range desktop.Words {
			found = found || w == want
		}
		if !found {
			t.Errorf("the desktop outcome is not found by %q: %v", want, desktop.Words)
		}
	}
	api, _ := byID(out, "api")
	if !reflect.DeepEqual(api.Languages, []string{"node", "python", "go"}) || api.Kind != "app" {
		t.Fatalf("an API keeps the backend languages that exist, in the order people look for them: %+v", api)
	}
	if lib, _ := byID(out, "library"); lib.Kind != "lib" || len(lib.Languages) != 3 {
		t.Fatalf("a library fits every language: %+v", lib)
	}
	if pkg, _ := byID(out, "package"); pkg.Kind != "package" {
		t.Fatalf("a package is published: %+v", pkg)
	}
}

func TestAnOutcomeWithNothingOnItsLayerIsLeftOut(t *testing.T) {
	var items []Item
	for _, it := range outcomeItems() {
		if it.Kind == KindLanguage {
			items = append(items, it)
		}
	}
	out := Outcomes(items)
	for _, gone := range []string{"web", "function", "desktop"} {
		if _, ok := byID(out, gone); ok {
			t.Errorf("%s is offered with no framework, host or shell to build it", gone)
		}
	}
	if _, ok := byID(out, "api"); !ok {
		t.Error("an API needs only a language")
	}
	if got := Outcomes(nil); len(got) != 0 {
		t.Fatalf("no language, no outcome: %+v", got)
	}
}
