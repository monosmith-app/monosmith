//go:build e2e

package goldenpath

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestKotlinWorkspace builds Gradle modules in Kotlin with the jvm plugin's
// kotlin stack: an app depending on a package, every Gradle task, ktlint,
// the cache, dist/ for every build, and a release plan.
func TestKotlinWorkspace(t *testing.T) {
	for _, tool := range []string{"java", "gradle"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skip(tool + " is not installed")
		}
	}
	base := t.TempDir()
	// The Gradle daemon holds the workspace's files open; stop it before
	// the directory is removed.
	t.Cleanup(func() { _ = exec.Command("gradle", "--stop", "-q").Run() })
	w := &workspace{t: t, bin: buildSmith(t, base), dir: base, env: identity()}
	w.smith(true, "new", "smith-e2e-kotlin", "--ci", "github")
	w.dir = filepath.Join(base, "smith-e2e-kotlin")
	w.smith(true, "add", "app", "api", "--stack", "kotlin")
	w.smith(true, "add", "package", "core", "--stack", "kotlin")

	// api uses core: a project dependency is a graph edge.
	script := "apps/api/build.gradle.kts"
	w.write(script, strings.Replace(w.read(script), "dependencies {\n", "dependencies {\n    implementation(project(\":packages:core\"))\n", 1))
	main := "apps/api/src/main/kotlin/dev/smithe2ekotlin/api/Main.kt"
	w.write(main, strings.NewReplacer(
		"import dev.smithe2ekotlin.api.greeting.GreetUseCase\n", "import dev.smithe2ekotlin.api.greeting.GreetUseCase\nimport dev.smithe2ekotlin.core.greeting.GreetUseCase as CoreGreetUseCase\n",
		`    println(GreetUseCase.greet("world"))`, `    println("${GreetUseCase.greet("world")} ${CoreGreetUseCase.greet("core")}")`,
	).Replace(w.read(main)))
	if !hasEdge(w.smith(true, "graph", "--format", "json"), "api", "core") {
		t.Fatal("a project dependency is a graph edge")
	}
	out := w.smith(true, "run", "lint", "test", "build")
	for _, want := range []string{"api:build ✓", "core:test ✓", "api:test ✓", "api:lint ✓", "core:lint ✓"} {
		if !strings.Contains(out, want) {
			t.Fatalf("%s did not pass:\n%s", want, out)
		}
	}
	if out := w.smith(true, "run", "lint", "test"); !strings.Contains(out, "(6 from cache") {
		t.Fatalf("a re-run replays from the cache:\n%s", out)
	}
	if out := w.smith(true, "run", "start", "-p", "api"); !strings.Contains(out, "Hello, world! Hello, core!") {
		t.Fatalf("start:\n%s", out)
	}
	// Builds go to dist/, never build/ in a source tree.
	for _, dir := range []string{"apps/api/build", "packages/core/build", "build"} {
		if _, err := os.Stat(filepath.Join(w.dir, filepath.FromSlash(dir))); err == nil {
			t.Errorf("%s exists: Gradle builds must go to dist/", dir)
		}
	}
	if _, err := os.Stat(filepath.Join(w.dir, "dist", "gradle", "smith-e2e-kotlin-core", "libs")); err != nil {
		t.Error("the builds are under dist/gradle")
	}
	// The app ships from dist/<project dir>: its jar is linked there.
	if shipped, _ := filepath.Glob(filepath.Join(w.dir, "dist", "apps", "api", "*.jar")); len(shipped) == 0 {
		t.Error("the api jar is not shipped to dist/apps/api")
	}
	if out := w.smith(true, "doctor"); !strings.Contains(out, "✓ Gradle is installed") || !strings.Contains(out, "✓ every build writes to dist/") {
		t.Fatalf("doctor:\n%s", out)
	}
	if ci := w.read(".github/workflows/ci.yml"); !strings.Contains(ci, "gradle-9.8.1-bin.zip") || !strings.Contains(ci, "MAVEN_REPOSITORY_URL") {
		t.Fatalf("ci.yml installs Gradle and passes the repository:\n%s", ci)
	}

	w.git("add", "-A")
	w.git("commit", "-qm", "feat: an api and its core")
	if plan := w.smith(true, "release", "--dry-run"); !strings.Contains(plan, "core") || !strings.Contains(plan, "0.1.0") {
		t.Fatalf("release plan:\n%s", plan)
	}
}
