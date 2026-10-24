package tui

import (
	"fmt"
	"slices"

	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
)

// confirmDetails are the lines a confirmation dialog shows between its
// sentence and the command line: what the command is going to touch, as far
// as the views already know it (the Release plan, the Doctor findings). The
// list is cut to fit the screen, with a count of what was left out.
func (s Shell) confirmDetails(argv []string) []string {
	if len(argv) == 0 {
		return nil
	}
	var lines []string
	switch {
	case argv[0] == "release":
		lines = s.releaseDetails()
	case argv[0] == "doctor" && slices.Contains(argv, "--fix"):
		only := ""
		if i := slices.Index(argv, "--only"); i >= 0 && i+1 < len(argv) {
			only = argv[i+1]
		}
		lines = s.doctorFixDetails(only)
	}
	return capLines(lines, max(s.h-16, 3))
}

// releaseDetails lists the packages and the versions the release publishes.
func (s Shell) releaseDetails() []string {
	if !s.release.Loaded || len(s.release.Plan.Items) == 0 {
		return []string{"The plan is not loaded: open the Release view first (" + commands.ShortcutOf("view.release") + ") to see the packages and versions."}
	}
	lines := []string{"Packages and versions to publish:"}
	for _, item := range s.release.Plan.Items {
		from := item.Current
		if from == "" {
			from = "new"
		}
		lines = append(lines, fmt.Sprintf("  %s  %s → %s", item.Project, from, item.Next))
	}
	return lines
}

// doctorFixDetails lists the files the fix of one check, or of every check
// with a safe fix, writes or removes.
func (s Shell) doctorFixDetails(only string) []string {
	if s.doctor.Report == nil {
		return nil
	}
	var lines []string
	seen := map[string]bool{}
	for _, f := range s.doctor.Report.Findings {
		if f.OK || f.Fix == nil || !f.Fix.Safe || (only != "" && f.Check != only) {
			continue
		}
		for _, e := range f.Fix.Files {
			verb := "writes "
			if e.Delete {
				verb = "removes "
			}
			if line := "  " + verb + e.Path; !seen[line] {
				seen[line] = true
				lines = append(lines, line)
			}
		}
	}
	if len(lines) == 0 {
		return []string{"This fix lists no files: `smith doctor --fix` shows what it changes first."}
	}
	return append([]string{"Files that will change:"}, lines...)
}

// capLines keeps at most n lines, replacing the rest with a count.
func capLines(lines []string, n int) []string {
	if len(lines) <= n {
		return lines
	}
	out := append([]string(nil), lines[:n-1]...)
	return append(out, fmt.Sprintf("  … and %d more", len(lines)-(n-1)))
}
