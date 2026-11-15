//go:build e2e

package goldenpath

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestPhpWorkspace builds Composer packages with the built-in php plugin: an
// app requiring a package in one workspace vendor/, PHPStan and PHPUnit,
// the cache, and a release plan.
func TestPhpWorkspace(t *testing.T) {
	for _, tool := range []string{"php", "composer"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skip(tool + " is not installed")
		}
	}
	base := t.TempDir()
	w := &workspace{t: t, bin: buildSmith(t, base), dir: base, env: identity()}
	// A vendor no package on Packagist has: the release checks ask Packagist.
	w.smith(true, "new", "smith-e2e-phpish", "--ci", "github")
	w.dir = filepath.Join(base, "smith-e2e-phpish")
	w.smith(true, "add", "package", "core", "--lang", "php")
	w.smith(true, "add", "app", "cli", "--lang", "php")

	// cli uses core: a requirement on a workspace package is a graph edge.
	manifest := "apps/cli/composer.json"
	w.write(manifest, strings.Replace(w.read(manifest), `"php": ">=8.2"`, `"php": ">=8.2",`+"\n        "+`"smith-e2e-phpish/core": "*"`, 1))
	bin := "apps/cli/bin/cli"
	w.write(bin, strings.Replace(strings.Replace(w.read(bin),
		"use SmithE2ePhpish\\Cli\\Greeting\\GreetUseCase;", "use SmithE2ePhpish\\Cli\\Greeting\\GreetUseCase;\nuse SmithE2ePhpish\\Core\\Greeting\\GreetUseCase as CoreGreeting;", 1),
		"echo GreetUseCase::greet('world'), PHP_EOL;", "echo GreetUseCase::greet('world'), ' ', CoreGreeting::greet('core'), PHP_EOL;", 1))
	w.smith(true, "install")
	if !hasEdge(w.smith(true, "graph", "--format", "json"), "cli", "core") {
		t.Fatal("a workspace requirement is a graph edge")
	}
	out := w.smith(true, "run", "lint", "test")
	for _, want := range []string{"cli:lint ✓", "core:lint ✓", "cli:test ✓", "core:test ✓"} {
		if !strings.Contains(out, want) {
			t.Fatalf("%s did not pass:\n%s", want, out)
		}
	}
	if out := w.smith(true, "run", "lint", "test"); !strings.Contains(out, "(6 from cache") {
		t.Fatalf("a re-run replays from the cache:\n%s", out)
	}
	if out := w.smith(true, "run", "start", "-p", "cli"); !strings.Contains(out, "Hello, world! Hello, core!") {
		t.Fatalf("start:\n%s", out)
	}
	if out := w.smith(true, "doctor"); !strings.Contains(out, "✓ Composer is installed") {
		t.Fatalf("doctor:\n%s", out)
	}
	if ci := w.read(".github/workflows/ci.yml"); !strings.Contains(ci, "shivammathur/setup-php") {
		t.Fatalf("ci.yml installs PHP:\n%s", ci)
	}

	w.git("add", "-A")
	w.git("commit", "-qm", "feat: a cli and its core")
	if plan := w.smith(true, "release", "--dry-run"); !strings.Contains(plan, "core") || !strings.Contains(plan, "0.1.0") {
		t.Fatalf("release plan:\n%s", plan)
	}
}
