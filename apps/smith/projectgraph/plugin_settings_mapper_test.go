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
	if got := PluginSettings("docker", workspace.ConventionConfig{}, workspace.ProjectOverride{}); got != nil {
		t.Errorf("nothing set: %v", got)
	}
	if got := PluginSettings("vertical-slices", conv, po); got != nil {
		t.Errorf("other conventions have no settings: %v", got)
	}
}
