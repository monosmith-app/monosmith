package lockcarry

import (
	"os"
	"path/filepath"
)

// npmLockfiles are the npm lockfiles smith reads, in npm's order of
// preference: a shrinkwrap wins over a package-lock.json.
var npmLockfiles = []string{"npm-shrinkwrap.json", "package-lock.json"}

// otherNodeLockfiles are the lockfiles of the package managers whose
// versions smith cannot carry into npm's lockfile.
var otherNodeLockfiles = []string{"yarn.lock", "pnpm-lock.yaml", "bun.lock", "bun.lockb"}

// Result says what CarryNpm did.
type Result struct {
	// Lockfile is the project's lockfile smith read, or "" when it read none.
	Lockfile string
	// Kept are the project's direct dependencies and their locked versions.
	Kept []Locked
	// Added is the number of entries added to the workspace lockfile.
	Added int
	// Unsupported is a lockfile smith cannot carry (yarn, pnpm, bun), or "".
	Unsupported string
	// NoWorkspaceLock is true when the workspace has no package-lock.json yet,
	// so there is nothing to merge into.
	NoWorkspaceLock bool
}

// CarryNpm merges the npm lockfile at the top of a project's source folder
// into the workspace's package-lock.json, under the project's folder, so the
// next install keeps the versions the project locked (MergeNpm).
//
// Parameters: root is the workspace root; source is the folder the project
// comes from, which still holds its lockfile; folder is the project's folder
// from the workspace root, with forward slashes, which holds its package.json.
// Returns: what was done, and an error when a lockfile cannot be read or
// written.
func CarryNpm(root, source, folder string) (Result, error) {
	var r Result
	var lock []byte
	for _, name := range npmLockfiles {
		raw, err := os.ReadFile(filepath.Join(source, name))
		if err == nil {
			r.Lockfile, lock = name, raw
			break
		}
	}
	if r.Lockfile == "" {
		for _, name := range otherNodeLockfiles {
			if _, err := os.Stat(filepath.Join(source, name)); err == nil {
				r.Unsupported = name
				break
			}
		}
		return r, nil
	}
	wsPath := filepath.Join(root, "package-lock.json")
	workspace, err := os.ReadFile(wsPath)
	if os.IsNotExist(err) {
		r.NoWorkspaceLock = true
		return r, nil
	}
	if err != nil {
		return r, err
	}
	manifest, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(folder), "package.json"))
	if err != nil {
		return r, err
	}
	merged, kept, added, err := MergeNpm(workspace, lock, manifest, folder)
	if err != nil {
		return r, err
	}
	r.Kept, r.Added = kept, added
	if added == 0 {
		return r, nil
	}
	return r, os.WriteFile(wsPath, merged, 0o644)
}
