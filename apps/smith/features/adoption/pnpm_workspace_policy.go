package adoption

import (
	"path"
	"strings"
)

// PnpmWorkspace decides whether the repository is a pnpm workspace: it has a
// pnpm-workspace.yaml, or a package.json that uses a workspace: or catalog:
// range. smith installs with npm, which reads neither, so smith install would
// fail on it.
//
// Parameters: files are the repository's tracked files (slash paths); read
// returns a file's text, "" when it is missing.
// Returns: the blocker for --init, and true when the repository is one.
func PnpmWorkspace(files []string, read func(string) string) (Item, bool) {
	found := false
	for _, f := range files {
		if f == "pnpm-workspace.yaml" {
			found = true
			break
		}
		if path.Base(f) == "package.json" {
			text := read(f)
			if strings.Contains(text, `"workspace:`) || strings.Contains(text, `"catalog:`) {
				found = true
				break
			}
		}
	}
	if !found {
		return Item{}, false
	}
	steps := []string{
		"smith convert --dry-run, when the repository also has an nx.json or a turbo.json",
		"otherwise replace the workspace: and catalog: ranges with versions, move the packages of pnpm-workspace.yaml into \"workspaces\" in package.json, then run smith adopt again",
	}
	return Item{
		Message: "this is a pnpm workspace: smith installs with npm, and npm cannot read workspace: or catalog: ranges, so smith install would fail",
		Step:    steps[0],
		Steps:   steps,
		Blocks:  BlocksAdoption,
	}, true
}
