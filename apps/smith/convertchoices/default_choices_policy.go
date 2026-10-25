package convertchoices

import "github.com/monosmith-app/monosmith/apps/smith/legacyinventory"

// Defaults are the choices a workspace's own tooling suggests.
//
// Lint is smith's when the root config can be moved over without losing the
// rules the team wrote (it imports a known shared config, so the rules added
// beside it are kept), and kept otherwise. The test runner is the one the
// workspace uses. Build and CI are smith's: the conversion has already
// replaced the old build and the pipeline.
func Defaults(u legacyinventory.Usage) Choices {
	c := Choices{Level: LevelFull, Lint: Keep, Test: Keep, Build: Smith, CI: Smith}
	if u.ESLint && u.SharedLintConfig != "" {
		c.Lint = Smith
	}
	switch {
	case u.Jest:
		c.Test = Jest
	case u.Vitest:
		c.Test = Vitest
	}
	return c
}
