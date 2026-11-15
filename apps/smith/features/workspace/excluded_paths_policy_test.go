package workspace

import "testing"

func TestExcludedCoversTheMatchedDirectoriesAndBelow(t *testing.T) {
	cfg := Config{Workspace: WorkspaceSection{Exclude: []string{"examples/*", "tools/scratch", "fixtures/**"}}}
	for file, want := range map[string]bool{
		"examples/movies/package.json":      true,
		"examples/movies/deep/package.json": true,
		"examples/README.md":                false,
		"tools/scratch/go.mod":              true,
		"tools/other/go.mod":                false,
		"fixtures/package.json":             true,
		"fixtures/a/b/Cargo.toml":           true,
		"apps/web/package.json":             false,
		"package.json":                      false,
	} {
		if got := cfg.Excluded(file); got != want {
			t.Errorf("Excluded(%s) = %v, want %v", file, got, want)
		}
	}
	deep := Config{Workspace: WorkspaceSection{Exclude: []string{"examples/**"}}}
	if !deep.Excluded("examples/README.md") || !deep.Excluded("examples/movies/package.json") {
		t.Error("examples/** must cover examples itself and everything below")
	}
	if (Config{}).Excluded("examples/movies/package.json") {
		t.Error("nothing is excluded without [workspace] exclude")
	}
}

func TestValidateRejectsAnExcludeThatIsNotAWorkspaceGlob(t *testing.T) {
	for _, glob := range []string{"/abs/*", `examples\*`, "examples/["} {
		cfg := Config{Workspace: WorkspaceSection{Name: "a", Exclude: []string{glob}}}
		if err := Validate(&cfg); err == nil {
			t.Errorf("%q accepted", glob)
		}
	}
	cfg := Config{Workspace: WorkspaceSection{Name: "a", Exclude: []string{"examples/**"}}}
	if err := Validate(&cfg); err != nil {
		t.Error(err)
	}
}
