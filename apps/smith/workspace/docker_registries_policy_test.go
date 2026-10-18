package workspace

import (
	"reflect"
	"strings"
	"testing"
)

func TestDockerRegistriesAreParsedAsListedAndNormalized(t *testing.T) {
	cfg, err := Parse([]byte(`[workspace]
name = "a"
[conventions.docker]
registries = [
  "ghcr.io/open-craw/",
  "docker.io/opencraw",
  { url = "registry.acme.dev/team/", username_env = "ACME_USER", password_env = "ACME_TOKEN" },
]
description_from = "README.md"
`))
	if err != nil {
		t.Fatal(err)
	}
	want := []any{
		"ghcr.io/open-craw",
		"docker.io/opencraw",
		map[string]any{"url": "registry.acme.dev/team", "username_env": "ACME_USER", "password_env": "ACME_TOKEN"},
	}
	if got := cfg.Conventions["docker"].Registries; !reflect.DeepEqual(got, want) {
		t.Errorf("registries %#v, want %#v", got, want)
	}
}

func TestAnAppsRegistriesAreChecked(t *testing.T) {
	cfg, err := Parse([]byte(`[workspace]
name = "a"
[conventions.docker]
registry = "ghcr.io/a"
[projects.api.docker]
registries = ["docker.io/b", "ghcr.io/a"]
`))
	if err != nil {
		t.Fatal(err)
	}
	if got := cfg.Projects["api"].Docker.Registries; len(got) != 2 {
		t.Errorf("the app's list, beside the convention's registry: %v", got)
	}
}

func TestDockerRegistryMistakesAreErrors(t *testing.T) {
	for name, tc := range map[string]struct{ toml, want string }{
		"both forms":      {"[conventions.docker]\nregistry = \"ghcr.io/a\"\nregistries = [\"docker.io/b\"]\n", "sets both registry and registries"},
		"both, per app":   {"[conventions.docker]\n[projects.api.docker]\nregistry = \"ghcr.io/a\"\nregistries = [\"docker.io/b\"]\n", "[projects.api.docker] sets both"},
		"a scheme":        {"[conventions.docker]\nregistries = [\"https://ghcr.io/a\"]\n", "must not start with a scheme"},
		"a Hub namespace": {"[conventions.docker]\nregistries = [\"opencraw\"]\n", "write docker.io/opencraw for Docker Hub"},
		"uppercase":       {"[conventions.docker]\nregistries = [\"ghcr.io/OpenCraw\"]\n", "must be lowercase"},
		"a duplicate":     {"[conventions.docker]\nregistries = [\"ghcr.io/a\", \"ghcr.io/a/\"]\n", "lists ghcr.io/a twice"},
		"empty":           {"[conventions.docker]\nregistries = []\n", "is empty"},
		"no url":          {"[conventions.docker]\nregistries = [{ username_env = \"U\", password_env = \"P\" }]\n", "must be host/namespace"},
		"half a pair":     {"[conventions.docker]\nregistries = [{ url = \"r.dev/x\", username_env = \"U\" }]\n", "names one of username_env and password_env"},
		"a bad variable":  {"[conventions.docker]\nregistries = [{ url = \"r.dev/x\", username_env = \"my user\", password_env = \"P\" }]\n", "is not an environment variable name"},
		"an unknown key":  {"[conventions.docker]\nregistries = [{ url = \"r.dev/x\", token = \"t\" }]\n", "unknown key \"token\""},
		"a number":        {"[conventions.docker]\nregistries = [3]\n", "must be a \"host/namespace\" string"},
		"a path out":      {"[conventions.docker]\ndescription_from = \"../README.md\"\n", "workspace-relative file"},
		"another convention": {
			"[conventions.vertical-slices]\nregistries = [\"ghcr.io/a\"]\n", "settings of the docker convention",
		},
	} {
		_, err := Parse([]byte("[workspace]\nname = \"a\"\n" + tc.toml))
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: want error containing %q, got %v", name, tc.want, err)
		}
	}
}

func TestValidRegistryURL(t *testing.T) {
	for _, ok := range []string{"ghcr.io/acme", "docker.io/acme", "localhost:5000/e2e", "127.0.0.1:5000", "acme.azurecr.io/team/app"} {
		if why := ValidRegistryURL(ok); why != "" {
			t.Errorf("%s: %s", ok, why)
		}
	}
	for _, bad := range []string{"", "acme", "ghcr.io/a b", " ghcr.io/a", "ghcr.io/a?x=1", "ghcr.io/../b"} {
		if ValidRegistryURL(bad) == "" {
			t.Errorf("%q should be rejected", bad)
		}
	}
}
