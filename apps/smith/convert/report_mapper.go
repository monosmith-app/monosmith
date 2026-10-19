package convert

import (
	"fmt"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/convertplan"
)

var sections = []struct {
	severity convertplan.Severity
	title    string
}{
	{convertplan.Blocker, "Blocking the conversion"},
	{convertplan.Manual, "Needs you"},
	{convertplan.Unsupported, "Not supported by smith yet"},
	{convertplan.Changed, "Behaves differently"},
	{convertplan.Mapped, "Converted"},
	{convertplan.Dropped, "Dropped"},
}

// sourceLabel names the workspace a plan converts from.
func sourceLabel(plan *convertplan.Plan) string {
	switch {
	case plan.Source == convertplan.SourceTurborepo:
		return "Turborepo workspace"
	case plan.Legacy:
		return "Nx workspace made by the old mnci"
	}
	return "Nx workspace"
}

// Report renders a plan for a terminal.
func Report(plan *convertplan.Plan, dryRun bool) string {
	var b strings.Builder
	kind := sourceLabel(plan)
	verb := "Converting"
	if dryRun {
		verb = "Conversion plan for"
	}
	fmt.Fprintf(&b, "%s %s %q\n\n", verb, kind, plan.Config.Workspace.Name)
	b.WriteString("Projects:\n")
	for _, p := range plan.Projects {
		if !p.Supported {
			fmt.Fprintf(&b, "  %-20s %-28s not supported\n", p.Name, p.Root)
			continue
		}
		fmt.Fprintf(&b, "  %-20s %-28s %s %s\n", p.Name, p.Root, p.Language, p.Kind)
	}
	for _, s := range sections {
		var lines []string
		for _, f := range plan.Findings {
			if f.Severity != s.severity {
				continue
			}
			where := f.Subject
			if f.Project != "" {
				where = f.Project + " " + f.Subject
			}
			lines = append(lines, fmt.Sprintf("  - %s: %s", where, f.Message))
		}
		if len(lines) == 0 {
			continue
		}
		sort.Strings(lines)
		title := s.title
		if s.severity == convertplan.Dropped {
			title += " (" + dropReason(plan) + ")"
		}
		fmt.Fprintf(&b, "\n%s (%d):\n%s\n", title, len(lines), strings.Join(lines, "\n"))
	}
	if len(plan.Delete) > 0 {
		fmt.Fprintf(&b, "\nRemoves: %s\n", strings.Join(plan.Delete, ", "))
	}
	return b.String()
}

// dropReason says why what was dropped is not needed.
func dropReason(plan *convertplan.Plan) string {
	if plan.Source == convertplan.SourceTurborepo {
		return "Turborepo, pnpm and yarn only"
	}
	return "Nx-only"
}
