//go:build e2e

package goldenpath

import (
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestReleaseDryRunPlansRustArchives dry-runs the release of a workspace with
// two Rust apps, each with its own archives: both are listed as separate
// assets with a checksums file, and a platform this machine cannot build is
// refused with the way to build it. Nothing is built or published.
func TestReleaseDryRunPlansRustArchives(t *testing.T) {
	if _, err := exec.LookPath("cargo"); err != nil {
		t.Skip("cargo is not installed")
	}
	base := t.TempDir()
	w := &workspace{t: t, bin: buildSmith(t, base), dir: base, env: identity()}
	w.smith(true, "new", "smith-e2e-archives", "--ci", "github")
	w.dir = filepath.Join(base, "smith-e2e-archives")
	w.smith(true, "add", "app", "cli", "--lang", "rust")
	w.smith(true, "add", "app", "vcw", "--lang", "rust")
	host := runtime.GOOS + "/" + runtime.GOARCH
	// The other operating system's platform: one this machine cannot build.
	other := "linux/amd64"
	if runtime.GOOS == "linux" {
		other = "windows/amd64"
	}
	w.write("smith.toml", w.read("smith.toml")+
		"\n[projects.cli.release]\nplatforms = [\""+host+"\"]\nasset_name = \"{product}-{version}-{os}-{arch}.{ext}\"\n"+
		"\n[projects.vcw.release]\nplatforms = [\""+host+"\"]\nproduct = \"videocraw-vcw\"\n")
	w.git("add", "-A")
	w.git("commit", "-qm", "feat: two apps")

	out := w.smith(true, "release", "--dry-run")
	for _, want := range []string{
		"dist/release/cli/cli-0.1.0-",
		"dist/release/vcw/videocraw-vcw_0.1.0_",
		"dist/release/cli/checksums.txt",
		"dist/release/vcw/checksums.txt",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("the dry run does not plan %q:\n%s", want, out)
		}
	}

	// A native app lists what this machine's native job would attach: its zip
	// and the zip's own checksum, no combined checksums file.
	w.write("apps/vcw/Cargo.toml", strings.Replace(w.read("apps/vcw/Cargo.toml"), "[package.metadata.smith]\n", "[package.metadata.smith]\nnative = true\n", 1))
	out = w.smith(true, "release", "--dry-run")
	if !strings.Contains(out, "videocraw-vcw_0.1.0_") || !strings.Contains(out, ".zip.sha256") || strings.Contains(out, "dist/release/vcw/checksums.txt") {
		t.Fatalf("a native app attaches its own zip and checksum:\n%s", out)
	}

	w.write("smith.toml", strings.Replace(w.read("smith.toml"), "[projects.cli.release]\nplatforms = [\""+host+"\"]", "[projects.cli.release]\nplatforms = [\""+host+"\", \""+other+"\"]", 1))
	out = w.smith(false, "release", "--dry-run")
	if !strings.Contains(out, other) || !strings.Contains(out, "native = true") {
		t.Fatalf("a platform this machine cannot build is refused with the way out:\n%s", out)
	}
}
