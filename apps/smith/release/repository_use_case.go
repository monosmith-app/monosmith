package release

import (
	"fmt"
	"os"

	"github.com/monosmith-app/monosmith/apps/smith/githubrelease"
	"github.com/monosmith-app/monosmith/apps/smith/session"
)

// Repository is the GitHub repository releases go to: GITHUB_REPOSITORY on
// Actions, otherwise the release remote's URL.
func Repository(s *session.Session) (githubrelease.Repository, error) {
	if env := os.Getenv("GITHUB_REPOSITORY"); env != "" {
		return githubrelease.ParseRepository(env)
	}
	remote, err := s.Git.RemoteURL(s.Config.Release.Remote)
	if err != nil {
		return githubrelease.Repository{}, fmt.Errorf("no GitHub repository: set GITHUB_REPOSITORY or add the %s remote", s.Config.Release.Remote)
	}
	return githubrelease.ParseRepository(remote)
}
