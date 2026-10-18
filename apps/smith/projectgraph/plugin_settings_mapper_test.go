package projectgraph

import (
	"reflect"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

func TestPluginSettingsLayersTheProjectOnTheConvention(t *testing.T) {
	no, yes := false, true
	conv := workspace.ConventionConfig{Registry: "ghcr.io/acme", Push: &yes}
	if got := PluginSettings("docker", conv, workspace.ProjectOverride{}); !reflect.DeepEqual(got, map[string]any{"registry": "ghcr.io/acme", "push": true}) {
		t.Errorf("convention only: %v", got)
	}
	po := workspace.ProjectOverride{Docker: &workspace.DockerOverride{Image: "acme-api", Push: &no}}
	if got := PluginSettings("docker", conv, po); !reflect.DeepEqual(got, map[string]any{"registry": "ghcr.io/acme", "push": false, "image": "acme-api"}) {
		t.Errorf("project on top: %v", got)
	}
	po = workspace.ProjectOverride{Docker: &workspace.DockerOverride{Context: "apps/api/"}}
	if got := PluginSettings("docker", workspace.ConventionConfig{}, po); !reflect.DeepEqual(got, map[string]any{"context": "apps/api"}) {
		t.Errorf("context: %v", got)
	}
	po = workspace.ProjectOverride{Docker: &workspace.DockerOverride{Verify: &yes}}
	if got := PluginSettings("docker", workspace.ConventionConfig{Verify: &no}, po); !reflect.DeepEqual(got, map[string]any{"verify": true}) {
		t.Errorf("verify: %v", got)
	}
	// The app's registries replace the convention's, whichever form each uses.
	list := workspace.ConventionConfig{Registries: []any{"ghcr.io/acme", "docker.io/acme"}, DescriptionFrom: "README.md"}
	if got := PluginSettings("docker", list, workspace.ProjectOverride{}); !reflect.DeepEqual(got, map[string]any{"registries": []any{"ghcr.io/acme", "docker.io/acme"}, "description_from": "README.md"}) {
		t.Errorf("a list: %v", got)
	}
	po = workspace.ProjectOverride{Docker: &workspace.DockerOverride{Registry: "registry.example.com/team"}}
	if got := PluginSettings("docker", list, po); !reflect.DeepEqual(got, map[string]any{"registry": "registry.example.com/team", "description_from": "README.md"}) {
		t.Errorf("one replaces the list: %v", got)
	}
	po = workspace.ProjectOverride{Docker: &workspace.DockerOverride{Registries: []any{"docker.io/acme"}}}
	if got := PluginSettings("docker", conv, po); !reflect.DeepEqual(got, map[string]any{"registries": []any{"docker.io/acme"}, "push": true}) {
		t.Errorf("a list replaces the one: %v", got)
	}
	if got := PluginSettings("docker", workspace.ConventionConfig{}, workspace.ProjectOverride{}); got != nil {
		t.Errorf("nothing set: %v", got)
	}
	if got := PluginSettings("vertical-slices", conv, po); got != nil {
		t.Errorf("other conventions have no settings: %v", got)
	}
}
