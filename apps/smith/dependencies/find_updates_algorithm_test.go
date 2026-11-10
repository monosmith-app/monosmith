package dependencies

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestFindUpdates(t *testing.T) {
	declared := protocol.DeclaredResult{Declarations: []protocol.Declaration{
		decl("web", "zod", "^3.22.0", true), decl("api", "zod", "~3.23.8", true),
		decl("web", "jest", "^29.7.0", true),
		decl("web", "chalk", "^5.3.0", true), decl("api", "chalk", "^5.3.1", true),
		decl("web", "pinned", "github:me/pinned", false),
		decl("web", "@acme/ui", "^1.0.0", true),
		{Project: "ui", Name: "react", Spec: "^17.0.0", Peer: true},
	}}
	latest := map[string]string{"zod": "3.24.1", "jest": "30.0.0", "chalk": "5.3.1", "pinned": "2.0.0", "@acme/ui": "9.9.9", "react": "19.0.0"}
	updates := FindUpdates("node", declared, latest, map[string]bool{"@acme/ui": true})
	var got []string
	for _, u := range updates {
		got = append(got, u.Level+":"+u.Name+":"+u.Current+"→"+u.Latest+":"+strings.Join(u.Projects(), ","))
	}
	want := "patch:chalk:5.3.0→5.3.1:web minor:zod:3.23.8→3.24.1:web,api major:jest:29.7.0→30.0.0:web"
	if strings.Join(got, " ") != want {
		t.Fatalf("got  %s\nwant %s", strings.Join(got, " "), want)
	}
	var specs []string
	for _, r := range rewrites(updates) {
		specs = append(specs, r.Declaration.Project+":"+r.Declaration.Name+"="+r.Spec)
	}
	if strings.Join(specs, " ") != "web:chalk=^5.3.1 web:zod=^3.24.1 api:zod=~3.24.1 web:jest=^30.0.0" {
		t.Fatalf("rewrites %v", specs)
	}
}

func TestPresetsComeFirstAsOneGroup(t *testing.T) {
	updates := []Update{
		{Name: "zod", Level: "patch"},
		{Name: "@acme/eslint-config", Level: "major"},
		{Name: "jest", Level: "minor"},
	}
	MarkPresets(updates, []string{"@acme/eslint-config"})
	if updates[0].Name != "@acme/eslint-config" || !updates[0].Preset || updates[0].Group() != "preset" {
		t.Fatalf("got %+v", updates)
	}
	if updates[1].Name != "zod" || updates[1].Group() != "patch" || updates[2].Name != "jest" {
		t.Errorf("the others changed order or group: %+v", updates)
	}
}
