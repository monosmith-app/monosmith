package legacyjestsupport

import (
	"fmt"
	"regexp"
	"strconv"
)

// StableRange is the range of the preset for a stable smith. The preset has
// its own stable line (0.1.0, 0.2.0, 0.2.1): it is released when its own folder
// changes, not with every smith release, so a stable smith 2.0.0 has no preset
// 2.0.0 to name. The range stays on the 0.x line until a stable preset of a
// higher major exists.
const StableRange = "^0.2.0"

// smithVersionPattern reads major, minor, patch and the prerelease of a smith
// version, with or without a leading v.
var smithVersionPattern = regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)(?:-([0-9A-Za-z-]+)(?:\.\d+)?)?`)

// PackageRangeFor returns the range of @monosmith/jest-preset that belongs to
// the running smith. A channel release of smith (2.0.0-beta.N) releases the
// preset as 2.0.0-beta.M when the preset changes, and the counters of the two
// differ, so the range starts at the first release of the channel
// (^2.0.0-beta.1) and npm resolves the newest one. A stable smith, a dev build
// and an unreadable version get StableRange: the stable preset line is 0.x
// whatever the major of smith (a range such as ^2.0.0 names a version npm does
// not have, and npm install fails with ETARGET).
//
// Parameters: smith is the version of the running smith.
// Returns: the npm range to write in a devDependencies entry.
func PackageRangeFor(smith string) string {
	m := smithVersionPattern.FindStringSubmatch(smith)
	if m == nil {
		return StableRange
	}
	major, _ := strconv.Atoi(m[1])
	if major < 2 || m[4] == "" {
		return StableRange
	}
	return fmt.Sprintf("^%s.%s.%s-%s.1", m[1], m[2], m[3], m[4])
}
