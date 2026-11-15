//go:build e2e

package goldenpath

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestDartWorkspace builds a pub workspace with the built-in dart plugin: a
// Dart package, a Dart app and a Flutter web app (the flutter stack) using
// the package, every dart and flutter task, the cache, dist/ for every
// build, and a release plan.
func TestDartWorkspace(t *testing.T) {
	for _, tool := range []string{"dart", "flutter"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skip(tool + " is not installed")
		}
	}
	base := t.TempDir()
	w := &workspace{t: t, bin: buildSmith(t, base), dir: base, env: identity()}
	// A name no package on pub.dev has: the release checks ask pub.dev.
	w.smith(true, "new", "smith-e2e-darty", "--ci", "github")
	w.dir = filepath.Join(base, "smith-e2e-darty")
	w.smith(true, "add", "package", "core", "--lang", "dart")
	w.smith(true, "add", "app", "tool", "--lang", "dart")
	w.smith(true, "add", "app", "web", "--stack", "flutter")

	// web uses core: a dependency on a workspace package is a graph edge.
	pubspec := "apps/web/pubspec.yaml"
	w.write(pubspec, strings.Replace(w.read(pubspec), "dependencies:\n  flutter:\n", "dependencies:\n  smith_e2e_darty_core: ^0.0.0\n  flutter:\n", 1))
	main := "apps/web/lib/main.dart"
	w.write(main, strings.Replace(strings.Replace(w.read(main),
		"import 'src/greeting/greeting.dart';", "import 'package:smith_e2e_darty_core/smith_e2e_darty_core.dart' as core;\n\nimport 'src/greeting/greeting.dart';", 1),
		"      home: Scaffold(body: Center(child: Text(greet('world')))),\n",
		"      home: Scaffold(\n        body: Center(child: Text('${greet('world')} ${core.greet('core')}')),\n      ),\n", 1))
	w.smith(true, "install")
	if !hasEdge(w.smith(true, "graph", "--format", "json"), "web", "core") {
		t.Fatal("a workspace dependency is a graph edge")
	}
	out := w.smith(true, "run", "format", "lint", "test", "build")
	for _, want := range []string{"web:build ✓", "tool:build ✓", "web:test ✓", "core:test ✓", "core:lint ✓", "web:format ✓"} {
		if !strings.Contains(out, want) {
			t.Fatalf("%s did not pass:\n%s", want, out)
		}
	}
	if out := w.smith(true, "run", "lint", "test"); !strings.Contains(out, "(9 from cache") {
		t.Fatalf("a re-run replays from the cache:\n%s", out)
	}
	if out := w.smith(true, "run", "start", "-p", "tool"); !strings.Contains(out, "Hello, world!") {
		t.Fatalf("start:\n%s", out)
	}
	// Builds go to dist/, never build/ beside the sources.
	for _, built := range []string{"dist/apps/web/index.html", "dist/apps/tool"} {
		if _, err := os.Stat(filepath.Join(w.dir, filepath.FromSlash(built))); err != nil {
			t.Errorf("%s was not built", built)
		}
	}
	// flutter test keeps its caches in build/ (the app's .gitignore covers
	// it); the web build itself is in dist/.
	if _, err := os.Stat(filepath.Join(w.dir, "apps", "web", "build", "web")); err == nil {
		t.Error("apps/web/build/web exists: Flutter builds must go to dist/")
	}
	if out := w.smith(true, "doctor"); !strings.Contains(out, "✓ Flutter is installed") || !strings.Contains(out, "✓ every build writes to dist/") {
		t.Fatalf("doctor:\n%s", out)
	}
	if ci := w.read(".github/workflows/ci.yml"); !strings.Contains(ci, "subosito/flutter-action") || !strings.Contains(ci, "PUB_TOKEN") {
		t.Fatalf("ci.yml installs Flutter and passes the pub token:\n%s", ci)
	}

	w.git("add", "-A")
	w.git("commit", "-qm", "feat: a web app, a tool and their core")
	if plan := w.smith(true, "release", "--dry-run"); !strings.Contains(plan, "core") || !strings.Contains(plan, "0.1.0") {
		t.Fatalf("release plan:\n%s", plan)
	}
}
