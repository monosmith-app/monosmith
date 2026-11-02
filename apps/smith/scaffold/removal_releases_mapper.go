package scaffold

import (
	"fmt"
	"strings"
)

// KeptReleasesLines say what smith remove leaves alone: the releases already
// made. A tag, a GitHub Release and a registry package outlive the folder, and
// smith asks no registry (no network call is made for this line).
//
// Parameters: r is the removal; its Releases are the tags found with
// `git tag --list '<project>@*'`.
// Returns: the lines to print, each starting with two spaces; empty when the
// project never released and so has nothing that stays.
func KeptReleasesLines(r Removal) []string {
	if len(r.Releases) == 0 {
		return nil
	}
	shown := r.Releases
	more := ""
	if len(shown) > 6 {
		more = fmt.Sprintf(" and %d more", len(shown)-6)
		shown = shown[len(shown)-6:]
	}
	return []string{
		"  its releases stay: tags " + strings.Join(shown, ", ") + more + ", GitHub Releases and registry packages are not touched",
		"  smith did not ask any registry; if the package is published, unpublish or deprecate it there yourself",
	}
}
