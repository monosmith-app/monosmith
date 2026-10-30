package kindchange

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/fileedits"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

var kindComment = regexp.MustCompile(`(?m)^(--\s*smithKind:\s*)\S+`)

// Change answers what making the project params.Project a params.To changes:
// the smithKind comment of its rockspec. Rocks refer to each other by name, not
// by path, so a move edits nothing else. A package is refused: there is no
// luarocks.org publishing yet.
func Change(workspaceRoot string, params protocol.ChangeKindParams) (protocol.ChangeKindResult, error) {
	if params.To == protocol.ProjectPackage {
		return protocol.ChangeKindResult{Refused: "a Lua project is an app or a lib: publishing to luarocks.org is not supported yet"}, nil
	}
	edits := fileedits.New(workspaceRoot)
	found := edits.List(params.Project.Root, func(name string) bool { return strings.HasSuffix(name, ".rockspec") })
	if len(found) == 0 {
		return protocol.ChangeKindResult{}, fmt.Errorf("%s has no rockspec", params.Project.Root)
	}
	src, _ := edits.Read(found[0])
	if kindComment.MatchString(src) {
		edits.Write(found[0], kindComment.ReplaceAllString(src, "${1}"+string(params.To)))
	} else {
		edits.Write(found[0], "-- smithKind: "+string(params.To)+"\n"+src)
	}
	return protocol.ChangeKindResult{Files: edits.Edits()}, nil
}
