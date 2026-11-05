package releaserehearsal

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/githubrelease"
)

// CheckGitHubRelease is the rehearsal of packages with registry =
// "github-release": no registry is asked, only GitHub, and with read-only
// requests. It records one result for the token (is there one, from the
// environment or gh) and one for access (the repository states that the token
// may write, which is what creating a release and uploading assets need:
// contents: write). Nothing is created and the token is never printed.
//
// Parameters: report receives the results; projects are the packages that
// attach to GitHub Releases; repo is the repository; checker reads the
// repository's permissions for the token; tokenErr is the error of looking for
// the token, nil when there is one.
func CheckGitHubRelease(report *Report, projects []string, repo githubrelease.Repository, checker interface {
	CanPush(githubrelease.Repository) (bool, error)
}, tokenErr error) {
	if len(projects) == 0 {
		return
	}
	where := "github.com/" + repo.String()
	add := func(check, status, message, fix string) {
		report.Add(Result{Project: strings.Join(projects, ", "), Registry: where, Check: check, Status: status, Message: message, Fix: fix})
	}
	if tokenErr != nil {
		add(CheckCredential, Fail, "no GitHub token: "+tokenErr.Error(), "set GITHUB_TOKEN (the generated workflow does, with permissions: contents: write), or sign in with gh auth login")
		return
	}
	add(CheckCredential, Pass, "a GitHub token is set (GITHUB_TOKEN, GH_TOKEN or gh)", "")
	ok, err := checker.CanPush(repo)
	switch {
	case err != nil:
		add(CheckAccess, Fail, fmt.Sprintf("GitHub did not show %s to the token: %v", repo, err), "the token must reach the repository; for a private repository GitHub answers 404 to a token without access")
	case !ok:
		add(CheckAccess, Fail, fmt.Sprintf("the token cannot write to %s, so it cannot create releases or upload assets", repo), "give the token contents: write (a workflow: permissions: contents: write; a fine-grained token: Contents, Read and write)")
	default:
		add(CheckAccess, Pass, fmt.Sprintf("the token may write to %s: it can create the releases and attach the package files", repo), "")
	}
}
