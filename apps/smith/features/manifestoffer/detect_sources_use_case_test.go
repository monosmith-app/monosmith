package manifestoffer

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/features/adoptmove"
)

var claims = []adoptmove.Claim{
	{Language: "node", Manifests: []string{"package.json"}, Sources: []string{"*.ts", "*.js"}},
	{Language: "python", Manifests: []string{"pyproject.toml"}, Sources: []string{"*.py"}},
	{Language: "dotnet", Manifests: []string{"*.csproj"}, Sources: []string{"*.cs"}},
	{Language: "go", Manifests: []string{"go.mod"}},
}

func write(t *testing.T, dir string, files ...string) {
	t.Helper()
	for _, f := range files {
		p := filepath.Join(dir, filepath.FromSlash(f))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestSourcesPutsTheLanguageWithMostFilesFirst(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "a.py", "b.py", "lib/c.py", "tool.js", "node_modules/x/index.js", ".venv/y.py", "main.go")
	finds := Sources(dir, claims)
	if len(finds) != 2 {
		t.Fatalf("got %d finds, want 2 (go names no sources): %+v", len(finds), finds)
	}
	if finds[0].Language != "python" || !slices.Equal(finds[0].Files, []string{"a.py", "b.py", "lib/c.py"}) {
		t.Errorf("first find = %+v, want python with a.py, b.py, lib/c.py", finds[0])
	}
	if finds[1].Language != "node" || len(finds[1].Files) != 1 {
		t.Errorf("second find = %+v, want node with tool.js only", finds[1])
	}
	if finds[0].Manifest != "pyproject.toml" {
		t.Errorf("manifest = %q", finds[0].Manifest)
	}
}

func TestSourcesNamesAPatternManifestInWords(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "Program.cs")
	finds := Sources(dir, claims)
	if len(finds) != 1 || finds[0].Manifest != "a .csproj file" {
		t.Fatalf("got %+v, want dotnet with a .csproj file", finds)
	}
}

func TestFindNestedListsTheProjectsBelowTheTop(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "README.md", "src/Lib/Lib.csproj", "src/Lib/Code.cs", "test/Lib.Tests/Lib.Tests.csproj", "bin/Old/Old.csproj", "Root.csproj")
	got := FindNested(dir, claims)
	want := []Nested{{Path: "src/Lib", Manifest: "Lib.csproj", Language: "dotnet"}, {Path: "test/Lib.Tests", Manifest: "Lib.Tests.csproj", Language: "dotnet"}}
	if !slices.Equal(got, want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}
