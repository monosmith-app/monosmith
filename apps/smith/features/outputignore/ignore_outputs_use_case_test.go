package outputignore

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"
)

func repo(t *testing.T, gitignore string) string {
	t.Helper()
	root := t.TempDir()
	if out, err := exec.Command("git", "-C", root, "init", "-q").CombinedOutput(); err != nil {
		t.Skipf("git is not available: %v %s", err, out)
	}
	if gitignore != "" {
		if err := os.WriteFile(filepath.Join(root, ".gitignore"), []byte(gitignore), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func TestMissingNamesThePatternsGitDoesNotIgnore(t *testing.T) {
	root := repo(t, "node_modules/\n*.egg-info/\n/packages/astar/obj/\n")
	got := Missing(root, "packages/astar", []string{"bin/", "obj/", "node_modules/", "*.egg-info/"})
	if want := []string{"/packages/astar/bin/"}; !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestMissingWithNoGitignoreNamesEveryPattern(t *testing.T) {
	root := repo(t, "")
	got := Missing(root, "packages/astar/", []string{"bin/", "obj/"})
	if want := []string{"/packages/astar/bin/", "/packages/astar/obj/"}; !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	if Missing(root, "packages/astar", nil) != nil {
		t.Fatal("no pattern, no line")
	}
}

func TestAppendMakesGitIgnoreTheFolderOnly(t *testing.T) {
	root := repo(t, "# smith\n.smith/")
	lines := Missing(root, "packages/astar", []string{"bin/", "obj/"})
	if err := Append(root, "packages/astar", lines); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(filepath.Join(root, ".gitignore"))
	if want := "# smith\n.smith/\n\n# build output of packages/astar\n/packages/astar/bin/\n/packages/astar/obj/\n"; string(raw) != want {
		t.Fatalf(".gitignore:\n%s", raw)
	}
	if again := Missing(root, "packages/astar", []string{"bin/", "obj/"}); len(again) != 0 {
		t.Fatalf("after Append git ignores both: %v", again)
	}
	if other := Missing(root, "packages/cli", []string{"bin/"}); len(other) != 1 {
		t.Fatalf("the bin/ of another project stays tracked: %v", other)
	}
}

func TestWithStartsAFileThatIsNotThere(t *testing.T) {
	if got := With("", "apps/api", []string{"/apps/api/bin/"}); got != "# build output of apps/api\n/apps/api/bin/\n" {
		t.Fatalf("got %q", got)
	}
}
