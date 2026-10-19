//go:build e2e

package goldenpath

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestJvmWorkspace builds Maven modules with the built-in jvm plugin: an app
// depending on a package, every Maven task, the cache, dist/ for every
// build, and a release plan.
func TestJvmWorkspace(t *testing.T) {
	for _, tool := range []string{"java", "mvn"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skip(tool + " is not installed")
		}
	}
	// The modules install into the user's Maven repository; leave it as found.
	if home, err := os.UserHomeDir(); err == nil {
		t.Cleanup(func() { _ = os.RemoveAll(filepath.Join(home, ".m2", "repository", "dev", "smithe2ejvm")) })
	}
	base := t.TempDir()
	w := &workspace{t: t, bin: buildSmith(t, base), dir: base, env: identity()}
	w.smith(true, "new", "smith-e2e-jvm", "--ci", "github")
	w.dir = filepath.Join(base, "smith-e2e-jvm")
	w.smith(true, "add", "app", "api", "--lang", "jvm")
	w.smith(true, "add", "package", "core", "--lang", "jvm")

	// api uses core: a dependency on a workspace module is a graph edge.
	pom := "apps/api/pom.xml"
	w.write(pom, strings.Replace(w.read(pom), "  <dependencies>\n", "  <dependencies>\n    <dependency>\n      <groupId>dev.smithe2ejvm</groupId>\n      <artifactId>smith-e2e-jvm-core</artifactId>\n      <version>0.0.0</version>\n    </dependency>\n", 1))
	main := "apps/api/src/main/java/dev/smithe2ejvm/api/Main.java"
	w.write(main, strings.Replace(w.read(main), `    System.out.println(GreetUseCase.greet("world"));`, `    String core = dev.smithe2ejvm.core.greeting.GreetUseCase.greet("core");
    System.out.println(GreetUseCase.greet("world") + " " + core);`, 1))
	if !hasEdge(w.smith(true, "graph", "--format", "json"), "api", "core") {
		t.Fatal("a module dependency is a graph edge")
	}
	out := w.smith(true, "run", "lint", "test", "build")
	for _, want := range []string{"api:build ✔", "core:test ✔", "api:test ✔", "core:lint ✔"} {
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
	// Builds go to dist/, never target/ in a source tree.
	for _, dir := range []string{"apps/api/target", "packages/core/target"} {
		if _, err := os.Stat(filepath.Join(w.dir, filepath.FromSlash(dir))); err == nil {
			t.Errorf("%s exists: Maven builds must go to dist/", dir)
		}
	}
	if _, err := os.Stat(filepath.Join(w.dir, "dist", "maven", "smith-e2e-jvm-core")); err != nil {
		t.Error("the builds are under dist/maven")
	}
	if out := w.smith(true, "doctor"); !strings.Contains(out, "✔ Maven is installed") || !strings.Contains(out, "✔ every build writes to dist/") {
		t.Fatalf("doctor:\n%s", out)
	}
	if ci := w.read(".github/workflows/ci.yml"); !strings.Contains(ci, "actions/setup-java") || !strings.Contains(ci, "MAVEN_REPOSITORY_URL") {
		t.Fatalf("ci.yml installs Java and passes the repository:\n%s", ci)
	}

	w.git("add", "-A")
	w.git("commit", "-qm", "feat: an api and its core")
	if plan := w.smith(true, "release", "--dry-run"); !strings.Contains(plan, "core") || !strings.Contains(plan, "0.1.0") {
		t.Fatalf("release plan:\n%s", plan)
	}
}

// TestJvmPublishesToGitHubPackagesWithTheJobsToken names GitHub Packages as
// the release registry and checks the generated workflow: the job's token
// with packages: write, and no Maven secret to create.
func TestJvmPublishesToGitHubPackagesWithTheJobsToken(t *testing.T) {
	base := t.TempDir()
	w := &workspace{t: t, bin: buildSmith(t, base), dir: base, env: identity()}
	w.smith(true, "new", "smith-e2e-jvm-gh", "--ci", "github")
	w.dir = filepath.Join(base, "smith-e2e-jvm-gh")
	w.smith(true, "add", "package", "core", "--lang", "jvm")
	w.write("smith.toml", w.read("smith.toml")+"\n[registries.github]\nauth = \"github-packages\"\nmaven = \"github\"\n\n[release]\nregistry = \"github\"\n")
	w.smith(true, "ci", "generate")
	ci := w.read(".github/workflows/ci.yml")
	if !strings.Contains(ci, "  packages: write\n") {
		t.Fatalf("ci.yml grants packages: write:\n%s", ci)
	}
	for _, secret := range []string{"MAVEN_REPOSITORY_URL", "MAVEN_USERNAME", "MAVEN_PASSWORD"} {
		if strings.Contains(ci, secret) {
			t.Fatalf("ci.yml reads %s, which GitHub Packages does not need:\n%s", secret, ci)
		}
	}
	if !strings.Contains(ci, "GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}") {
		t.Fatalf("ci.yml gives the release the job's token:\n%s", ci)
	}
	if out := w.smith(true, "ci", "generate", "--check"); strings.Contains(out, "MAVEN_") {
		t.Fatalf("nothing to create by hand:\n%s", out)
	}
}
