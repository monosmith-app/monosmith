package workspace

import (
	"strings"
	"testing"
)

func TestValidateCIAzure(t *testing.T) {
	cases := []struct {
		name string
		toml string
		err  string
	}{
		{"absent", "", ""},
		{"tags off", "build_tags = false", ""},
		{"tags renamed", `build_tags = { api = "api-release" }`, ""},
		{"empty tag name", `build_tags = { api = "" }`, "build_tags.api must be the tag's name"},
		{"tag name not a string", `build_tags = { api = 1 }`, "build_tags.api must be the tag's name"},
		{"tags as a string", `build_tags = "api"`, "build_tags must be true, false or a table"},
		{"tag with a semicolon", `build_tags = { api = "a;b" }`, "must not hold a semicolon"},
		{"deploy variables off", "deploy_variables = false", ""},
		{"pattern", `build_number = "{project}-v{version}"`, ""},
		{"number off", "build_number = false", ""},
		{"number empty", `build_number = ""`, ""},
		{"unknown placeholder", `build_number = "{name}_{version}"`, "unknown placeholder {name}"},
		{"forbidden character", `build_number = "{project}:{version}"`, "Azure Pipelines refuses"},
		{"number as a number", "build_number = 3", "build_number must be a pattern"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := Parse([]byte("[workspace]\nname = \"w\"\n\n[ci.azure]\n" + c.toml + "\n"))
			if c.err == "" && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if c.err != "" && (err == nil || !strings.Contains(err.Error(), c.err)) {
				t.Fatalf("error %v, want one containing %q", err, c.err)
			}
		})
	}
}

func TestAzureSectionValues(t *testing.T) {
	no := false
	cases := []struct {
		name    string
		section AzureSection
		tag     string
		tagOn   bool
		deploy  bool
		pattern string
	}{
		{"defaults", AzureSection{}, "api", true, true, DefaultBuildNumber},
		{"tags off", AzureSection{BuildTags: false}, "api", false, true, DefaultBuildNumber},
		{"tag renamed", AzureSection{BuildTags: map[string]any{"api": "api-release"}}, "api-release", true, true, DefaultBuildNumber},
		{"other project renamed", AzureSection{BuildTags: map[string]any{"app": "app-release"}}, "api", true, true, DefaultBuildNumber},
		{"variables off", AzureSection{DeployVariables: &no}, "api", true, false, DefaultBuildNumber},
		{"number off", AzureSection{BuildNumber: false}, "api", true, true, ""},
		{"number empty", AzureSection{BuildNumber: ""}, "api", true, true, ""},
		{"number pattern", AzureSection{BuildNumber: "{version}"}, "api", true, true, "{version}"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			tag, on := c.section.BuildTag("api")
			if tag != c.tag || on != c.tagOn {
				t.Errorf("BuildTag = %q, %v; want %q, %v", tag, on, c.tag, c.tagOn)
			}
			if got := c.section.DeployVariablesOn(); got != c.deploy {
				t.Errorf("DeployVariablesOn = %v, want %v", got, c.deploy)
			}
			if got := c.section.BuildNumberPattern(); got != c.pattern {
				t.Errorf("BuildNumberPattern = %q, want %q", got, c.pattern)
			}
		})
	}
}
