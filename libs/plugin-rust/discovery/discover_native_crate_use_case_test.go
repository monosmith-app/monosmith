package discovery

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestACrateMarkedNativeIsTaggedNative(t *testing.T) {
	ws := t.TempDir()
	for name, extra := range map[string]string{"cli": "[package.metadata.smith]\nnative = true\n", "tool": ""} {
		dir := filepath.Join(ws, "apps", name)
		_ = os.MkdirAll(filepath.Join(dir, "src"), 0o755)
		_ = os.WriteFile(filepath.Join(dir, "src", "main.rs"), []byte("fn main() {}\n"), 0o644)
		_ = os.WriteFile(filepath.Join(dir, "Cargo.toml"), []byte("[package]\nname = \""+name+"\"\nversion = \"0.1.0\"\n\n"+extra), 0o644)
	}
	got, err := Discover(ws, []string{"apps/cli/Cargo.toml", "apps/tool/Cargo.toml"})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(got.Projects[0].Tags, []string{NativeTag}) || len(got.Projects[1].Tags) != 0 {
		t.Errorf("%+v", got.Projects)
	}
}
