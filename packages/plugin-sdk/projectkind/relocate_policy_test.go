package projectkind

import (
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestRelocateMovesAProjectBetweenTheKindDirectories(t *testing.T) {
	cases := []struct {
		root  string
		to    protocol.ProjectKind
		want  string
		moves bool
	}{
		{"libs/core", protocol.ProjectPackage, "packages/core", true},
		{"packages/core", protocol.ProjectLib, "libs/core", true},
		{"libs/core", protocol.ProjectApp, "apps/core", true},
		{"libs/core", protocol.ProjectLib, "libs/core", false},
		{"tools/core", protocol.ProjectPackage, "tools/core", false},
		{"libs/group/core", protocol.ProjectPackage, "libs/group/core", false},
	}
	for _, c := range cases {
		got, moves := Relocate(c.root, c.to)
		if got != c.want || moves != c.moves {
			t.Errorf("Relocate(%q, %s) = %q, %v; want %q, %v", c.root, c.to, got, moves, c.want, c.moves)
		}
	}
}
