package pipeline

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// shippingApp is a workspace with an app that has a Dockerfile.
func shippingApp(t *testing.T, settings map[string]any) (string, protocol.Project) {
	t.Helper()
	ws := t.TempDir()
	_ = os.MkdirAll(filepath.Join(ws, "apps", "api"), 0o755)
	_ = os.WriteFile(filepath.Join(ws, "apps", "api", "Dockerfile"), nil, 0o644)
	api := protocol.Project{Name: "api", Root: "apps/api", Kind: protocol.ProjectApp}
	if settings != nil {
		api.Settings = map[string]map[string]any{"docker": settings}
	}
	return ws, api
}

func mustCI(t *testing.T, ws string, projects ...protocol.Project) protocol.CIResult {
	t.Helper()
	res, err := CI(ws, protocol.CIParams{Projects: projects})
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func TestCIAndDoctor(t *testing.T) {
	ws := t.TempDir()
	api := protocol.Project{Name: "api", Root: "apps/api", Kind: protocol.ProjectApp}
	if res := mustCI(t, ws, api); len(res.Credentials) != 0 {
		t.Fatal("no Dockerfile, no registry")
	}
	ws, api = shippingApp(t, nil)
	if res := mustCI(t, ws, api); len(res.Credentials) != 3 || res.Credentials[0].Env != "CONTAINER_REGISTRY" || res.Credentials[0].Required {
		t.Fatalf("no registry setting keeps CONTAINER_REGISTRY, optional: %+v", res)
	}
	api.Settings = map[string]map[string]any{"docker": {"registry": "ghcr.io/acme"}}
	if res := mustCI(t, ws, api); len(res.Credentials) != 1 || res.Credentials[0].Env != "GITHUB_TOKEN" || len(res.Permissions) != 1 || res.Permissions[0] != "packages: write" {
		t.Fatalf("ghcr.io: %+v", res)
	}
	api.Settings["docker"]["push"] = false
	if res := mustCI(t, ws, api); len(res.Credentials) != 0 || len(res.Permissions) != 0 {
		t.Fatalf("an image not pushed needs nothing: %+v", res)
	}
	real := probe
	defer func() { probe = real }()
	probe = func(string, []string) (string, error) { return "", errors.New("no daemon") }
	if res := Doctor(ws, protocol.DoctorParams{Projects: []protocol.Project{api}}); len(res.Findings) != 1 || res.Findings[0].OK {
		t.Fatalf("%+v", res)
	}
}

func TestEveryRegistrysSecretsReachThePipeline(t *testing.T) {
	ws, api := shippingApp(t, map[string]any{"registries": []any{
		"ghcr.io/open-craw",
		"docker.io/opencraw",
		map[string]any{"url": "registry.acme.dev/team", "username_env": "ACME_USER", "password_env": "ACME_TOKEN"},
	}})
	res := mustCI(t, ws, api)
	var envs []string
	for _, c := range res.Credentials {
		envs = append(envs, c.Env)
		if !c.Required {
			t.Errorf("%s should be required: preflight must name it when missing", c.Env)
		}
	}
	if got := strings.Join(envs, " "); got != "GITHUB_TOKEN DOCKERHUB_USERNAME DOCKERHUB_TOKEN ACME_USER ACME_TOKEN" {
		t.Errorf("credentials %q", got)
	}
	if len(res.Permissions) != 1 || res.Permissions[0] != "packages: write" {
		t.Errorf("permissions %v", res.Permissions)
	}
}

func TestContradictoryRegistriesAreAnErrorAndAFinding(t *testing.T) {
	ws, api := shippingApp(t, map[string]any{"registry": "ghcr.io/a", "registries": []any{"docker.io/b"}})
	if _, err := CI(ws, protocol.CIParams{Projects: []protocol.Project{api}}); err == nil || !strings.Contains(err.Error(), "api:") {
		t.Fatalf("the error names the app: %v", err)
	}
	real := probe
	defer func() { probe = real }()
	probe = func(string, []string) (string, error) { return "", nil }
	res := Doctor(ws, protocol.DoctorParams{Projects: []protocol.Project{api}})
	if len(res.Findings) != 2 || res.Findings[1].OK || !strings.Contains(res.Findings[1].Check, "api") {
		t.Fatalf("%+v", res)
	}
}
