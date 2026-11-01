package adoption

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkspaceName(t *testing.T) {
	for dir, want := range map[string]string{
		"Legacy_Repo": "legacy-repo",
		"my.app":      "my-app",
		"42-things":   "workspace-42-things",
		"acme":        "acme",
	} {
		if got := workspaceName(filepath.Join("/repos", dir)); got != want {
			t.Errorf("%s: %s, want %s", dir, got, want)
		}
	}
}

func TestScopeAndProviders(t *testing.T) {
	root := t.TempDir()
	if got := scopeOf(root, "acme"); got != "@acme" {
		t.Errorf("no package.json: %s", got)
	}
	_ = os.WriteFile(filepath.Join(root, "package.json"), []byte(`{"name": "@globex/monorepo", "dependencies": {"zod": "^3", "chalk": "^5"}}`), 0o644)
	if got := scopeOf(root, "acme"); got != "@globex" {
		t.Errorf("scoped root: %s", got)
	}
	if deps := rootRuntimeDependencies(root); len(deps) != 2 || deps[0] != "chalk" {
		t.Errorf("deps %v", deps)
	}
	exists := func(set ...string) func(string) bool {
		return func(p string) bool {
			for _, s := range set {
				if s == p {
					return true
				}
			}
			return false
		}
	}
	if got := providersOf(exists()); len(got) != 1 || got[0] != "github" {
		t.Errorf("default %v", got)
	}
	if got := providersOf(exists("azure-pipelines.yml")); len(got) != 1 || got[0] != "azure" {
		t.Errorf("azure %v", got)
	}
	if got := providersOf(exists(".github/workflows", "azure-pipelines.yml")); len(got) != 2 {
		t.Errorf("both %v", got)
	}
}

func TestAdoptable(t *testing.T) {
	r := Report{Blockers: []Item{{Blocks: BlocksRelease}}}
	if !r.Adoptable() {
		t.Error("a release blocker does not stop --init")
	}
	r.Blockers = append(r.Blockers, Item{Blocks: BlocksAdoption})
	if r.Adoptable() {
		t.Error("an adoption blocker does")
	}
}

func inspectWith(t *testing.T, files ...string) Report {
	t.Helper()
	dir := t.TempDir()
	if out, err := exec.Command("git", "-C", dir, "init", "-q").CombinedOutput(); err != nil {
		t.Skipf("git unavailable: %v %s", err, out)
	}
	for _, f := range files {
		if err := os.WriteFile(filepath.Join(dir, f), []byte("{}"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	r, err := Inspect(dir, "0.0.0", io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestAdoptSendsNxAndTurborepoToConvert(t *testing.T) {
	for _, c := range []struct{ file, say, step string }{
		{"turbo.json", "Turborepo", "smith convert --dry-run"},
		{"nx.json", "Nx", "npm ci"},
	} {
		r := inspectWith(t, c.file)
		if len(r.Blockers) != 1 || !strings.Contains(r.Blockers[0].Message, c.say) || !strings.Contains(strings.Join(append(r.Blockers[0].Steps, r.Blockers[0].Step), " "), c.step) || r.Adoptable() {
			t.Errorf("%s: %+v", c.file, r.Blockers)
		}
	}
}

func TestNoProjectWarningExplainsARootPackage(t *testing.T) {
	has := func(set ...string) func(string) bool {
		return func(p string) bool {
			for _, s := range set {
				if s == p {
					return true
				}
			}
			return false
		}
	}
	root := noProjectWarning(has("package.json"))
	if !strings.Contains(root.Message, "package.json at the root describes the workspace, not a project") || len(root.Steps) != 2 {
		t.Errorf("root package: %+v", root)
	}
	goRoot := noProjectWarning(has("go.mod"))
	if !strings.Contains(goRoot.Message, "go.mod at the root is one module for the whole repository") || len(goRoot.Steps) != 2 || !strings.Contains(goRoot.Steps[0], "go work use") {
		t.Errorf("root go module: %+v", goRoot)
	}
	if both := noProjectWarning(has("package.json", "go.mod")); !strings.Contains(both.Message, "package.json at the root") {
		t.Errorf("a root package.json still comes first: %+v", both)
	}
	if bare := noProjectWarning(has()); !strings.Contains(bare.Steps[0], "smith adopt --init makes an empty workspace") {
		t.Errorf("no manifest: %+v", bare)
	}
}

func TestWhatSmithItselfLeavesInTheTreeDoesNotMakeItDirty(t *testing.T) {
	for status, want := range map[string]bool{
		"":                                  true,
		"?? .smith/\n":                      true,
		"?? .smith/defaults/golangci.yml\n": true,
		"?? .smith/\n M go.mod\n":           false,
		"?? notes.txt\n":                    false,
		"?? .smithy/\n":                     false,
	} {
		if got := cleanApartFromSmith(status); got != want {
			t.Errorf("%q: got %v, want %v", status, got, want)
		}
	}
}
