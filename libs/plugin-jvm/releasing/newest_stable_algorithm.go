package releasing

import (
	"strconv"
	"strings"
)

// NewestStable is the version a maven-metadata.xml reports as the newest release,
// for a package smith publishes with semver versions. Maven sets <release> to the
// last version deployed, so after a prerelease (2.0.0-beta.1) it names the
// prerelease; then the newest version of <versions> without a "-" qualifier is
// the release. release is the metadata's <release> and versions its <versions>.
func NewestStable(release string, versions []string) string {
	if !strings.Contains(release, "-") {
		return release
	}
	newest := ""
	for _, v := range versions {
		if strings.Contains(v, "-") {
			continue
		}
		if newest == "" || numericLess(newest, v) {
			newest = v
		}
	}
	return newest
}

// numericLess reports whether the dotted number a is below b; a segment that is
// not a number counts as 0.
func numericLess(a, b string) bool {
	as, bs := strings.Split(a, "."), strings.Split(b, ".")
	for i := 0; i < len(as) || i < len(bs); i++ {
		var x, y int
		if i < len(as) {
			x, _ = strconv.Atoi(as[i])
		}
		if i < len(bs) {
			y, _ = strconv.Atoi(bs[i])
		}
		if x != y {
			return x < y
		}
	}
	return false
}
