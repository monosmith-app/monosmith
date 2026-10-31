package apisurface

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func writeTree(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for name, body := range files {
		full := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func surface(t *testing.T, root, project, ref string) protocol.APISurfaceResult {
	t.Helper()
	got, err := List(root, protocol.APISurfaceParams{Project: protocol.Project{Root: project}, Ref: ref})
	if err != nil {
		t.Fatal(err)
	}
	return got
}

func TestListReadsTheSourceBehindTheBuiltTypesAndFollowsReExports(t *testing.T) {
	dir := t.TempDir()
	writeTree(t, dir, map[string]string{
		"packages/lib/package.json":       `{"name":"lib","version":"1.0.0","types":"./dist/index.d.ts","exports":{".":{"types":"./dist/index.d.ts","import":"./dist/index.js"},"./cli":"./dist/cli.js"}}`,
		"packages/lib/src/index.ts":       "export * from './math.js'\nexport * from './extra'\nexport * from 'other-package'\nexport { run } from './run'\nexport const version = '1'\n",
		"packages/lib/src/math.ts":        "export function add() {}\nexport function sub() {}\n",
		"packages/lib/src/extra/index.ts": "export const extra = 1\nexport * from '../index'\n",
		"packages/lib/src/run.ts":         "export function run() {}\n",
		"packages/lib/src/cli.ts":         "export function main() {}\n",
	})
	got := surface(t, dir, "packages/lib", "")
	want := []string{"* from other-package", "./cli:main", "add", "extra", "run", "sub", "version"}
	if got.Unreadable != "" || !slices.Equal(got.Names, want) {
		t.Errorf("surface = %v (%q), want %v", got.Names, got.Unreadable, want)
	}
}

func TestListPrefersTheBuiltDeclarationWhenItIsInTheRepository(t *testing.T) {
	dir := t.TempDir()
	writeTree(t, dir, map[string]string{
		"package.json":    `{"name":"lib","types":"dist/index.d.ts"}`,
		"dist/index.d.ts": "export declare function built(): void\nexport {}\n",
		"src/index.ts":    "export function fromSource() {}\n",
	})
	if got := surface(t, dir, ".", ""); !slices.Equal(got.Names, []string{"built"}) {
		t.Errorf("surface = %v, want [built]", got.Names)
	}
}

func TestListSaysWhyThereIsNothingToList(t *testing.T) {
	dir := t.TempDir()
	writeTree(t, dir, map[string]string{
		"private/package.json": `{"name":"p","private":true,"main":"index.js"}`,
		"bare/package.json":    `{"name":"b","main":"dist/index.js"}`,
		"broken/package.json":  `{`,
	})
	for project, fragment := range map[string]string{"private": "private", "bare": "no entry point", "broken": "not valid JSON", "missing": "no package.json"} {
		got := surface(t, dir, project, "")
		if !strings.Contains(got.Unreadable, fragment) || len(got.Names) != 0 {
			t.Errorf("%s: Unreadable = %q, names %v; want a reason with %q", project, got.Unreadable, got.Names, fragment)
		}
	}
}

func TestListAnEmptyPackageIsAnAnswerNotAFailure(t *testing.T) {
	dir := t.TempDir()
	writeTree(t, dir, map[string]string{"package.json": `{"name":"e","main":"index.js"}`, "index.js": "// nothing\n"})
	if got := surface(t, dir, ".", ""); got.Unreadable != "" || got.Names == nil || len(got.Names) != 0 {
		t.Errorf("surface = %#v, want an empty list and no reason", got)
	}
}

func TestListReadsAGitRevisionWithoutACheckout(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}
	dir := t.TempDir()
	git := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-c", "user.name=t", "-c", "user.email=t@t", "-c", "commit.gpgsign=false"}, args...)...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	git("init", "-q")
	writeTree(t, dir, map[string]string{
		"pkg/package.json": `{"name":"lib","main":"index.js"}`,
		"pkg/index.js":     "exports.keep = 1\nexports.drop = 2\n",
	})
	git("add", ".")
	git("commit", "-q", "-m", "release")
	git("tag", "lib@1.0.0")
	writeTree(t, dir, map[string]string{"pkg/index.js": "exports.keep = 1\nexports.fresh = 3\n"})

	released := surface(t, dir, "pkg", "lib@1.0.0")
	if want := []string{"drop", "keep"}; !slices.Equal(released.Names, want) {
		t.Errorf("at the tag = %v, want %v", released.Names, want)
	}
	if now := surface(t, dir, "pkg", ""); !slices.Equal(now.Names, []string{"fresh", "keep"}) {
		t.Errorf("working tree = %v, want [fresh keep]", now.Names)
	}
	if gone := surface(t, dir, "pkg", "no-such-tag"); gone.Unreadable == "" {
		t.Errorf("an unknown revision must say it is unreadable, got %#v", gone)
	}
}
