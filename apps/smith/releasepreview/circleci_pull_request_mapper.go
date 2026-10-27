package releasepreview

import (
	"net/url"
	"strconv"
	"strings"
)

// circleciPullRequest finds the pull request a CircleCI run builds. CircleCI
// has no comment API of its own: the pull request lives on the repository's
// host, which the pull request's URL (CIRCLE_PULL_REQUEST) names. A GitHub
// repository gets the comment, with a GITHUB_TOKEN the project keeps as an
// environment variable; on any other host the preview stays in the job's log.
func circleciPullRequest(getenv func(string) string) (pr PullRequest, ok bool, why string) {
	raw := getenv("CIRCLE_PULL_REQUEST")
	if raw == "" {
		return pr, false, "this CircleCI run is not for a pull request"
	}
	notGitHub := "CircleCI comments on GitHub pull requests only: " + raw + " is not one, so the preview stays in the log"
	u, err := url.Parse(raw)
	if err != nil {
		return pr, false, notGitHub
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) != 4 || parts[2] != "pull" {
		return pr, false, notGitHub
	}
	n, _ := strconv.Atoi(parts[3])
	api := "https://api.github.com"
	if u.Host != "github.com" {
		api = u.Scheme + "://" + u.Host + "/api/v3" // GitHub Enterprise Server
	}
	token := getenv("GITHUB_TOKEN")
	if token == "" {
		token = getenv("GH_TOKEN")
	}
	pr = PullRequest{Provider: "github", Number: n, Token: token, API: api, Repo: parts[0] + "/" + parts[1]}
	if token == "" {
		return pr, false, "GITHUB_TOKEN is not set (a token that can comment on pull requests, as a project environment variable)"
	}
	return pr, true, ""
}
