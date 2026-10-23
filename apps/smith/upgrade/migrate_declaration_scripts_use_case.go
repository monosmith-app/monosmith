package upgrade

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/declarationspecifiers"
	"github.com/monosmith-app/monosmith/apps/smith/session"
)

// migrateDeclarationScripts moves a workspace converted by an older smith off
// the script it left at the root: every package.json that ran
// `node ../../declaration-specifiers.mjs` runs `smith declaration-specifiers`
// instead, and the root file goes once nothing calls it and it is still
// exactly what smith wrote. An edited one stays, named in a hint. With check
// it changes nothing.
func migrateDeclarationScripts(s *session.Session, check bool, r *Result) error {
	stillCalled := false
	for _, p := range s.Graph.Projects {
		if p.Plugin != "node" {
			continue
		}
		manifest := filepath.Join(s.Root, filepath.FromSlash(p.Root), "package.json")
		raw, err := os.ReadFile(manifest)
		if err != nil {
			continue
		}
		out, found := declarationspecifiers.MigrateScript(string(raw))
		if !found {
			continue
		}
		addFile(r, filepath.ToSlash(filepath.Join(p.Root, "package.json")))
		if check {
			stillCalled = true
			continue
		}
		if err := os.WriteFile(manifest, []byte(out), 0o644); err != nil {
			return err
		}
	}
	legacy := filepath.Join(s.Root, declarationspecifiers.LegacyFile)
	raw, err := os.ReadFile(legacy)
	if err != nil || stillCalled {
		return nil
	}
	if strings.ReplaceAll(string(raw), "\r\n", "\n") != declarationspecifiers.LegacyContent {
		r.Hints = append(r.Hints, declarationspecifiers.LegacyFile+" is not what smith wrote, so it was left; builds now call `smith declaration-specifiers`, so delete it if your edits are not needed")
		return nil
	}
	addFile(r, declarationspecifiers.LegacyFile)
	if check {
		return nil
	}
	return os.Remove(legacy)
}
