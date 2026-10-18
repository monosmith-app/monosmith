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

func TestTheDescriptionStepRunsAfterTheReleaseOnlyForADockerHubDescription(t *testing.T) {
	cases := map[string]struct {
		settings map[string]any
		want     int
	}{
		"docker.io with description_from": {map[string]any{"registries": []any{"docker.io/opencraw"}, "description_from": "README.md"}, 1},
		"also pushed to ghcr.io":          {map[string]any{"registries": []any{"ghcr.io/open-craw", "docker.io/opencraw"}, "description_from": "README.md"}, 1},
		"no description_from":             {map[string]any{"registries": []any{"docker.io/opencraw"}}, 0},
		"description_from, only ghcr.io":  {map[string]any{"registry": "ghcr.io/open-craw", "description_from": "README.md"}, 0},
		"description_from, no registry":   {map[string]any{"description_from": "README.md"}, 0},
	}
	for name, c := range cases {
		ws, api := shippingApp(t, c.settings)
		res := mustCI(t, ws, api)
		if len(res.PostRelease) != c.want {
			t.Errorf("%s: post-release steps %+v, want %d", name, res.PostRelease, c.want)
		}
		if c.want == 1 {
			step := res.PostRelease[0]
			if strings.Join(step.Command, " ") != "{smith} docker description" || step.Blocking {
				t.Errorf("%s: step %+v runs smith docker description and only warns", name, step)
			}
		}
	}
	// Two apps with a description share one step.
	ws, api := shippingApp(t, map[string]any{"registries": []any{"docker.io/opencraw"}, "description_from": "README.md"})
	other := api
	other.Name, other.Root = "worker", "apps/worker"
	_ = os.MkdirAll(filepath.Join(ws, "apps", "worker"), 0o755)
	_ = os.WriteFile(filepath.Join(ws, "apps", "worker", "Dockerfile"), nil, 0o644)
	if res := mustCI(t, ws, api, other); len(res.PostRelease) != 1 {
		t.Errorf("one step for the workspace: %+v", res.PostRelease)
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

func TestAShortDescriptionAloneAddsTheDescriptionStep(t *testing.T) {
	ws, api := shippingApp(t, map[string]any{"registries": []any{"docker.io/opencraw"}})
	if res := mustCI(t, ws, api); len(res.PostRelease) != 0 {
		t.Fatalf("no description anywhere: %+v", res.PostRelease)
	}
	api.Description = "What the API does."
	if res := mustCI(t, ws, api); len(res.PostRelease) != 1 {
		t.Errorf("the project's own description is sent: %+v", res.PostRelease)
	}
	ws, api = shippingApp(t, map[string]any{"registries": []any{"ghcr.io/open-craw"}, "short_description": "x"})
	if res := mustCI(t, ws, api); len(res.PostRelease) != 0 {
		t.Errorf("only docker.io has a short description: %+v", res.PostRelease)
	}
}

func TestDoctorReportsAShortDescriptionOverTheLimit(t *testing.T) {
	real := probe
	defer func() { probe = real }()
	probe = func(string, []string) (string, error) { return "", nil }
	ws, api := shippingApp(t, map[string]any{"registries": []any{"docker.io/opencraw"}, "short_description": strings.Repeat("a", 101)})
	res := Doctor(ws, protocol.DoctorParams{Projects: []protocol.Project{api}})
	if len(res.Findings) != 2 || res.Findings[1].OK || !strings.Contains(res.Findings[1].Check, "short description") || !strings.Contains(res.Findings[1].Detail, "100") {
		t.Fatalf("%+v", res)
	}
	api.Settings["docker"]["short_description"] = strings.Repeat("é", 100)
	if res := Doctor(ws, protocol.DoctorParams{Projects: []protocol.Project{api}}); len(res.Findings) != 1 {
		t.Errorf("100 characters fit, bytes do not count: %+v", res)
	}
	// A long project description is cut, not reported: it is not the setting.
	api.Settings["docker"]["short_description"] = ""
	api.Description = strings.Repeat("a", 127)
	if res := Doctor(ws, protocol.DoctorParams{Projects: []protocol.Project{api}}); len(res.Findings) != 1 {
		t.Errorf("%+v", res)
	}
}
