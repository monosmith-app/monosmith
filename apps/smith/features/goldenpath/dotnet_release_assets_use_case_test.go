//go:build e2e

package goldenpath

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestReleaseDryRunPlansDotnetArchives dry-runs the release of a workspace
// with two .NET apps: a .NET program is managed code, so one machine plans
// every platform, each app's archives and checksums apart. Nothing is built
// or published.
func TestReleaseDryRunPlansDotnetArchives(t *testing.T) {
	if _, err := exec.LookPath("dotnet"); err != nil {
		t.Skip("the .NET SDK is not installed")
	}
	base := t.TempDir()
	w := &workspace{t: t, bin: buildSmith(t, base), dir: base, env: identity()}
	w.smith(true, "new", "smith-e2e-dotnet-archives", "--ci", "github")
	w.dir = filepath.Join(base, "smith-e2e-dotnet-archives")
	w.smith(true, "add", "app", "cli", "--lang", "dotnet")
	w.smith(true, "add", "app", "tool", "--lang", "dotnet")
	w.write("smith.toml", w.read("smith.toml")+
		"\n[projects.cli.release]\nplatforms = [\"windows/amd64\", \"linux/arm64\", \"darwin/arm64\"]\n"+
		"\n[projects.tool.release]\nproduct = \"acme-tool\"\nasset_name = \"{product}-{version}-{os}-{arch}.{ext}\"\n")
	w.git("add", "-A")
	w.git("commit", "-qm", "feat: two apps")

	out := w.smith(true, "release", "--dry-run")
	for _, want := range []string{
		"dist/release/cli/cli_0.1.0_windows_amd64.zip",
		"dist/release/cli/cli_0.1.0_linux_arm64.zip",
		"dist/release/cli/cli_0.1.0_darwin_arm64.zip",
		"dist/release/cli/checksums.txt",
		"dist/release/tool/acme-tool-0.1.0-linux-amd64.zip",
		"dist/release/tool/acme-tool-0.1.0-windows-arm64.zip",
		"dist/release/tool/checksums.txt",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("the dry run does not plan %q:\n%s", want, out)
		}
	}

	w.write("smith.toml", strings.Replace(w.read("smith.toml"), "[projects.cli.release]\nplatforms = [", "[projects.cli.release]\nplatforms = [\"linux/mips\", ", 1))
	if out := w.smith(false, "release", "--dry-run"); !strings.Contains(out, "linux/mips") || !strings.Contains(out, "no .NET runtime") {
		t.Fatalf("a platform .NET has no runtime for is refused:\n%s", out)
	}
}
