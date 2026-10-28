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

// publishedTo lists where a package went: its registry, container images, a
// GitHub Release, a split repository; else why nothing was published.
func publishedTo(res release.ReportResult) string {
	var where []string
	if res.Published && res.Registry != "" {
		where = append(where, registryHost(res.Registry))
	}
	if n := len(res.Images); n > 0 {
		where = append(where, plural(n, "image"))
	}
	if res.GitHubRelease != "" {
		where = append(where, "GitHub Release")
	}
	if res.Mirror != "" {
		where = append(where, res.Mirror)
	}
	switch {
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

// registryHost is a registry URL as a reader names it: registry.npmjs.org.
func registryHost(url string) string {
	_, rest, found := strings.Cut(url, "://")
	if !found {
		rest = url
	}
	return strings.TrimSuffix(rest, "/")
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
