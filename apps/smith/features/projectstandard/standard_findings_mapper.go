package projectstandard

import (
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// CheckSuffix ends the check of every standard finding: "<folder> uses smith's
// standard configuration".
const CheckSuffix = " uses smith's standard configuration"

// WorkspaceCheck is the check of the workspace's own files (the root
// tsconfig.base.json).
const WorkspaceCheck = "the workspace" + CheckSuffix

// FixDescription is the description every standard fix shares, so doctor
// folds them into one ("applies to 12 projects").
const FixDescription = "use smith's standard configuration files (a replaced file is kept as <file>.pre-smith)"

// WorkspaceFixDescription is the description of the fix of the workspace's own
// files.
const WorkspaceFixDescription = "use smith's standard configuration in the workspace's root files"

// Findings turns an answer into doctor findings: one for each project the
// plugin has a standard for, and one for the workspace's own files, passed
// when every file is as the standard has it. A failed one is a warning, since
// the project works as it is, and its fix writes the standard.
//
// Parameters: a is the answer of one plugin.
// Returns: the findings, projects in the order asked, the workspace last.
func Findings(a Answer) []protocol.Finding {
	byProject := map[string][]Difference{}
	for _, d := range a.Differences {
		byProject[d.File.Project] = append(byProject[d.File.Project], d)
	}
	var out []protocol.Finding
	for _, p := range a.Projects {
		if len(a.Standard[p.Name]) == 0 {
			continue
		}
		out = append(out, finding(p.Root+CheckSuffix, p.Name, byProject[p.Name]))
	}
	if len(a.Standard[""]) > 0 {
		out = append(out, finding(WorkspaceCheck, "", byProject[""]))
	}
	return out
}

func finding(check, project string, diffs []Difference) protocol.Finding {
	f := protocol.Finding{Check: check, Project: project, OK: len(diffs) == 0}
	if project == "" {
		f.Area = "config"
	}
	if f.OK {
		return f
	}
	var files, reasons []string
	fix := &protocol.Fix{Description: WorkspaceFixDescription}
	if project != "" {
		fix.Description, fix.AppliesTo = FixDescription, []string{project}
	}
	for _, d := range diffs {
		files = append(files, d.Summary())
		if d.File.Reason != "" && !contains(reasons, d.File.Reason) {
			reasons = append(reasons, d.File.Reason)
		}
		fix.Files = append(fix.Files, d.Edits()...)
	}
	f.Warning = true
	f.Problem = "differs from smith's standard: " + strings.Join(files, "; ")
	f.Detail = strings.Join(reasons, " ")
	accept := "[[doctor.accept]]\ncheck = \"" + check + "\""
	if project != "" {
		accept += "\nproject = \"" + project + "\""
	}
	f.Remedy = "smith doctor --fix writes smith's standard (" + strings.Join(files, "; ") + ")\nor keep yours: add to smith.toml\n" + accept + "\nreason = \"<why yours stays>\""
	f.Fix = fix
	return f
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
