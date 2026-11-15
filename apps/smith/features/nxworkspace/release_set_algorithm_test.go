package nxworkspace

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestReleaseSet(t *testing.T) {
	var nx NxJSON
	if err := json.Unmarshal([]byte(`{"release": {"projects": ["packages/*", "tag:release:go", "!tag:type:go-lib", "@acme/extra"]}}`), &nx); err != nil {
		t.Fatal(err)
	}
	w := &Workspace{Nx: nx, Projects: []*Project{
		{Name: "@acme/kit", Root: "packages/kit"},
		{Name: "gokit", Root: "packages/gokit", Tags: []string{"type:go-lib"}},
		{Name: "tool", Root: "apps/tool", Tags: []string{"release:go"}},
		{Name: "web", Root: "apps/web"},
		{Name: "@acme/extra", Root: "libs/extra"},
	}}
	got := ReleaseSet(w)
	want := map[string]bool{"@acme/kit": true, "tool": true, "@acme/extra": true}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v", got)
	}

	w.Nx.Release = nil
	w.Projects = []*Project{
		{Name: "pub", Package: &PackageJSON{}},
		{Name: "priv", Package: &PackageJSON{Private: true}},
		{Name: "go"},
	}
	if got := ReleaseSet(w); !reflect.DeepEqual(got, map[string]bool{"pub": true}) {
		t.Fatalf("without release.projects only public packages publish: %v", got)
	}
}

func TestContracts(t *testing.T) {
	var nx NxJSON
	err := json.Unmarshal([]byte(`{
		"plugins": ["@nx/js", {"plugin": "@nx-go/nx-go", "options": {"modulePrefix": "github.com/a/b"}}],
		"release": {"releaseTagPattern": "v{version}", "changelog": {"projectChangelogs": {"createRelease": "github"}}}
	}`), &nx)
	if err != nil {
		t.Fatal(err)
	}
	if nx.Plugins[0].Plugin != "@nx/js" || nx.Plugins[1].Plugin != "@nx-go/nx-go" || string(nx.Plugins[1].Options["modulePrefix"]) != `"github.com/a/b"` {
		t.Fatalf("plugins %+v", nx.Plugins)
	}
	if nx.Release.TagPattern() != "v{version}" || !nx.Release.CreatesGitHubReleases() {
		t.Fatalf("release %+v", nx.Release)
	}
	var none *Release
	if none.TagPattern() != "{projectName}@{version}" || none.CreatesGitHubReleases() {
		t.Fatal("defaults")
	}

	var tg Target
	err = json.Unmarshal([]byte(`{
		"dependsOn": ["^build", "lint", {"target": "gen", "dependencies": true}, {"target": "x", "projects": ["other"]}],
		"inputs": ["default", "{projectRoot}/**/*.go", {"env": "VERSION"}, {"externalDependencies": ["jest"]}, {"input": "production", "dependencies": true}]
	}`), &tg)
	if err != nil {
		t.Fatal(err)
	}
	d := tg.DependsOn
	if !d[0].Dependencies || d[0].Target != "build" || d[1].Target != "lint" || !d[2].Dependencies || d[3].Projects[0] != "other" {
		t.Fatalf("dependsOn %+v", d)
	}
	in := tg.Inputs
	if in[0].Text != "default" || in[1].Text != "{projectRoot}/**/*.go" || in[2].Env != "VERSION" || in[3].External[0] != "jest" || in[4].Named != "production" || !in[4].Dependencies {
		t.Fatalf("inputs %+v", in)
	}
	var list StringList
	if json.Unmarshal([]byte(`"one"`), &list) != nil || list[0] != "one" {
		t.Fatal(list)
	}
}
