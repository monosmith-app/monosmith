package projectgraph

import (
	"slices"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

// CoversLanguage says whether a convention that declares the languages
// supported applies to a project of language. A convention that declares none
// applies to every language: the handshake's languages are the one place that
// says where a convention applies.
func CoversLanguage(supported []string, language string) bool {
	return len(supported) == 0 || slices.Contains(supported, language)
}

// SplitByLanguage divides projects into those a convention declaring supported
// languages covers, and those it does not because of their language.
func SplitByLanguage(projects []*Project, supported []string) (covered, skipped []*Project) {
	for _, p := range projects {
		if CoversLanguage(supported, p.Language) {
			covered = append(covered, p)
		} else {
			skipped = append(skipped, p)
		}
	}
	return covered, skipped
}

// Coverage is what a convention does to the projects of a workspace.
type Coverage struct {
	// Checked are the projects it applies to.
	Checked []*Project
	// Skipped are the projects it leaves alone because their language is not
	// one it declares. Projects the person left out in smith.toml (exclude, or
	// a narrower languages list) are neither: that was their choice.
	Skipped []*Project
}

// ConventionCoverage says which projects a convention declaring the languages
// declared applies to under cfg, and which it skips for want of the language.
func ConventionCoverage(projects []*Project, declared []string, cfg workspace.ConventionConfig) Coverage {
	var chosen []*Project
	for _, p := range projects {
		if slices.Contains(cfg.Exclude, p.Name) || (len(cfg.Languages) > 0 && !slices.Contains(cfg.Languages, p.Language)) {
			continue
		}
		chosen = append(chosen, p)
	}
	checked, skipped := SplitByLanguage(chosen, declared)
	return Coverage{Checked: checked, Skipped: skipped}
}
