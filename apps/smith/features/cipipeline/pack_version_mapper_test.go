package cipipeline

import (
	"reflect"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/features/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/features/release"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestPackVersionLine(t *testing.T) {
	cases := []struct {
		name  string
		stamp release.Stamp
		want  string
	}{
		{"stable", release.Stamp{Project: "api", Was: "6.1.0", Version: "6.2.0", Files: []string{"apps/api/package.json"}},
			"api: 6.1.0 -> 6.2.0 for the pack (the planned release; 1 file put back after the pack)"},
		{"channel, no manifest version", release.Stamp{Project: "app", Version: "1.0.0-beta.3", Files: []string{"a", "b"}},
			"app: no version -> 1.0.0-beta.3 for the pack (the planned release; 2 files put back after the pack)"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := packVersionLine(c.stamp); got != c.want {
				t.Errorf("got  %q\nwant %q", got, c.want)
			}
		})
	}
}

func TestDeployableNames(t *testing.T) {
	project := func(name string, kind protocol.ProjectKind, packs bool) *projectgraph.Project {
		p := &projectgraph.Project{}
		p.Name, p.Kind = name, kind
		if packs {
			p.Tasks = map[string]protocol.Task{PackTask: {}}
		}
		return p
	}
	got := deployableNames([]*projectgraph.Project{
		project("web", protocol.ProjectApp, false),
		project("kit", protocol.ProjectPackage, false),
		project("func", protocol.ProjectPackage, true),
		project("api", protocol.ProjectApp, true),
	})
	if want := []string{"api", "func", "web"}; !reflect.DeepEqual(got, want) {
		t.Errorf("deployableNames = %v, want %v", got, want)
	}
}
