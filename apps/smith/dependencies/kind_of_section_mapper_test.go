package dependencies

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestKindOf(t *testing.T) {
	for section, want := range map[string]string{
		"dependencies": "prod", "require": "prod", "PackageReference": "prod", "gradle": "prod",
		"devDependencies": "dev", "dev_dependencies": "dev", "dev-dependencies": "dev",
		"development_dependencies": "dev", "require-dev": "dev", "dev": "dev",
		"optionalDependencies": "optional", "peerDependencies": "peer",
		"build-dependencies": "build", "workspace.dependencies": "managed",
		"dependencyManagement": "managed", "property:junit.version": "managed",
	} {
		if got := KindOf(protocol.Declaration{Section: section}); got != want {
			t.Errorf("%s: got %s, want %s", section, got, want)
		}
	}
	if got := KindOf(protocol.Declaration{Section: "dependencies", Peer: true}); got != "peer" {
		t.Errorf("a peer declaration is peer, got %s", got)
	}
}

func TestFindUpdatesKindsAcrossProjects(t *testing.T) {
	dev := decl("api", "zod", "^3.22.0", true)
	dev.Section = "devDependencies"
	jest := decl("web", "jest", "^29.0.0", true)
	jest.Section = "devDependencies"
	declared := protocol.DeclaredResult{Declarations: []protocol.Declaration{decl("web", "zod", "^3.22.0", true), dev, jest}}
	updates := FindUpdates("node", declared, map[string]string{"zod": "3.24.1", "jest": "30.0.0"}, nil)
	got := map[string]string{}
	for _, u := range updates {
		got[u.Name] = strings.Join(u.Kinds, ", ")
	}
	if got["zod"] != "dev, prod" || got["jest"] != "dev" {
		t.Fatalf("got %v", got)
	}
}
