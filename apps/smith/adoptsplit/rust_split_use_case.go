package adoptsplit

import (
	"os"
	"path/filepath"
	"regexp"
)

// crateEdition matches the edition line of a crate's [package] table.
var crateEdition = regexp.MustCompile(`(?m)^edition[ \t]*=[ \t]*"([^"]+)"[ \t]*`)

// SplitRust makes the root a Cargo workspace with the moved crate as its
// member: a root Cargo.toml with [workspace], the crate's Cargo.lock moved to
// the root (cargo keeps one lockfile for the workspace and ignores a member's),
// the crate's edition shared in [workspace.package] (the crate then inherits it),
// and a root .cargo/config.toml that builds every crate into dist/cargo, the
// folder smith caches. A file the root has already is kept. The crate's own
// Cargo.toml, with its dependencies and dev-dependencies, stays whole except
// its edition line: cargo scopes both to the crate. Nothing is staged or committed.
//
// Parameters: req is the move; Dir is the crate's folder.
// Returns: the notes for the person, or an error when a file cannot be written.
func SplitRust(req Request) (Result, error) {
	var res Result
	missing := func(rel string) bool {
		_, err := os.Stat(filepath.Join(req.Root, filepath.FromSlash(rel)))
		return err != nil
	}
	if missing("Cargo.toml") {
		text := "[workspace]\nresolver = \"2\"\nmembers = [\"" + req.Dir + "\"]\n"
		crateManifest := filepath.Join(req.Root, filepath.FromSlash(req.Dir), "Cargo.toml")
		if raw, err := os.ReadFile(crateManifest); err == nil {
			if m := crateEdition.FindStringSubmatch(string(raw)); m != nil {
				// The workspace shares the crate's edition, as a new crate does.
				text += "\n[workspace.package]\nedition = \"" + m[1] + "\"\n"
				if err := os.WriteFile(crateManifest, []byte(crateEdition.ReplaceAllString(string(raw), "edition.workspace = true")), 0o644); err != nil {
					return res, err
				}
			}
		}
		if err := os.WriteFile(filepath.Join(req.Root, "Cargo.toml"), []byte(text), 0o644); err != nil {
			return res, err
		}
		res.Notes = append(res.Notes, "the root Cargo.toml is a workspace with "+req.Dir+" as its member.")
	}
	lock := filepath.Join(req.Root, filepath.FromSlash(req.Dir), "Cargo.lock")
	if _, err := os.Stat(lock); err == nil && missing("Cargo.lock") {
		if err := os.Rename(lock, filepath.Join(req.Root, "Cargo.lock")); err != nil {
			return res, err
		}
		res.Notes = append(res.Notes, "Cargo.lock moved to the root, where the workspace keeps its one lockfile.")
	}
	if missing(".cargo/config.toml") && missing(".cargo/config") {
		if err := os.MkdirAll(filepath.Join(req.Root, ".cargo"), 0o755); err != nil {
			return res, err
		}
		text := "[build]\ntarget-dir = \"dist/cargo\"\n"
		if err := os.WriteFile(filepath.Join(req.Root, ".cargo", "config.toml"), []byte(text), 0o644); err != nil {
			return res, err
		}
		res.Notes = append(res.Notes, "cargo builds into dist/cargo (.cargo/config.toml), where smith caches it.")
	}
	return res, nil
}
