package prompt

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSaveThenLoadReturnsTheSummary(t *testing.T) {
	root := t.TempDir()
	want := Summary{Schema: schemaVersion, Workspace: root, CheckedAt: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC), Failed: 1, Warnings: 2,
		Projects: []Project{{Name: "api", Root: "apps/api", Dependents: []string{"web"}}}}
	if err := Save(root, want); err != nil {
		t.Fatal(err)
	}
	got, ok := Load(root)
	if !ok || got.Failed != 1 || got.Warnings != 2 || len(got.Projects) != 1 || got.Projects[0].Dependents[0] != "web" || !got.CheckedAt.Equal(want.CheckedAt) {
		t.Fatalf("got %+v, %v", got, ok)
	}
}

func TestLoadIsFalseWithoutAFile(t *testing.T) {
	if _, ok := Load(t.TempDir()); ok {
		t.Fatal("no file must read as no summary")
	}
}

func TestLoadIgnoresASummaryOfAnotherWorkspaceOrShape(t *testing.T) {
	root := t.TempDir()
	if err := Save(root, Summary{Schema: schemaVersion, Workspace: "/somewhere/else"}); err != nil {
		t.Fatal(err)
	}
	if _, ok := Load(root); ok {
		t.Fatal("a summary of another workspace must not be used")
	}
	if err := Save(root, Summary{Schema: schemaVersion + 1, Workspace: root}); err != nil {
		t.Fatal(err)
	}
	if _, ok := Load(root); ok {
		t.Fatal("a summary of another shape must not be used")
	}
	if err := os.WriteFile(filepath.Join(root, ".smith", "prompt.json"), []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, ok := Load(root); ok {
		t.Fatal("a broken file must read as no summary")
	}
}
