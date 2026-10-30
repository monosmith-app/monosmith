//go:build e2e

package goldenpath

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"
)

type workspace struct {
	t   *testing.T
	bin string
	dir string
	env []string
}

// smith runs the binary and returns its output; wantOK says whether it must succeed.
func (w *workspace) smith(wantOK bool, args ...string) string {
	w.t.Helper()
	cmd := exec.Command(w.bin, args...)
	cmd.Dir = w.dir
	// A smith home of its own, so the trust a test records is not the developer's.
	cmd.Env = append(append(os.Environ(), "SMITH_HOME="+filepath.Join(filepath.Dir(w.bin), "smith-home")), w.env...)
	began := time.Now()
	out, err := cmd.CombinedOutput()
	w.t.Logf("smith %s (%s)", strings.Join(args, " "), time.Since(began).Round(time.Millisecond))
	if (err == nil) != wantOK {
		w.t.Fatalf("smith %s: err=%v\n%s", strings.Join(args, " "), err, out)
	}
	return string(out)
}

// trust trusts an external plugin the workspace declares, and pins its digest:
// smith starts an external plugin only when it is trusted, and in CI only when
// smith.toml pins it. Call it again after smith.toml is rewritten.
func (w *workspace) trust(name string) {
	w.t.Helper()
	w.smith(true, "plugin", "trust", name, "--yes", "--pin")
}

func (w *workspace) write(rel, body string) {
	w.t.Helper()
	p := filepath.Join(w.dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		w.t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		w.t.Fatal(err)
	}
}

func (w *workspace) read(rel string) string {
	w.t.Helper()
	b, err := os.ReadFile(filepath.Join(w.dir, filepath.FromSlash(rel)))
	if err != nil {
		w.t.Fatal(err)
	}
	return string(b)
}

func (w *workspace) git(args ...string) string {
	w.t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = w.dir
	cmd.Env = append(os.Environ(), w.env...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		w.t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return string(out)
}

func buildSmith(t *testing.T, dir string) string {
	t.Helper()
	bin := filepath.Join(dir, "smith")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	wd, _ := os.Getwd()
	run(t, filepath.Dir(wd), "go", "build", "-ldflags", "-X main.version=0.0.0-e2e", "-o", bin, ".")
	return bin
}

func TestGoldenPath(t *testing.T) {
	base := t.TempDir()
	w := &workspace{t: t, bin: buildSmith(t, base), dir: base}
	for _, kv := range [][2]string{{"user.email", "e2e@example.com"}, {"user.name", "e2e"}} {
		w.env = append(w.env, "GIT_AUTHOR_"+strings.ToUpper(strings.TrimPrefix(kv[0], "user."))+"="+kv[1], "GIT_COMMITTER_"+strings.ToUpper(strings.TrimPrefix(kv[0], "user."))+"="+kv[1])
	}

	// 1. A new workspace is committed and CI-ready.
	w.smith(true, "new", "acme", "--ci", "github,azure")
	w.dir = filepath.Join(base, "acme")
	for _, f := range []string{"smith.toml", ".github/workflows/ci.yml", "azure-pipelines.yml", ".devcontainer/devcontainer.json"} {
		if _, err := os.Stat(filepath.Join(w.dir, filepath.FromSlash(f))); err != nil {
			t.Fatalf("new did not write %s", f)
		}
	}
	// Editors validate and complete smith.toml from its schema.
	if cfg, _ := os.ReadFile(filepath.Join(w.dir, "smith.toml")); !strings.HasPrefix(string(cfg), "#:schema https://") {
		t.Fatalf("smith.toml does not name its schema:\n%s", cfg)
	}
	var schemaDoc map[string]any
	if err := json.Unmarshal([]byte(w.smith(true, "schema")), &schemaDoc); err != nil || schemaDoc["title"] != "smith.toml" {
		t.Fatalf("smith schema: %v %v", err, schemaDoc["title"])
	}
	// One setting changes; the file's comments and schema line stay.
	w.smith(true, "config", "set", "cache.max_size_mb", "4096")
	if got := strings.TrimSpace(w.smith(true, "config", "get", "cache.max_size_mb")); got != "4096" {
		t.Fatalf("config get: %q", got)
	}
	if cfg, _ := os.ReadFile(filepath.Join(w.dir, "smith.toml")); !strings.HasPrefix(string(cfg), "#:schema https://") || !strings.Contains(string(cfg), "# Pipelines to generate") {
		t.Fatalf("config set rewrote the file:\n%s", cfg)
	}
	w.smith(false, "config", "set", "ci.providers", "jenkins")

	// 2. An app, a lib and a package in each language. Each project picks its
	// stack: the language default (node-jest), a published one (node-vitest),
	// or one smith.toml declares (go-sliced).
	toml, _ := os.ReadFile(filepath.Join(w.dir, "smith.toml"))
	w.write("smith.toml", string(toml)+"\n[stacks.go-sliced]\nlanguage = \"go\"\nstructure = [\"vertical-slices\"]\n")
	for _, add := range [][]string{
		{"app", "web", "--lang", "node"}, {"lib", "ui", "--lang", "node"}, {"package", "sdk", "--stack", "node-vitest"}, {"app", "hello", "--lang", "node", "--stack", "none"},
		{"app", "api", "--lang", "go"}, {"lib", "core", "--lang", "go"}, {"package", "gokit", "--stack", "go-sliced"}, {"app", "probe", "--lang", "go"},
		{"app", "worker", "--lang", "python"}, {"python-lib", "pyutil"}, {"package", "pykit", "--lang", "python"},
	} {
		w.smith(true, append([]string{"add"}, add...)...)
	}
	for dir, want := range map[string][]string{
		"apps/web/package.json":     {`"stack": "node-jest"`, `"test": "jest"`, `"lint": "eslint ."`},
		"packages/sdk/package.json": {`"stack": "node-vitest"`, `"test": "vitest run"`, `"lint": "eslint ."`},
		"packages/gokit/go.mod":     {"// smith:stack go-sliced"},
		"apps/hello/package.json":   {`"stack": "none"`, `"start": "node src/main.js"`},
	} {
		manifest, _ := os.ReadFile(filepath.Join(w.dir, filepath.FromSlash(dir)))
		for _, s := range want {
			if !strings.Contains(string(manifest), s) {
				t.Fatalf("%s lacks %s:\n%s", dir, s, manifest)
			}
		}
	}
	if stacks := w.smith(true, "stacks"); !strings.Contains(stacks, "node-vitest  (plugin vitest)") || !strings.Contains(stacks, "projects   sdk") || !strings.Contains(stacks, "go-sliced  (smith.toml)") {
		t.Fatalf("stacks:\n%s", stacks)
	}
	// What an editor builds its menus from (the VS Code extension's contract).
	if kinds := w.smith(true, "kinds", "--json"); !strings.Contains(kinds, `"kind": "node-app"`) {
		t.Fatalf("kinds:\n%s", kinds)
	}
	var commands []struct{ Name, Group string }
	if err := json.Unmarshal([]byte(w.smith(true, "commands", "--json")), &commands); err != nil || len(commands) < 10 {
		t.Fatalf("commands --json: %v %v", err, commands)
	}
	projects := w.smith(true, "projects", "--json")
	if !strings.Contains(projects, `"ecosystem": "pip"`) || !strings.Contains(projects, `"targets": [`) {
		t.Fatalf("projects --json lacks the extension's fields:\n%s", projects)
	}
	var list []struct{ Name, Kind, Language string }
	if err := json.Unmarshal([]byte(projects[strings.Index(projects, "["):]), &list); err != nil || len(list) != 11 {
		t.Fatalf("projects: %v\n%s", err, projects)
	}

	// 3. Imports between projects work immediately.
	w.write("apps/web/src/main.ts", "import { greet as ui } from '@acme/ui'\nimport { greet } from './greeting/index.ts'\n\nprocess.stdout.write(greet('world').message + ' ' + ui('ui').message + '\\n')\n")
	w.write("apps/api/main.go", "package main\n\nimport (\n\t\"fmt\"\n\n\t\"example.com/acme/apps/api/greeting\"\n\tcore \"example.com/acme/libs/core/greeting\"\n)\n\nfunc main() {\n\tfmt.Println(greeting.Greet(\"world\").Message, core.Greet(\"core\").Message)\n}\n")
	w.write("apps/worker/src/acme_worker/__main__.py", "from acme_pyutil.greeting import greet as util\n\nfrom .greeting import greet\n\nprint(greet(\"world\").message, util(\"util\").message)\n")
	graph := w.smith(true, "graph", "--format", "json")
	for _, edge := range [][2]string{{"web", "ui"}, {"api", "core"}, {"worker", "pyutil"}} {
		if !strings.Contains(graph, fmt.Sprintf(`"name": %q`, edge[0])) || !hasEdge(graph, edge[0], edge[1]) {
			t.Fatalf("graph lacks %s → %s:\n%s", edge[0], edge[1], graph)
		}
	}
	// A dependency belongs to one project: worker requires its sibling, then
	// stops; pip resolves it from the venv's editable install, offline.
	pyproject, _ := os.ReadFile(filepath.Join(w.dir, "apps", "worker", "pyproject.toml"))
	if out := w.smith(false, "install", "acme-pyutil"); !strings.Contains(out, "needs a target project") {
		t.Fatalf("install without -w:\n%s", out)
	}
	if out := w.smith(true, "install", "-w", "apps/worker", "acme-pyutil"); !strings.Contains(out, "✔ worker: added acme-pyutil (.venv") && !strings.Contains(out, "✔ worker: added acme-pyutil (apps/worker/pyproject.toml)") {
		t.Fatalf("install:\n%s", out)
	}
	if after, _ := os.ReadFile(filepath.Join(w.dir, "apps", "worker", "pyproject.toml")); !strings.Contains(string(after), `dependencies = ["acme-pyutil"]`) {
		t.Fatalf("pyproject:\n%s", after)
	}
	w.smith(true, "uninstall", "-w", "worker", "acme-pyutil")
	if after, _ := os.ReadFile(filepath.Join(w.dir, "apps", "worker", "pyproject.toml")); string(after) != string(pyproject) {
		t.Fatalf("uninstall did not restore pyproject.toml:\n%s", after)
	}
	// Two projects disagree on a shared dependency; sync converges them on
	// the highest one declared.
	pykit, _ := os.ReadFile(filepath.Join(w.dir, "packages", "pykit", "pyproject.toml"))
	w.write("apps/worker/pyproject.toml", strings.Replace(string(pyproject), "dependencies = []", `dependencies = ["requests>=2.31"]`, 1))
	w.write("packages/pykit/pyproject.toml", strings.Replace(string(pykit), "dependencies = []", `dependencies = ["requests>=2.32"]`, 1))
	if out := w.smith(false, "sync", "--check"); !strings.Contains(out, "requests (python) → >=2.32") || !strings.Contains(out, "worker") {
		t.Fatalf("sync --check:\n%s", out)
	}
	w.smith(true, "sync", "--no-install")
	if after, _ := os.ReadFile(filepath.Join(w.dir, "apps", "worker", "pyproject.toml")); !strings.Contains(string(after), `dependencies = ["requests>=2.32"]`) {
		t.Fatalf("sync did not converge worker:\n%s", after)
	}
	if out := w.smith(true, "sync", "--check"); !strings.Contains(out, "Every project agrees") {
		t.Fatalf("sync --check after sync:\n%s", out)
	}
	w.write("apps/worker/pyproject.toml", string(pyproject))
	w.write("packages/pykit/pyproject.toml", string(pykit))
	// up reports; offline its registries and feeds answer nothing, but every
	// language still lists the tools it pins.
	upOut := w.smith(true, "up", "--json")
	var up struct {
		Tools []struct{ Plugin, Name string }
	}
	if err := json.Unmarshal([]byte(upOut[strings.Index(upOut, "{"):]), &up); err != nil {
		t.Fatalf("up --json: %v\n%s", err, upOut)
	}
	var tools []string
	for _, tool := range up.Tools {
		tools = append(tools, tool.Plugin+":"+tool.Name)
	}
	for _, want := range []string{"node:Node.js (LTS)", "python:Python", "go:Go", "go:golangci-lint"} {
		if !slices.Contains(tools, want) {
			t.Fatalf("up lists no %s: %v", want, tools)
		}
	}

	// 4. Everything lints, type-checks, tests and builds; a re-run is all cache.
	out := w.smith(true, "run", "lint", "typecheck", "test", "build")
	if !strings.Contains(out, "tasks succeeded") {
		t.Fatal(out)
	}
	// The vertical-slices convention checks Go and Python layouts before lint.
	for _, want := range []string{"api:slice-check", "pykit:slice-check"} {
		if !strings.Contains(out, want) {
			t.Fatalf("%s did not run:\n%s", want, out)
		}
	}
	w.write("libs/core/helpers/format.go", "package helpers\n")
	if out := w.smith(false, "run", "slice-check", "-p", "core"); !strings.Contains(out, `"helpers" is a junk drawer`) {
		t.Fatalf("a junk-drawer folder passes the slice check:\n%s", out)
	}
	if err := os.RemoveAll(filepath.Join(w.dir, "libs", "core", "helpers")); err != nil {
		t.Fatal(err)
	}
	out = w.smith(true, "run", "lint", "typecheck", "test", "build")
	if strings.Contains(out, "▶") {
		t.Fatalf("second run executed tasks:\n%s", out)
	}
	// why-miss names what would make a task run again.
	if out := w.smith(true, "why-miss", "api:build"); !strings.Contains(out, "would come from the cache") {
		t.Fatalf("why-miss on a cached task:\n%s", out)
	}
	w.write("libs/core/greeting/extra.go", "package greeting\n")
	if out := w.smith(true, "why-miss", "api:build"); !strings.Contains(out, "upstream   core changed") {
		t.Fatalf("why-miss must name the changed upstream project:\n%s", out)
	}
	if err := os.Remove(filepath.Join(w.dir, "libs", "core", "greeting", "extra.go")); err != nil {
		t.Fatal(err)
	}
	out = w.smith(true, "run", "start", "-p", "web,api,worker,hello", "--parallel", "1")
	for _, want := range []string{"Hello, ui!", "Hello, core!", "Hello, util!", "hello:start │ Hello, world!"} {
		if !strings.Contains(out, want) {
			t.Fatalf("apps must run with their internal deps (%s):\n%s", want, out)
		}
	}
	// dev starts several apps at once, after building what they use.
	if out := w.smith(true, "dev", "web", "hello"); !strings.Contains(out, "→ smith run start -p web,hello --parallel 6") || !strings.Contains(out, "hello:start │ Hello, world!") || !strings.Contains(out, "Hello, ui!") {
		t.Fatalf("dev:\n%s", out)
	}
	if out := w.smith(false, "dev", "ui"); !strings.Contains(out, "ui has no start task") {
		t.Fatalf("dev on a library:\n%s", out)
	}
	w.smith(true, "doctor")
	// Ready-made pipelines: written once, beside the generated CI.
	if out := w.smith(true, "pipeline"); !strings.Contains(out, "deploy-pages (--project <app>)") {
		t.Fatalf("pipeline list:\n%s", out)
	}
	if out := w.smith(true, "pipeline", "e2e"); !strings.Contains(out, "wrote .github/workflows/e2e.yml") || !strings.Contains(out, "wrote azure-pipelines/e2e.yml") {
		t.Fatalf("pipeline e2e:\n%s", out)
	}
	if out := w.smith(false, "pipeline", "e2e"); !strings.Contains(out, "exists and is yours") {
		t.Fatalf("a second e2e must not overwrite:\n%s", out)
	}
	if out := w.smith(false, "pipeline", "deploy-pages", "--project", "ui"); !strings.Contains(out, "deploys an app") {
		t.Fatalf("deploy a lib:\n%s", out)
	}
	if out := w.smith(true, "pipeline", "deploy-pages", "--project", "web"); !strings.Contains(out, "wrote .github/workflows/deploy-pages.yml") || strings.Contains(out, "azure-pipelines/deploy-pages") {
		t.Fatalf("deploy-pages is GitHub's only:\n%s", out)
	}
	if pages := w.read(".github/workflows/deploy-pages.yml"); !strings.Contains(pages, "smith run build -p web") || !strings.Contains(pages, "smith ci setup") {
		t.Fatalf("deploy-pages.yml:\n%s", pages)
	}
	w.smith(true, "ci", "generate", "--check")
	for _, f := range []string{".github/workflows/e2e.yml", "azure-pipelines/e2e.yml", ".github/workflows/deploy-pages.yml"} {
		_ = os.Remove(filepath.Join(w.dir, filepath.FromSlash(f)))
	}
	// doctor fixes what smith and the plugins own, showing the diff first.
	work := w.read("go.work")
	w.write("go.work", strings.Replace(work, "./libs/core", "./libs/gone", 1))
	// The report lists the fixes for editors, and does not mix with applying them.
	if out := w.smith(false, "doctor", "--json"); !strings.Contains(out, `"fixes"`) || !strings.Contains(out, `"ask": false`) || !strings.Contains(out, `"checks"`) {
		t.Fatalf("doctor --json lists the fixes:\n%s", out)
	}
	if out := w.smith(false, "doctor", "--json", "--fix"); !strings.Contains(out, "apply them with --fix, without --json") {
		t.Fatalf("doctor --json --fix:\n%s", out)
	}
	if out := w.smith(false, "doctor", "--fix", "--only", "make go.work use exactly the workspace's Go modules", "--choose", "1"); !strings.Contains(out, "has none") {
		t.Fatalf("--choose on a fix with no question:\n%s", out)
	}
	if out := w.smith(false, "doctor", "--fix"); !strings.Contains(out, "-\t./libs/gone") || !strings.Contains(out, "+\t./libs/core") || !strings.Contains(out, "Pass --yes") {
		t.Fatalf("doctor --fix without a terminal shows the diff only:\n%s", out)
	}
	if out := w.smith(true, "doctor", "--fix", "--yes"); !strings.Contains(out, "✔ make go.work use exactly the workspace's Go modules") || !strings.Contains(out, "after fixing:") {
		t.Fatalf("doctor --fix --yes:\n%s", out)
	}
	if !strings.Contains(w.read("go.work"), "./libs/core") || strings.Contains(w.read("go.work"), "libs/gone") {
		t.Fatalf("go.work:\n%s", w.read("go.work"))
	}
	// GitLab gets the same phases, each toolchain installed by its script.
	w.smith(true, "config", "set", "ci.providers", "github", "azure", "gitlab", "circleci", "bitbucket")
	w.smith(true, "ci", "generate")
	circle := w.read(".circleci/config.yml")
	for _, want := range []string{"image: debian:bookworm", "deb.nodesource.com/setup_24.x", "smith ci verify", "smith ci release", "pipeline.git.branch"} {
		if !strings.Contains(circle, want) {
			t.Fatalf(".circleci/config.yml lacks %q:\n%s", want, circle)
		}
	}
	if out := w.smith(true, "ci", "generate", "--check"); !strings.Contains(out, "up to date") {
		t.Fatalf("a fresh CircleCI config drifts:\n%s", out)
	}
	bitbucket := w.read("bitbucket-pipelines.yml")
	for _, want := range []string{"image: debian:bookworm", "deb.nodesource.com/setup_24.x", "smith ci verify", "smith ci release", "BITBUCKET_PR_ID"} {
		if !strings.Contains(bitbucket, want) {
			t.Fatalf("bitbucket-pipelines.yml lacks %q:\n%s", want, bitbucket)
		}
	}
	if out := w.smith(true, "ci", "generate", "--check"); !strings.Contains(out, "up to date") {
		t.Fatalf("a fresh Bitbucket config drifts:\n%s", out)
	}
	gitlab := w.read(".gitlab-ci.yml")
	for _, want := range []string{"deb.nodesource.com/setup_24.x", "uv python install 3.12", ".linux-amd64.tar.gz | tar -C /usr/local -xz", "smith ci verify", "smith ci release"} {
		if !strings.Contains(gitlab, want) {
			t.Fatalf(".gitlab-ci.yml lacks %q:\n%s", want, gitlab)
		}
	}
	w.smith(true, "config", "set", "ci.providers", "github", "azure")
	w.smith(true, "ci", "generate")
	if _, err := os.Stat(filepath.Join(w.dir, ".gitlab-ci.yml")); err == nil {
		t.Fatal("dropping gitlab deletes the pipeline smith wrote")
	}
	if _, err := os.Stat(filepath.Join(w.dir, "bitbucket-pipelines.yml")); err == nil {
		t.Fatal("dropping bitbucket deletes the config smith wrote")
	}
	if _, err := os.Stat(filepath.Join(w.dir, ".circleci", "config.yml")); err == nil {
		t.Fatal("dropping circleci deletes the config smith wrote")
	}
	// An AI assistant sees the same workspace over MCP.
	mcp := exec.Command(w.bin, "mcp")
	mcp.Dir, mcp.Env = w.dir, append(os.Environ(), w.env...)
	mcp.Stdin = strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18"}}` + "\n" +
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"smith_projects"}}` + "\n")
	if out, err := mcp.Output(); err != nil || !strings.Contains(string(out), `"serverInfo"`) || !strings.Contains(string(out), `$ smith projects --json`) || !strings.Contains(string(out), `\"name\": \"worker\"`) {
		t.Fatalf("smith mcp: %v\n%s", err, out)
	}
	// The dev container installs what CI installs, at the same versions.
	devcontainer, _ := os.ReadFile(filepath.Join(w.dir, ".devcontainer", "devcontainer.json"))
	for _, want := range []string{"features/node:1", "features/go:1", "features/python:1", `"golangciLintVersion": "2.14.0"`, "&& ", "/.local/bin/smith", ` ci setup"`} {
		if !strings.Contains(string(devcontainer), want) {
			t.Fatalf("devcontainer lacks %s:\n%s", want, devcontainer)
		}
	}
	w.smith(true, "ci", "generate", "--check")

	// smith upgrade keeps the files smith owns in line with smith.toml: dropping
	// a CI provider deletes its pipeline, which --check reports first.
	toml, _ = os.ReadFile(filepath.Join(w.dir, "smith.toml"))
	w.write("smith.toml", strings.Replace(string(toml), `providers = ["github", "azure"]`, `providers = ["github"]`, 1))
	if out := w.smith(false, "upgrade", "--check"); !strings.Contains(out, "would change: azure-pipelines.yml") {
		t.Fatalf("upgrade --check:\n%s", out)
	}
	if out := w.smith(true, "upgrade"); !strings.Contains(out, "upgraded: azure-pipelines.yml") {
		t.Fatalf("upgrade:\n%s", out)
	}
	if _, err := os.Stat(filepath.Join(w.dir, "azure-pipelines.yml")); !os.IsNotExist(err) {
		t.Fatal("the dropped provider's pipeline is deleted")
	}
	if out := w.smith(true, "upgrade", "--check"); !strings.Contains(out, "up to date") {
		t.Fatalf("upgrade --check after upgrade:\n%s", out)
	}

	// The phases the generated pipeline runs, exactly as CI runs them.
	// probe is native (cgo): verify leaves it to the native job, which
	// builds it on each OS.
	gomod, _ := os.ReadFile(filepath.Join(w.dir, "apps", "probe", "go.mod"))
	w.write("apps/probe/go.mod", "// smith:native\n"+string(gomod))
	if out := w.smith(true, "ci", "generate"); !strings.Contains(w.read(".github/workflows/ci.yml"), "run: smith ci native") {
		t.Fatalf("no native job:\n%s", out)
	}
	// CI starts from a clean checkout, and verify fails on a tracked file
	// changed under it, so the workspace is committed first.
	w.git("add", "-A")
	w.git("commit", "-qm", "feat: add every kind of project")
	for _, phase := range []string{"setup", "audit", "verify", "pack", "native"} {
		out := w.smith(true, "ci", phase)
		if !strings.Contains(out, "passed in") {
			t.Fatalf("ci %s:\n%s", phase, out)
		}
		if phase == "verify" && strings.Contains(out, "probe:") {
			t.Fatalf("verify ran the native project:\n%s", out)
		}
		if phase == "native" && !strings.Contains(out, "probe:build") {
			t.Fatalf("native did not build probe:\n%s", out)
		}
	}
	// pack leaves each app's deliverable in dist/drop: an npm tarball with
	// the built code, the Go binary, the Python wheel.
	drop := filepath.Join(w.dir, "dist", "drop")
	tarball := filepath.Join(drop, "acme-web-0.0.0.tgz")
	if !tarContains(t, tarball, "package/main.js") || tarContains(t, tarball, "package/src/main.ts") {
		t.Fatalf("%s must hold the built app, staged from dist/apps/web, and not its sources", tarball)
	}
	exe := ""
	if runtime.GOOS == "windows" {
		exe = ".exe"
	}
	for _, want := range []string{filepath.Join("api", "api"+exe), filepath.Join("probe", "probe"+exe), "acme_worker-0.0.0-py3-none-any.whl", "acme-hello-0.0.0.tgz"} {
		if _, err := os.Stat(filepath.Join(drop, want)); err != nil {
			entries, _ := os.ReadDir(drop)
			t.Fatalf("dist/drop lacks %s: %v", want, entries)
		}
	}
	if err := os.RemoveAll(filepath.Join(w.dir, "dist")); err != nil {
		t.Fatal(err)
	}

	// 5. Affected follows dependents and stays narrow.
	// The phases changed no tracked file, which verify now guarantees.
	if status := strings.TrimSpace(w.git("status", "--porcelain", "--untracked-files=no")); status != "" {
		t.Fatalf("the CI phases changed tracked files:\n%s", status)
	}
	w.git("add", "-A")
	w.git("commit", "--allow-empty", "-qm", "chore: after the ci phases")
	w.git("branch", "base")
	w.write("libs/ui/src/greeting/extra.ts", "export const extra = 1\n")
	affected := w.smith(true, "affected", "--base", "base")
	if !strings.Contains(affected, "ui") || !strings.Contains(affected, "web") || strings.Contains(affected, "api") {
		t.Fatalf("affected:\n%s", affected)
	}
	// The graph shows the same set, the changed project marked, in any format.
	var view struct {
		Projects []struct {
			Name   string
			Marked bool
		}
	}
	graphJSON := w.smith(true, "graph", "--affected", "--base", "base", "--format", "json")
	if err := json.Unmarshal([]byte(graphJSON[strings.Index(graphJSON, "{"):]), &view); err != nil {
		t.Fatal(err)
	}
	got := map[string]bool{}
	for _, p := range view.Projects {
		got[p.Name] = p.Marked
	}
	if marked, ok := got["ui"]; !ok || !marked || got["web"] || len(got) != 2 {
		t.Fatalf("graph --affected: %v", got)
	}
	if text := w.smith(true, "graph", "--focus", "core"); !strings.Contains(text, "core *") || !strings.Contains(text, "api") || strings.Contains(text, "web") {
		t.Fatalf("graph --focus core:\n%s", text)
	}
	_ = os.Remove(filepath.Join(w.dir, "libs", "ui", "src", "greeting", "extra.ts"))

	// 6. Release: preview, a release interrupted halfway, and a resume.
	reg := startRegistries(t, base)
	remote := filepath.Join(base, "remote.git")
	run(t, base, "git", "init", "-q", "--bare", remote)
	w.git("remote", "add", "origin", remote)
	npmrc, _ := os.ReadFile(filepath.Join(w.dir, ".npmrc"))
	host := strings.TrimPrefix(reg.npmURL, "http:")
	w.write(".npmrc", string(npmrc)+"registry="+reg.npmURL+"\n"+host+":_authToken=${NODE_AUTH_TOKEN}\n")
	w.git("add", "-A")
	w.git("commit", "-qm", "ci: publish to the test registries")
	w.git("push", "-q", "origin", "main")

	preview := w.smith(true, "release", "--dry-run")
	for _, p := range []string{"gokit", "pykit", "sdk"} {
		if !strings.Contains(preview, p) {
			t.Fatalf("preview lacks %s:\n%s", p, preview)
		}
	}
	w.env = append(w.env, "NODE_AUTH_TOKEN="+reg.npmToken, "TWINE_USERNAME=ci", "TWINE_PASSWORD=ci", "TWINE_REPOSITORY_URL=http://localhost:1/")
	failed := w.smith(false, "release", "--yes")
	if !strings.Contains(failed, "Stopped") || !strings.Contains(w.git("ls-remote", "--tags", "origin"), "packages/gokit/v0.1.0") {
		t.Fatalf("the release must stop at pykit, keeping gokit:\n%s", failed)
	}
	if status := w.git("status", "--porcelain"); strings.TrimSpace(status) != "" {
		t.Fatalf("a failed release must restore manifests:\n%s", status)
	}
	w.env[len(w.env)-1] = "TWINE_REPOSITORY_URL=" + reg.pypiURL
	resumed := w.smith(true, "release", "--yes")
	if !strings.Contains(resumed, "Unchanged: gokit") {
		t.Fatalf("the resume must keep gokit:\n%s", resumed)
	}
	tags := w.git("ls-remote", "--tags", "origin")
	for _, tag := range []string{"packages/gokit/v0.1.0", "pykit@0.1.0", "sdk@0.1.0"} {
		if !strings.Contains(tags, tag) {
			t.Fatalf("remote lacks %s:\n%s", tag, tags)
		}
	}
	if body := get(t, reg.npmURL+"@acme%2fsdk"); !strings.Contains(body, `"0.1.0"`) {
		t.Fatalf("npm registry lacks @acme/sdk@0.1.0:\n%s", body)
	}
	if body := get(t, reg.pypiURL+"simple/acme-pykit/"); !strings.Contains(body, "acme_pykit-0.1.0") {
		t.Fatalf("index lacks acme-pykit 0.1.0:\n%s", body)
	}
	if again := w.smith(true, "release", "--yes"); !strings.Contains(again, "Nothing to release") {
		t.Fatalf("a third run releases nothing:\n%s", again)
	}

	// A published version whose tag is gone (here and on the remote) would
	// restart the package from its manifest: release refuses, even a dry run,
	// and doctor reports it, until the tag is back.
	commit := strings.TrimSpace(w.git("rev-parse", "sdk@0.1.0^{commit}"))
	w.git("tag", "-d", "sdk@0.1.0")
	w.git("push", "-q", "origin", ":refs/tags/sdk@0.1.0")
	if out := w.smith(false, "release", "--dry-run"); !strings.Contains(out, "sdk is published at 0.1.0 according to") {
		t.Fatalf("a published, untagged package must block the release:\n%s", out)
	}
	if out := w.smith(false, "doctor"); !strings.Contains(out, "release tags are not behind the registry") {
		t.Fatalf("doctor must report it:\n%s", out)
	}
	w.git("tag", "sdk@0.1.0", commit)
	w.git("push", "-q", "origin", "sdk@0.1.0")
	if out := w.smith(true, "doctor"); strings.Contains(out, "published at") {
		t.Fatalf("doctor still reports the restored tag:\n%s", out)
	}

	// 7. A fix releases a patch.
	w.write("packages/sdk/src/fix.ts", "export const fixed = true\n")
	w.git("add", "-A")
	w.git("commit", "-qm", "fix(sdk): a patch")
	if plan := w.smith(true, "release", "--dry-run", "--notes"); !strings.Contains(plan, "0.1.0 → 0.1.1") || !strings.Contains(plan, "## Fixes\n\n- **sdk:** a patch (") || !strings.Contains(plan, "Thanks to ") {
		t.Fatalf("a fix must plan a patch, with its notes:\n%s", plan)
	}

	// 8. Removing projects: one others import is refused; the rest go with
	// their traces (go.work, the lockfile, the venv) and the pipelines follow.
	if out := w.smith(false, "remove", "core", "--dry-run"); !strings.Contains(out, "core is imported by api") {
		t.Fatalf("removing an imported project must be refused:\n%s", out)
	}
	if out := w.smith(true, "remove", "gokit", "--dry-run"); !strings.Contains(out, "Would remove gokit (packages/gokit)") || !strings.Contains(out, "changes go.work") {
		t.Fatalf("dry run:\n%s", out)
	}
	if _, err := os.Stat(filepath.Join(w.dir, "packages", "gokit")); err != nil {
		t.Fatal("a dry run changes nothing")
	}
	// npm installs from the public registry again, not the test one.
	w.write(".npmrc", string(npmrc))
	for _, name := range []string{"gokit", "hello", "pykit"} {
		if out := w.smith(true, "remove", name, "--yes"); strings.Contains(out, "warning") {
			t.Fatalf("remove %s:\n%s", name, out)
		}
	}
	// The graph's history: compared with the last commit, three projects went.
	if out := w.smith(true, "graph", "--diff", "HEAD"); !strings.Contains(out, "- gokit (go package, packages/gokit)") || !strings.Contains(out, "- hello (node app, apps/hello)") || !strings.Contains(out, "- pykit (python package, packages/pykit)") {
		t.Fatalf("graph --diff HEAD:\n%s", out)
	}
	if work, _ := os.ReadFile(filepath.Join(w.dir, "go.work")); strings.Contains(string(work), "gokit") {
		t.Fatalf("go.work still uses gokit:\n%s", work)
	}
	if lock, _ := os.ReadFile(filepath.Join(w.dir, "package-lock.json")); strings.Contains(string(lock), "apps/hello") {
		t.Fatal("the lockfile still has hello")
	}
	for _, gone := range []string{"packages/gokit", "apps/hello", "packages/pykit"} {
		if _, err := os.Stat(filepath.Join(w.dir, filepath.FromSlash(gone))); err == nil {
			t.Fatalf("%s still exists", gone)
		}
	}
	w.smith(true, "ci", "generate", "--check")
	if out := w.smith(true, "run", "build"); !strings.Contains(out, "tasks succeeded") {
		t.Fatalf("the workspace still builds:\n%s", out)
	}
}

func hasEdge(graphJSON, from, to string) bool {
	var g struct {
		Projects []struct {
			Name string   `json:"name"`
			Deps []string `json:"deps"`
		} `json:"projects"`
	}
	if json.Unmarshal([]byte(graphJSON), &g) != nil {
		return false
	}
	for _, p := range g.Projects {
		if p.Name == from {
			for _, d := range p.Deps {
				if d == to {
					return true
				}
			}
		}
	}
	return false
}

func get(t *testing.T, url string) string {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	var b strings.Builder
	buf := make([]byte, 32<<10)
	for {
		n, err := resp.Body.Read(buf)
		b.Write(buf[:n])
		if err != nil {
			break
		}
	}
	return b.String()
}

// tarContains reports whether a .tgz holds a file.
func tarContains(t *testing.T, archive, name string) bool {
	t.Helper()
	f, err := os.Open(archive)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	gz, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	r := tar.NewReader(gz)
	for {
		h, err := r.Next()
		if err != nil {
			return false
		}
		if h.Name == name {
			return true
		}
	}
}
