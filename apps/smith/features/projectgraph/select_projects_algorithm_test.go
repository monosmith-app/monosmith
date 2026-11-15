package projectgraph

import (
	"reflect"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestSelect(t *testing.T) {
	mk := func(name, root, lang string, kind protocol.ProjectKind) *Project {
		return &Project{Project: protocol.Project{Name: name, Root: root, Language: lang, Kind: kind, Tags: []string{"team:" + lang}}, Plugin: lang}
	}
	g, _ := New([]*Project{mk("web", "apps/web", "node", protocol.ProjectApp), mk("ui", "libs/ui", "node", protocol.ProjectLib), mk("api", "apps/api", "go", protocol.ProjectApp), mk("sdk", "packages/sdk", "python", protocol.ProjectPackage)})
	cases := map[string][]string{
		"web":             {"web"},
		"apps/*":          {"api", "web"},
		"kind:app,!api":   {"web"},
		"lang:node":       {"ui", "web"},
		"tag:team:python": {"sdk"},
		"!ui":             {"api", "sdk", "web"},
		"":                {"api", "sdk", "ui", "web"},
	}
	for pattern, want := range cases {
		got, err := g.Select([]string{pattern})
		if err != nil || !reflect.DeepEqual(got, want) {
			t.Errorf("%q: got %v %v want %v", pattern, got, err, want)
		}
	}
	if _, err := g.Select([]string{"nope"}); err == nil {
		t.Fatal("an unmatched pattern must fail")
	}
}
