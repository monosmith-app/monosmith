package main

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/monosmith-app/monosmith/libs/words/surface"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// fixture resolves the fixture catalogue for one fixture surface manifest.
func fixture(t *testing.T, name string) Resolved {
	t.Helper()
	catalogs, err := words.LoadCatalog(os.DirFS(filepath.Join("fixtures", "catalog")))
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join("fixtures", "surface", name+".json"))
	if err != nil {
		t.Fatal(err)
	}
	m, err := surface.Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	r, err := Resolve(catalogs, m)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestResolveClaimsKeysByExactNameAndByPrefix(t *testing.T) {
	for name, want := range map[string][]string{
		"vscode":    {"cmd.add.title", "cmd.remove.title", "vscode.quote"},
		"jetbrains": {"cmd.items.one", "cmd.items.other", "cmd.remove.title"},
	} {
		if got := fixture(t, name).Keys; !slices.Equal(got, want) {
			t.Errorf("%s keys = %v, want %v", name, got, want)
		}
	}
}

func TestResolveFallsBackToEnglishKeyByKey(t *testing.T) {
	r := fixture(t, "vscode")
	pt := r.Texts["pt-BR"]
	if got := pt["cmd.remove.title"]; !got.Own || got.Text != "Remover {project} de {name} — ação" {
		t.Errorf("translated key = %+v", got)
	}
	if got := pt["cmd.add.title"]; got.Own || got.Text != "Add a package" {
		t.Errorf("a key pt-BR lacks must show English: %+v", got)
	}
	if !slices.Equal(r.Locales, []string{"pt-BR"}) {
		t.Errorf("locales = %v", r.Locales)
	}
}

func TestAnEmptySurfaceResolvesToNothing(t *testing.T) {
	r, err := Resolve(map[string]map[string]string{"en": {"a": "A"}, "pt": {"a": "Á"}}, surface.Manifest{})
	if err != nil || len(r.Keys) != 0 || len(r.Locales) != 0 {
		t.Fatalf("got %+v, %v", r, err)
	}
}

func TestResolveRefusesWhatTheCatalogueCannotServe(t *testing.T) {
	en := map[string]string{"a": "A {name}"}
	for name, c := range map[string]struct {
		catalogs map[string]map[string]string
		m        surface.Manifest
	}{
		"a claimed key English lacks":               {map[string]map[string]string{"en": en}, surface.Manifest{Keys: []string{"missing"}}},
		"a translation with an unknown placeholder": {map[string]map[string]string{"en": en, "pt": {"a": "A {other}"}}, surface.Manifest{Keys: []string{"a"}}},
	} {
		if _, err := Resolve(c.catalogs, c.m); err == nil {
			t.Errorf("%s must be an error", name)
		}
	}
}
