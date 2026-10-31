package scaffold

import (
	"encoding/json"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// nodeWorkspaceDirectory is the parent folder of a new Node project. It is the
// kind's own folder (conventional) when the root package.json "workspaces"
// already cover a project there, so npm links it. When they do not (a layout
// converted from Nx lists only packages/* and apps/*), it is the first folder
// those globs name that exists, packages first, so the new project is part of
// the npm workspace at once. It returns conventional when the root has no
// workspaces or no glob names a folder.
func nodeWorkspaceDirectory(root, conventional, name string) string {
	raw, err := os.ReadFile(filepath.Join(root, "package.json"))
	if err != nil {
		return conventional
	}
	var pkg struct {
		Workspaces json.RawMessage `json:"workspaces"`
	}
	if json.Unmarshal(raw, &pkg) != nil {
		return conventional
	}
	var globs []string
	if json.Unmarshal(pkg.Workspaces, &globs) != nil {
		var object struct {
			Packages []string `json:"packages"`
		}
		if json.Unmarshal(pkg.Workspaces, &object) != nil {
			return conventional
		}
		globs = object.Packages
	}
	var parents []string
	for _, glob := range globs {
		glob = strings.TrimPrefix(glob, "./")
		if ok, _ := path.Match(glob, conventional+"/"+name); ok {
			return conventional
		}
		parent, found := strings.CutSuffix(glob, "/*")
		if found && !strings.ContainsAny(parent, "*?[") {
			parents = append(parents, parent)
		}
	}
	for _, want := range []string{"packages", ""} {
		for _, parent := range parents {
			if want != "" && parent != want {
				continue
			}
			if info, err := os.Stat(filepath.Join(root, filepath.FromSlash(parent))); err == nil && info.IsDir() {
				return parent
			}
		}
	}
	return conventional
}
