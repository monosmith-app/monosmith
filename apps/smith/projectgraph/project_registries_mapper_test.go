package projectgraph

import (
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

func TestProjectRegistriesFollowTheirPrecedence(t *testing.T) {
	cfg := workspace.Config{
		Workspace:  workspace.WorkspaceSection{Registry: "feed"},
		Release:    workspace.ReleaseSection{Registry: "releases"},
		Projects:   map[string]workspace.ProjectOverride{"special": {Registry: "other"}},
		Registries: map[string]workspace.RegistryConfig{"feed": {Auth: "azure-artifacts", NPM: "https://a/npm/"}, "releases": {Auth: "token", TokenEnv: "T", NPM: "https://r/npm/"}, "other": {Auth: "github-packages", NPM: "https://npm.pkg.github.com/"}},
	}
	r := ProjectRegistries(cfg, "lib")
	if r.Install.Name != "feed" || r.Publish.Name != "releases" || r.Publish.TokenEnv != "T" {
		t.Errorf("lib: %+v %+v", r.Install, r.Publish)
	}
	if r := ProjectRegistries(cfg, "special"); r.Publish.Name != "other" || r.Install.Name != "feed" {
		t.Errorf("special: %+v %+v", r.Install, r.Publish)
	}
	cfg.Release.Registry = ""
	if r := ProjectRegistries(cfg, "lib"); r.Publish.Name != "feed" {
		t.Errorf("publishing defaults to the install registry: %+v", r.Publish)
	}
	if r := ProjectRegistries(workspace.Config{}, "lib"); r != nil {
		t.Errorf("no registries named: %+v", r)
	}
}
