package kindchange

import (
	"fmt"
	"path"
	"regexp"

	"github.com/monosmith-app/monosmith/libs/plugin-swift/discovery"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/fileedits"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/pathrefs"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// pathDependency matches a local package dependency, whose path is group 2.
var pathDependency = regexp.MustCompile(`(\.package\(\s*(?:name:\s*"[^"]*"\s*,\s*)?path:\s*")([^"]+)(")`)

// Change answers what making the project params.Project a params.To changes:
// the smithKind comment of its Package.swift, and, when its directory moves,
// the path dependencies of its own manifest that leave it and those other
// projects' manifests make on it.
func Change(workspaceRoot string, params protocol.ChangeKindParams) (protocol.ChangeKindResult, error) {
	p := params.Project
	edits := fileedits.New(workspaceRoot)
	manifest := path.Join(p.Root, discovery.Manifest)
	src, ok := edits.Read(manifest)
	if !ok {
		return protocol.ChangeKindResult{}, fmt.Errorf("%s not found", manifest)
	}
	out, _ := discovery.SetField(src, "Kind", string(params.To))
	edits.Write(manifest, out)

	if params.NewRoot != p.Root {
		out = pathDependency.ReplaceAllStringFunc(out, func(m string) string {
			g := pathDependency.FindStringSubmatch(m)
			return g[1] + pathrefs.RebaseFrom(g[2], p.Root, p.Root, params.NewRoot) + g[3]
		})
		edits.Write(manifest, out)
		for _, other := range params.Projects {
			if other.Language != "swift" || other.Root == p.Root {
				continue
			}
			file := path.Join(other.Root, discovery.Manifest)
			text, ok := edits.Read(file)
			if !ok {
				continue
			}
			edits.Write(file, pathDependency.ReplaceAllStringFunc(text, func(m string) string {
				g := pathDependency.FindStringSubmatch(m)
				return g[1] + pathrefs.Follow(g[2], other.Root, p.Root, params.NewRoot) + g[3]
			}))
		}
	}
	return protocol.ChangeKindResult{Files: edits.Edits()}, nil
}
