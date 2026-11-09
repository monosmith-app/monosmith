package cipipeline

import (
	"fmt"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/release"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// releaseIntro says what the release is about to do and the plan in one line.
func releaseIntro(plan release.Plan, branch string, projects int) []string {
	names := make([]string, len(plan.Items))
	for i, item := range plan.Items {
		names[i] = item.Project + " " + item.Next
	}
	return []string{
		fmt.Sprintf("Releasing %s of %d from %s, in dependency order: each is versioned, built, published and tagged.",
			plural(len(plan.Items), "project"), projects, branch),
		"Plan: " + listSome(names, 8),
	}
}

// releaseTable is the one summary of a release, a row per package: the
// version it was and becomes, how far it moved and why, where it went, its tag
// and the time. Failures come first, then the rest in release order.
func releaseTable(plan release.Plan, report release.Report, icons theme.Icons) resultTable {
	t := resultTable{
		Headers: []string{"Project", "Was", "Now", "Bump", "Why", "Published", "Tag", "Time"},
		Right:   []bool{false, false, false, false, false, false, false, true},
	}
	results := map[string]release.ReportResult{}
	for _, res := range report.Results {
		results[res.Project] = res
	}
	var failed, rest []tableRow
	for _, item := range plan.Items {
		res, ran := results[item.Project]
		was := item.Current
		if was == "" {
			was = "new"
		}
		bump := bumpOf(item)
		why := item.Why()
		whyTint := tintNone
		switch {
		case item.Forced:
			why, whyTint = "version() directive", tintYellow
		case len(item.Changes) == 0 && strings.HasPrefix(item.Reason, "dependency "):
			whyTint = tintYellow
		}
		row := tableRow{Tints: map[int]tint{1: tintDim, 2: tintGreen, 3: bumpTint(bump), 4: whyTint}}
		row.Cells = []string{item.Project, was, icons.Pointer + " " + item.Next, bump, why, "", "-", "-"}
		switch {
		case !ran:
			row.Cells[5], row.Tints[5] = icons.Queued+" not reached", tintDim
		case res.Status == "failed":
			row.Cells[5], row.Bad = icons.Failed+" failed", true
		case res.Status == "already-released":
			row.Cells[5], row.Tints[5] = icons.Skipped+" already released", tintYellow
			row.Cells[6] = tagOf(res)
		default:
			row.Cells[5], row.Cells[6] = icons.Passed+" "+publishedTo(res), tagOf(res)
		}
		if ran && res.TookMs > 0 {
			row.Cells[7] = formatDuration(time.Duration(res.TookMs) * time.Millisecond)
		}
		if row.Bad {
			failed = append(failed, row)
		} else {
			rest = append(rest, row)
		}
	}
	t.Rows = append(failed, rest...)
	return t
}

// bumpOf is how far the version moved: major, minor or patch.
func bumpOf(item release.Item) string {
	was, errWas := release.ParseVersion(item.Current)
	now, errNow := release.ParseVersion(item.Next)
	switch {
	case errWas != nil || errNow != nil:
		if item.Level == "forced" || item.Level == "" {
			return "-"
		}
		return item.Level
	case now.Major != was.Major:
		return "major"
	case now.Minor != was.Minor:
		return "minor"
	}
	return "patch"
}

func bumpTint(bump string) tint {
	switch bump {
	case "major":
		return tintRed
	case "minor":
		return tintYellow
	case "patch":
		return tintGreen
	}
	return tintNone
}

// unchangedLine says which packages had nothing to release: their names when
// there are few, else a count.
func unchangedLine(names []string) string {
	if len(names) == 0 {
		return ""
	}
	if len(names) <= 10 {
		return fmt.Sprintf("Unchanged (%d): %s", len(names), strings.Join(names, ", "))
	}
	return fmt.Sprintf("Unchanged: %s have no releasing commit", plural(len(names), "project"))
}

// publishedTo lists every target a package reached: its registries, container
// images, GitHub Release, split repository and pub.dev. Up to three are listed
// by name; four or more are counted first ("5 targets: npm, PyPI, ..."). When
// there are none it says why nothing was published.
func publishedTo(res release.ReportResult) string {
	where := release.TargetLabels(res)
	switch {
	case len(where) > 3:
		return fmt.Sprintf("%d targets: %s", len(where), strings.Join(where, ", "))
	case len(where) > 0:
		return strings.Join(where, ", ")
	case res.Publish == "published by its tag":
		return "its tag"
	case res.Publish != "":
		return "nothing (" + res.Publish + ")"
	case res.Status == "failed":
		return "-"
	}
	return "nothing to publish"
}

func tagOf(res release.ReportResult) string {
	switch {
	case res.Tag == "":
		return "-"
	case res.Pushed:
		return res.Tag + " (pushed)"
	}
	return res.Tag + " (local)"
}

// releaseConclusion is the end of the closing line: what was released.
func releaseConclusion(report release.Report) string {
	var names []string
	failed, again := 0, 0
	for _, res := range report.Results {
		switch res.Status {
		case "released":
			names = append(names, res.Project+" "+res.Next)
		case "failed":
			failed++
		case "already-released":
			again++
		}
	}
	line := fmt.Sprintf("%d released", len(names))
	if len(names) > 0 {
		line += " (" + strings.Join(names, ", ") + ")"
	}
	if again > 0 {
		line += fmt.Sprintf(", %d already released", again)
	}
	return line + fmt.Sprintf(", %d unchanged, %d failed", len(report.Unchanged), failed)
}

// releasedSummary is the "what was released" block under the result table:
// a banner, then each released project's targets. The platform lines are dim.
func releasedSummary(report release.Report, icons theme.Icons, colour bool) []string {
	var out []string
	for _, line := range report.SummaryLines(icons.Passed == theme.ASCII.Passed) {
		text := line.Text
		if line.Detail {
			text = paint(text, tintDim, colour)
		}
		out = append(out, text)
	}
	return out
}
