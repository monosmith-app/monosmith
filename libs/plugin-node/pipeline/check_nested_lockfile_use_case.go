package pipeline

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// nestedLockfiles are the lockfiles a package manager writes beside a
// package.json. In an npm workspace only the root's counts.
var nestedLockfiles = []string{"package-lock.json", "yarn.lock", "pnpm-lock.yaml", "bun.lock", "bun.lockb"}

// CheckNestedLockfile reports a lockfile inside a project that the root
// package.json workspaces cover: npm installs the whole workspace from the
// root lockfile, so the nested one is dead and only misleads. A project that
// moved or came in from its own repository often keeps one.
//
// Parameters: workspaceRoot is the workspace folder; p is the project, which
// the caller found covered by the root workspaces.
// Returns: a warning with a fix that deletes the files, and true, when the
// project folder holds a lockfile; false when it holds none or is the root.
func CheckNestedLockfile(workspaceRoot string, p protocol.Project) (protocol.Finding, bool) {
	if p.Root == "" || p.Root == "." {
		return protocol.Finding{}, false
	}
	var found []string
	for _, name := range nestedLockfiles {
		rel := path.Join(p.Root, name)
		if _, err := os.Stat(filepath.Join(workspaceRoot, filepath.FromSlash(rel))); err == nil {
			found = append(found, rel)
		}
	}
	if len(found) == 0 {
		return protocol.Finding{}, false
	}
	fix := &protocol.Fix{Description: "delete the lockfile(s) in " + p.Root + ": the root package-lock.json holds the workspace's versions"}
	for _, rel := range found {
		fix.Files = append(fix.Files, protocol.FileEdit{Path: rel, Delete: true})
	}
	return protocol.Finding{
		Check:   p.Root + " has no lockfile of its own",
		Project: p.Name,
		Warning: true,
		Problem: "has a lockfile that npm does not use: " + strings.Join(found, ", "),
		Detail:  strings.Join(found, ", ") + " is not used: npm installs the workspace from the root package-lock.json",
		Remedy:  "delete " + strings.Join(found, " and ") + ", then run `smith install` so the root lockfile holds the project's dependencies",
		Fix:     fix,
	}, true
}
