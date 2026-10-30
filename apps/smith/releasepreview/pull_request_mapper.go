package releasepreview

import (
	"regexp"
	"strconv"
	"strings"
)

// Marker identifies smith's comment, so a later run updates it.
const Marker = "<!-- smith:release-preview -->"

// PullRequest is where the preview goes.
type PullRequest struct {
	// Provider is "github", "azure", "gitlab" or "bitbucket".
	Provider string
	Number   int
	Token    string
	// GitHub: the REST root and "owner/name".
	API, Repo string
	// Azure: the collection URL, team project and repository id.
	Collection, Project, RepoID string
	// GitLab: the API root (API) and the project id (ProjectID).
	ProjectID string
	// Bitbucket: the API root (API) and "workspace/repository" (Repo).
}

var githubPullRef = regexp.MustCompile(`^refs/pull/(\d+)/`)

// FromEnvironment finds the pull request a CI run builds, from the variables
// GitHub Actions and Azure Pipelines set. ok is false outside a pull request
// build; why says what is missing.
func FromEnvironment(getenv func(string) string) (pr PullRequest, ok bool, why string) {
	switch {
	case getenv("GITHUB_ACTIONS") == "true":
		event := getenv("GITHUB_EVENT_NAME")
		m := githubPullRef.FindStringSubmatch(getenv("GITHUB_REF"))
		if !strings.HasPrefix(event, "pull_request") || m == nil {
			return pr, false, "this GitHub Actions run is not for a pull request"
		}
		n, _ := strconv.Atoi(m[1])
		token := getenv("GITHUB_TOKEN")
		if token == "" {
			token = getenv("GH_TOKEN")
		}
		api := strings.TrimSuffix(getenv("GITHUB_API_URL"), "/")
		if api == "" {
			api = "https://api.github.com"
		}
		pr = PullRequest{Provider: "github", Number: n, Token: token, API: api, Repo: getenv("GITHUB_REPOSITORY")}
		if token == "" {
			return pr, false, "GITHUB_TOKEN is not set for the step (it needs pull-requests: write)"
		}
		return pr, true, ""
	case strings.EqualFold(getenv("TF_BUILD"), "true"):
		if getenv("BUILD_REASON") != "PullRequest" {
			return pr, false, "this Azure Pipelines run is not for a pull request"
		}
		if provider := getenv("BUILD_REPOSITORY_PROVIDER"); provider != "" && provider != "TfsGit" {
			return pr, false, "the repository is on " + provider + ", not Azure Repos"
		}
		n, _ := strconv.Atoi(getenv("SYSTEM_PULLREQUEST_PULLREQUESTID"))
		pr = PullRequest{
			Provider: "azure", Number: n, Token: getenv("SYSTEM_ACCESSTOKEN"),
			Collection: strings.TrimSuffix(getenv("SYSTEM_COLLECTIONURI"), "/") + "/", Project: getenv("SYSTEM_TEAMPROJECT"), RepoID: getenv("BUILD_REPOSITORY_ID"),
		}
		if pr.Token == "" {
			return pr, false, "SYSTEM_ACCESSTOKEN is not set for the step (env: SYSTEM_ACCESSTOKEN: $(System.AccessToken))"
		}
		if n == 0 {
			return pr, false, "SYSTEM_PULLREQUEST_PULLREQUESTID is not set"
		}
		return pr, true, ""
	case getenv("GITLAB_CI") == "true":
		if getenv("CI_PIPELINE_SOURCE") != "merge_request_event" {
			return pr, false, "this GitLab CI pipeline is not for a merge request"
		}
		n, _ := strconv.Atoi(getenv("CI_MERGE_REQUEST_IID"))
		pr = PullRequest{Provider: "gitlab", Number: n, Token: getenv("GITLAB_TOKEN"), API: strings.TrimSuffix(getenv("CI_API_V4_URL"), "/"), ProjectID: getenv("CI_PROJECT_ID")}
		if pr.Token == "" {
			return pr, false, "GITLAB_TOKEN is not set (a project access token with the api scope, as a masked CI/CD variable)"
		}
		return pr, true, ""
	case getenv("BITBUCKET_BUILD_NUMBER") != "":
		if getenv("BITBUCKET_PR_ID") == "" {
			return pr, false, "this Bitbucket Pipelines run is not for a pull request"
		}
		n, _ := strconv.Atoi(getenv("BITBUCKET_PR_ID"))
		repo := getenv("BITBUCKET_WORKSPACE") + "/" + getenv("BITBUCKET_REPO_SLUG")
		pr = PullRequest{Provider: "bitbucket", Number: n, Token: getenv("BITBUCKET_TOKEN"), API: "https://api.bitbucket.org/2.0", Repo: repo}
		if pr.Token == "" {
			return pr, false, "BITBUCKET_TOKEN is not set (a repository access token with pull request write, as a secured repository variable)"
		}
		return pr, true, ""
	case getenv("CIRCLECI") == "true":
		return circleciPullRequest(getenv)
	}
	return pr, false, "not running in GitHub Actions, Azure Pipelines, GitLab CI, Bitbucket Pipelines or CircleCI"
}
