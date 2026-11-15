package workspace

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSetEnabledPluginsKeepsEverythingElse(t *testing.T) {
	cases := map[string]string{
		"no table":      "# top comment\n[workspace]\nname = \"a\" # inline\n",
		"empty table":   "[workspace]\nname = \"a\"\n\n[plugins] # mine\n\n[cache]\ndir = \"x\"\n",
		"existing line": "[workspace]\nname = \"a\"\n[plugins]\nenabled = [\"node\"] # trimmed\n[plugins.external.r]\ncommand = [\"r\"]\n",
		"no final eol":  "[workspace]\nname = \"a\"",
	}
	for name, src := range cases {
		root := t.TempDir()
		_ = os.WriteFile(filepath.Join(root, FileName), []byte(src), 0o644)
		if err := SetEnabledPlugins(root, []string{"go", "python"}); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		out, _ := os.ReadFile(filepath.Join(root, FileName))
		cfg, err := Load(root)
		if err != nil {
			t.Fatalf("%s: %v\n%s", name, err, out)
		}
		if strings.Join(cfg.Plugins.Enabled, ",") != "go,python" {
			t.Fatalf("%s: enabled %v\n%s", name, cfg.Plugins.Enabled, out)
		}
		for _, keep := range []string{"# top comment", "# inline", "# mine", "dir = \"x\"", "command = [\"r\"]"} {
			if strings.Contains(src, keep) && !strings.Contains(string(out), keep) {
				t.Fatalf("%s: lost %q\n%s", name, keep, out)
			}
		}
		if strings.Count(string(out), "enabled =") != 1 {
			t.Fatalf("%s: duplicate lines\n%s", name, out)
		}
	}
}
