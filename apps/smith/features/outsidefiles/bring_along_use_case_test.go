package outsidefiles

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBringCopiesTheFileAndWritesItsNewPath(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "Lib.csproj", `<Project><ItemGroup><None Include="..\LICENSE" /></ItemGroup><PropertyGroup><AssemblyOriginatorKeyFile>..\keys\k.snk</AssemblyOriginatorKeyFile></PropertyGroup></Project>`)
	write(t, dir, "README.md", "mine")
	files := []File{
		{Ref: `..\LICENSE`, Path: "LICENSE", Manifest: "Lib.csproj", Kind: KindFile, Content: []byte("MIT")},
		{Ref: `..\keys\k.snk`, Path: "keys/k.snk", Manifest: "Lib.csproj", Kind: KindFile, Content: []byte("key")},
		{Ref: `..\README.md`, Path: "README.md", Manifest: "Lib.csproj", Kind: KindFile, Content: []byte("theirs")},
		{Ref: `..\Other\Other.csproj`, Path: "Other/Other.csproj", Manifest: "Lib.csproj", Kind: KindProject},
		{Path: "Directory.Build.props", Kind: KindAbove},
	}
	brought, failed := Bring(dir, files)
	if len(brought) != 2 || len(failed) != 1 || failed["README.md"] == nil {
		t.Fatalf("brought %+v, failed %v", brought, failed)
	}
	if raw, _ := os.ReadFile(filepath.Join(dir, "LICENSE")); string(raw) != "MIT" {
		t.Errorf("LICENSE: %q", raw)
	}
	if raw, _ := os.ReadFile(filepath.Join(dir, "README.md")); string(raw) != "mine" {
		t.Errorf("a file of the project is never replaced: %q", raw)
	}
	raw, _ := os.ReadFile(filepath.Join(dir, "Lib.csproj"))
	if text := string(raw); !strings.Contains(text, `Include="LICENSE"`) || !strings.Contains(text, "<AssemblyOriginatorKeyFile>k.snk<") || strings.Contains(text, `..\LICENSE`) {
		t.Errorf("the project file names the new paths:\n%s", text)
	}
}
