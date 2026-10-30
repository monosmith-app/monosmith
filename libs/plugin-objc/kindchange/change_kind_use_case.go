package kindchange

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-objc/discovery"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/fileedits"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Change answers what making the project params.Project a params.To changes:
// the smithKind comment of its podspec. A podspec's paths are relative to it
// and move with it, and pods refer to each other by name, so a move edits
// nothing else. An app is refused: a podspec describes a library.
func Change(workspaceRoot string, params protocol.ChangeKindParams) (protocol.ChangeKindResult, error) {
	if params.To == protocol.ProjectApp {
		return protocol.ChangeKindResult{Refused: "an Objective-C project is a CocoaPods library (a lib or a package), not an app"}, nil
	}
	edits := fileedits.New(workspaceRoot)
	found := edits.List(params.Project.Root, func(name string) bool { return strings.HasSuffix(name, ".podspec") })
	if len(found) == 0 {
		return protocol.ChangeKindResult{}, fmt.Errorf("%s has no podspec", params.Project.Root)
	}
	src, _ := edits.Read(found[0])
	out, _ := discovery.SetKind(src, string(params.To))
	edits.Write(found[0], out)
	return protocol.ChangeKindResult{Files: edits.Edits()}, nil
}
