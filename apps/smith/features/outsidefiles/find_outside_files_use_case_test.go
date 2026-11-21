package outsidefiles

import (
	"os"
	"path/filepath"
	"testing"
)

func write(t *testing.T, root, rel, body string) {
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

const csproj = `<Project Sdk="Microsoft.NET.Sdk">
  <PropertyGroup><AssemblyOriginatorKeyFile>..\keys\signing.snk</AssemblyOriginatorKeyFile></PropertyGroup>
  <ItemGroup>
    <None Include="..\LICENSE" Pack="true" PackagePath="" />
    <None Include="..\gone.txt" />
    <None Include="..\docs" />
    <None Include="..\..\..\outside.txt" />
    <ProjectReference Include="..\Other\Other.csproj" />
  </ItemGroup>
</Project>`

func TestFindListsWhatTheProjectUsesOutsideItsFolder(t *testing.T) {
	top := t.TempDir()
	write(t, top, "src/LICENSE", "MIT")
	write(t, top, "src/keys/signing.snk", "key")
	write(t, top, "src/docs/a.md", "")
	write(t, top, "Directory.Build.props", "<Project />")
	write(t, top, "src/Directory.Packages.props", "<Project />")
	write(t, top, "src/Lib/Lib.csproj", csproj)
	write(t, top, "src/Other/Other.csproj", "<Project />")

	got := Find(top, "src/Lib")
	byPath := map[string]File{}
	for _, f := range got {
		byPath[f.Path] = f
	}
	if len(got) != 7 {
		t.Fatalf("seven files: %+v", got)
	}
	if f := byPath["src/LICENSE"]; !f.Movable() || string(f.Content) != "MIT" {
		t.Errorf("..\\LICENSE of src/Lib is src/LICENSE: %+v", f)
	}
	if f := byPath["src/keys/signing.snk"]; !f.Movable() || f.Manifest != "Lib.csproj" || f.Ref != `..\keys\signing.snk` {
		t.Errorf("the key path resolves from the project folder: %+v", f)
	}
	if f := byPath["src/docs"]; f.Kind != KindFolder {
		t.Errorf("a folder: %+v", f)
	}
	if f := byPath["src/gone.txt"]; !f.Missing || f.Movable() {
		t.Errorf("a missing file cannot come along: %+v", f)
	}
	if f := byPath["src/Other/Other.csproj"]; f.Kind != KindProject || f.Movable() {
		t.Errorf("a project reference cannot come along: %+v", f)
	}
	if f := byPath["Directory.Build.props"]; f.Kind != KindAbove {
		t.Errorf("a Directory.Build.props at the top: %+v", f)
	}
	if f := byPath["src/Directory.Packages.props"]; f.Kind != KindAbove {
		t.Errorf("a Directory.Packages.props in the parent: %+v", f)
	}
	if _, ok := byPath["../outside.txt"]; ok {
		t.Error("a path that leaves the source is left out")
	}
}

func TestFindReadsAFileAtTheTopOfTheSource(t *testing.T) {
	top := t.TempDir()
	write(t, top, "LICENSE", "MIT")
	write(t, top, "docs/a.md", "")
	write(t, top, "Roy-T.AStar/Roy-T.AStar.csproj", `<Project><ItemGroup><None Include="..\LICENSE"><Pack>True</Pack></None><None Include="..\docs" /></ItemGroup></Project>`)
	got := Find(top, "Roy-T.AStar")
	if len(got) != 2 {
		t.Fatalf("the licence and the folder: %+v", got)
	}
	lic, docs := got[0], got[1]
	if lic.Path != "LICENSE" || lic.Ref != `..\LICENSE` || string(lic.Content) != "MIT" || !lic.Movable() || lic.Name() != "LICENSE" {
		t.Errorf("the licence comes along: %+v", lic)
	}
	if docs.Kind != KindFolder || docs.Movable() {
		t.Errorf("a folder does not come along as a file: %+v", docs)
	}
}

func TestFindOfANodePackageAndOfTheTop(t *testing.T) {
	top := t.TempDir()
	write(t, top, "LICENSE", "MIT")
	write(t, top, "Directory.Build.props", "<Project />")
	write(t, top, "packages/a/package.json", `{"name":"a","files":["dist","../../LICENSE"]}`)
	got := Find(top, "packages/a")
	if len(got) != 1 || got[0].Path != "LICENSE" || got[0].Manifest != "package.json" {
		t.Fatalf("the licence only, no MSBuild file above a Node package: %+v", got)
	}
	if Find(top, ".") != nil || Find(top, "") != nil {
		t.Fatal("the top of the source uses nothing outside itself")
	}
}
