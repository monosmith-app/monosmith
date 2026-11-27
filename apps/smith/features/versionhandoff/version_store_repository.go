package versionhandoff

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"

	"github.com/monosmith-app/monosmith/apps/smith/integrations/smithreleases"
)

// installedName is the name of a version folder: it starts with three numbers.
var installedName = regexp.MustCompile(`^\d+\.\d+\.\d+`)

// WorkspacePin is the pin of the workspace a folder is in: [workspace] smith in
// the nearest smith.toml at or above it.
//
// Parameters: from is the folder to start from (the current folder).
// Returns: the pinned version, or "" when the nearest smith.toml pins none or
// latest, or no smith.toml is found.
func WorkspacePin(from string) string {
	for dir := from; dir != ""; {
		raw, err := os.ReadFile(filepath.Join(dir, "smith.toml"))
		if err == nil {
			return PinIn(string(raw))
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
	return ""
}

// NewestInstalled is the newest version in <home>/versions that has its program.
//
// Parameters: home is the smith home; goos is runtime.GOOS.
// Returns: the version, or "" when none is installed.
func NewestInstalled(home, goos string) string {
	entries, err := os.ReadDir(filepath.Join(home, "versions"))
	if err != nil {
		return ""
	}
	var installed []string
	for _, e := range entries {
		if !installedName.MatchString(e.Name()) {
			continue
		}
		if _, err := os.Stat(smithreleases.ProgramPath(home, e.Name(), goos)); err == nil {
			installed = append(installed, e.Name())
		}
	}
	if len(installed) == 0 {
		return ""
	}
	return slices.MaxFunc(installed, CompareVersions)
}
