package kindchange

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-applescript/discovery"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/fileedits"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Change answers what making the project params.Project a params.To changes:
// the smithKind comment of its main script. Scripts do not refer to each
// other by path in a way smith reads, so a move edits nothing else. A package
// is refused: there is no registry to publish to.
func Change(workspaceRoot string, params protocol.ChangeKindParams) (protocol.ChangeKindResult, error) {
	if params.To == protocol.ProjectPackage {
		return protocol.ChangeKindResult{Refused: "an AppleScript project is an app or a lib: there is no registry to publish to"}, nil
	}
	edits := fileedits.New(workspaceRoot)
	found := edits.List(params.Project.Root, func(name string) bool { return strings.HasSuffix(name, ".applescript") })
	names := make([]string, len(found))
	for i, f := range found {
		names[i] = f[strings.LastIndex(f, "/")+1:]
	}
	main := discovery.PickMain(names)
	if main == "" {
		return protocol.ChangeKindResult{}, fmt.Errorf("%s has no .applescript source", params.Project.Root)
	}
	rel := params.Project.Root + "/" + main
	src, _ := edits.Read(rel)
	out, _ := discovery.SetField(src, "Kind", string(params.To))
	edits.Write(rel, out)
	return protocol.ChangeKindResult{Files: edits.Edits()}, nil
}
