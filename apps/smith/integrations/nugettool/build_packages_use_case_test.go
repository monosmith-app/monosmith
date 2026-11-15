package nugettool

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func archive(t *testing.T, root, name, file string) string {
	t.Helper()
	f, err := os.Create(filepath.Join(root, name))
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	w, _ := zw.Create(file)
	_, _ = w.Write([]byte("binary " + file))
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()
	return name
}

func release(t *testing.T, root string) Release {
	return Release{
		ID: "acme", Binary: "acme", Version: "5.0.0", Description: "Does things", License: "MIT", Authors: "acme-org", OutDir: "nuget",
		Archives: []Archive{
			{Path: archive(t, root, "acme_windows_amd64.zip", "acme.exe"), OS: "windows", Arch: "amd64"},
			{Path: archive(t, root, "acme_linux_arm64.zip", "acme"), OS: "linux", Arch: "arm64"},
			{Path: archive(t, root, "acme_freebsd_amd64.zip", "acme"), OS: "freebsd", Arch: "amd64"},
		},
	}
}

// entries reads a package's files, failing on an entry NuGet would reject.
func entries(t *testing.T, file string) (map[string]string, map[string]os.FileMode) {
	t.Helper()
	zr, err := zip.OpenReader(file)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = zr.Close() }()
	files, modes := map[string]string{}, map[string]os.FileMode{}
	for _, f := range zr.File {
		if f.Flags&0x8 != 0 {
			t.Errorf("%s has a data descriptor", f.Name)
		}
		in, _ := f.Open()
		raw, _ := io.ReadAll(in)
		_ = in.Close()
		files[f.Name] = string(raw)
		modes[f.Name] = f.Mode()
	}
	return files, modes
}

func TestBuildMakesOnePackagePerPlatformThenTheToolPackage(t *testing.T) {
	root := t.TempDir()
	packages, skipped, err := Build(root, release(t, root))
	if err != nil {
		t.Fatal(err)
	}
	if len(skipped) != 1 || skipped[0] != "freebsd/amd64" {
		t.Errorf("skipped = %v", skipped)
	}
	var ids []string
	for _, p := range packages {
		ids = append(ids, p.ID+" "+filepath.Base(p.File))
	}
	if got := strings.Join(ids, ", "); got != "acme.linux-arm64 acme.linux-arm64.5.0.0.nupkg, acme.win-x64 acme.win-x64.5.0.0.nupkg, acme acme.5.0.0.nupkg" {
		t.Fatalf("packages = %s", got)
	}

	files, modes := entries(t, packages[0].File)
	if files["tools/any/linux-arm64/acme"] != "binary acme" || modes["tools/any/linux-arm64/acme"]&0o111 == 0 {
		t.Errorf("binary = %q, mode %v", files["tools/any/linux-arm64/acme"], modes["tools/any/linux-arm64/acme"])
	}
	if !strings.Contains(files["tools/any/linux-arm64/DotnetToolSettings.xml"], `<Command Name="acme" EntryPoint="acme" Runner="executable" />`) {
		t.Errorf("settings = %s", files["tools/any/linux-arm64/DotnetToolSettings.xml"])
	}
	nuspec := files["acme.linux-arm64.nuspec"]
	for _, want := range []string{"<id>acme.linux-arm64</id>", "<version>5.0.0</version>", "<authors>acme-org</authors>", `<packageType name="DotnetToolRidPackage" />`, `<license type="expression">MIT</license>`, `<licenseUrl>https://licenses.nuget.org/MIT</licenseUrl>`} {
		if !strings.Contains(nuspec, want) {
			t.Errorf("nuspec lacks %s:\n%s", want, nuspec)
		}
	}
	if !strings.Contains(files["[Content_Types].xml"], `<Override PartName="/tools/any/linux-arm64/acme" ContentType="application/octet" />`) {
		t.Errorf("content types = %s", files["[Content_Types].xml"])
	}
	if !strings.Contains(files["_rels/.rels"], `Target="/acme.linux-arm64.nuspec"`) {
		t.Errorf("rels = %s", files["_rels/.rels"])
	}

	files, _ = entries(t, packages[1].File)
	if !strings.Contains(files["tools/any/win-x64/DotnetToolSettings.xml"], `EntryPoint="acme.exe"`) || files["tools/any/win-x64/acme.exe"] != "binary acme.exe" {
		t.Errorf("windows package = %v", files)
	}

	files, _ = entries(t, packages[2].File)
	settings := files["tools/any/any/DotnetToolSettings.xml"]
	for _, want := range []string{`<Command Name="acme" />`, `<RuntimeIdentifierPackage RuntimeIdentifier="linux-arm64" Id="acme.linux-arm64" />`, `<RuntimeIdentifierPackage RuntimeIdentifier="win-x64" Id="acme.win-x64" />`} {
		if !strings.Contains(settings, want) {
			t.Errorf("tool settings lack %s:\n%s", want, settings)
		}
	}
	if !strings.Contains(files["acme.nuspec"], `<packageType name="DotnetTool" />`) || !strings.Contains(files["acme.nuspec"], "<readme>README.md</readme>") || !strings.Contains(files["README.md"], "dotnet tool install -g acme") {
		t.Errorf("tool package = %v", files)
	}
}

func TestBuildFailsWhenAnArchiveLacksTheBinary(t *testing.T) {
	root := t.TempDir()
	r := release(t, root)
	r.Binary = "other"
	if _, _, err := Build(root, r); err == nil || !strings.Contains(err.Error(), "has no other") {
		t.Errorf("err = %v", err)
	}
}
