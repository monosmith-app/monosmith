package images

import (
	"strings"
	"testing"
)

func TestLoginFor(t *testing.T) {
	for registry, want := range map[string]Login{
		"ghcr.io/acme":              LoginGitHub,
		"docker.io/acme":            LoginDockerHub,
		"index.docker.io/acme":      LoginDockerHub,
		"registry.example.com/team": LoginPassword,
		"acme.azurecr.io":           LoginPassword,
		"localhost:5000/e2e":        LoginNone,
		"127.0.0.1:32768/smith-e2e": LoginNone,
		"[::1]:5000/e2e":            LoginNone,
	} {
		if got := LoginFor(registry); got != want {
			t.Errorf("%s: %v, want %v", registry, got, want)
		}
	}
}

// target builds a pushing target from registry URLs, as smith.toml lists them.
func target(t *testing.T, urls ...any) Target {
	t.Helper()
	got, err := TargetOf(projectWith(map[string]any{"registries": urls}), noEnv)
	if err != nil {
		t.Fatal(err)
	}
	return got
}

func TestCredentialsFollowTheRegistries(t *testing.T) {
	creds, perms := Credentials(target(t, "ghcr.io/acme"))
	if len(creds) != 1 || creds[0].Env != "GITHUB_TOKEN" || !creds[0].Required || len(perms) != 1 || perms[0] != "packages: write" {
		t.Errorf("ghcr.io: %+v %v", creds, perms)
	}
	creds, perms = Credentials(target(t, "registry.example.com/team"))
	if len(creds) != 2 || creds[0].Env != "CONTAINER_USERNAME" || !creds[1].Required || perms != nil ||
		!strings.Contains(creds[1].Description, "registry.example.com/team") {
		t.Errorf("another registry: %+v %v", creds, perms)
	}
	if creds, _ := Credentials(target(t, "localhost:5000/e2e")); creds != nil {
		t.Errorf("a local registry needs no login: %+v", creds)
	}
	noPush := target(t, "ghcr.io/acme")
	noPush.Push = false
	if creds, _ := Credentials(noPush); creds != nil {
		t.Errorf("no push, no credentials: %+v", creds)
	}
	creds, _ = Credentials(Target{Legacy: true})
	if len(creds) != 3 || creds[0].Env != LegacyRegistryEnv || creds[0].Required {
		t.Errorf("CONTAINER_REGISTRY keeps its optional credentials: %+v", creds)
	}
}

func TestEveryRegistrysSecretsAreListed(t *testing.T) {
	creds, perms := Credentials(target(t,
		"ghcr.io/open-craw",
		"docker.io/opencraw",
		"docker.io/opencraw-labs",
		map[string]any{"url": "registry.acme.dev/team", "username_env": "ACME_USER", "password_env": "ACME_TOKEN"},
	))
	var envs []string
	for _, c := range creds {
		if !c.Required {
			t.Errorf("%s should be required", c.Env)
		}
		envs = append(envs, c.Env)
	}
	if got := strings.Join(envs, " "); got != "GITHUB_TOKEN DOCKERHUB_USERNAME DOCKERHUB_TOKEN ACME_USER ACME_TOKEN" {
		t.Errorf("secrets %q: a variable two registries share is listed once", got)
	}
	if len(perms) != 1 || perms[0] != "packages: write" {
		t.Errorf("permissions %v", perms)
	}
}
