package images

import (
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// projectWith is an app whose docker settings are s.
func projectWith(s map[string]any) protocol.Project {
	return protocol.Project{Name: "api", Kind: protocol.ProjectApp, Settings: map[string]map[string]any{"docker": s}}
}

func mustTarget(t *testing.T, p protocol.Project, env func(string) string) Target {
	t.Helper()
	got, err := TargetOf(p, env)
	if err != nil {
		t.Fatal(err)
	}
	return got
}

func TestTargetOf(t *testing.T) {
	none := func(string) string { return "" }
	legacy := func(k string) string {
		if k == LegacyRegistryEnv {
			return "registry.example.com/team/"
		}
		return ""
	}
	api := protocol.Project{Name: "API", Kind: protocol.ProjectApp}
	if got := mustTarget(t, api, none); got.Push || len(got.Registries) != 0 || got.Image != "api" || got.LocalReference() != "api" {
		t.Errorf("no registry builds without pushing: %+v", got)
	}
	if got := mustTarget(t, api, legacy); !got.Push || !got.Legacy || got.LocalReference() != "registry.example.com/team/api" {
		t.Errorf("CONTAINER_REGISTRY still names one: %+v", got)
	}
	api.Settings = map[string]map[string]any{"docker": {"registry": "ghcr.io/acme/", "image": "acme-api"}}
	if got := mustTarget(t, api, legacy); !got.Push || got.Legacy || got.LocalReference() != "ghcr.io/acme/acme-api" || got.Where() != "ghcr.io/acme" {
		t.Errorf("the setting wins over the env: %+v", got)
	}
	api.Settings["docker"]["push"] = false
	if got := mustTarget(t, api, none); got.Push || !PushOff(api) || !Configured(api) {
		t.Errorf("push = false: %+v", got)
	}
}

func TestSeveralRegistriesShareTheImageName(t *testing.T) {
	api := protocol.Project{Name: "api", Kind: protocol.ProjectApp, Settings: map[string]map[string]any{"docker": {
		"registries": []any{"ghcr.io/open-craw", "docker.io/opencraw"}, "image": "opencraw-api", "description_from": "apps/api/README.md",
	}}}
	got := mustTarget(t, api, func(string) string { return "" })
	if !got.Push || len(got.Registries) != 2 || got.Where() != "ghcr.io/open-craw, docker.io/opencraw" || !Configured(api) {
		t.Fatalf("%+v", got)
	}
	if got.Reference(got.Registries[1]) != "docker.io/opencraw/opencraw-api" || got.LocalReference() != "ghcr.io/open-craw/opencraw-api" {
		t.Errorf("references %q, local %q", got.Reference(got.Registries[1]), got.LocalReference())
	}
	if got.DescriptionFrom != "apps/api/README.md" {
		t.Errorf("description_from %q", got.DescriptionFrom)
	}
}
