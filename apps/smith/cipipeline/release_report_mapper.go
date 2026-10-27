package cipipeline

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/release"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// planTable is the release decision, one row per package: the version it was,
// the version it becomes and the commit that decided it.
func planTable(plan release.Plan) resultTable {
	t := resultTable{Headers: []string{"Package", "Was", "Now", "Why"}}
	for _, item := range plan.Items {
		was := item.Current
		if was == "" {
			was = "new"
		}
		t.Rows = append(t.Rows, tableRow{Cells: []string{item.Project, was, item.Next, item.Why()}})
	}
	return t
}

// releaseIntro says what the release is about to do.
func releaseIntro(plan release.Plan, branch string, projects int) []string {
	return []string{fmt.Sprintf("Releasing %s of %d from %s, in dependency order: each is versioned, built, published and tagged.",
		plural(len(plan.Items), "project"), projects, branch)}
}

// releaseResultTable is what happened to each released package: the outcome,
// where it was published and the tag.
func releaseResultTable(report release.Report, icons theme.Icons) resultTable {
	t := resultTable{Headers: []string{"Package", "Result", "Published to", "Tag"}}
	for _, res := range report.Results {
		row := tableRow{Cells: []string{res.Project, icons.Passed + " released", publishedTo(res), tagOf(res)}}
		switch res.Status {
		case "already-released":
			row.Cells[1] = icons.Skipped + " already released"
		case "failed":
			row.Cells[1], row.Bad = icons.Failed+" failed", true
		}
		t.Rows = append(t.Rows, row)
	}
	return t
}

// publishedTo lists where a package went: its registry, container images, a
// GitHub Release, a split repository; else why nothing was published.
func publishedTo(res release.ReportResult) string {
	var where []string
	if res.Published && res.Registry != "" {
		where = append(where, res.Registry)
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
	for _, res := range report.Results {
		if res.Status == "released" {
			names = append(names, res.Project+" "+res.Next)
		}
	}
	line := fmt.Sprintf("%s released", plural(len(names), "project"))
	if len(names) > 0 {
		line += " (" + strings.Join(names, ", ") + ")"
	}
	if n := len(report.Unchanged); n > 0 {
		line += fmt.Sprintf(", %s unchanged", plural(n, "project"))
	}
	return line
}
