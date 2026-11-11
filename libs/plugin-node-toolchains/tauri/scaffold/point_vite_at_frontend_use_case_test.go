package scaffold

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPointViteAtFrontend(t *testing.T) {
	cases := map[string]struct{ config, want, changed string }{
		"smith's outDir moves": {"build: { outDir: '../../dist/apps/studio' }", "build: { outDir: '../../dist/tauri/apps/studio/frontend' }", "apps/studio/vite.config.ts"},
		"another outDir stays": {"build: { outDir: 'out' }", "build: { outDir: 'out' }", ""},
		"no config":            {"", "", ""},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			ws := t.TempDir()
			dir := filepath.Join(ws, "apps", "studio")
			if err := os.MkdirAll(dir, 0o755); err != nil {
				t.Fatal(err)
			}
			if c.config != "" {
				if err := os.WriteFile(filepath.Join(dir, "vite.config.ts"), []byte(c.config), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			changed, err := PointViteAtFrontend(ws, "apps/studio")
			if err != nil || changed != c.changed {
				t.Fatalf("changed %q, %v", changed, err)
			}
			if got, _ := os.ReadFile(filepath.Join(dir, "vite.config.ts")); string(got) != c.want {
				t.Errorf("config %q, want %q", got, c.want)
			}
		})
	}
}
