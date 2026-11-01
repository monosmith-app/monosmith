package divergences

import (
	"bytes"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/derive"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestFromListsOverridesEjectedDriftAndDifferingFiles(t *testing.T) {
	no := false
	cfg := workspace.Config{
		Tasks: map[string]workspace.TaskOverride{"build": {Outputs: []string{"dist"}}},
		Projects: map[string]workspace.ProjectOverride{
			"web": {Stack: "react", Tasks: map[string]workspace.TaskOverride{"test": {Cache: &no, Env: map[string]string{"A": "1"}}}},
		},
	}
	items := []derive.Item{
		{File: protocol.DerivedFile{Path: "a.json"}, Status: derive.StatusEjected},
		{File: protocol.DerivedFile{Path: "b.json"}, Status: derive.StatusDrift, Detail: "edited by hand"},
		{File: protocol.DerivedFile{Path: "c.json"}, Status: derive.StatusCurrent},
	}
	diverging := []derive.Divergence{{DivergentFile: protocol.DivergentFile{Path: "d.json", Detail: "has no extends"}}}
	got := From(cfg, items, diverging)
	var subjects []string
	for _, d := range got {
		subjects = append(subjects, string(d.Kind)+":"+d.Subject)
		if d.Why == "" || d.Undo == "" {
			t.Errorf("%s lacks a why or an undo", d.Subject)
		}
	}
	want := "override:all projects, build outputs override:web stack override:web test cache override:web test env ejected:a.json drift:b.json differs:d.json"
	if strings.Join(subjects, " ") != want {
		t.Errorf("got  %s\nwant %s", strings.Join(subjects, " "), want)
	}
}

func TestAnEmptyListIsGoodNews(t *testing.T) {
	var b bytes.Buffer
	Print(&b, From(workspace.Config{}, nil, nil))
	if !strings.Contains(b.String(), "No divergences") {
		t.Errorf("%q", b.String())
	}
}
