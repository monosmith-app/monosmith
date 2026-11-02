package cipipeline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/operationplan"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

func render(p operationplan.Plan) string {
	var b strings.Builder
	operationplan.Render(&b, p)
	return b.String()
}

// A first generation creates files, writes nothing, and a second preview is empty
// once the files are written.
func TestPreviewListsWhatWriteWouldDoAndWritesNothing(t *testing.T) {
	root := t.TempDir()
	cfg := workspace.Config{CI: workspace.CISection{Providers: []string{"github"}}}
	plan, _, err := Preview(root, cfg, model())
	if err != nil || plan.Empty() {
		t.Fatalf("%v %v", plan, err)
	}
	if out := render(plan); !strings.Contains(out, "create  "+GitHubPath) {
		t.Fatalf("the plan names the pipeline it creates:\n%s", out)
	}
	if _, err := os.Stat(filepath.Join(root, GitHubPath)); !os.IsNotExist(err) {
		t.Fatal("Preview must not write")
	}
	if _, _, err := Write(root, cfg, model()); err != nil {
		t.Fatal(err)
	}
	if again, _, _ := Preview(root, cfg, model()); !again.Empty() {
		t.Fatalf("up to date means an empty plan:\n%s", render(again))
	}
}

// A hand edit outside a slot shows as a removed line in the diff, and a provider
// that was dropped shows as a deletion with the content it had.
func TestPreviewShowsTheEditsItWouldLoseAndTheFilesItWouldDelete(t *testing.T) {
	root := t.TempDir()
	both := workspace.Config{CI: workspace.CISection{Providers: []string{"github", "azure"}}}
	if _, _, err := Write(root, both, model()); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, GitHubPath)
	raw, _ := os.ReadFile(path)
	edited := strings.Replace(string(raw), "\n", "\n# my hand edit\n", 1)
	if err := os.WriteFile(path, []byte(edited), 0o644); err != nil {
		t.Fatal(err)
	}
	githubOnly := workspace.Config{CI: workspace.CISection{Providers: []string{"github"}}}
	plan, _, err := Preview(root, githubOnly, model())
	if err != nil {
		t.Fatal(err)
	}
	out := render(plan)
	for _, want := range []string{"remove  " + AzurePath, "change  " + GitHubPath, "# my hand edit"} {
		if !strings.Contains(out, want) {
			t.Errorf("the plan lacks %q:\n%s", want, out)
		}
	}
	if _, err := os.Stat(filepath.Join(root, AzurePath)); err != nil {
		t.Fatal("Preview must not delete")
	}
}
