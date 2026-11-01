package vscode

import (
	"fmt"
	"strconv"
	"strings"
)

// MarketplaceVersion is the version a VS Code extension is published under.
// The Marketplace takes plain major.minor.patch numbers and no semver
// prerelease tag, and tells a pre-release from a release by a flag, so a
// channel's 2.0.0-beta.3 becomes 2.1.3: the major, the odd minor after the
// channel's (2m+1, the Marketplace's own convention: even minors release, odd
// ones pre-release), and the channel's counter as the patch. A version with no
// prerelease part is returned as it is. A prerelease with no numeric counter
// has no spelling and is an error.
func MarketplaceVersion(version string) (string, error) {
	core, pre, found := strings.Cut(version, "-")
	if !found {
		return version, nil
	}
	parts := strings.Split(core, ".")
	_, counter, _ := strings.Cut(pre, ".")
	major, errMajor := strconv.Atoi(parts[0])
	minor, errMinor := 0, error(nil)
	if len(parts) > 1 {
		minor, errMinor = strconv.Atoi(parts[1])
	}
	n, errN := strconv.Atoi(counter)
	if len(parts) != 3 || errMajor != nil || errMinor != nil || errN != nil {
		return "", fmt.Errorf("%q has no Marketplace version: it needs major.minor.patch-<word>.<number>", version)
	}
	return fmt.Sprintf("%d.%d.%d", major, 2*minor+1, n), nil
}

// isPrerelease says whether the version has a semver prerelease part.
func isPrerelease(version string) bool { return strings.Contains(version, "-") }
