package workspace

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestClearEnabledPluginsDropsOnlyThatLine(t *testing.T) {
	src := "[workspace]\nname = \"a\" # inline\n\n[plugins] # mine\nenabled = [\"go\", \"node\"] # frozen\nmarketplace = \"x\"\n\n[cache]\ndir = \"x\"\n"
	root := t.TempDir()
	path := filepath.Join(root, FileName)
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := ClearEnabledPlugins(root); err != nil {
		t.Fatal(err)
	}
	out, _ := os.ReadFile(path)
	if strings.Contains(string(out), "enabled") {
		t.Fatalf("the line stayed:\n%s", out)
	}
	for _, keep := range []string{"# inline", "# mine", "marketplace = \"x\"", "dir = \"x\""} {
		if !strings.Contains(string(out), keep) {
			t.Fatalf("lost %q:\n%s", keep, out)
		}
	}
	if cfg, err := Load(root); err != nil || len(cfg.Plugins.Enabled) != 0 {
		t.Fatalf("enabled %v, err %v", cfg.Plugins.Enabled, err)
	}
}

func TestClearEnabledPluginsLeavesAFileWithoutTheLineAlone(t *testing.T) {
	for _, src := range []string{"[workspace]\nname = \"a\"\n", "[workspace]\nname = \"a\"\n[plugins]\nmarketplace = \"x\"\n"} {
		root := t.TempDir()
		path := filepath.Join(root, FileName)
		_ = os.WriteFile(path, []byte(src), 0o644)
		if err := ClearEnabledPlugins(root); err != nil {
			t.Fatal(err)
		}
		if out, _ := os.ReadFile(path); string(out) != src {
			t.Fatalf("file changed:\n%s", out)
		}
	}
}
