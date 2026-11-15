package dependencies

import (
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func decl(project, name, spec string, rewritable bool) protocol.Declaration {
	return protocol.Declaration{Project: project, File: project + "/package.json", Name: name, Spec: spec, Section: "dependencies", Rewritable: rewritable}
}

func TestFindDrift(t *testing.T) {
	declared := protocol.DeclaredResult{
		Declarations: []protocol.Declaration{
			decl("web", "zod", "^3.22.0", true), decl("api", "zod", "^3.20.0", true), decl("cli", "zod", "~3.21.0", true), decl("docs", "zod", "github:colinhacks/zod", false),
			decl("web", "lodash", "4.17.20", true), decl("api", "lodash", "^4.17.21", true),
			decl("web", "same", "^1.0.0", true), decl("api", "same", "^1.0.0", true),
			decl("web", "@acme/ui", "*", false), decl("api", "@acme/ui", "^1.0.0", true),
			{Project: "ui", Name: "react", Spec: "^17.0.0", Peer: true, Rewritable: false}, decl("web", "react", "^18.2.0", true),
		},
		Installed: map[string]string{"zod": "3.23.8"},
	}
	drift := FindDrift("node", declared, map[string]bool{"@acme/ui": true})
	if len(drift) != 2 {
		t.Fatalf("drift %+v", drift)
	}
	lodash, zod := drift[0], drift[1]
	if lodash.Name != "lodash" || lodash.Target != "^4.17.21" || lodash.Reason != "highest" || len(lodash.Fixable) != 1 || lodash.Fixable[0].Project != "web" {
		t.Errorf("lodash %+v", lodash)
	}
	if zod.Target != "^3.23.8" || zod.Reason != "installed" || len(zod.Fixable) != 3 || len(zod.Blocked) != 1 || zod.Blocked[0].Project != "docs" {
		t.Errorf("zod %+v", zod)
	}
}

func TestNewer(t *testing.T) {
	for _, c := range []struct {
		a, b string
		want bool
	}{
		{"1.10.0", "1.9.9", true}, {"2.31", "2.31.0", false}, {"2.32", "2.31.5", true},
		{"1.0.0", "1.0.0-rc.1", true}, {"1.0.0-rc.2", "1.0.0-rc.1", true}, {"1.0.0-rc.1", "1.0.0", false},
	} {
		if got := newer(c.a, c.b); got != c.want {
			t.Errorf("newer(%s, %s) = %v", c.a, c.b, got)
		}
	}
	for spec, ok := range map[string]bool{"^1.2.3": true, ">=2.31": true, "==8.2.0": true, "1.0.0-beta.1": true, "*": false, "latest": false, ">=1 <2": false} {
		if _, _, got := splitSpec(spec); got != ok {
			t.Errorf("splitSpec(%q) ok = %v", spec, got)
		}
	}
}
