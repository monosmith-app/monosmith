package removalplan

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/semver"
)

// ESLintMeetsConfig says whether an eslint version is one smith's config runs
// with: the major of LintESLintFloor, at least that version.
//
// Parameters: version is a plain version ("9.39.3"), without a range prefix.
// Returns true when it is inside LintESLintRange.
func ESLintMeetsConfig(version string) bool {
	major, _, _ := strings.Cut(LintESLintFloor, ".")
	got, _, _ := strings.Cut(version, ".")
	return got == major && semver.Compare(version, LintESLintFloor) >= 0
}
