package convertplan

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/nxworkspace"
)

func TestExistingTsconfigPicksTheFileTheProjectHas(t *testing.T) {
	for name, tc := range map[string]struct {
		files []string
		want  string
	}{
		"lib config":   {[]string{"tsconfig.lib.json", "tsconfig.json"}, "tsconfig.lib.json"},
		"plain config": {[]string{"tsconfig.json"}, "tsconfig.json"},
		"none":         {nil, "tsconfig.lib.json"},
	} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			dir := filepath.Join(root, "libs", "kit")
			if err := os.MkdirAll(dir, 0o755); err != nil {
				t.Fatal(err)
			}
			for _, f := range tc.files {
				if err := os.WriteFile(filepath.Join(dir, f), []byte("{}"), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			got := existingTsconfig(&nxworkspace.Workspace{Root: root}, &nxworkspace.Project{Root: "libs/kit"})
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}
