package projectgraph

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

func TestAGroupAddsTagsAndALaterGroupCanSelectByThem(t *testing.T) {
	cfg := workspace.Config{
		Groups: []workspace.Group{
			{Select: []string{"libs/*"}, Tags: []string{"shared"}},
			{Select: []string{"tag:shared"}, Tags: []string{"published"}},
			{Select: []string{"lang:go"}, Tags: []string{"never"}},
		},
		Projects: map[string]workspace.ProjectOverride{"own": {Tags: []string{"mine"}}},
	}
	plain, own := nodeLib("plain"), nodeLib("own")
	for _, p := range []*Project{plain, own} {
		if err := ApplyOverrides(p, cfg); err != nil {
			t.Fatal(err)
		}
	}
	if got := strings.Join(plain.Tags, ","); got != "published,shared" {
		t.Errorf("plain tags = %q", got)
	}
	if got := strings.Join(own.Tags, ","); got != "mine,published,shared" {
		t.Errorf("a project keeps its own tags beside the group's: %q", got)
	}
}

func TestAGroupReleaseServesAProjectThatSetsNone(t *testing.T) {
	group := &workspace.ProjectRelease{Platforms: []string{"linux/amd64"}}
	own := &workspace.ProjectRelease{Platforms: []string{"darwin/arm64"}}
	cfg := workspace.Config{
		Groups:   []workspace.Group{{Select: []string{"libs/*"}, Release: group}},
		Projects: map[string]workspace.ProjectOverride{"own": {Release: own}},
	}
	plain, mine := nodeLib("plain"), nodeLib("own")
	for _, p := range []*Project{plain, mine} {
		if err := ApplyOverrides(p, cfg); err != nil {
			t.Fatal(err)
		}
	}
	if plain.Release == nil || strings.Join(plain.Release.Platforms, ",") != "linux/amd64" {
		t.Errorf("plain release = %+v", plain.Release)
	}
	if mine.Release == nil || strings.Join(mine.Release.Platforms, ",") != "darwin/arm64" {
		t.Errorf("a project's own release replaces the group's: %+v", mine.Release)
	}
}
