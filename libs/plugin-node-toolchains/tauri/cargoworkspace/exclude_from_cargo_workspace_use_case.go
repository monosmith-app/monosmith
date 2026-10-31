package cargoworkspace

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/tauri/tauriapp"
)

// ExcludeFromCargoWorkspace lists the project's src-tauri under `exclude` in the
// root Cargo.toml's [workspace] when its `members` would otherwise take it in,
// and says whether it edited the file. A workspace without such a Cargo.toml is
// left alone.
func ExcludeFromCargoWorkspace(workspaceRoot, root string) (bool, error) {
	if len(workspaceMembers(workspaceRoot, []string{root})) == 0 {
		return false, nil
	}
	file := filepath.Join(workspaceRoot, "Cargo.toml")
	raw, err := os.ReadFile(file)
	if err != nil {
		return false, err
	}
	src, entry := string(raw), `"`+path.Join(strings.Trim(root, "/"), tauriapp.SrcDir)+`"`
	if loc := excludeList.FindStringSubmatchIndex(src); loc != nil {
		list := src[loc[2]:loc[3]]
		trimmed := strings.TrimRight(list, " \t\r\n")
		sep := ", "
		switch {
		case strings.TrimSpace(trimmed) == "":
			sep = ""
		case strings.HasSuffix(trimmed, ","):
			sep = " "
		}
		at := loc[2] + len(trimmed)
		src = src[:at] + sep + entry + src[at:]
	} else {
		at := strings.Index(src, "[workspace]") + len("[workspace]")
		src = src[:at] + "\nexclude = [" + entry + "]" + src[at:]
	}
	return true, os.WriteFile(file, []byte(src), 0o644)
}
