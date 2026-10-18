package upgrade

import (
	"fmt"
	"slices"

	"github.com/monosmith-app/monosmith/apps/smith/doctor"
	"github.com/monosmith-app/monosmith/apps/smith/recipedrift"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

// upgradeProjects brings the projects made from a recipe up to date with the
// recipe as it is now. It rewrites their recipe record, settings and shared
// stack in smith.toml, then has their plugins restore the files they own
// from the new settings: the safe fixes their doctor checks offer, which only
// restore what a plugin owns. Starter files and files you edited are never
// written. With check it changes nothing and reports the same plan, with the
// change to smith.toml as a diff.
func upgradeProjects(s *session.Session, check bool, r *Result) error {
	plan, err := recipedrift.PlanUpgrade(s, recipedrift.Check(s))
	if err != nil {
		return err
	}
	r.Projects, r.Skipped, r.Starters = plan.Changes, plan.Skipped, plan.Starters
	if !plan.Changed() {
		return nil
	}
	addFile(r, workspace.FileName)
	if check {
		r.Diff = doctor.UnifiedDiff(workspace.FileName, plan.Before, plan.After)
		return nil
	}
	if err := plan.Apply(s); err != nil {
		return fmt.Errorf("writing the recipes' changes to smith.toml: %w", err)
	}
	return restoreOwnedFiles(s, plan.Projects(), r)
}

// restoreOwnedFiles applies the safe fixes the plugins of the given projects
// offer, which rewrite the files each owns from the settings smith.toml now
// holds. A fix that is not safe is left, and named.
func restoreOwnedFiles(s *session.Session, projects []string, r *Result) error {
	only := map[string]bool{}
	for _, name := range projects {
		only[name] = true
	}
	safe, manual := doctor.Fixes(doctor.Report{Findings: doctor.PluginChecks(s, only)})
	for _, fix := range safe {
		if err := doctor.ApplyFix(s.Root, fix); err != nil {
			return err
		}
		for _, e := range fix.Files {
			addFile(r, e.Path)
		}
	}
	for _, fix := range manual {
		r.Hints = append(r.Hints, fmt.Sprintf("%s: not applied, because it could change files you edit; `smith doctor --fix` shows the diff", fix.Description))
	}
	return nil
}

func addFile(r *Result, path string) {
	if !slices.Contains(r.Files, path) {
		r.Files = append(r.Files, path)
	}
}
