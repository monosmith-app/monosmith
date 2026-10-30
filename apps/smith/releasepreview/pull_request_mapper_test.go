package releasepreview

import (
	"strings"
	"testing"
)

func env(vars map[string]string) func(string) string {
	return func(k string) string { return vars[k] }
}

func TestFromEnvironment(t *testing.T) {
	pr, ok, _ := FromEnvironment(env(map[string]string{
		"GITHUB_ACTIONS": "true", "GITHUB_EVENT_NAME": "pull_request", "GITHUB_REF": "refs/pull/42/merge",
		"GITHUB_TOKEN": "t", "GITHUB_REPOSITORY": "acme/repo",
	}))
	if !ok || pr.Provider != "github" || pr.Number != 42 || pr.API != "https://api.github.com" || pr.Repo != "acme/repo" {
		t.Fatalf("github: %+v %v", pr, ok)
	}
	pr, ok, _ = FromEnvironment(env(map[string]string{
		"TF_BUILD": "True", "BUILD_REASON": "PullRequest", "BUILD_REPOSITORY_PROVIDER": "TfsGit",
		"SYSTEM_PULLREQUEST_PULLREQUESTID": "7", "SYSTEM_COLLECTIONURI": "https://dev.azure.com/acme",
		"SYSTEM_TEAMPROJECT": "Platform", "BUILD_REPOSITORY_ID": "r1", "SYSTEM_ACCESSTOKEN": "a",
	}))
	if !ok || pr.Provider != "azure" || pr.Number != 7 || pr.Collection != "https://dev.azure.com/acme/" {
		t.Fatalf("azure: %+v %v", pr, ok)
	}
	for want, vars := range map[string]map[string]string{
		"not for a pull request":  {"GITHUB_ACTIONS": "true", "GITHUB_EVENT_NAME": "push", "GITHUB_REF": "refs/heads/main"},
		"GITHUB_TOKEN is not set": {"GITHUB_ACTIONS": "true", "GITHUB_EVENT_NAME": "pull_request", "GITHUB_REF": "refs/pull/1/merge"},
		"not Azure Repos":         {"TF_BUILD": "True", "BUILD_REASON": "PullRequest", "BUILD_REPOSITORY_PROVIDER": "GitHub"},
		"not running in":          {},
	} {
		if _, ok, why := FromEnvironment(env(vars)); ok || !strings.Contains(why, want) {
			t.Errorf("%v: ok=%v why=%q", vars, ok, why)
		}
	}
}

func TestBitbucketPullRequest(t *testing.T) {
	vars := map[string]string{"BITBUCKET_BUILD_NUMBER": "9", "BITBUCKET_WORKSPACE": "acme", "BITBUCKET_REPO_SLUG": "repo"}
	if _, ok, why := FromEnvironment(env(vars)); ok || !strings.Contains(why, "not for a pull request") {
		t.Fatalf("a branch build: %v %s", ok, why)
	}
	vars["BITBUCKET_PR_ID"] = "12"
	if _, ok, why := FromEnvironment(env(vars)); ok || !strings.Contains(why, "BITBUCKET_TOKEN") {
		t.Fatalf("without a token: %v %s", ok, why)
	}
	vars["BITBUCKET_TOKEN"] = "t"
	pr, ok, _ := FromEnvironment(env(vars))
	if !ok || pr.Provider != "bitbucket" || pr.Number != 12 || pr.Repo != "acme/repo" || pr.API != "https://api.bitbucket.org/2.0" {
		t.Fatalf("%+v", pr)
	}
}

func TestGitLabMergeRequest(t *testing.T) {
	env := map[string]string{"GITLAB_CI": "true", "CI_PIPELINE_SOURCE": "merge_request_event", "CI_MERGE_REQUEST_IID": "7", "CI_API_V4_URL": "https://gitlab.example/api/v4", "CI_PROJECT_ID": "42"}
	if _, ok, why := FromEnvironment(func(k string) string { return env[k] }); ok || !strings.Contains(why, "GITLAB_TOKEN") {
		t.Fatalf("without a token: %v %s", ok, why)
	}
	env["GITLAB_TOKEN"] = "t"
	pr, ok, _ := FromEnvironment(func(k string) string { return env[k] })
	if !ok || pr.Provider != "gitlab" || pr.Number != 7 || pr.ProjectID != "42" || pr.API != "https://gitlab.example/api/v4" {
		t.Fatalf("%+v", pr)
	}
	env["CI_PIPELINE_SOURCE"] = "push"
	if _, ok, _ := FromEnvironment(func(k string) string { return env[k] }); ok {
		t.Fatal("a push pipeline has no merge request")
	}
}
