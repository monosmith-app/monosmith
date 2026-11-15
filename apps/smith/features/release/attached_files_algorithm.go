package release

import (
	"fmt"
	"path"
	"strings"
)

// fileKinds are the kinds of file a GitHub Release holds, in the order they
// are listed. A counted kind prints its number ("6 zips"); the others name the
// kind ("checksums") because a release holds one of each, or one group.
var fileKinds = []struct {
	key              string
	singular, plural string
	counted          bool
}{
	{"zip", "zip", "zips", true},
	{"tarball", "tarball", "tarballs", true},
	{"package", "deb/rpm", "deb/rpm", true},
	{"checksums", "checksums", "checksums", false},
	{"sbom", "SBOM", "SBOM", false},
	{"winget", "winget", "winget", false},
	{"brew", "brew", "brew", false},
	{"manifest", "manifest", "manifest", false},
	{"other", "other file", "other files", true},
}

// fileKindOf says what kind of release file a name is.
func fileKindOf(name string) string {
	lower := strings.ToLower(path.Base(strings.ReplaceAll(name, "\\", "/")))
	switch {
	case strings.HasSuffix(lower, ".zip"):
		return "zip"
	case strings.HasSuffix(lower, ".tar.gz"), strings.HasSuffix(lower, ".tgz"):
		return "tarball"
	case strings.HasSuffix(lower, ".deb"), strings.HasSuffix(lower, ".rpm"):
		return "package"
	case strings.Contains(lower, "sbom"), strings.HasSuffix(lower, ".cdx.json"), strings.HasSuffix(lower, ".spdx.json"):
		return "sbom"
	case strings.Contains(lower, "checksum"), strings.Contains(lower, "sha256sums"):
		return "checksums"
	case strings.HasSuffix(lower, ".yaml"), strings.HasSuffix(lower, ".yml"):
		return "winget"
	case strings.HasSuffix(lower, ".rb"):
		return "brew"
	case strings.HasSuffix(lower, ".json"):
		return "manifest"
	}
	return "other"
}

// summariseFiles counts files by kind: "6 zips, 4 deb/rpm, checksums, SBOM".
func summariseFiles(names []string) string {
	counts := map[string]int{}
	for _, name := range names {
		counts[fileKindOf(name)]++
	}
	var parts []string
	for _, kind := range fileKinds {
		n := counts[kind.key]
		switch {
		case n == 0:
		case !kind.counted:
			parts = append(parts, kind.singular)
		case n == 1:
			parts = append(parts, "1 "+kind.singular)
		default:
			parts = append(parts, fmt.Sprintf("%d %s", n, kind.plural))
		}
	}
	return strings.Join(parts, ", ")
}
