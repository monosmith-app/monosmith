package manifestfixes

import (
	"path"

	"github.com/monosmith-app/monosmith/libs/plugin-node/linking"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// InternalRangesDescription is the same for every package, so the host folds
// the fix of all of them into one.
const InternalRangesDescription = "use \"*\" for the dependencies on packages of this workspace"

// InternalRangesFix rewrites the pinned ranges of one manifest to the range
// smith writes itself, "*". root is the workspace-relative folder of the
// package ("" for the root package.json) and project its name for the host's
// folding ("" for the root). internal holds the package names of the
// workspace. With hasLockfile it also regenerates package-lock.json. The text
// is edited in place, so only the range values change. It returns nil when the
// file cannot be read or has nothing to change.
func InternalRangesFix(workspaceRoot, root, project string, internal map[string]bool, hasLockfile bool) *protocol.Fix {
	manifest := path.Join(root, "package.json")
	raw, err := readFile(workspaceRoot, manifest)
	if err != nil {
		return nil
	}
	pinned, err := linking.PinnedInternalRanges(raw, internal)
	if err != nil || len(pinned) == 0 {
		return nil
	}
	fix := &protocol.Fix{
		Description: InternalRangesDescription,
		Files:       []protocol.FileEdit{{Path: manifest, Content: string(linking.UseAnySpec(raw, pinned))}},
	}
	if project != "" {
		fix.AppliesTo = []string{project}
	}
	if hasLockfile {
		fix.Steps = []protocol.FixStep{{Description: "regenerate package-lock.json for the new ranges", Argv: []string{"npm", "install", "--package-lock-only", "--no-audit", "--no-fund"}}}
	}
	return fix
}
