package release

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

// smithSample is the release of smith 1.129.0 as the report of the real run
// held it: eight npm packages, six wheels, seven NuGet packages, eight gems, 17
// files on the GitHub Release and a queued pub.dev publish.
func smithSample() ReportResult {
	return ReportResult{
		Item:   Item{Project: "smith", Current: "1.128.0", Next: "1.129.0", Tag: "smith@1.129.0", Level: "minor"},
		Status: "released", Pushed: true, GitHubRelease: "https://github.com/monosmith-app/monosmith/releases/tag/smith%401.129.0",
		Packages: []string{
			"@monosmith/cli-darwin-x64@1.129.0", "@monosmith/cli-darwin-arm64@1.129.0", "@monosmith/cli-linux-x64@1.129.0",
			"@monosmith/cli-linux-arm64@1.129.0", "@monosmith/cli-win32-x64@1.129.0", "@monosmith/cli-win32-arm64@1.129.0",
			"@monosmith/cli@1.129.0", "monosmith@1.129.0",
			"monosmith-1.129.0-py3-none-macosx_12_0_x86_64.whl", "monosmith-1.129.0-py3-none-macosx_12_0_arm64.whl",
			"monosmith-1.129.0-py3-none-manylinux_2_17_x86_64.manylinux2014_x86_64.musllinux_1_1_x86_64.whl",
			"monosmith-1.129.0-py3-none-manylinux_2_17_aarch64.manylinux2014_aarch64.musllinux_1_1_aarch64.whl",
			"monosmith-1.129.0-py3-none-win_amd64.whl", "monosmith-1.129.0-py3-none-win_arm64.whl",
			"monosmith.osx-x64.1.129.0.nupkg", "monosmith.osx-arm64.1.129.0.nupkg", "monosmith.linux-x64.1.129.0.nupkg",
			"monosmith.linux-arm64.1.129.0.nupkg", "monosmith.win-x64.1.129.0.nupkg", "monosmith.win-arm64.1.129.0.nupkg", "monosmith.1.129.0.nupkg",
			"monosmith-1.129.0-x86_64-darwin.gem", "monosmith-1.129.0-arm64-darwin.gem", "monosmith-1.129.0-x86_64-linux.gem",
			"monosmith-1.129.0-x86_64-linux-musl.gem", "monosmith-1.129.0-aarch64-linux.gem", "monosmith-1.129.0-aarch64-linux-musl.gem",
			"monosmith-1.129.0-x64-mingw-ucrt.gem", "monosmith-1.129.0-aarch64-mingw-ucrt.gem",
		},
		Attached: []string{
			"monosmith_1.129.0_linux_amd64.zip", "monosmith_1.129.0_linux_arm64.zip", "monosmith_1.129.0_darwin_amd64.zip",
			"monosmith_1.129.0_darwin_arm64.zip", "monosmith_1.129.0_windows_amd64.zip", "monosmith_1.129.0_windows_arm64.zip",
			"checksums.txt", "monosmith_1.129.0_amd64.deb", "monosmith-1.129.0-1.x86_64.rpm", "monosmith_1.129.0_arm64.deb",
			"monosmith-1.129.0-1.aarch64.rpm", "monosmith-app.monosmith.installer.yaml", "monosmith-app.monosmith.locale.en-US.yaml",
			"monosmith-app.monosmith.yaml", "monosmith.json", "monosmith.rb", "monosmith_1.129.0_sbom.cdx.json",
		},
		PubDispatch: "monosmith 1.129.0 from smith-v1.129.0",
	}
}

func TestTargetsOfTheRealSampleNameEveryTarget(t *testing.T) {
	got := TargetsOf(smithSample())
	want := []struct{ kind, what, where string }{
		{"npm", "@monosmith/cli + monosmith + 6 platform packages", "registry.npmjs.org"},
		{"PyPI", "6 wheels (macOS, Linux, Windows x x64, arm64)", "pypi.org"},
		{"NuGet", "monosmith + 6 platform packages", "nuget.org"},
		{"RubyGems", "8 gems (macOS, Linux, Windows x x64, arm64)", "rubygems.org"},
		{"GitHub Release", "release smith@1.129.0 with 17 files: 6 zips, 4 deb/rpm, checksums, SBOM, winget, brew, manifest", "https://github.com/monosmith-app/monosmith/releases/tag/smith%401.129.0"},
		{"pub.dev", "publish queued (pub.yml), not finished yet", ""},
	}
	if len(got) != len(want) {
		t.Fatalf("%d targets: %+v", len(got), got)
	}
	for i, w := range want {
		if got[i].Kind != w.kind || got[i].What != w.what || got[i].Where != w.where {
			t.Errorf("target %d: %+v, want %+v", i, got[i], w)
		}
	}
	if got[0].detail != "darwin x64 arm64 · linux x64 arm64 · win32 x64 arm64" || got[0].Count != 8 {
		t.Errorf("npm: %q, count %d", got[0].detail, got[0].Count)
	}
	if got[2].detail != "darwin x64 arm64 · linux x64 arm64 · windows x64 arm64" {
		t.Errorf("NuGet: %q", got[2].detail)
	}
}

func TestPlatformsAreReadFromFileNames(t *testing.T) {
	for name, want := range map[string]string{
		"monosmith-1.129.0-py3-none-macosx_12_0_arm64.whl":                          "darwin/arm64",
		"monosmith-1.129.0-py3-none-manylinux_2_17_x86_64.musllinux_1_1_x86_64.whl": "linux/x64",
		"monosmith-1.129.0-py3-none-win_amd64.whl":                                  "windows/x64",
		"monosmith-1.129.0-aarch64-mingw-ucrt.gem":                                  "windows/arm64",
		"monosmith.osx-x64.1.129.0.nupkg":                                           "darwin/x64",
		"@monosmith/cli-win32-arm64@1.129.0":                                        "windows/arm64",
		"monosmith_1.129.0_amd64.deb":                                               "linux/x64",
	} {
		system, arch, ok := platformOf(name)
		if got := system + "/" + arch; !ok || got != want {
			t.Errorf("%s: %s, want %s", name, got, want)
		}
	}
	if _, _, ok := platformOf("monosmith@1.129.0"); ok {
		t.Error("the main package is for every platform")
	}
}

func TestTargetLabelsCountEveryTargetOnce(t *testing.T) {
	got := TargetLabels(smithSample())
	want := []string{"npm", "PyPI", "NuGet", "RubyGems", "GitHub Release", "pub.dev (queued)"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%v", got)
	}
	if len(TargetLabels(ReportResult{Item: Item{Project: "web", Next: "1"}})) != 0 {
		t.Error("a result that published nothing has no target")
	}
}

func TestARegistryPublishNamesTheHostAndTheExistingPackagesCount(t *testing.T) {
	res := ReportResult{Item: Item{Project: "kit", Next: "1.1.0"}, Published: true, Registry: "https://registry.npmjs.org/", Package: "@acme/kit"}
	got := TargetsOf(res)
	if len(got) != 1 || got[0].Kind != "npm" || got[0].Where != "registry.npmjs.org" || got[0].pkg != "@acme/kit" || got[0].fresh != 1 {
		t.Fatalf("%+v", got)
	}
	if labels := TargetLabels(res); len(labels) != 1 || labels[0] != "registry.npmjs.org" {
		t.Fatalf("%v", labels)
	}
	again := ReportResult{Item: Item{Next: "1.1.0"}, ExistingPackages: []string{"acme.1.1.0.nupkg"}}
	if got := TargetsOf(again); len(got) != 1 || got[0].fresh != 0 || !strings.Contains(got[0].What, "1 already published") {
		t.Fatalf("%+v", got)
	}
	pub := ReportResult{ExistingPackages: []string{"monosmith@1.4.0"}, PubExisting: "monosmith@1.4.0"}
	if got := TargetsOf(pub); len(got) != 1 || got[0].Kind != "pub.dev" || got[0].What != "already published" {
		t.Fatalf("pub.dev must not count as an npm package: %+v", got)
	}
}

func TestTheJSONReportHoldsTheTargets(t *testing.T) {
	res := smithSample()
	res.Targets = TargetsOf(res)
	raw, err := json.Marshal(Report{Results: []ReportResult{res}})
	if err != nil {
		t.Fatal(err)
	}
	var back struct {
		Results []struct {
			Targets []struct {
				Kind      string   `json:"kind"`
				What      string   `json:"what"`
				Platforms []string `json:"platforms"`
				Where     string   `json:"where"`
				Count     int      `json:"count"`
			} `json:"targets"`
		} `json:"results"`
	}
	if err := json.Unmarshal(raw, &back); err != nil {
		t.Fatal(err)
	}
	targets := back.Results[0].Targets
	if len(targets) != 6 || targets[1].Kind != "PyPI" || targets[1].Count != 6 || targets[1].Where != "pypi.org" ||
		!reflect.DeepEqual(targets[1].Platforms, []string{"darwin/x64", "darwin/arm64", "linux/x64", "linux/arm64", "windows/x64", "windows/arm64"}) {
		t.Fatalf("%s", raw)
	}
}
