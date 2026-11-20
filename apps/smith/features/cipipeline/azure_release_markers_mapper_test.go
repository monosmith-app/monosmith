package cipipeline

import (
	"reflect"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/features/release"
	"github.com/monosmith-app/monosmith/apps/smith/features/workspace"
)

func TestAzureReleaseMarkers(t *testing.T) {
	no := false
	plan := release.Plan{Items: []release.Item{{Project: "api", Next: "6.2.0"}, {Project: "app", Next: "6.2.0"}}}
	deployables := []string{"api", "app", "web-admin"}
	cases := []struct {
		name string
		cfg  workspace.AzureSection
		plan release.Plan
		want []string
	}{
		{"defaults", workspace.AzureSection{}, plan, []string{
			"##vso[build.addbuildtag]api",
			"##vso[build.addbuildtag]app",
			"##vso[task.setvariable variable=DEPLOY_API]true",
			"##vso[task.setvariable variable=DEPLOY_APP]true",
			"##vso[task.setvariable variable=DEPLOY_WEB_ADMIN]false",
			"##vso[build.updatebuildnumber]api_6.2.0-app_6.2.0",
		}},
		{"nothing planned", workspace.AzureSection{}, release.Plan{}, []string{
			"##vso[task.setvariable variable=DEPLOY_API]false",
			"##vso[task.setvariable variable=DEPLOY_APP]false",
			"##vso[task.setvariable variable=DEPLOY_WEB_ADMIN]false",
		}},
		{"renamed tag, all else off", workspace.AzureSection{BuildTags: map[string]any{"api": "api-release"}, DeployVariables: &no, BuildNumber: false}, plan, []string{
			"##vso[build.addbuildtag]api-release",
			"##vso[build.addbuildtag]app",
		}},
		{"tags off, own pattern", workspace.AzureSection{BuildTags: false, DeployVariables: &no, BuildNumber: "{project}-v{version}"}, plan, []string{
			"##vso[build.updatebuildnumber]api-v6.2.0-app-v6.2.0",
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := azureReleaseMarkers(c.cfg, c.plan, deployables); !reflect.DeepEqual(got, c.want) {
				t.Errorf("got\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(c.want, "\n"))
			}
		})
	}
}

func TestBuildNumber(t *testing.T) {
	cases := []struct {
		name    string
		pattern string
		items   []release.Item
		want    string
	}{
		{"off", "", []release.Item{{Project: "api", Next: "1.0.0"}}, ""},
		{"no project", "{project}_{version}", nil, ""},
		{"one project", "{project}_{version}", []release.Item{{Project: "api", Next: "1.0.0-beta.2"}}, "api_1.0.0-beta.2"},
		{"refused characters", "{project}_{version}", []release.Item{{Project: "@acme/api", Next: "1.0.0"}}, "_acme_api_1.0.0"},
		{"no trailing dot", "{version}.", []release.Item{{Project: "api", Next: "1.0.0"}}, "1.0.0"},
		{"cut to the limit", "{project}", []release.Item{{Project: strings.Repeat("a", 300), Next: "1.0.0"}}, strings.Repeat("a", buildNumberLimit)},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := buildNumber(c.pattern, release.Plan{Items: c.items}); got != c.want {
				t.Errorf("buildNumber = %q, want %q", got, c.want)
			}
		})
	}
}

func TestDeployVariable(t *testing.T) {
	cases := map[string]string{"api": "DEPLOY_API", "web-admin": "DEPLOY_WEB_ADMIN", "@acme/kit.v2": "DEPLOY__ACME_KIT_V2"}
	for in, want := range cases {
		if got := deployVariable(in); got != want {
			t.Errorf("deployVariable(%q) = %q, want %q", in, got, want)
		}
	}
}
