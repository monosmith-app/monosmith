package pipeline

import (
	"encoding/json"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-node/linking"
	"github.com/monosmith-app/monosmith/libs/plugin-node/manifestfixes"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// CheckInternalRanges is the name of the check of the ranges a package uses to
// depend on another package of the same workspace.
const CheckInternalRanges = `internal dependencies use "*"`

// CheckInternalRangeFindings checks the package.json of every project (and the
// root one) for a dependency, devDependency or optionalDependency on another
// package of this workspace with a range other than "*", the range smith
// writes itself. Each manifest with such entries is one warning with a fix
// that rewrites the ranges and regenerates the lockfile; none gives one
// passing finding. peerDependencies are not read: a peer range is a contract
// for consumers.
func CheckInternalRangeFindings(workspaceRoot string, projects []protocol.Project) []protocol.Finding {
	internal := map[string]bool{}
	for _, p := range projects {
		if name := packageName(workspaceRoot, p.Root); name != "" {
			internal[name] = true
		}
	}
	_, lockErr := os.Stat(filepath.Join(workspaceRoot, "package-lock.json"))
	hasLockfile := lockErr == nil

	var findings []protocol.Finding
	add := func(root, project, label string) {
		raw, err := os.ReadFile(filepath.Join(workspaceRoot, filepath.FromSlash(path.Join(root, "package.json"))))
		if err != nil {
			return
		}
		pinned, err := linking.PinnedInternalRanges(raw, internal)
		if err != nil || len(pinned) == 0 {
			return
		}
		entries := make([]string, len(pinned))
		for i, p := range pinned {
			entries[i] = p.String()
		}
		findings = append(findings, protocol.Finding{
			Check:   CheckInternalRanges,
			Project: project,
			Area:    "dependencies",
			Warning: true,
			Problem: label + " depends on packages of this workspace with a pinned range: " + strings.Join(entries, ", "),
			Detail:  strings.Join(entries, "; "),
			Remedy: "use \"*\" for each, the range smith writes when it installs an internal dependency. " +
				"Internal packages come from the workspace, so a pinned range goes stale when the sibling's version moves, " +
				"and npm may then fetch the registry copy instead of the workspace one",
			Fix: manifestfixes.InternalRangesFix(workspaceRoot, root, project, internal, hasLockfile),
		})
	}
	add("", "", "the root package.json")
	for _, p := range projects {
		add(p.Root, p.Name, p.Name)
	}
	if len(findings) == 0 {
		return []protocol.Finding{{Check: CheckInternalRanges, Area: "dependencies", OK: true, Warning: true}}
	}
	return findings
}

// packageName reads the "name" of the package.json in root, or "" when the
// file has none.
func packageName(workspaceRoot, root string) string {
	raw, err := os.ReadFile(filepath.Join(workspaceRoot, filepath.FromSlash(root), "package.json"))
	if err != nil {
		return ""
	}
	var pkg struct {
		Name string `json:"name"`
	}
	if json.Unmarshal(raw, &pkg) != nil {
		return ""
	}
	return pkg.Name
}
