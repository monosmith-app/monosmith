package deriving

import (
	"bytes"
	"path"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// TSConfigReferences is the id of the renderer of a tsconfig's references.
const TSConfigReferences = "node.tsconfig.references"

// libConfig is the tsconfig that carries the references.
const libConfig = "tsconfig.lib.json"

// tsconfigReferences derives the references of a project's tsconfig.lib.json
// from its internal dependencies in the project graph: one reference to the
// tsconfig of each dependency that is a Node project with one. The renderer
// owns the "references" key; the rest of the file is the person's. A project
// whose file is not there, and that the record does not own, gets nothing.
func tsconfigReferences(workspaceRoot string, p protocol.Project, byName map[string]protocol.Project, params protocol.DeriveParams) (protocol.DerivedFile, bool) {
	target := path.Join(p.Root, libConfig)
	content, exists := readFile(workspaceRoot, target)
	if !exists && !owns(params, target) {
		return protocol.DerivedFile{}, false
	}
	refs := []any{}
	for _, dep := range params.Graph[p.Name] {
		d, ok := byName[dep]
		if !ok {
			continue
		}
		for _, name := range []string{libConfig, "tsconfig.json"} {
			// An owned file counts although it is gone: deriving it comes first or
			// last, and the reference must not depend on the order.
			if _, there := readFile(workspaceRoot, path.Join(d.Root, name)); there || owns(params, path.Join(d.Root, name)) {
				refs = append(refs, map[string]any{"path": relative(p.Root, path.Join(d.Root, name))})
				break
			}
		}
	}
	if len(refs) == 0 && !owns(params, target) && !bytes.Contains(content, []byte(`"references"`)) {
		return protocol.DerivedFile{}, false
	}
	return protocol.DerivedFile{
		Project: p.Name, Path: target, Mode: protocol.DeriveKeys, Format: "json",
		Renderer: TSConfigReferences, Version: 1, Keys: map[string]any{"references": refs},
	}, true
}
