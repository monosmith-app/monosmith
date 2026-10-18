package images

import (
	"strings"
	"testing"
)

func noEnv(string) string { return "" }

func TestRegistriesGetTheCredentialsTheirHostImplies(t *testing.T) {
	regs, legacy, err := RegistriesOf(map[string]any{"registries": []any{
		"ghcr.io/open-craw/",
		"docker.io/opencraw",
		"localhost:5000/e2e",
		map[string]any{"url": "registry.acme.dev/team", "username_env": "ACME_USER", "password_env": "ACME_TOKEN"},
	}}, noEnv)
	if err != nil || legacy || len(regs) != 4 {
		t.Fatalf("%+v %v %v", regs, legacy, err)
	}
	want := []struct {
		url   string
		login Login
		user  string
		pass  string
	}{
		{"ghcr.io/open-craw", LoginGitHub, "GITHUB_ACTOR", "GITHUB_TOKEN"},
		{"docker.io/opencraw", LoginDockerHub, "DOCKERHUB_USERNAME", "DOCKERHUB_TOKEN"},
		{"localhost:5000/e2e", LoginNone, "", ""},
		{"registry.acme.dev/team", LoginPassword, "ACME_USER", "ACME_TOKEN"},
	}
	for i, w := range want {
		r := regs[i]
		if r.URL != w.url || r.Login != w.login || r.UserEnv != w.user || r.PasswordEnv != w.pass {
			t.Errorf("registry %d: %+v, want %+v", i, r, w)
		}
	}
	if !regs[3].Explicit || regs[0].Explicit {
		t.Errorf("only the table form is explicit: %+v", regs)
	}
}

func TestOneUnknownHostReadsContainerCredentials(t *testing.T) {
	regs, _, err := RegistriesOf(map[string]any{"registry": "registry.example.com/team"}, noEnv)
	if err != nil || len(regs) != 1 || regs[0].UserEnv != "CONTAINER_USERNAME" || regs[0].PasswordEnv != "CONTAINER_PASSWORD" {
		t.Fatalf("%+v %v", regs, err)
	}
	// With docker.io and ghcr.io beside it, it is still the only one that
	// would read CONTAINER_*.
	regs, _, err = RegistriesOf(map[string]any{"registries": []any{"ghcr.io/a", "docker.io/b", "registry.example.com/team"}}, noEnv)
	if err != nil || regs[2].UserEnv != "CONTAINER_USERNAME" {
		t.Fatalf("%+v %v", regs, err)
	}
}

func TestTwoUnknownHostsMustNameTheirOwnCredentials(t *testing.T) {
	_, _, err := RegistriesOf(map[string]any{"registries": []any{"registry.a.dev/x", "registry.b.dev/y"}}, noEnv)
	if err == nil || !strings.Contains(err.Error(), "registry.a.dev/x and registry.b.dev/y") || !strings.Contains(err.Error(), "[conventions.docker] registries") {
		t.Fatalf("the error names both registries and the setting: %v", err)
	}
	// Naming the variables of one of them is enough.
	if _, _, err := RegistriesOf(map[string]any{"registries": []any{
		"registry.a.dev/x",
		map[string]any{"url": "registry.b.dev/y", "username_env": "B_USER", "password_env": "B_TOKEN"},
	}}, noEnv); err != nil {
		t.Fatal(err)
	}
}

func TestContradictoryRegistrySettingsAreErrors(t *testing.T) {
	for name, s := range map[string]map[string]any{
		"both forms":        {"registry": "ghcr.io/a", "registries": []any{"docker.io/b"}},
		"a duplicate":       {"registries": []any{"ghcr.io/a", "ghcr.io/a/"}},
		"an empty string":   {"registries": []any{""}},
		"a table, no url":   {"registries": []any{map[string]any{"username_env": "U", "password_env": "P"}}},
		"half a credential": {"registries": []any{map[string]any{"url": "registry.a.dev/x", "username_env": "U"}}},
		"not a list":        {"registries": "ghcr.io/a"},
		"a number":          {"registries": []any{3}},
	} {
		if _, _, err := RegistriesOf(s, noEnv); err == nil {
			t.Errorf("%s: no error", name)
		}
	}
}

func TestContainerRegistryIsOnlyReadWhenNothingIsSet(t *testing.T) {
	env := func(k string) string {
		if k == LegacyRegistryEnv {
			return "registry.example.com/team"
		}
		return ""
	}
	regs, legacy, err := RegistriesOf(nil, env)
	if err != nil || !legacy || len(regs) != 1 || regs[0].URL != "registry.example.com/team" {
		t.Fatalf("%+v %v %v", regs, legacy, err)
	}
	regs, legacy, _ = RegistriesOf(map[string]any{"registries": []any{"ghcr.io/a"}}, env)
	if legacy || len(regs) != 1 || regs[0].URL != "ghcr.io/a" {
		t.Errorf("the setting wins: %+v %v", regs, legacy)
	}
}
