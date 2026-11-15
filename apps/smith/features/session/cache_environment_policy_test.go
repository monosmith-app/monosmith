package session

import (
	"reflect"
	"testing"
)

func env(values map[string]string) func(string) string {
	return func(name string) string { return values[name] }
}

func TestUnexpandedMacrosAreNoValue(t *testing.T) {
	got := unexpandedMacros(env(map[string]string{
		"SMITH_CACHE_KEY":         "$(SMITH_CACHE_KEY)",
		"AZURE_STORAGE_SAS_TOKEN": "sv=2024&sig=abc",
		"AWS_REGION":              "$(AWS_REGION)",
		"SMITH_CACHE_TOKEN":       "",
	}), cacheEnvironment)
	if want := []string{"SMITH_CACHE_KEY", "AWS_REGION"}; !reflect.DeepEqual(got, want) {
		t.Errorf("unexpanded = %v, want %v", got, want)
	}
}

func TestARealValueThatContainsADollarIsKept(t *testing.T) {
	got := unexpandedMacros(env(map[string]string{"SMITH_CACHE_KEY": "pass$(word"}), cacheEnvironment)
	if len(got) != 0 {
		t.Errorf("unexpanded = %v, want none", got)
	}
}

func TestPullRequestsAreReadOnlyOnEveryProvider(t *testing.T) {
	cases := map[string]map[string]string{
		"github pull request":        {"GITHUB_EVENT_NAME": "pull_request"},
		"github pull_request_target": {"GITHUB_EVENT_NAME": "pull_request_target"},
		"azure pull request":         {"BUILD_REASON": "PullRequest"},
		"gitlab merge request":       {"CI_PIPELINE_SOURCE": "merge_request_event"},
		"circleci pull request":      {"CIRCLE_PULL_REQUEST": "https://github.com/o/r/pull/3"},
		"circleci fork pull request": {"CIRCLE_PR_NUMBER": "3"},
		"bitbucket pull request":     {"BITBUCKET_PR_ID": "3"},
	}
	for name, values := range cases {
		if !pullRequest(env(values)) {
			t.Errorf("%s is writable", name)
		}
	}
	for name, values := range map[string]map[string]string{
		"github push":      {"GITHUB_EVENT_NAME": "push"},
		"azure ci":         {"BUILD_REASON": "IndividualCI"},
		"gitlab pipeline":  {"CI_PIPELINE_SOURCE": "push"},
		"circleci branch":  {"CIRCLECI": "true", "CIRCLE_BRANCH": "main"},
		"bitbucket branch": {"BITBUCKET_BUILD_NUMBER": "9", "BITBUCKET_BRANCH": "main"},
		"a laptop":         {},
	} {
		if pullRequest(env(values)) {
			t.Errorf("%s is read-only", name)
		}
	}
}
