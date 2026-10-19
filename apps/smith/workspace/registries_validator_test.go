package workspace

import (
	"strings"
	"testing"
)

func validConfig() Config {
	return Config{Workspace: WorkspaceSection{Name: "w"}}
}

func TestRegistriesAreValidated(t *testing.T) {
	cases := map[string]struct {
		registries map[string]RegistryConfig
		install    string
		want       string
	}{
		"no auth":                   {map[string]RegistryConfig{"f": {NPM: "https://f/npm/"}}, "", "needs auth"},
		"unknown auth":              {map[string]RegistryConfig{"f": {Auth: "basic", NPM: "https://f/npm/"}}, "", "unknown auth"},
		"token no env":              {map[string]RegistryConfig{"f": {Auth: "token", NPM: "https://f/npm/"}}, "", "needs token_env"},
		"env not token":             {map[string]RegistryConfig{"f": {Auth: "azure-artifacts", TokenEnv: "X", NPM: "https://f/npm/"}}, "", "only for auth"},
		"no url":                    {map[string]RegistryConfig{"f": {Auth: "azure-artifacts"}}, "", "no npm, pypi, nuget or maven"},
		"http url":                  {map[string]RegistryConfig{"f": {Auth: "azure-artifacts", NPM: "http://f/npm/"}}, "", "https URL"},
		"github pypi":               {map[string]RegistryConfig{"f": {Auth: "github-packages", PyPI: "https://f/simple/"}}, "", "no Python registry"},
		"unknown name":              {map[string]RegistryConfig{"f": {Auth: "azure-artifacts", NPM: "https://f/npm/"}}, "g", `names no [registries.g]`},
		"bad name":                  {map[string]RegistryConfig{"my feed": {Auth: "azure-artifacts", NPM: "https://f/npm/"}}, "", "must start with a letter"},
		"valid, used":               {map[string]RegistryConfig{"f": {Auth: "token", TokenEnv: "FEED_TOKEN", NPM: "https://f/npm/"}}, "f", ""},
		"maven github":              {map[string]RegistryConfig{"gh": {Auth: "github-packages", Maven: "github"}}, "gh", ""},
		"maven url":                 {map[string]RegistryConfig{"gh": {Auth: "github-packages", Maven: "https://maven.pkg.github.com/acme/tools"}}, "gh", ""},
		"maven http":                {map[string]RegistryConfig{"f": {Auth: "token", TokenEnv: "T", Maven: "http://f/maven"}}, "", "https URL"},
		"maven shorthand elsewhere": {map[string]RegistryConfig{"f": {Auth: "token", TokenEnv: "T", Maven: "github"}}, "", "needs auth = \"github-packages\""},
		"nothing at all":            {nil, "", ""},
	}
	for name, c := range cases {
		cfg := validConfig()
		cfg.Registries = c.registries
		cfg.Workspace.Registry = c.install
		err := Validate(&cfg)
		if c.want == "" && err != nil {
			t.Errorf("%s: %v", name, err)
		}
		if c.want != "" && (err == nil || !strings.Contains(err.Error(), c.want)) {
			t.Errorf("%s: err = %v, want %q", name, err, c.want)
		}
	}
}

func TestPyPIUploadDefaultsToAzureArtifactsSpelling(t *testing.T) {
	cfg := validConfig()
	cfg.Registries = map[string]RegistryConfig{"f": {Auth: "azure-artifacts", PyPI: "https://pkgs.dev.azure.com/acme/_packaging/libs/pypi/simple/"}}
	if err := Validate(&cfg); err != nil {
		t.Fatal(err)
	}
	if got := cfg.Registries["f"].PyPIUpload; got != "https://pkgs.dev.azure.com/acme/_packaging/libs/pypi/upload/" {
		t.Errorf("upload = %q", got)
	}
}

func TestAReleaseOrProjectRegistryMustExist(t *testing.T) {
	cfg := validConfig()
	cfg.Registries = map[string]RegistryConfig{"f": {Auth: "azure-artifacts", NPM: "https://f/npm/"}}
	cfg.Projects = map[string]ProjectOverride{"lib": {Registry: "nope"}}
	if err := Validate(&cfg); err == nil || !strings.Contains(err.Error(), "[projects.lib] registry") {
		t.Errorf("err = %v", err)
	}
	cfg.Projects = nil
	cfg.Release.Registry = "nope"
	if err := Validate(&cfg); err == nil || !strings.Contains(err.Error(), "[release] registry") {
		t.Errorf("err = %v", err)
	}
}
