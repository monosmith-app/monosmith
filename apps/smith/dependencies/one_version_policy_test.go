package dependencies

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func devDecl(project, name, spec string) protocol.Declaration {
	return protocol.Declaration{Project: project, File: project + "/package.json", Name: name, Spec: spec, Section: "devDependencies", Rewritable: true}
}

func TestFindFragmentationIgnoresDevToolsAndExceptions(t *testing.T) {
	declared := protocol.DeclaredResult{Declarations: []protocol.Declaration{
		decl("web", "axios", "^1.6.0", true), decl("api", "axios", "^1.7.0", true),
		decl("web", "left-pad", "^1.0.0", true), decl("api", "left-pad", "^1.3.0", true),
		devDecl("web", "jest", "^29.0.0"), devDecl("", "jest", "^30.0.0"),
	}}
	got := FindFragmentation("node", declared, nil, []string{"left-pad"})
	if len(got) != 1 || got[0].Name != "axios" || got[0].Target != "^1.7.0" {
		t.Fatalf("only axios is fragmented, left-pad is excepted and jest is a dev tool: %+v", got)
	}
	if got := FindFragmentation("node", declared, nil, []string{"left-pad", "axios"}); len(got) != 0 {
		t.Fatalf("every exception listed: %+v", got)
	}
}

func TestFindMisplacedDevTools(t *testing.T) {
	declared := protocol.DeclaredResult{Declarations: []protocol.Declaration{
		devDecl("", "jest", "^30.0.0"), devDecl("web", "jest", "^30.0.0"),
		devDecl("", "eslint", "^9.0.0"),
		devDecl("web", "vite", "^5.0.0"),
		devDecl("web", "tsx", "^4.0.0"), devDecl("api", "tsx", "^4.0.0"),
		devDecl("web", "pinned", "^1.0.0"), devDecl("api", "pinned", "^1.0.0"),
	}}
	got := FindMisplacedDevTools("node", declared, nil, []string{"pinned"})
	var names []string
	for _, m := range got {
		names = append(names, m.Name+":"+strings.Join(m.Sites, "+"))
	}
	if want := "jest:(root)+web,tsx:api+web"; strings.Join(names, ",") != want {
		t.Fatalf("got %v want %s: a tool at the root alone or in a single project is fine, an exception is left alone", names, want)
	}
}
