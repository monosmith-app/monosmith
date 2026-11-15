package scaffold

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNodeWorkspaceDirectoryFollowsTheWorkspaceGlobs(t *testing.T) {
	for name, tc := range map[string]struct {
		manifest string
		folders  []string
		want     string
	}{
		"libs glob present":       {`{"workspaces":["apps/*","libs/*"]}`, []string{"apps", "libs"}, "libs"},
		"nx layout, no libs glob": {`{"workspaces":["packages/*","apps/*"]}`, []string{"packages", "apps"}, "packages"},
		"another folder":          {`{"workspaces":["apps/*","tools/*"]}`, []string{"tools"}, "tools"},
		"object form":             {`{"workspaces":{"packages":["packages/*"]}}`, []string{"packages"}, "packages"},
		"no workspaces":           {`{"name":"x"}`, nil, "libs"},
		"glob names no folder":    {`{"workspaces":["packages/*"]}`, nil, "libs"},
		"glob matches the folder": {`{"workspaces":["libs/*"]}`, nil, "libs"},
	} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			if err := os.WriteFile(filepath.Join(root, "package.json"), []byte(tc.manifest), 0o644); err != nil {
				t.Fatal(err)
			}
			for _, f := range tc.folders {
				if err := os.MkdirAll(filepath.Join(root, f), 0o755); err != nil {
					t.Fatal(err)
				}
			}
			if got := nodeWorkspaceDirectory(root, "libs", "foo"); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestNodeWorkspaceDirectoryWithoutPackageJSON(t *testing.T) {
	if got := nodeWorkspaceDirectory(t.TempDir(), "libs", "foo"); got != "libs" {
		t.Errorf("got %q, want libs", got)
	}
}
