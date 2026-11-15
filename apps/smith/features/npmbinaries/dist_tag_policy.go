package npmbinaries

import (
	"cmp"
	"fmt"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/npmpropagation"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"strings"
	"time"
)

// NotServedNote says why a dist-tag was not moved when the wait ended: npm said
// it published the packages, the registry still does not serve them. It names
// each package and says how to finish, by hand or by running the release again.
func NotServedNote(m protocol.TagMove, out npmpropagation.Outcome) string {
	missing := "the platform packages"
	if len(out.Missing) > 0 {
		var named []string
		for _, n := range out.Missing {
			named = append(named, n+"@"+m.Version)
		}
		missing = strings.Join(named, ", ")
	}
	return fmt.Sprintf("the dist-tag %q of %s was not moved: npm said it published %s but the registry still does not serve it after %s. "+
		"Run the release again: it republishes what is missing and moves the tag. Or, when npm serves it, run: %s",
		m.Tag, m.Package, missing, out.Waited.Round(time.Second), m.Command())
}

// intendedTag is the dist-tag the release moves: a prerelease's channel (npm
// refuses a prerelease without a tag, so it never becomes latest), and none
// for a stable version, which takes latest.
func intendedTag(r Release) string {
	if strings.Contains(r.Version, "-") {
		return cmp.Or(r.Tag, "next")
	}
	return ""
}
