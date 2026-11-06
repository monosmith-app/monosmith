package tui

import (
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/release"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/libs/words/words"
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
	case isForcedRelease(argv):
		lines = forcedReleaseDetails(argv)
	case argv[0] == "uninstall":
		lines = uninstallDetails(argv)
	case argv[0] == "release":
		lines = s.releaseDetails(argv)
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
func (s Shell) releaseDetails(argv []string) []string {
	if !s.release.Loaded || len(s.release.Plan.Items) == 0 {
		return []string{words.Say("tui.confirm_details.the_plan_is_not_loaded_open", "shortcutOf", commands.ShortcutOf("view.release"))}
	}
	var lines []string
	if s.release.Plan.Channel != "" {
		lines = append(lines, release.ChannelNotice(s.release.Plan.Channel), "")
	}
	plan := s.release.Plan
	var held []string
	if adjusted, _, err := release.Adjust(plan, adjustmentsIn(argv)); err == nil {
		for _, item := range plan.Items {
			if !slices.ContainsFunc(adjusted.Items, func(a release.Item) bool { return a.Project == item.Project }) {
				held = append(held, item.Project)
			}
		}
		plan = adjusted
	}
	lines = append(lines, words.Say("tui.confirm_details.packages_and_versions_to_publish"))
	for _, item := range plan.Items {
		from := item.Current
		if from == "" {
			from = "new"
		}
		lines = append(lines, words.Say("tui.confirm_details.text", "project", item.Project, "from", from, "next", item.Next))
	}
	if len(held) > 0 {
		lines = append(lines, "", words.Say("tui.confirm_details.held_back_not_released", "list", strings.Join(held, ", ")))
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
			verb := words.Say("tui.confirm_details.writes")
			if e.Delete {
				verb = words.Say("tui.confirm_details.removes")
			}
			if line := "  " + verb + e.Path; !seen[line] {
				seen[line] = true
				lines = append(lines, line)
			}
		}
	}
	if len(lines) == 0 {
		return []string{words.Say("tui.confirm_details.no_files")}
	}
	return append([]string{words.Say("tui.confirm_details.files_change")}, lines...)
}

// capLines keeps at most n lines, replacing the rest with a count.
func capLines(lines []string, n int) []string {
	if len(lines) <= n {
		return lines
	}
	out := append([]string(nil), lines[:n-1]...)
	return append(out, words.Say("tui.confirm_details.and_more", "value", len(lines)-(n-1)))
}

// adjustmentsIn reads the --hold and --bump flags of a release command line.
func adjustmentsIn(argv []string) release.Adjustments {
	var a release.Adjustments
	var bumps []string
	for i := 0; i+1 < len(argv); i++ {
		switch argv[i] {
		case "--hold":
			a.Hold = append(a.Hold, strings.Split(argv[i+1], ",")...)
		case "--bump":
			bumps = append(bumps, argv[i+1])
		}
	}
	a.Bump, _ = release.ParseBumps(bumps)
	return a
}
