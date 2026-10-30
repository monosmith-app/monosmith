package kindchange

import (
	"fmt"
	"path"

	"github.com/monosmith-app/monosmith/libs/plugin-vba/discovery"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/fileedits"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Change answers what making the project params.Project a params.To changes:
// the smithKind comment of its main module. Modules do not refer to each
// other by path, so a move edits nothing else. A package is refused: there is
// no registry to publish to.
func Change(workspaceRoot string, params protocol.ChangeKindParams) (protocol.ChangeKindResult, error) {
	if params.To == protocol.ProjectPackage {
		return protocol.ChangeKindResult{Refused: "a VBA project is an app or a lib: there is no registry to publish to"}, nil
	}
	edits := fileedits.New(workspaceRoot)
	var names []string
	for _, f := range edits.List(params.Project.Root, func(string) bool { return true }) {
		names = append(names, path.Base(f))
	}
	main, src := discovery.FindMain(names, func(name string) (string, bool) {
		return edits.Read(path.Join(params.Project.Root, name))
	})
	if main == "" {
		return protocol.ChangeKindResult{}, fmt.Errorf("%s has no VBA module", params.Project.Root)
	}
	out, _ := discovery.SetField(src, "Kind", string(params.To))
	edits.Write(path.Join(params.Project.Root, main), out)
	return protocol.ChangeKindResult{Files: edits.Edits()}, nil
}
