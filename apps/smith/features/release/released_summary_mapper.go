package release

import (
	"fmt"
	"strings"
)

// bannerWidth keeps a banner readable in a narrow log viewer.
const bannerWidth = 60

// maxTargetLines bounds the lines one project takes in the summary.
const maxTargetLines = 12

// SummaryLine is one line of the "what was released" block. Detail lines
// continue the line above (the platforms of a package list); a caller may
// paint them dim.
type SummaryLine struct {
	Text   string
	Detail bool
}

// Rule is a banner rule with the title inside it: "══ Released ═══" or, in
// ASCII, "== Released ==".
func Rule(title string, ascii bool) string {
	if ascii {
		return "== " + title + " =="
	}
	rule := "══ " + title + " "
	return rule + strings.Repeat("═", max(bannerWidth-len([]rune(rule)), 3))
}

// bumpName is how far a result's version moved: major, minor, patch, or "".
func bumpName(item Item) string {
	was, errWas := ParseVersion(item.Current)
	now, errNow := ParseVersion(item.Next)
	switch {
	case errWas != nil || errNow != nil:
		if item.Level == "forced" {
			return ""
		}
		return item.Level
	case now.Major != was.Major:
		return "major"
	case now.Minor != was.Minor:
		return "minor"
	}
	return "patch"
}

// releasedHeader names a project and its version step: "smith 1.128.0 ›
// 1.129.0 (minor)".
func releasedHeader(item Item) string {
	was := item.Current
	if was == "" {
		was = "new"
	}
	line := fmt.Sprintf("%s %s › %s", item.Project, was, item.Next)
	if bump := bumpName(item); bump != "" {
		line += " (" + bump + ")"
	}
	return line
}

// releasedBlocks pairs each released project with the targets it reached; a
// dry run, a failure and a project that reached nothing have no block.
func (r Report) releasedBlocks() (items []Item, targets [][]ReleasedTarget) {
	if r.DryRun {
		return nil, nil
	}
	for _, res := range r.Results {
		if t := TargetsOf(res); res.Status == "released" && len(t) > 0 {
			items, targets = append(items, res.Item), append(targets, t)
		}
	}
	return items, targets
}

// shortKind is the name of a target in the first column.
func shortKind(kind string) string {
	if kind == KindGitHubRelease {
		return "GitHub"
	}
	return kind
}

// SummaryLines is the "what was released" block for a terminal or a log: a
// banner, then for each released project its version step and a line per
// target (type, what, where) with the platforms under a package line. It is
// empty when nothing was released to a target.
func (r Report) SummaryLines(ascii bool) []SummaryLine {
	items, all := r.releasedBlocks()
	if len(items) == 0 {
		return nil
	}
	lines := []SummaryLine{{Text: Rule("Released", ascii)}}
	for i, item := range items {
		lines = append(lines, SummaryLine{Text: releasedHeader(item)})
		targets := all[i]
		for n, t := range targets {
			if n == maxTargetLines {
				lines = append(lines, SummaryLine{Text: fmt.Sprintf("  ... %d more targets", len(targets)-n)})
				break
			}
			text := fmt.Sprintf("  %-10s %s", shortKind(t.Kind), t.What)
			if t.Where != "" && t.Kind != KindGitHubRelease {
				text += "    " + t.Where
			}
			lines = append(lines, SummaryLine{Text: text})
			if t.detail != "" {
				lines = append(lines, SummaryLine{Text: strings.Repeat(" ", 15) + t.detail, Detail: true})
			}
		}
	}
	return lines
}

// SummaryMarkdown is the same block for a job summary: a heading, then for
// each released project its version step and a list of its targets. It is empty
// when nothing was released to a target.
func (r Report) SummaryMarkdown() string {
	items, all := r.releasedBlocks()
	if len(items) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("\n#### Released\n")
	for i, item := range items {
		was := item.Current
		if was == "" {
			was = "new"
		}
		fmt.Fprintf(&b, "\n**%s** %s → **%s**", item.Project, was, item.Next)
		if bump := bumpName(item); bump != "" {
			fmt.Fprintf(&b, " (%s)", bump)
		}
		b.WriteString("\n\n")
		for _, t := range all[i] {
			fmt.Fprintf(&b, "- **%s** %s", shortKind(t.Kind), t.What)
			if t.Where != "" && t.Kind != KindGitHubRelease {
				fmt.Fprintf(&b, " (%s)", t.Where)
			}
			if t.Kind == KindGitHubRelease && t.Where != "" {
				fmt.Fprintf(&b, " ([open](%s))", t.Where)
			}
			if t.detail != "" {
				b.WriteString("  \n  " + t.detail)
			}
			b.WriteString("\n")
		}
	}
	return b.String()
}
