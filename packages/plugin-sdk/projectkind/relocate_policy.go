package projectkind

import (
	"path"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Relocate is where a project of the kind belongs: the project's directory
// moves from its conventional kind directory to the target kind's
// (libs/core to packages/core). A project that sits anywhere else (tools/x,
// libs/group/x) stays where it is, and moves is false.
func Relocate(root string, to protocol.ProjectKind) (newRoot string, moves bool) {
	parent, name := path.Split(root)
	switch strings.TrimSuffix(parent, "/") {
	case "apps", "libs", "packages":
		target := path.Join(Directory(to), name)
		return target, target != root
	}
	return root, false
}
