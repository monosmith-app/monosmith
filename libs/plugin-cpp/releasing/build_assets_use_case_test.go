package releasing

import (
	"archive/zip"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// fakeCMake stands in for cmake: the build writes the program into the build
// tree, the install writes what the project's install rules would (when
// installs is set) into the prefix. It returns the commands seen.
func fakeCMake(t *testing.T, host string, installs bool) *[]string {
	t.Helper()
	realHost, realRunner := hostPlatform, CMakeRunner
	t.Cleanup(func() { hostPlatform, CMakeRunner = realHost, realRunner })
	hostPlatform = host
	var seen []string
	CMakeRunner = func(dir string, argv []string) error {
		seen = append(seen, strings.Join(argv[1:], " "))
		switch {
		case slices.Contains(argv, "--build"):
			_ = os.MkdirAll(argv[2], 0o755)
			_ = os.WriteFile(filepath.Join(argv[2], "acme_tool"), []byte("program"), 0o755)
		case slices.Contains(argv, "--install") && installs:
			prefix := argv[slices.Index(argv, "--prefix")+1]
			_ = os.MkdirAll(filepath.Join(prefix, "bin"), 0o755)
			_ = os.WriteFile(filepath.Join(prefix, "bin", "acme_tool"), []byte("program"), 0o755)
			_ = os.WriteFile(filepath.Join(prefix, "bin", "plugin.dat"), []byte("data"), 0o644)
		}
		return nil
	}
	return &seen
}

func cppApp(t *testing.T, workspace string, tags ...string) protocol.Project {
	t.Helper()
	dir := filepath.Join(workspace, "apps", "tool")
	_ = os.MkdirAll(dir, 0o755)
	_ = os.WriteFile(filepath.Join(dir, "CMakeLists.txt"), []byte("cmake_minimum_required(VERSION 3.25)\nproject(acme_tool VERSION 1.0.0)\nadd_executable(acme_tool main.cpp)\n"), 0o644)
	_ = os.WriteFile(filepath.Join(workspace, "LICENSE"), []byte("MIT"), 0o644)
	return protocol.Project{Name: "tool", Root: "apps/tool", Language: "cpp", Kind: protocol.ProjectApp, Tags: tags}
}

func spec(platforms ...string) *protocol.AssetSpec {
	return &protocol.AssetSpec{Platforms: platforms, OutDir: "dist/release/tool", Product: "acme"}
}

func zipNames(t *testing.T, file string) []string {
	t.Helper()
	zr, err := zip.OpenReader(file)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = zr.Close() }()
	var names []string
	for _, f := range zr.File {
		names = append(names, f.Name)
	}
	slices.Sort(names)
	return names
}

func TestBuildAssetsZipsTheProgramWhenThereAreNoInstallRules(t *testing.T) {
	ws := t.TempDir()
	seen := fakeCMake(t, "linux/amd64", false)
	res, err := BuildAssets(ws, protocol.PublishParams{Project: cppApp(t, ws), Version: "1.0.0", Assets: spec("linux/amd64")})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"dist/release/tool/acme_1.0.0_linux_amd64.zip", "dist/release/tool/checksums.txt"}
	if !slices.Equal(res.Assets, want) || !res.Published || res.Binary != "acme_tool" || res.AssetPlatforms[want[0]] != "linux/amd64" {
		t.Fatalf("%+v", res)
	}
	if len(*seen) != 3 || !strings.Contains((*seen)[0], "-G Ninja -DCMAKE_BUILD_TYPE=Release") || !strings.HasPrefix((*seen)[2], "--install ") {
		t.Errorf("cmake: %v", *seen)
	}
	if got := zipNames(t, filepath.Join(ws, filepath.FromSlash(want[0]))); !slices.Equal(got, []string{"LICENSE", "acme_tool"}) {
		t.Errorf("archive holds %v", got)
	}
	sums, _ := os.ReadFile(filepath.Join(ws, "dist", "release", "tool", "checksums.txt"))
	if !strings.HasSuffix(string(sums), "  acme_1.0.0_linux_amd64.zip\n") {
		t.Errorf("checksums: %q", sums)
	}
	if _, err := os.Stat(filepath.Join(ws, "dist", "release", "tool", "build")); err == nil {
		t.Error("the build tree is removed")
	}
}

func TestBuildAssetsZipsWhatTheInstallRulesInstall(t *testing.T) {
	ws := t.TempDir()
	fakeCMake(t, "darwin/arm64", true)
	res, err := BuildAssets(ws, protocol.PublishParams{Project: cppApp(t, ws), Version: "1.0.0", Assets: spec("darwin/arm64")})
	if err != nil {
		t.Fatal(err)
	}
	if got := zipNames(t, filepath.Join(ws, filepath.FromSlash(res.Assets[0]))); !slices.Equal(got, []string{"LICENSE", "bin/acme_tool", "bin/plugin.dat"}) {
		t.Errorf("archive holds %v", got)
	}
}

func TestBuildAssetsOfANativeAppChecksumsItsOwnArchive(t *testing.T) {
	ws := t.TempDir()
	fakeCMake(t, "windows/amd64", false)
	_ = os.WriteFile(filepath.Join(ws, "never"), nil, 0o644)
	CMakeRunner = func(dir string, argv []string) error {
		if slices.Contains(argv, "--build") {
			_ = os.MkdirAll(argv[2], 0o755)
			_ = os.WriteFile(filepath.Join(argv[2], "acme_tool.exe"), []byte("program"), 0o755)
		}
		return nil
	}
	res, err := BuildAssets(ws, protocol.PublishParams{Project: cppApp(t, ws, "native"), Version: "1.0.0", Assets: spec("windows/amd64")})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"dist/release/tool/acme_1.0.0_windows_amd64.zip", "dist/release/tool/acme_1.0.0_windows_amd64.zip.sha256"}
	if !slices.Equal(res.Assets, want) {
		t.Fatalf("%v", res.Assets)
	}
	if got := zipNames(t, filepath.Join(ws, filepath.FromSlash(want[0]))); !slices.Equal(got, []string{"LICENSE", "acme_tool.exe"}) {
		t.Errorf("archive holds %v", got)
	}
}

func TestBuildAssetsDryRunChecksButBuildsNothing(t *testing.T) {
	ws := t.TempDir()
	seen := fakeCMake(t, "linux/amd64", false)
	res, err := BuildAssets(ws, protocol.PublishParams{Project: cppApp(t, ws), Version: "1.0.0", DryRun: true, Assets: spec("linux/amd64")})
	if err != nil || res.Published || len(*seen) != 0 || len(res.Assets) != 2 {
		t.Fatalf("%+v %v %v", res, err, *seen)
	}
	if _, err := os.Stat(filepath.Join(ws, "dist")); err == nil {
		t.Error("a dry run writes nothing")
	}
}

func TestBuildAssetsRefusesAnotherSystemAndSaysHowToBuildIt(t *testing.T) {
	ws := t.TempDir()
	fakeCMake(t, "linux/amd64", false)
	_, err := BuildAssets(ws, protocol.PublishParams{Project: cppApp(t, ws), Version: "1.0.0", DryRun: true, Assets: spec("linux/amd64", "darwin/arm64")})
	if err == nil || !strings.Contains(err.Error(), "darwin/arm64") || !strings.Contains(err.Error(), "SMITH_NATIVE") {
		t.Fatalf("%v", err)
	}
	for _, platform := range []string{"plan9/amd64", "linux", "linux/mips"} {
		if _, err := BuildAssets(ws, protocol.PublishParams{Project: cppApp(t, ws), Version: "1.0.0", DryRun: true, Assets: spec(platform)}); err == nil {
			t.Errorf("%s was accepted", platform)
		}
	}
}

func TestBuildAssetsExplainsABuildThatMakesNoProgram(t *testing.T) {
	ws := t.TempDir()
	fakeCMake(t, "linux/amd64", false)
	CMakeRunner = func(string, []string) error { return nil }
	_, err := BuildAssets(ws, protocol.PublishParams{Project: cppApp(t, ws), Version: "1.0.0", Assets: spec("linux/amd64")})
	if err == nil || !strings.Contains(err.Error(), "install(TARGETS acme_tool)") {
		t.Fatalf("%v", err)
	}
}

func TestDoctorChecksThePlatformsAndTheProgram(t *testing.T) {
	ws := t.TempDir()
	fakeCMake(t, "linux/amd64", false)
	p := cppApp(t, ws)
	p.ReleaseAssets = spec("linux/amd64", "windows/arm64", "beos/ppc")
	var failed []string
	for _, f := range CheckReleasePlatforms(ws, p) {
		if !f.OK {
			failed = append(failed, f.Check)
		}
	}
	if len(failed) != 2 || !strings.Contains(failed[0], "windows/arm64") || !strings.Contains(failed[1], "beos/ppc") {
		t.Fatalf("%v", failed)
	}
	p.Tags = []string{"native"}
	if got := CheckReleasePlatforms(ws, p); len(got) != 2 || got[0].OK || !strings.Contains(got[0].Check, "beos") || !got[1].OK {
		t.Errorf("a native app is checked for names and its program only: %+v", got)
	}
	_ = os.WriteFile(filepath.Join(ws, "apps", "tool", "CMakeLists.txt"), []byte("project(acme_tool)\n"), 0o644)
	found := false
	for _, f := range CheckReleasePlatforms(ws, p) {
		found = found || (strings.Contains(f.Check, "has a program to release") && !f.OK)
	}
	if !found {
		t.Error("a project with no program and no install rule fails")
	}
	if CheckReleasePlatforms(ws, protocol.Project{Name: "x"}) != nil {
		t.Error("an app without a release has nothing to check")
	}
}
