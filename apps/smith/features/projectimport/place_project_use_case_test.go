package projectimport

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func write(t *testing.T, root, rel, text string) {
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

func git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-c", "user.name=T", "-c", "user.email=t@example.com", "-c", "commit.gpgsign=false", "-c", "core.autocrlf=false"}, args...)...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func TestPlaceCopiesAPlainFolderWithoutDependenciesAndCaches(t *testing.T) {
	src, root := t.TempDir(), t.TempDir()
	write(t, src, "package.json", "{}")
	write(t, src, "src/index.js", "x")
	write(t, src, "node_modules/a/index.js", "x")
	write(t, src, "pkg/__pycache__/m.pyc", "x")
	p := Plan{Name: "web", Folder: "apps/web", History: HistoryCopy, Source: Source{Dir: src}}
	if err := Place(root, p, io.Discard); err != nil {
		t.Fatal(err)
	}
	for rel, want := range map[string]bool{"package.json": true, "src/index.js": true, "node_modules": false, "pkg/__pycache__": false} {
		_, err := os.Stat(filepath.Join(root, "apps/web", rel))
		if (err == nil) != want {
			t.Errorf("%s copied = %v, want %v", rel, err == nil, want)
		}
	}
}

func TestDropLockfilesRemovesOnlyTheLockfilesAtTheTop(t *testing.T) {
	root := t.TempDir()
	write(t, root, "apps/web/package-lock.json", "{}")
	write(t, root, "apps/web/package.json", "{}")
	write(t, root, "apps/web/sub/yarn.lock", "")
	got, err := DropLockfiles(root, "apps/web", Lockfiles("node"))
	if err != nil || !reflect.DeepEqual(got, []string{"package-lock.json"}) {
		t.Fatalf("dropped %v %v", got, err)
	}
	if _, err := os.Stat(filepath.Join(root, "apps/web/sub/yarn.lock")); err != nil {
		t.Error("a lockfile below the top stays")
	}
}

func TestKeepVersionsMergesTheSourceLockIntoTheWorkspaceLock(t *testing.T) {
	src, root := t.TempDir(), t.TempDir()
	write(t, src, "package-lock.json", `{"lockfileVersion": 3, "packages": {"node_modules/react-icons": {"version": "4.3.1"}}}`)
	write(t, root, "package-lock.json", `{"lockfileVersion": 3, "packages": {"": {}}}`)
	write(t, root, "apps/web/package.json", `{"dependencies": {"react-icons": "^4.3.1"}}`)
	var out strings.Builder
	left := KeepVersions(root, Plan{Name: "web", Folder: "apps/web", Ecosystem: "node", Source: Source{Dir: src}}, &out)
	if len(left) != 0 {
		t.Fatalf("left = %+v", left)
	}
	if !strings.Contains(out.String(), "kept the versions package-lock.json locked (1 packages): react-icons 4.3.1") {
		t.Errorf("output = %q", out.String())
	}
	raw, _ := os.ReadFile(filepath.Join(root, "package-lock.json"))
	if !strings.Contains(string(raw), "apps/web/node_modules/react-icons") {
		t.Error("the workspace lockfile does not hold the project's entry")
	}
}

func TestKeepVersionsLeavesAnItemForALockfileItCannotCarry(t *testing.T) {
	src, root := t.TempDir(), t.TempDir()
	write(t, src, "yarn.lock", "")
	if left := KeepVersions(root, Plan{Name: "web", Folder: "apps/web", Ecosystem: "node", Source: Source{Dir: src}}, io.Discard); len(left) != 1 || !strings.Contains(left[0].Problem, "yarn.lock") {
		t.Errorf("left = %+v", left)
	}
	if left := KeepVersions(root, Plan{Name: "lib", Folder: "libs/lib", Ecosystem: "python", Lockfiles: []string{"uv.lock"}}, io.Discard); len(left) != 1 || !strings.Contains(left[0].Problem, "uv.lock") {
		t.Errorf("python left = %+v", left)
	}
}

func TestPlaceKeepsTheHistoryUnderTheFolder(t *testing.T) {
	src, root := t.TempDir(), t.TempDir()
	git(t, src, "init", "-q", "-b", "main")
	write(t, src, "package.json", "{}\n")
	git(t, src, "add", "-A")
	git(t, src, "commit", "-q", "-m", "first")
	write(t, src, "index.js", "x\n")
	git(t, src, "add", "-A")
	git(t, src, "commit", "-q", "-m", "second")
	git(t, root, "init", "-q", "-b", "main")
	write(t, root, "smith.toml", "schema = 1\n")
	git(t, root, "add", "-A")
	git(t, root, "commit", "-q", "-m", "workspace")
	t.Setenv("GIT_AUTHOR_NAME", "T")
	t.Setenv("GIT_AUTHOR_EMAIL", "t@example.com")
	t.Setenv("GIT_COMMITTER_NAME", "T")
	t.Setenv("GIT_COMMITTER_EMAIL", "t@example.com")

	p := Plan{Name: "web", Folder: "apps/web", History: HistoryKeep, Source: Source{Arg: src, Dir: src, Git: true, Ref: "main", Commit: "abc", Commits: 2}}
	if err := Place(root, p, io.Discard); err != nil {
		t.Fatal(err)
	}
	log := git(t, root, "log", "--format=%s", "--", "apps/web")
	if log != "second\nfirst" {
		t.Errorf("git log -- apps/web = %q, want the source's two commits", log)
	}
	if files := git(t, root, "ls-files", "apps/web"); files != "apps/web/index.js\napps/web/package.json" {
		t.Errorf("files = %q", files)
	}
	if subject := git(t, root, "log", "-1", "--format=%s"); subject != "chore: import web into apps/web" {
		t.Errorf("merge commit = %q", subject)
	}
}
