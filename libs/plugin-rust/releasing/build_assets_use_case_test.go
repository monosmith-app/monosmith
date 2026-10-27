package releasing

import (
	"archive/zip"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// fakeToolchain makes cargo "build" a program per call and rustup list the
// given targets, and returns the cargo commands it saw.
func fakeToolchain(t *testing.T, host string, installed ...string) *[]string {
	t.Helper()
	realHost, realBuild, realInstalled, realLinker := hostPlatform, CargoBuild, InstalledTargets, LinkerAvailable
	t.Cleanup(func() { hostPlatform, CargoBuild, InstalledTargets, LinkerAvailable = realHost, realBuild, realInstalled, realLinker })
	hostPlatform = host
	InstalledTargets = func(string) ([]string, bool) { return installed, true }
	LinkerAvailable = func(string, string, string) bool { return false }
	var seen []string
	CargoBuild = func(dir string, argv []string, env []string) (string, error) {
		seen = append(seen, strings.Join(argv, " ")+" "+strings.Join(env, " "))
		triple := argv[slices.Index(argv, "--target")+1]
		crate := argv[slices.Index(argv, "-p")+1]
		name := crate
		if strings.Contains(triple, "windows") {
			name += ".exe"
		}
		exe := filepath.Join(dir, "dist", "cargo", triple, "release", name)
		_ = os.MkdirAll(filepath.Dir(exe), 0o755)
		_ = os.WriteFile(exe, []byte("program for "+triple), 0o755)
		message, _ := json.Marshal(map[string]any{"reason": "compiler-artifact", "executable": exe})
		other, _ := json.Marshal(map[string]any{"reason": "compiler-artifact", "executable": nil})
		return string(other) + "\n" + string(message) + "\n" + `{"reason":"build-finished","success":true}` + "\n", nil
	}
	return &seen
}

func rustApp(workspace string, tags ...string) protocol.Project {
	_ = os.MkdirAll(filepath.Join(workspace, "apps", "cli"), 0o755)
	_ = os.WriteFile(filepath.Join(workspace, "apps", "cli", "README.md"), []byte("the cli"), 0o644)
	_ = os.WriteFile(filepath.Join(workspace, "LICENSE"), []byte("MIT"), 0o644)
	_ = os.WriteFile(filepath.Join(workspace, "README.md"), []byte("the workspace"), 0o644)
	return protocol.Project{Name: "cli", Root: "apps/cli", Language: "rust", Kind: protocol.ProjectApp, Exports: []string{"acme-cli"}, Tags: tags}
}

func spec(platforms ...string) *protocol.AssetSpec {
	return &protocol.AssetSpec{Platforms: platforms, OutDir: "dist/release/cli", Product: "acme"}
}

func names(r *zip.ReadCloser) []string {
	var out []string
	for _, f := range r.File {
		out = append(out, f.Name)
	}
	return out
}

func TestBuildAssetsMakesAnArchivePerPlatformWithChecksums(t *testing.T) {
	ws := t.TempDir()
	seen := fakeToolchain(t, "windows/amd64", "x86_64-pc-windows-msvc", "aarch64-pc-windows-msvc")
	res, err := BuildAssets(ws, protocol.PublishParams{Project: rustApp(ws), Version: "1.2.0", Assets: spec("windows/amd64", "windows/arm64")})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"dist/release/cli/acme_1.2.0_windows_amd64.zip", "dist/release/cli/acme_1.2.0_windows_arm64.zip", "dist/release/cli/checksums.txt"}
	if !slices.Equal(res.Assets, want) || !res.Published || res.Binary != "acme-cli" {
		t.Fatalf("%+v", res)
	}
	if res.AssetPlatforms[want[1]] != "windows/arm64" || len(res.AssetPlatforms) != 2 {
		t.Errorf("platforms: %v", res.AssetPlatforms)
	}
	if !strings.Contains((*seen)[1], "cargo build --release -p acme-cli --target aarch64-pc-windows-msvc") {
		t.Errorf("cargo: %v", *seen)
	}
	zr, err := zip.OpenReader(filepath.Join(ws, filepath.FromSlash(want[0])))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = zr.Close() }()
	if got := names(zr); !slices.Equal(got, []string{"LICENSE", "README.md", "acme-cli.exe"}) && !slices.Equal(got, []string{"acme-cli.exe", "LICENSE", "README.md"}) {
		t.Errorf("archive holds %v", got)
	}
	sums, _ := os.ReadFile(filepath.Join(ws, "dist", "release", "cli", "checksums.txt"))
	if strings.Count(string(sums), "\n") != 2 || !strings.Contains(string(sums), "  acme_1.2.0_windows_amd64.zip\n") {
		t.Errorf("checksums:\n%s", sums)
	}
	if readme := zipText(t, zr, "README.md"); readme != "the cli" {
		t.Errorf("the crate's README wins over the workspace's: %q", readme)
	}
}

func zipText(t *testing.T, zr *zip.ReadCloser, name string) string {
	t.Helper()
	for _, f := range zr.File {
		if f.Name == name {
			r, _ := f.Open()
			defer func() { _ = r.Close() }()
			b := make([]byte, 100)
			n, _ := r.Read(b)
			return string(b[:n])
		}
	}
	return ""
}

func TestBuildAssetsDryRunChecksButBuildsNothing(t *testing.T) {
	ws := t.TempDir()
	seen := fakeToolchain(t, "linux/amd64", "x86_64-unknown-linux-gnu")
	res, err := BuildAssets(ws, protocol.PublishParams{Project: rustApp(ws), Version: "1.2.0", DryRun: true, Assets: spec("linux/amd64")})
	if err != nil || res.Published || len(*seen) != 0 || len(res.Assets) != 2 {
		t.Fatalf("%+v %v %v", res, err, *seen)
	}
	if _, err := os.Stat(filepath.Join(ws, "dist")); err == nil {
		t.Error("a dry run writes nothing")
	}
}

func TestBuildAssetsRefusesAnotherSystemAndSaysHowToBuildIt(t *testing.T) {
	ws := t.TempDir()
	fakeToolchain(t, "windows/amd64", "x86_64-pc-windows-msvc", "x86_64-apple-darwin")
	_, err := BuildAssets(ws, protocol.PublishParams{Project: rustApp(ws), Version: "1.0.0", DryRun: true, Assets: spec("windows/amd64", "darwin/amd64")})
	if err == nil || !strings.Contains(err.Error(), "darwin/amd64") || !strings.Contains(err.Error(), "native = true") {
		t.Fatalf("%v", err)
	}
}

func TestBuildAssetsNeedsALinkerForAnotherLinuxProcessor(t *testing.T) {
	ws := t.TempDir()
	seen := fakeToolchain(t, "linux/amd64", "x86_64-unknown-linux-gnu", "aarch64-unknown-linux-gnu")
	params := protocol.PublishParams{Project: rustApp(ws), Version: "1.0.0", Assets: spec("linux/arm64")}
	if _, err := BuildAssets(ws, params); err == nil || !strings.Contains(err.Error(), "aarch64-linux-gnu-gcc") || !strings.Contains(err.Error(), "CARGO_TARGET_AARCH64_UNKNOWN_LINUX_GNU_LINKER") {
		t.Fatalf("no linker: %v", err)
	}
	LinkerAvailable = func(string, string, string) bool { return true }
	if _, err := BuildAssets(ws, params); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains((*seen)[0], "CARGO_TARGET_AARCH64_UNKNOWN_LINUX_GNU_LINKER=aarch64-linux-gnu-gcc") {
		t.Errorf("the linker is handed to cargo: %v", *seen)
	}
}

func TestBuildAssetsNamesAMissingRustTarget(t *testing.T) {
	ws := t.TempDir()
	fakeToolchain(t, "darwin/arm64", "aarch64-apple-darwin")
	_, err := BuildAssets(ws, protocol.PublishParams{Project: rustApp(ws), Version: "1.0.0", DryRun: true, Assets: spec("darwin/arm64", "darwin/amd64")})
	if err == nil || !strings.Contains(err.Error(), "rustup target add x86_64-apple-darwin") {
		t.Fatalf("%v", err)
	}
}

func TestBuildAssetsRejectsAPlatformRustHasNoTargetFor(t *testing.T) {
	ws := t.TempDir()
	fakeToolchain(t, "linux/amd64")
	for _, platform := range []string{"plan9/amd64", "linux", "linux/mips"} {
		if _, err := BuildAssets(ws, protocol.PublishParams{Project: rustApp(ws), Version: "1.0.0", DryRun: true, Assets: spec(platform)}); err == nil {
			t.Errorf("%s was accepted", platform)
		}
	}
}

func TestBuildAssetsOfANativeAppChecksumsEachArchiveOnItsOwn(t *testing.T) {
	ws := t.TempDir()
	fakeToolchain(t, "linux/amd64", "x86_64-unknown-linux-gnu")
	res, err := BuildAssets(ws, protocol.PublishParams{Project: rustApp(ws, "native"), Version: "1.0.0", Assets: spec("linux/amd64")})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"dist/release/cli/acme_1.0.0_linux_amd64.zip", "dist/release/cli/acme_1.0.0_linux_amd64.zip.sha256"}
	if !slices.Equal(res.Assets, want) {
		t.Fatalf("%v", res.Assets)
	}
	sum, _ := os.ReadFile(filepath.Join(ws, filepath.FromSlash(want[1])))
	if !strings.HasSuffix(string(sum), "  acme_1.0.0_linux_amd64.zip\n") {
		t.Errorf("%q", sum)
	}
}

func TestBuildAssetsFailsWhenTheCrateHasNoProgram(t *testing.T) {
	ws := t.TempDir()
	fakeToolchain(t, "linux/amd64", "x86_64-unknown-linux-gnu")
	CargoBuild = func(string, []string, []string) (string, error) { return `{"reason":"build-finished"}` + "\n", nil }
	if _, err := BuildAssets(ws, protocol.PublishParams{Project: rustApp(ws), Version: "1.0.0", Assets: spec("linux/amd64")}); err == nil || !strings.Contains(err.Error(), "no binary target") {
		t.Fatalf("%v", err)
	}
}

func TestAssetNamePatternAndAliasesApply(t *testing.T) {
	ws := t.TempDir()
	fakeToolchain(t, "darwin/arm64", "aarch64-apple-darwin")
	s := spec("darwin/arm64")
	s.Name, s.OSAlias, s.ArchAlias = "{product}-{version}-{os}-{arch}.{ext}", map[string]string{"darwin": "macos"}, map[string]string{"arm64": "aarch64"}
	res, err := BuildAssets(ws, protocol.PublishParams{Project: rustApp(ws), Version: "2.0.0", DryRun: true, Assets: s})
	if err != nil || res.Assets[0] != "dist/release/cli/acme-2.0.0-macos-aarch64.zip" {
		t.Fatalf("%+v %v", res, err)
	}
}
