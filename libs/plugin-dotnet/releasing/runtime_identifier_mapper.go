package releasing

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// runtimeIdentifiers are .NET's runtime identifiers (RIDs) of the platforms
// an app is released for, by Go's "os/arch" spelling (the one smith.toml
// uses).
var runtimeIdentifiers = map[string]string{
	"windows/amd64": "win-x64",
	"windows/arm64": "win-arm64",
	"linux/amd64":   "linux-x64",
	"linux/arm64":   "linux-arm64",
	"darwin/amd64":  "osx-x64",
	"darwin/arm64":  "osx-arm64",
}

// RuntimeIdentifier is the RID of a platform ("os/arch").
func RuntimeIdentifier(platform string) (string, error) {
	if rid, ok := runtimeIdentifiers[platform]; ok {
		return rid, nil
	}
	known := make([]string, 0, len(runtimeIdentifiers))
	for p := range runtimeIdentifiers {
		known = append(known, p)
	}
	slices.Sort(known)
	return "", fmt.Errorf("%q has no .NET runtime: release a .NET app for %s", platform, strings.Join(known, ", "))
}

// RuntimePackCached reports whether the runtime pack of a RID is already in
// the NuGet package cache. A self-contained app carries the runtime, so the
// first publish for a RID downloads that pack from NuGet; later ones, and
// machines without network, need it cached. Tests replace it.
var RuntimePackCached = func(rid string) bool {
	root := os.Getenv("NUGET_PACKAGES")
	if root == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return false
		}
		root = filepath.Join(home, ".nuget", "packages")
	}
	_, err := os.Stat(filepath.Join(root, "microsoft.netcore.app.runtime."+rid))
	return err == nil
}
