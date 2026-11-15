//go:build e2e

package goldenpath

import (
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestReleaseDryRunPlansCppArchives dry-runs the release of a workspace with
// two C++ apps: each is planned for this machine's platform, one is native
// (its own checksum file) and a platform of another system is refused with
// the way out. Nothing is built or published.
func TestReleaseDryRunPlansCppArchives(t *testing.T) {
	if _, err := exec.LookPath("cmake"); err != nil {
		t.Skip("cmake is not installed")
	}
	base := t.TempDir()
	w := &workspace{t: t, bin: buildSmith(t, base), dir: base, env: identity()}
	w.smith(true, "new", "smith-e2e-cpp-archives", "--ci", "github")
	w.dir = filepath.Join(base, "smith-e2e-cpp-archives")
	w.smith(true, "add", "app", "cli", "--lang", "cpp")
	w.smith(true, "add", "app", "tool", "--lang", "cpp")
	host := runtime.GOOS + "/" + runtime.GOARCH
	other := "linux/amd64"
	if runtime.GOOS == "linux" {
		other = "windows/amd64"
	}
	w.write("apps/tool/CMakeLists.txt", w.read("apps/tool/CMakeLists.txt")+"\nset(SMITH_NATIVE ON)\n")
	w.write("smith.toml", w.read("smith.toml")+
		"\n[projects.cli.release]\nplatforms = [\""+host+"\"]\n"+
		"\n[projects.tool.release]\nplatforms = [\""+host+"\"]\nproduct = \"acme-tool\"\n")
	w.git("add", "-A")
	w.git("commit", "-qm", "feat: two apps")

	out := w.smith(true, "release", "--dry-run")
	for _, want := range []string{"dist/release/cli/cli_0.1.0_", "dist/release/cli/checksums.txt", "dist/release/tool/acme-tool_0.1.0_", ".zip.sha256"} {
		if !strings.Contains(out, want) {
			t.Fatalf("the dry run does not plan %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "dist/release/tool/checksums.txt") {
		t.Fatalf("a native app has no combined checksums:\n%s", out)
	}

	w.write("smith.toml", strings.Replace(w.read("smith.toml"), "[projects.cli.release]\nplatforms = [\""+host+"\"]", "[projects.cli.release]\nplatforms = [\""+host+"\", \""+other+"\"]", 1))
	if out := w.smith(false, "release", "--dry-run"); !strings.Contains(out, other) || !strings.Contains(out, "SMITH_NATIVE") {
		t.Fatalf("a platform this machine cannot build is refused with the way out:\n%s", out)
	}
}
