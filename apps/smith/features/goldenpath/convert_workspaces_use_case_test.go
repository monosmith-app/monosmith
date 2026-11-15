//go:build e2e

package goldenpath

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// The generators are pinned so the fixtures do not drift under the test.
const (
	legacyMnci        = "@mnci/cli@4.117.0"
	createNxWorkspace = "create-nx-workspace@23.2.0"
	nxVersion         = "23.2.0"
)

func identity() []string {
	return []string{"GIT_AUTHOR_NAME=e2e", "GIT_AUTHOR_EMAIL=e2e@example.com", "GIT_COMMITTER_NAME=e2e", "GIT_COMMITTER_EMAIL=e2e@example.com"}
}

// TestConvertLegacyWorkspace converts a workspace made by the old mnci (an
// npm library, a Node app, a Python package, a releasable Go app and a Go
// library), then builds it, keeps its release tags and plans the next
// release with the old asset names.
func TestConvertLegacyWorkspace(t *testing.T) {
	base := t.TempDir()
	w := &workspace{t: t, bin: buildSmith(t, base), dir: base, env: identity()}
	// The old CLI runs from its own install: under npx, its nested npx calls
	// fail on Windows.
	tools := filepath.Join(base, "old-mnci")
	run(t, base, "npm", "install", "--prefix", tools, "--no-audit", "--no-fund", legacyMnci)
	oldBin := filepath.Join(tools, "node_modules", ".bin", "mnci")
	if runtime.GOOS == "windows" {
		oldBin += ".cmd"
	}
	old := func(args ...string) {
		w.t.Helper()
		cmd := exec.Command(oldBin, args...)
		cmd.Dir = w.dir
		cmd.Env = append(os.Environ(), w.env...)
		if out, err := cmd.CombinedOutput(); err != nil {
			w.t.Fatalf("old mnci %s: %v\n%s", strings.Join(args, " "), err, out)
		}
	}

	old("new", "legacy", "--yes", "--scope", "@legacy", "--registry", "npm", "--ci", "both")
	w.dir = filepath.Join(base, "legacy")
	for _, add := range [][]string{{"npm-lib", "kit"}, {"node-app", "api"}, {"python-lib", "pykit"}, {"go-app", "tool", "--release"}, {"go-lib", "gokit"}} {
		old(append([]string{"add"}, add...)...)
	}
	w.git("add", "-A")
	w.git("commit", "-qm", "chore: projects", "--no-verify")
	for _, tag := range []string{"@legacy/kit@0.1.0", "pykit@0.2.0", "tool@1.0.0", "packages/gokit/v0.3.0"} {
		w.git("tag", tag)
	}

	plan := w.smith(true, "convert", "--dry-run")
	for _, want := range []string{"made by the old mnci", "kit ", "tool ", "Rollup is replaced by esbuild", "Nothing changed (dry run)"} {
		if !strings.Contains(plan, want) {
			t.Fatalf("plan lacks %q:\n%s", want, plan)
		}
	}
	w.smith(true, "convert")
	for _, gone := range []string{"nx.json", "apps/tool/project.json", "tools/go-app-release.cjs"} {
		if _, err := os.Stat(filepath.Join(w.dir, filepath.FromSlash(gone))); err == nil {
			t.Fatalf("%s must be gone after convert", gone)
		}
	}
	if root, _ := os.ReadFile(filepath.Join(w.dir, "package.json")); strings.Contains(string(root), `"nx"`) || strings.Contains(string(root), "@nx/") {
		t.Fatalf("package.json still depends on Nx:\n%s", root)
	}
	// Nor is Nx installed, not even as @mnci/eslint-config's optional peer.
	for _, gone := range []string{"node_modules/nx", "node_modules/@nx/eslint-plugin", "node_modules/@nx/devkit"} {
		if _, err := os.Stat(filepath.Join(w.dir, filepath.FromSlash(gone))); err == nil {
			t.Fatalf("%s is still installed after convert", gone)
		}
	}
	config, _ := os.ReadFile(filepath.Join(w.dir, "smith.toml"))
	for _, want := range []string{`tag_prefix = "@legacy/kit@"`, "[projects.tool.release]", `asset_name = "go-app-{product}-{os}-{arch}.{ext}"`, "github_releases = true"} {
		if !strings.Contains(string(config), want) {
			t.Fatalf("smith.toml lacks %q:\n%s", want, config)
		}
	}

	w.smith(true, "ci", "setup")
	w.smith(true, "doctor")
	if out := w.smith(true, "run", "lint", "typecheck", "test", "build"); !strings.Contains(out, "tasks succeeded") {
		t.Fatal(out)
	}
	if out := w.smith(true, "run", "start", "-p", "tool"); !strings.Contains(out, "Hello") {
		t.Fatalf("the Go app runs:\n%s", out)
	}
	w.smith(true, "ci", "generate", "--check")

	w.git("add", "-A")
	w.git("commit", "-qm", "chore: convert to smith", "--no-verify")
	for _, f := range []string{"packages/kit/src/index.ts", "apps/tool/main.go", "packages/gokit/doc.go"} {
		appendTo(t, filepath.Join(w.dir, filepath.FromSlash(f)), "\n// touched\n")
	}
	appendTo(t, filepath.Join(w.dir, "python-packages", "pykit", "pykit", "__init__.py"), "\n# touched\n")
	w.git("commit", "-qam", "feat: a feature everywhere", "--no-verify")
	preview := w.smith(true, "release", "--dry-run")
	for _, want := range []string{"0.1.0 → **0.2.0**", "0.2.0 → **0.3.0**", "1.0.0 → **1.1.0**", "0.3.0 → **0.4.0**", "go-app-tool-linux-amd64.zip"} {
		if !strings.Contains(preview, want) {
			t.Fatalf("release preview lacks %q:\n%s", want, preview)
		}
	}
}

// TestConvertPlainNxWorkspace converts a workspace create-nx-workspace makes,
// with a Jest library, a publishable Vitest library and a Node app.
func TestConvertPlainNxWorkspace(t *testing.T) {
	base := t.TempDir()
	w := &workspace{t: t, bin: buildSmith(t, base), dir: base, env: identity()}
	w.npx("-y", createNxWorkspace, "plain", "--preset=ts", "--pm=npm", "--nxCloud=skip", "--interactive=false")
	w.dir = filepath.Join(base, "plain")
	run(t, w.dir, "npm", "install", "--no-audit", "--no-fund", "-D", "@nx/node@"+nxVersion, "@nx/jest@"+nxVersion, "@nx/vite@"+nxVersion)
	w.npx("nx", "g", "@nx/js:lib", "libs/util", "--bundler=tsc", "--unitTestRunner=jest", "--linter=eslint", "--no-interactive")
	w.npx("nx", "g", "@nx/js:lib", "packages/pub", "--publishable", "--importPath=@plain/pub", "--bundler=tsc", "--unitTestRunner=vitest", "--linter=eslint", "--no-interactive")
	w.npx("nx", "g", "@nx/node:app", "apps/api", "--framework=none", "--unitTestRunner=jest", "--linter=eslint", "--e2eTestRunner=none", "--no-interactive")
	w.git("add", "-A")
	w.git("commit", "-qm", "chore: projects", "--no-verify")

	plan := w.smith(true, "convert", "--dry-run")
	for _, want := range []string{"Conversion plan for Nx workspace", "pub ", "rewritten without @nx/eslint-plugin"} {
		if !strings.Contains(plan, want) {
			t.Fatalf("plan lacks %q:\n%s", want, plan)
		}
	}
	w.smith(true, "convert")
	if _, err := os.Stat(filepath.Join(w.dir, "nx.json")); err == nil {
		t.Fatal("nx.json must be gone")
	}
	w.smith(true, "doctor")
	if out := w.smith(true, "run", "lint", "typecheck", "test", "build"); !strings.Contains(out, "tasks succeeded") {
		t.Fatal(out)
	}
	if again := w.smith(true, "run", "lint", "typecheck", "test", "build"); strings.Contains(again, "▶") {
		t.Fatalf("a second run is all cache:\n%s", again)
	}
	if projects := w.smith(true, "projects"); !strings.Contains(projects, "util") || !strings.Contains(projects, "api") {
		t.Fatal(projects)
	}
}

// TestConvertTurborepoWorkspace converts a pnpm Turborepo (testdata/turborepo-pnpm:
// an app and two packages whose builds need each other's, no registry
// needed), then installs, builds and tests it as an npm workspace with smith.
func TestConvertTurborepoWorkspace(t *testing.T) {
	base := t.TempDir()
	w := &workspace{t: t, bin: buildSmith(t, base), dir: filepath.Join(base, "turbo"), env: identity()}
	copyTree(t, filepath.Join("testdata", "turborepo-pnpm"), w.dir)
	w.git("init", "-q", "-b", "main")
	w.git("add", "-A")
	w.git("commit", "-qm", "chore: turborepo", "--no-verify")

	plan := w.smith(true, "convert", "--dry-run")
	for _, want := range []string{
		"Conversion plan for Turborepo workspace",
		"pnpm-workspace.yaml: its packages become the package.json \"workspaces\" field",
		"2 workspace: and catalog: ranges become what npm accepts",
		"task test: env [\"NODE_ENV\"]",
		"web build: the build writes to dist inside the package",
		"Nothing changed (dry run)",
	} {
		if !strings.Contains(plan, want) {
			t.Fatalf("plan lacks %q:\n%s", want, plan)
		}
	}
	w.smith(true, "convert")
	for _, gone := range []string{"turbo.json", "pnpm-lock.yaml", "pnpm-workspace.yaml"} {
		if _, err := os.Stat(filepath.Join(w.dir, gone)); err == nil {
			t.Fatalf("%s must be gone after convert", gone)
		}
	}
	root := w.read("package.json")
	for _, gone := range []string{"packageManager", "turbo", "pnpm"} {
		if strings.Contains(root, gone) {
			t.Fatalf("package.json still mentions %s:\n%s", gone, root)
		}
	}
	if !strings.Contains(root, `"workspaces"`) || !strings.Contains(root, `"smith run build --bail"`) || !strings.Contains(root, `"semver": "7.6.3"`) {
		t.Fatalf("package.json is not the npm workspace root:\n%s", root)
	}
	if web := w.read("apps/web/package.json"); !strings.Contains(web, `"@acme/ui": "*"`) {
		t.Fatalf("workspace:* must become *:\n%s", web)
	}
	if ui := w.read("packages/ui/package.json"); !strings.Contains(ui, `"@acme/util": "^1.0.0"`) {
		t.Fatalf("workspace:^ must become ^<version>:\n%s", ui)
	}
	if _, err := os.Stat(filepath.Join(w.dir, "package-lock.json")); err != nil {
		t.Fatal("convert installs with npm and writes package-lock.json")
	}
	config := w.read("smith.toml")
	for _, want := range []string{"[projects.web.tasks.test]", `depends_on = ["^build", "build"]`, `outputs = ["dist"]`, "[projects.web.tasks.dev]", "cache = false"} {
		if !strings.Contains(config, want) {
			t.Fatalf("smith.toml lacks %q:\n%s", want, config)
		}
	}

	// The builds write into each package, which doctor refuses; the report said so.
	if out := w.smith(false, "doctor"); !strings.Contains(out, "every build writes to dist/") {
		t.Fatalf("doctor:\n%s", out)
	}
	// smith i installs the npm workspace; the packages build in dependency order.
	if out := w.smith(true, "i"); !strings.Contains(out, "node: installed 3 project(s)") {
		t.Fatalf("smith i:\n%s", out)
	}
	if out := w.smith(true, "run", "build", "test"); !strings.Contains(out, "tasks succeeded") {
		t.Fatal(out)
	}
	if again := w.smith(true, "run", "build", "test"); strings.Contains(again, "▶") {
		t.Fatalf("a second run is all cache:\n%s", again)
	}
}

// copyTree copies a directory of fixtures.
func copyTree(t *testing.T, from, to string) {
	t.Helper()
	err := filepath.WalkDir(from, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(from, path)
		dest := filepath.Join(to, rel)
		if d.IsDir() {
			return os.MkdirAll(dest, 0o755)
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(dest, raw, 0o644)
	})
	if err != nil {
		t.Fatal(err)
	}
}

// npx runs npx in the workspace directory with the test's git identity.
func (w *workspace) npx(args ...string) {
	w.t.Helper()
	cmd := exec.Command("npx", args...)
	cmd.Dir = w.dir
	cmd.Env = append(os.Environ(), w.env...)
	if out, err := cmd.CombinedOutput(); err != nil {
		w.t.Fatalf("npx %s: %v\n%s", strings.Join(args, " "), err, out)
	}
}

func appendTo(t *testing.T, file, text string) {
	t.Helper()
	f, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	if _, err := f.WriteString(text); err != nil {
		t.Fatal(err)
	}
}
