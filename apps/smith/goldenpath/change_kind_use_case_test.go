//go:build e2e

package goldenpath

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// kindScenario is one language's workspace for smith kind: a lib core that an
// app cli uses.
type kindScenario struct {
	lang string
	// tools are programs the scenario needs; it is skipped without them.
	tools []string
	// wire makes cli use core, as the language does it.
	wire func(w *workspace)
	// install is what the language needs after the folder moved (npm links
	// workspace packages, pip installs them editable).
	install bool
	// untracked are files the language's tools rewrite on their own, left out
	// of the "as it was" comparison.
	untracked []string
	// tag is a release tag of core as a package; Go resolves a module's version
	// from the tags of its folder.
	tag string
	// coreFiles are files that name core's folder and must follow the move.
	coreFiles map[string]string
}

func TestKindChange(t *testing.T) {
	for _, sc := range []kindScenario{
		{
			lang:  "go",
			tools: []string{"go"},
			tag:   "packages/core/v1.0.0",
			wire: func(w *workspace) {
				w.write("apps/cli/main.go", "package main\n\nimport (\n\t\"fmt\"\n\n\tcore \"example.com/kcgo/libs/core/greeting\"\n)\n\nfunc main() { fmt.Println(core.Greet(\"core\").Message) }\n")
			},
			coreFiles: map[string]string{"go.work": "./packages/core", "apps/cli/main.go": "example.com/kcgo/packages/core/greeting", "packages/core/go.mod": "module example.com/kcgo/packages/core"},
		},
		{
			lang:  "rust",
			tools: []string{"cargo"},
			wire: func(w *workspace) {
				w.write("apps/cli/Cargo.toml", strings.Replace(w.read("apps/cli/Cargo.toml"), "[dependencies]\n", "[dependencies]\nkcrust-core = { path = \"../../libs/core\", version = \"0.0.0\" }\n", 1))
				w.write("apps/cli/src/main.rs", "fn main() {\n    println!(\"{}\", kcrust_core::greeting::greet(\"core\"));\n}\n")
			},
			coreFiles: map[string]string{"apps/cli/Cargo.toml": "../../packages/core", "Cargo.toml": "packages/core"},
		},
		{
			lang:    "node",
			tools:   []string{"node", "npm"},
			install: true,
			wire: func(w *workspace) {
				w.write("apps/cli/package.json", strings.Replace(w.read("apps/cli/package.json"), "\"scripts\"", "\"dependencies\": { \"@kcnode/core\": \"*\" },\n  \"scripts\"", 1))
				w.write("apps/cli/src/main.js", "const core = require('@kcnode/core')\n\nprocess.stdout.write(core.greet('core').message + '\\n')\n")
			},
			untracked: []string{"package-lock.json"},
			coreFiles: map[string]string{"package.json": "packages/*"},
		},
		{
			lang:    "python",
			tools:   []string{"python"},
			install: true,
			wire: func(w *workspace) {
				w.write("apps/cli/src/kcpython_cli/__main__.py", "from kcpython_core.greeting import greet\n\nprint(greet('core').message)\n")
			},
			untracked: []string{"uv.lock"},
			coreFiles: map[string]string{},
		},
	} {
		t.Run(sc.lang, func(t *testing.T) { runKindScenario(t, sc) })
	}
}

func runKindScenario(t *testing.T, sc kindScenario) {
	for _, tool := range sc.tools {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skip(tool + " is not installed")
		}
	}
	base := t.TempDir()
	name := "kc" + sc.lang
	w := &workspace{t: t, bin: buildSmith(t, base), dir: base, env: identity()}
	w.smith(true, "new", name, "--ci", "github")
	w.dir = filepath.Join(base, name)
	w.smith(true, "add", "lib", "core", "--lang", sc.lang)
	w.smith(true, "add", "app", "cli", "--lang", sc.lang)
	sc.wire(w)
	if sc.install {
		w.smith(true, "install")
	}
	healthy := func(stage string) {
		t.Helper()
		if out := w.smith(true, "doctor"); strings.Contains(out, "✖") {
			t.Fatalf("%s: doctor is not clean:\n%s", stage, out)
		}
		w.smith(true, "run", "build", "test")
	}
	healthy("before")
	releases := func() string { return w.smith(true, "release", "--dry-run", "--verbose") }
	w.git("add", "-A")
	w.git("commit", "-qm", "feat: a cli and its core")
	start := strings.TrimSpace(w.git("rev-parse", "HEAD"))
	if plan := releases(); strings.Contains(plan, "core") {
		t.Fatalf("an internal lib is not in the release plan:\n%s", plan)
	}

	// A dry run says everything and changes nothing.
	dry := w.smith(true, "kind", "core", "package", "--dry-run")
	for _, want := range []string{"Would make core a package", "moves   libs/core -> packages/core", "edits   libs/core/"} {
		if !strings.Contains(dry, want) {
			t.Fatalf("dry run lacks %q:\n%s", want, dry)
		}
	}
	if _, err := os.Stat(filepath.Join(w.dir, "libs", "core")); err != nil {
		t.Fatal("a dry run moved the folder")
	}
	if status := strings.TrimSpace(w.git("status", "--porcelain")); status != "" {
		t.Fatalf("a dry run changed the workspace:\n%s", status)
	}

	// lib -> package: moved with history, references follow, still healthy,
	// and the release plan now publishes it.
	w.smith(true, "kind", "core", "package")
	if _, err := os.Stat(filepath.Join(w.dir, "libs", "core")); err == nil {
		t.Fatal("libs/core is still there")
	}
	if renames := w.git("status", "--short"); !strings.Contains(renames, "packages/core/") || !strings.Contains(renames, "R") {
		t.Fatalf("the folder moved with git mv:\n%s", renames)
	}
	for file, want := range sc.coreFiles {
		if got := w.read(file); !strings.Contains(got, want) {
			t.Fatalf("%s lacks %q:\n%s", file, want, got)
		}
	}
	if sc.install {
		w.smith(true, "install")
	}
	healthy("as a package")
	if plan := releases(); !strings.Contains(plan, "core") {
		t.Fatalf("a package is in the release plan:\n%s", plan)
	}
	if out := w.smith(true, "projects"); !strings.Contains(out, "package") {
		t.Fatalf("projects:\n%s", out)
	}

	// A released package is not quietly orphaned.
	tag := sc.tag
	if tag == "" {
		tag = "core@1.0.0"
	}
	w.git("tag", tag)
	if out := w.smith(false, "kind", "core", "lib"); !strings.Contains(out, tag) || !strings.Contains(out, "--force") {
		t.Fatalf("a released package is refused, naming why:\n%s", out)
	}
	w.git("tag", "-d", tag)

	// package -> lib: back where it was.
	w.smith(true, "kind", "core", "lib")
	if sc.install {
		w.smith(true, "install")
	}
	healthy("as a lib again")
	if plan := releases(); strings.Contains(plan, "core") {
		t.Fatalf("core left the release plan:\n%s", plan)
	}
	w.git("add", "-A")
	args := []string{"diff", "--cached", "--stat", start, "--", "."}
	for _, f := range sc.untracked {
		args = append(args, ":(exclude)"+f)
	}
	if diff := strings.TrimSpace(w.git(args...)); diff != "" {
		t.Fatalf("lib -> package -> lib left the workspace changed:\n%s", diff)
	}

	// An app that others import is refused, and the plugin's own limits too.
	if out := w.smith(false, "kind", "core", "app"); !strings.Contains(out, "an app is run, not imported") {
		t.Fatalf("an imported lib cannot become an app:\n%s", out)
	}
	if out := w.smith(false, "kind", "core", "lib"); !strings.Contains(out, "already a lib") {
		t.Fatalf("the same kind:\n%s", out)
	}
}
