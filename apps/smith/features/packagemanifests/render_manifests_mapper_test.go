package packagemanifests

import (
	"encoding/json"
	"strings"
	"testing"
)

func release() Release {
	r := Release{Product: "my-tool", Binary: "mytool", Version: "1.2.3", Description: "Does things", License: "MIT", Owner: "Acme", Repo: "tools", Tag: "my-tool@1.2.3"}
	for _, p := range []struct{ os, arch string }{{"darwin", "arm64"}, {"darwin", "amd64"}, {"linux", "amd64"}, {"windows", "amd64"}, {"windows", "arm64"}} {
		r.Archives = append(r.Archives, Archive{Name: "my-tool_1.2.3_" + p.os + "_" + p.arch + ".zip", OS: p.os, Arch: p.arch, SHA256: strings.Repeat("ab", 32)})
	}
	return r
}

func TestRender(t *testing.T) {
	files, skipped := Render(release(), []string{"homebrew", "scoop", "winget"})
	if len(skipped) != 0 {
		t.Fatal(skipped)
	}
	formula := files["my-tool.rb"]
	for _, want := range []string{
		"class MyTool < Formula", `desc "Does things"`, `homepage "https://github.com/Acme/tools"`, `license "MIT"`,
		"  on_macos do\n    on_arm do\n      url \"https://github.com/Acme/tools/releases/download/my-tool%401.2.3/my-tool_1.2.3_darwin_arm64.zip\"",
		"  on_linux do\n    on_intel do", `bin.install "mytool"`, `system "#{bin}/mytool", "--version"`,
	} {
		if !strings.Contains(formula, want) {
			t.Errorf("formula lacks %q:\n%s", want, formula)
		}
	}
	if strings.Contains(formula, "linux_arm64") {
		t.Error("only the archives that exist")
	}
	var scoop struct {
		Version      string
		Bin          string
		Architecture map[string]struct{ URL, Hash string }
	}
	if err := json.Unmarshal([]byte(files["my-tool.json"]), &scoop); err != nil || scoop.Bin != "mytool.exe" || !strings.HasSuffix(scoop.Architecture["64bit"].URL, "windows_amd64.zip") || scoop.Architecture["arm64"].Hash == "" {
		t.Fatalf("scoop %v %+v", err, scoop)
	}
	installer := files["Acme.my-tool.installer.yaml"]
	for _, want := range []string{"PackageIdentifier: Acme.my-tool", "PackageVersion: 1.2.3", "NestedInstallerType: portable", "RelativeFilePath: mytool.exe", "Architecture: x64", "InstallerSha256: " + strings.Repeat("AB", 32)} {
		if !strings.Contains(installer, want) {
			t.Errorf("winget installer lacks %q:\n%s", want, installer)
		}
	}
	if !strings.Contains(files["Acme.my-tool.locale.en-US.yaml"], "License: MIT") || !strings.Contains(files["Acme.my-tool.yaml"], "ManifestType: version") {
		t.Fatal("winget version and locale manifests")
	}
}

func TestRenderSkipsWhatHasNoArchive(t *testing.T) {
	r := release()
	r.Archives = r.Archives[:3] // no Windows
	files, skipped := Render(r, []string{"homebrew", "scoop", "winget"})
	if len(files) != 1 || len(skipped) != 2 || !strings.Contains(skipped[0], "no Windows archive") {
		t.Fatalf("%v %v", keys(files), skipped)
	}
}

func keys(m map[string]string) []string {
	var out []string
	for k := range m {
		out = append(out, k)
	}
	return out
}
