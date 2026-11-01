package workspace

import (
	"slices"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// parseAndResolve parses smith.toml and resolves the release of project app
// without a plugin's defaults.
func parseAndResolve(input string) (*ProjectRelease, error) {
	cfg, err := Parse([]byte(input))
	if err != nil {
		return nil, err
	}
	return ResolveRelease("app", cfg.Projects["app"], nil, cfg.Release.GitHubReleases)
}

var tauriDefaults = &protocol.ReleaseDefaults{
	Build:     "smith run tauri-build -p app",
	Assets:    []string{"dist/tauri/apps/app/release/bundle/**/*.exe"},
	Platforms: []string{"windows-x64", "linux-x64"},
	AssetName: "{product}-{version}-{platform}.{ext}",
	Checksums: true,
}

func resolveWithDefaults(t *testing.T, table string) (*ProjectRelease, error) {
	t.Helper()
	cfg, err := Parse([]byte(releaseHead + table))
	if err != nil {
		t.Fatal(err)
	}
	return ResolveRelease("app", cfg.Projects["app"], tauriDefaults, true)
}

func TestResolveReleaseFillsAnEmptyTableFromThePluginsDefaults(t *testing.T) {
	r, err := resolveWithDefaults(t, "")
	if err != nil {
		t.Fatal(err)
	}
	if r.Build != tauriDefaults.Build || !slices.Equal(r.Assets, tauriDefaults.Assets) || !slices.Equal(r.Platforms, tauriDefaults.Platforms) ||
		r.AssetName != tauriDefaults.AssetName || r.Checksums == nil || !*r.Checksums {
		t.Fatalf("%+v", r)
	}
}

func TestResolveReleaseKeepsWhatTheTableSets(t *testing.T) {
	r, err := resolveWithDefaults(t, "build = \"make installers\"\nassets = [\"out/*.msi\"]\nplatforms = [\"windows-x64\"]\nasset_name = \"{product}.{ext}\"\nchecksums = false\n")
	if err != nil {
		t.Fatal(err)
	}
	if r.Build != "make installers" || r.Assets[0] != "out/*.msi" || len(r.Platforms) != 1 || r.AssetName != "{product}.{ext}" || *r.Checksums {
		t.Fatalf("%+v", r)
	}
	r, err = resolveWithDefaults(t, "platforms = [\"macos-arm64\"]\n")
	if err != nil || r.Build != tauriDefaults.Build || !slices.Equal(r.Platforms, []string{"macos-arm64"}) {
		t.Fatalf("%+v %v", r, err)
	}
}

func TestResolveReleaseLeavesATableOfArchivesToThePlugin(t *testing.T) {
	r, err := resolveWithDefaults(t, "os_alias = { windows = \"win\" }\n")
	if err != nil || r.Build != "" || len(r.Platforms) != len(DefaultPlatforms) {
		t.Fatalf("%+v %v", r, err)
	}
}

func TestResolveReleaseNeedsGitHubReleasesForTheDefaults(t *testing.T) {
	cfg, err := Parse([]byte("[workspace]\nname = \"a\"\n[projects.app.release]\n"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = ResolveRelease("app", cfg.Projects["app"], tauriDefaults, false)
	if err == nil || !strings.Contains(err.Error(), "github_releases = true") {
		t.Fatalf("got %v", err)
	}
}

func TestResolveReleaseKeepsInstallAndRejectsItWithoutABuild(t *testing.T) {
	r, err := parseAndResolve(releaseHead + "build = \"zip -r out.zip app\"\nassets = [\"out.zip\"]\nplatforms = [\"linux-x64\"]\ninstall = false\n")
	if err != nil || r.Install == nil || *r.Install {
		t.Fatalf("install = false is kept: %+v %v", r, err)
	}
	if _, err := parseAndResolve(releaseHead + "install = false\n"); err == nil || !strings.Contains(err.Error(), "install") {
		t.Fatalf("install without build: %v", err)
	}
}

func TestResolveReleaseChecksAssetNamePlaceholders(t *testing.T) {
	_, err := parseAndResolve(releaseHead + "build = \"make\"\nassets = [\"a\"]\nplatforms = [\"linux-x64\"]\nasset_name = \"{os}.{ext}\"\n")
	if err == nil || !strings.Contains(err.Error(), "{os}") {
		t.Fatalf("got %v", err)
	}
	if _, err := parseAndResolve(releaseHead + "checksums = true\n"); err == nil || !strings.Contains(err.Error(), "checksums") {
		t.Fatalf("checksums without build: %v", err)
	}
}

// A release table without platforms builds six: the docs name them. A native app
// is the exception (one archive per native runner, not these).
func TestResolveReleaseDefaultsToTheSixPlatformsTheDocsName(t *testing.T) {
	want := []string{"linux/amd64", "linux/arm64", "darwin/amd64", "darwin/arm64", "windows/amd64", "windows/arm64"}
	r, err := resolveWithDefaults(t, "os_alias = { windows = \"win\" }\n")
	if err != nil || !slices.Equal(r.Platforms, want) {
		t.Fatalf("%+v %v", r, err)
	}
}
