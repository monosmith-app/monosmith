package releasepreview

import (
	"strings"
	"testing"
)

func TestCircleCIPullRequestsOnGitHub(t *testing.T) {
	vars := map[string]string{"CIRCLECI": "true", "CIRCLE_PULL_REQUEST": "https://github.com/acme/repo/pull/42"}
	get := func(k string) string { return vars[k] }
	if _, ok, why := FromEnvironment(get); ok || !strings.Contains(why, "GITHUB_TOKEN") {
		t.Fatalf("without a token: %v %s", ok, why)
	}
	vars["GITHUB_TOKEN"] = "t"
	pr, ok, _ := FromEnvironment(get)
	if !ok || pr.Provider != "github" || pr.Number != 42 || pr.Repo != "acme/repo" || pr.API != "https://api.github.com" {
		t.Fatalf("%+v", pr)
	}
	vars["CIRCLE_PULL_REQUEST"] = "https://ghe.example.com/acme/repo/pull/7"
	if pr, _, _ := FromEnvironment(get); pr.API != "https://ghe.example.com/api/v3" || pr.Number != 7 {
		t.Fatalf("enterprise: %+v", pr)
	}
	vars["CIRCLE_PULL_REQUEST"] = "https://bitbucket.org/acme/repo/pull-requests/3"
	if _, ok, why := FromEnvironment(get); ok || !strings.Contains(why, "stays in the log") {
		t.Fatalf("a Bitbucket pull request: %v %s", ok, why)
	}
	delete(vars, "CIRCLE_PULL_REQUEST")
	if _, ok, why := FromEnvironment(get); ok || !strings.Contains(why, "not for a pull request") {
		t.Fatalf("a branch build: %v %s", ok, why)
	}
}
