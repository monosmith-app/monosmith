//go:build e2e

package goldenpath

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestCppWorkspace builds CMake projects with the built-in cpp plugin: an
// app adding a package with add_subdirectory, configure, build, CTest and
// cppcheck, the cache, dist/ for every build tree, and a release plan.
func TestCppWorkspace(t *testing.T) {
	for _, tool := range []string{"cmake", "ninja", "cppcheck"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skip(tool + " is not installed")
		}
	}
	if _, err := exec.LookPath("c++"); err != nil {
		if _, err := exec.LookPath("g++"); err != nil {
			t.Skip("no C++ compiler (c++ or g++) is installed")
		}
	}
	// Some cppcheck builds (the MinGW bundle on the Windows runners) cannot
	// find their own std.cfg and fail on any file.
	probe := filepath.Join(t.TempDir(), "probe.cpp")
	_ = os.WriteFile(probe, []byte("int main() { return 0; }\n"), 0o644)
	if out, err := exec.Command("cppcheck", "--quiet", "--error-exitcode=1", probe).CombinedOutput(); err != nil {
		t.Skipf("cppcheck cannot check a file here: %v\n%s", err, out)
	}
	base := t.TempDir()
	w := &workspace{t: t, bin: buildSmith(t, base), dir: base, env: identity()}
	w.smith(true, "new", "smith-e2e-cpp", "--ci", "github")
	w.dir = filepath.Join(base, "smith-e2e-cpp")
	w.smith(true, "add", "package", "core", "--lang", "cpp")
	w.smith(true, "add", "app", "api", "--lang", "cpp")

	// api uses core: add_subdirectory of a workspace project is a graph edge.
	lists := "apps/api/CMakeLists.txt"
	w.write(lists, strings.Replace(w.read(lists), "target_link_libraries(smith_e2e_cpp_api PRIVATE smith_e2e_cpp_api_lib)",
		"add_subdirectory(${CMAKE_CURRENT_SOURCE_DIR}/../../packages/core smith_e2e_cpp_core)\ntarget_link_libraries(smith_e2e_cpp_api PRIVATE smith_e2e_cpp_api_lib smith_e2e_cpp_core)", 1))
	main := "apps/api/src/main.cpp"
	w.write(main, strings.Replace(strings.Replace(w.read(main),
		"#include \"smith_e2e_cpp_api/greeting.hpp\"\n", "#include \"smith_e2e_cpp_api/greeting.hpp\"\n#include \"smith_e2e_cpp_core/greeting.hpp\"\n", 1),
		"<< '\\n';", "<< ' ' << smith_e2e_cpp_core::greeting::greet(\"core\") << '\\n';", 1))
	if !hasEdge(w.smith(true, "graph", "--format", "json"), "api", "core") {
		t.Fatal("add_subdirectory of a workspace project is a graph edge")
	}
	out := w.smith(true, "run", "lint", "test", "build")
	for _, want := range []string{"api:build ✓", "core:test ✓", "api:test ✓", "core:lint ✓", "api:lint ✓"} {
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
	// The build trees are in dist/, never build/ beside the sources.
	for _, tree := range []string{"dist/apps/api/CMakeCache.txt", "dist/packages/core/CMakeCache.txt"} {
		if _, err := os.Stat(filepath.Join(w.dir, filepath.FromSlash(tree))); err != nil {
			t.Errorf("%s was not configured", tree)
		}
	}
	if out := w.smith(true, "doctor"); !strings.Contains(out, "✓ cppcheck is installed") || !strings.Contains(out, "✓ every build writes to dist/") {
		t.Fatalf("doctor:\n%s", out)
	}
	if ci := w.read(".github/workflows/ci.yml"); !strings.Contains(ci, "ninja-build g++ cppcheck") {
		t.Fatalf("ci.yml installs the C++ toolchain:\n%s", ci)
	}

	w.git("add", "-A")
	w.git("commit", "-qm", "feat: an api and its core")
	if plan := w.smith(true, "release", "--dry-run"); !strings.Contains(plan, "core") || !strings.Contains(plan, "0.1.0") {
		t.Fatalf("release plan:\n%s", plan)
	}
}
