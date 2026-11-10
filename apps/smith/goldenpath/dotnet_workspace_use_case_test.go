//go:build e2e

package goldenpath

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestDotnetWorkspace builds .NET projects with the built-in dotnet plugin:
// an app referencing a package, each with its tests project, every dotnet
// task, the cache, dist/ for every build, and a release plan.
func TestDotnetWorkspace(t *testing.T) {
	if _, err := exec.LookPath("dotnet"); err != nil {
		t.Skip("the .NET SDK is not installed")
	}
	base := t.TempDir()
	w := &workspace{t: t, bin: buildSmith(t, base), dir: base, env: identity()}
	// A name no package on nuget.org has: the release checks ask the feed.
	w.smith(true, "new", "smith-e2e-dotnety", "--ci", "github")
	w.dir = filepath.Join(base, "smith-e2e-dotnety")
	w.smith(true, "add", "app", "api", "--lang", "dotnet")
	w.smith(true, "add", "package", "core", "--lang", "dotnet")

	// api uses core: a project reference is a graph edge.
	csproj := "apps/api/SmithE2eDotnety.Api.csproj"
	w.write(csproj, strings.Replace(w.read(csproj), "  <!-- The tests project", "  <ItemGroup>\n    <ProjectReference Include=\"../../packages/core/SmithE2eDotnety.Core.csproj\" />\n  </ItemGroup>\n\n  <!-- The tests project", 1))
	w.write("apps/api/Program.cs", "Console.WriteLine($\"{SmithE2eDotnety.Api.Greeting.GreetUseCase.Greet(\"world\").Message} {SmithE2eDotnety.Core.Greeting.GreetUseCase.Greet(\"core\").Message}\");\n")
	if !hasEdge(w.smith(true, "graph", "--format", "json"), "api", "core") {
		t.Fatal("a project reference is a graph edge")
	}
	out := w.smith(true, "run", "lint", "test", "build")
	for _, want := range []string{"api:build ✓", "core:test ✓", "api:test ✓", "core:lint ✓"} {
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
	// bin/ and obj/ go to dist/, never into a source tree.
	for _, dir := range []string{"apps/api/bin", "apps/api/obj", "packages/core/tests/obj"} {
		if _, err := os.Stat(filepath.Join(w.dir, filepath.FromSlash(dir))); err == nil {
			t.Errorf("%s exists: .NET builds must go to dist/", dir)
		}
	}
	if _, err := os.Stat(filepath.Join(w.dir, "dist", "dotnet")); err != nil {
		t.Error("the builds are under dist/dotnet")
	}
	if out := w.smith(true, "doctor"); !strings.Contains(out, "✓ the .NET SDK is installed") || !strings.Contains(out, "✓ every build writes to dist/") {
		t.Fatalf("doctor:\n%s", out)
	}
	if ci := w.read(".github/workflows/ci.yml"); !strings.Contains(ci, "actions/setup-dotnet") || !strings.Contains(ci, "NUGET_API_KEY") {
		t.Fatalf("ci.yml installs .NET and passes the NuGet key:\n%s", ci)
	}

	w.git("add", "-A")
	w.git("commit", "-qm", "feat: an api and its core")
	if plan := w.smith(true, "release", "--dry-run"); !strings.Contains(plan, "core") || !strings.Contains(plan, "0.1.0") {
		t.Fatalf("release plan:\n%s", plan)
	}
}
