package legacyjesthelpers

import (
	"fmt"
	"regexp"
	"strconv"
)

// StableRange is the range of the preset for a stable 0.x or 1.x smith. The
// preset has its own stable line (0.1.0, 0.2.0), so it keeps a 0.x range until
// smith's stable line reaches the major version of the preset's next channel.
const StableRange = "^0.2.0"

// smithVersionPattern reads major, minor, patch and the prerelease of a smith
// version, with or without a leading v.
var smithVersionPattern = regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)(?:-([0-9A-Za-z-]+)(?:\.\d+)?)?`)

// PackageRangeFor returns the range of @monosmith/jest-preset that belongs to
// the running smith. The preset is released with smith on the same release
// line (ADR 0051): a channel release of smith (2.0.0-beta.N) releases the
// preset as 2.0.0-beta.M, and the counters of the two can differ, so the range
// starts at the first release of the channel (^2.0.0-beta.1) and npm resolves
// the newest one. After the channel's major is stable, the range is ^MAJOR.0.0.
// A stable 0.x or 1.x smith, a dev build and an unreadable version get
// StableRange.
//
// Parameters: smith is the version of the running smith.
// Returns: the npm range to write in a devDependencies entry.
func PackageRangeFor(smith string) string {
	m := smithVersionPattern.FindStringSubmatch(smith)
	if m == nil {
		return StableRange
	}
	major, _ := strconv.Atoi(m[1])
	if major < 2 {
		return StableRange
	}
	if m[4] != "" {
		return fmt.Sprintf("^%s.%s.%s-%s.1", m[1], m[2], m[3], m[4])
	}
	return fmt.Sprintf("^%d.0.0", major)
}
