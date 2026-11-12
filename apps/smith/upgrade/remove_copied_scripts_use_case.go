package upgrade

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/monosmith-app/monosmith/apps/smith/doctor"
	"github.com/monosmith-app/monosmith/apps/smith/doctorfix"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/functionhost"
)

// removeCopiedScripts removes the smith scripts that older smith copied into
// function and Electron apps (scripts/assemble.mjs, package.mjs, start.mjs),
// and points the package.json scripts at smith's own commands. A copy still as
// smith wrote it goes. A copy that was edited by hand stays, and a hint names
// it: smith doctor --fix asks whether to keep it. With check it changes
// nothing and reports the files it would change.
func removeCopiedScripts(s *session.Session, check bool, r *Result) error {
	only := map[string]bool{}
	for _, p := range s.Graph.Projects {
		for _, rel := range functionhost.CopiedScripts {
			if _, err := os.Stat(filepath.Join(s.Root, filepath.FromSlash(p.Root), filepath.FromSlash(rel))); err == nil {
				only[p.Name] = true
			}
		}
	}
	if len(only) == 0 {
		return nil
	}
	var findings []doctor.Finding
	for _, f := range doctor.PluginChecks(s, only) {
		if f.Fix != nil && f.Fix.Description == functionhost.CopiedScriptsFix {
			findings = append(findings, f)
		}
	}
	safe, manual := doctor.Fixes(doctor.Report{Findings: findings})
	for _, fix := range safe {
		if !check {
			if err := doctorfix.Apply(context.Background(), s.Root, fix, io.Discard); err != nil {
				return err
			}
		}
		for _, e := range fix.Files {
			addFile(r, e.Path)
		}
	}
	for _, fix := range manual {
		r.Hints = append(r.Hints, fmt.Sprintf("%s (%s): a copied script was edited by hand, so smith kept it.\n→ smith doctor --fix asks whether to keep it or let smith run the command", fix.Description, fix.AppliesTo))
	}
	return nil
}
