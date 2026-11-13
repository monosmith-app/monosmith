package removalplan

import "github.com/monosmith-app/monosmith/packages/plugin-sdk/eslintpeer"

// ESLintMeetsConfig says whether an eslint version is one smith's config runs
// with: the major of LintESLintFloor, at least that version.
//
// Parameters: version is a plain version ("9.39.3"), without a range prefix.
// Returns true when it is inside LintESLintRange.
func ESLintMeetsConfig(version string) bool { return eslintpeer.Meets(version) }
