package workspace

import (
	"fmt"
	"regexp"
)

var channelPrerelease = regexp.MustCompile(`^[A-Za-z][0-9A-Za-z-]*$`)

// validateReleaseChannels refuses a [[release.channels]] entry that would
// release from the release branch, from the same branch as another channel, or
// under a version that is not a semantic prerelease. r.Branch is already set.
func validateReleaseChannels(r ReleaseSection) error {
	seen := map[string]bool{r.Branch: true}
	for i, c := range r.Channels {
		at := fmt.Sprintf("%s: [[release.channels]] entry %d", FileName, i+1)
		switch {
		case c.Branch == "":
			return fmt.Errorf("%s: branch must name the branch the channel releases from", at)
		case seen[c.Branch] && c.Branch == r.Branch:
			return fmt.Errorf("%s: branch %q is the [release] branch, which releases stable versions", at, c.Branch)
		case seen[c.Branch]:
			return fmt.Errorf("%s: branch %q is the branch of another channel", at, c.Branch)
		case !channelPrerelease.MatchString(c.Prerelease):
			return fmt.Errorf("%s: prerelease must be an identifier such as \"beta\" or \"rc\", not %q", at, c.Prerelease)
		case c.Major < 1:
			return fmt.Errorf("%s: major must be 1 or more, not %d", at, c.Major)
		}
		seen[c.Branch] = true
	}
	return nil
}
