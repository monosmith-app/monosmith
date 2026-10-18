package cipipeline

import (
	"reflect"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func at(name, root string) *projectgraph.Project {
	return &projectgraph.Project{Project: protocol.Project{Name: name, Root: root}}
}

func TestTrackedChangesNamesTheProjectAndItsTasks(t *testing.T) {
	projects := []*projectgraph.Project{at("eslint-config", "packages/eslint-config"), at("smith", "apps/smith")}
	before := []string{"package-lock.json"}
	after := []string{"packages/eslint-config/tsconfig.typecheck.tsbuildinfo", "package-lock.json"}
	ran := []string{"eslint-config:typecheck", "eslint-config:lint", "smith:test"}
	got := TrackedChanges(before, after, projects, ran)
	want := []TrackedChange{
		{File: "package-lock.json", Before: true},
		{File: "packages/eslint-config/tsconfig.typecheck.tsbuildinfo", Project: "eslint-config", Tasks: []string{"lint", "typecheck"}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v\nwant %+v", got, want)
	}
	msg := ExplainTrackedChanges(got)
	for _, part := range []string{
		"packages/eslint-config/tsconfig.typecheck.tsbuildinfo (in eslint-config, whose lint, typecheck ran)",
		"package-lock.json (already changed before verify started",
		"add it to .gitignore and `git rm --cached` it",
	} {
		if !strings.Contains(msg, part) {
			t.Errorf("message lacks %q:\n%s", part, msg)
		}
	}
}

func TestTrackedChangesPicksTheDeepestProject(t *testing.T) {
	projects := []*projectgraph.Project{at("web", "apps/web"), at("web-e2e", "apps/web/e2e")}
	got := TrackedChanges(nil, []string{"apps/web/e2e/snap.txt"}, projects, nil)
	if len(got) != 1 || got[0].Project != "web-e2e" {
		t.Errorf("got %+v", got)
	}
}
