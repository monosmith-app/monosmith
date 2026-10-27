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

// fakeDotnet makes dotnet publish write a single-file program, its symbols
// and a settings file into the -o directory, and returns the commands seen.
func fakeDotnet(t *testing.T) *[]string {
	t.Helper()
	real := DotnetRunner
	t.Cleanup(func() { DotnetRunner = real })
	var seen []string
	DotnetRunner = func(dir string, argv []string) error {
		seen = append(seen, strings.Join(argv, " "))
		out := argv[slices.Index(argv, "-o")+1]
		program := "Acme.Cli"
		if slices.Contains(argv, "win-x64") || slices.Contains(argv, "win-arm64") {
			program += ".exe"
		}
		_ = os.MkdirAll(out, 0o755)
		_ = os.WriteFile(filepath.Join(out, program), []byte("program"), 0o755)
		_ = os.WriteFile(filepath.Join(out, "Acme.Cli.pdb"), []byte("symbols"), 0o644)
		_ = os.WriteFile(filepath.Join(out, "appsettings.json"), []byte("{}"), 0o644)
		return nil
	}
	return &seen
}

func dotnetApp(t *testing.T, workspace string) protocol.Project {
	t.Helper()
	dir := filepath.Join(workspace, "apps", "cli")
	_ = os.MkdirAll(dir, 0o755)
	_ = os.WriteFile(filepath.Join(dir, "Cli.csproj"), []byte("<Project Sdk=\"Microsoft.NET.Sdk\"><PropertyGroup><OutputType>Exe</OutputType><AssemblyName>Acme.Cli</AssemblyName></PropertyGroup></Project>"), 0o644)
	_ = os.WriteFile(filepath.Join(workspace, "LICENSE"), []byte("MIT"), 0o644)
	return protocol.Project{Name: "cli", Root: "apps/cli", Language: "dotnet", Kind: protocol.ProjectApp}
}

func spec(platforms ...string) *protocol.AssetSpec {
	return &protocol.AssetSpec{Platforms: platforms, OutDir: "dist/release/cli", Product: "acme"}
}

func TestBuildAssetsPublishesSelfContainedForEveryPlatformFromOneMachine(t *testing.T) {
	ws := t.TempDir()
	seen := fakeDotnet(t)
	res, err := BuildAssets(ws, protocol.PublishParams{Project: dotnetApp(t, ws), Version: "1.2.0", Assets: spec("windows/amd64", "linux/arm64", "darwin/arm64")})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"dist/release/cli/acme_1.2.0_windows_amd64.zip", "dist/release/cli/acme_1.2.0_linux_arm64.zip",
		"dist/release/cli/acme_1.2.0_darwin_arm64.zip", "dist/release/cli/checksums.txt",
	}
	if !slices.Equal(res.Assets, want) || res.Binary != "Acme.Cli" || res.AssetPlatforms[want[1]] != "linux/arm64" {
		t.Fatalf("%+v", res)
	}
	if !strings.Contains((*seen)[1], "publish Cli.csproj --nologo -c Release -r linux-arm64 --self-contained true -p:PublishSingleFile=true") || !strings.Contains((*seen)[1], "-p:Version=1.2.0") {
		t.Errorf("dotnet: %v", *seen)
	}
	zr, err := zip.OpenReader(filepath.Join(ws, filepath.FromSlash(want[0])))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = zr.Close() }()
	var names []string
	for _, f := range zr.File {
		names = append(names, f.Name)
		if f.Name == "Acme.Cli.exe" && f.Mode()&0o111 == 0 {
			t.Error("the program is executable")
		}
	}
	slices.Sort(names)
	if !slices.Equal(names, []string{"Acme.Cli.exe", "LICENSE", "appsettings.json"}) {
		t.Errorf("symbols are left out, settings and the licence kept: %v", names)
	}
	if _, err := os.Stat(filepath.Join(ws, "dist", "release", "cli", "win-x64")); err == nil {
		t.Error("the staging folder is removed")
	}
	sums, _ := os.ReadFile(filepath.Join(ws, "dist", "release", "cli", "checksums.txt"))
	if strings.Count(string(sums), "\n") != 3 {
		t.Errorf("checksums:\n%s", sums)
	}
}

func TestBuildAssetsDryRunBuildsNothing(t *testing.T) {
	ws := t.TempDir()
	seen := fakeDotnet(t)
	res, err := BuildAssets(ws, protocol.PublishParams{Project: dotnetApp(t, ws), Version: "1.2.0", DryRun: true, Assets: spec("linux/amd64")})
	if err != nil || res.Published || len(*seen) != 0 || len(res.Assets) != 2 {
		t.Fatalf("%+v %v %v", res, err, *seen)
	}
}

func TestBuildAssetsRejectsAPlatformWithoutARuntime(t *testing.T) {
	ws := t.TempDir()
	fakeDotnet(t)
	for _, platform := range []string{"linux/mips", "plan9/amd64", "linux"} {
		if _, err := BuildAssets(ws, protocol.PublishParams{Project: dotnetApp(t, ws), Version: "1.0.0", DryRun: true, Assets: spec(platform)}); err == nil {
			t.Errorf("%s was accepted", platform)
		}
	}
}

func TestBuildAssetsExplainsAProjectThatIsNotAnApp(t *testing.T) {
	ws := t.TempDir()
	fakeDotnet(t)
	DotnetRunner = func(dir string, argv []string) error {
		_ = os.MkdirAll(argv[slices.Index(argv, "-o")+1], 0o755)
		return nil
	}
	_, err := BuildAssets(ws, protocol.PublishParams{Project: dotnetApp(t, ws), Version: "1.0.0", Assets: spec("linux/amd64")})
	if err == nil || !strings.Contains(err.Error(), "OutputType Exe") {
		t.Fatalf("%v", err)
	}
}

func TestDoctorWarnsOfARuntimeNotYetDownloaded(t *testing.T) {
	real := RuntimePackCached
	defer func() { RuntimePackCached = real }()
	RuntimePackCached = func(rid string) bool { return rid == "linux-x64" }
	p := protocol.Project{Name: "cli", Root: "apps/cli", ReleaseAssets: spec("linux/amd64", "windows/arm64", "beos/ppc")}
	got := CheckReleaseRuntimes(p)
	if len(got) != 3 || !got[0].OK || got[1].OK || !got[1].Warning || !strings.Contains(got[1].Check, "win-arm64") || got[2].OK || got[2].Warning {
		t.Fatalf("%+v", got)
	}
	if CheckReleaseRuntimes(protocol.Project{Name: "x"}) != nil {
		t.Error("an app without a release has nothing to check")
	}
}
