package doctor

import (
	"github.com/monosmith-app/monosmith/apps/smith/doctorchecks"
	"os"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/ciannotations"
	"github.com/monosmith-app/monosmith/apps/smith/githubrelease"
	"github.com/monosmith-app/monosmith/apps/smith/pluginmarket"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// githubToken finds the GitHub login (GITHUB_TOKEN, GH_TOKEN or `gh auth
// token`); tests replace it.
var githubToken = pluginmarket.GitHubToken

// notStartedAnnotation reads the "job was not started" annotation of the
// latest default-branch run; tests replace it with a fake GitHub.
var notStartedAnnotation = func(repo githubrelease.Repository, token string) (string, error) {
	return ciannotations.NotStarted(nil, repo.Owner, repo.Name, token)
}

// notStartedChecks warns when the latest run of the default branch has a
// check-run annotation saying GitHub did not start a job because of billing.
// Such a run has no logs, so the annotation is the only clue. It only reports:
// it says nothing without a GitHub remote, a GitHub login or a network, and
// SMITH_NO_UPDATE_CHECK turns the lookup off. It is a warning, so it never
// fails the local checks.
func notStartedChecks(s *session.Session) []protocol.Finding {
	if os.Getenv("SMITH_NO_UPDATE_CHECK") != "" || s.Git == nil {
		return nil
	}
	remote, err := s.Git.RemoteURL(s.Config.Release.Remote)
	if err != nil || !strings.Contains(remote, "github.com") {
		return nil
	}
	repo, err := githubrelease.ParseRepository(remote)
	if err != nil {
		return nil
	}
	token := githubToken()
	if token == "" {
		return nil
	}
	message, err := notStartedAnnotation(repo, token)
	if err != nil || message == "" {
		return nil
	}
	return []protocol.Finding{{
		Check:   "the latest CI run on the default branch started its jobs",
		Warning: true,
		Area:    doctorchecks.AreaCI,
		Problem: "GitHub did not start the jobs of the latest CI run: " + message,
		Detail:  "GitHub did not start a job, so the run has no logs. GitHub says: " + message,
		Remedy:  "open the GitHub billing page of the account that owns " + repo.String() + ", check the payment method and the spending limit, then run the failed jobs again",
	}}
}
