package workspace

import (
	"os"
	"path/filepath"

	"github.com/monosmith-app/monosmith/apps/smith/problem"
)

// MissingWorkspace is the error for a directory with no smith.toml at or above
// it. The advice depends on what is in the directory: a repository that
// already exists is brought in with adopt (or convert, for Nx and Turborepo),
// and smith new is for an empty folder.
func MissingWorkspace(dir string) error {
	has := func(names ...string) bool {
		for _, n := range names {
			if _, err := os.Stat(filepath.Join(dir, n)); err == nil {
				return true
			}
		}
		return false
	}
	e := &problem.Error{Cause: ErrNotFound}
	switch {
	case has("nx.json", "turbo.json"):
		e.Problem = "this folder is not a smith workspace yet (there is no smith.toml); it is an Nx or Turborepo workspace"
		e.Remedy = []string{"smith convert --dry-run shows what moving it to smith involves"}
	case has(".git", "package.json", "go.mod", "pyproject.toml", "Cargo.toml", "pom.xml", "pubspec.yaml", "Gemfile", "composer.json"):
		e.Problem = "this folder is not a smith workspace yet (there is no smith.toml); it already holds a repository"
		e.Remedy = []string{"smith adopt --dry-run shows what bringing it under smith involves"}
	default:
		e.Problem = "there is no smith workspace here (no smith.toml in this folder or above it)"
		e.Remedy = []string{"smith new <name> creates one", "in an existing repository, smith adopt --dry-run brings it in"}
	}
	return e
}
