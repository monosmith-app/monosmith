package dashboard

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// boxGap is the cells between two boxes of a row.
const boxGap = 1

// box is one summary box: a muted label, the value with its mark in the
// colour of its state, and a muted sub-line.
type box struct {
	label, value, sub string
}

// boxes draws the four summary boxes, joined side by side with Lip Gloss: in
// one row on a wide page, in a 2x2 grid on a narrow one, where they keep only
// the label and the value. The Last run box opens the newest run on a click.
func (c *canvas) boxes(f Facts) {
	all := []box{c.changedBox(f), c.runBox(f), c.healthBox(f), c.releaseBox(f)}
	per, h := 4, 5
	if !c.wide {
		per, h = 2, 4
	}
	// The row fills the page: the first boxes take the cells the division leaves.
	room := c.w - 2 - (per-1)*boxGap
	for start := 0; start < len(all); start += per {
		row, x := len(c.lines), 1
		var blocks []string
		for i, b := range all[start:min(start+per, len(all))] {
			if i > 0 {
				blocks = append(blocks, strings.Repeat(" ", boxGap))
				x += boxGap
			}
			bw := max(room/per, 4)
			if i < room%per {
				bw++
			}
			body := []string{" " + c.th.Muted(b.label), " " + b.value}
			if c.wide {
				body = append(body, " "+c.th.Muted(b.sub))
			}
			blocks = append(blocks, strings.Join(pane.Frame(c.th, "", body, bw, h, false), "\n"))
			if start+i == 1 && len(f.Recent) > 0 {
				for y := range h {
					c.targets = append(c.targets, Target{Row: row + y, From: x, To: x + bw, Kind: TargetRun, Run: f.Recent[0].Index, Item: -1})
				}
			}
			x += bw
		}
		for _, line := range strings.Split(lipgloss.JoinHorizontal(lipgloss.Top, blocks...), "\n") {
			c.add(" " + line)
		}
	}
}

// waiting is the value of a box whose facts are still loading: the spinner
// and a muted word, as the other views wait.
func (c *canvas) waiting(f Facts) string {
	return c.th.Accent(c.th.Spin(f.Tick)) + " " + c.th.Muted(words.Say("dashboard.box.loading"))
}

// value is a box value: the mark and the text in bold, in the colour of s.
func (c *canvas) value(s Severity, mark, text string) string {
	return c.th.Bold(c.paint(s, mark+" "+text))
}

// changedBox tells how many projects the changes affect.
func (c *canvas) changedBox(f Facts) box {
	b := box{label: words.Say("dashboard.box.changed")}
	switch {
	case !f.ProjectsLoaded:
		b.value = c.waiting(f)
	case f.Projects == 0:
		b.value = c.value(SeverityMuted, c.th.Icons.Queued, words.Say("dashboard.box.changed.empty"))
	case f.Affected == 0:
		b.value = c.value(SeveritySuccess, c.th.Icons.Passed, words.Say("dashboard.box.changed.none"))
	default:
		b.value = c.value(SeverityInfo, c.th.Icons.Dot, words.Count("dashboard.project", f.Affected))
	}
	if f.ProjectsLoaded && f.Projects > 0 && f.Base != "" {
		b.sub = words.Say("dashboard.box.changed.since", "base", f.Base)
	}
	return b
}

// runBox tells how the newest run ended: what failed, or what passed and was cached.
func (c *canvas) runBox(f Facts) box {
	th := c.th
	b := box{label: words.Say("dashboard.box.run")}
	if len(f.Recent) == 0 {
		b.value = c.value(SeverityMuted, th.Icons.Queued, words.Say("dashboard.box.run.none"))
		return b
	}
	r := f.Recent[0]
	counted := r.Passed+r.Cached+r.Failed > 0
	tally := []string{}
	if r.Passed > 0 {
		tally = append(tally, words.Count("dashboard.box.run.passed", r.Passed))
	}
	if r.Cached > 0 {
		tally = append(tally, words.Count("dashboard.box.run.cached", r.Cached))
	}
	switch {
	case r.Cancelled:
		b.value = c.value(SeverityMuted, th.Icons.Cancelled, words.Say("dashboard.box.run.stopped"))
		b.sub = shortCommand(r.Command)
	case !r.OK && r.Failed > 0:
		b.value = c.value(SeverityFailure, th.Icons.Failed, words.Count("dashboard.box.run.failed", r.Failed))
		b.sub = strings.Join(tally, " · ")
	case !r.OK:
		b.value = c.value(SeverityFailure, th.Icons.Failed, words.Say("dashboard.box.run.failed_plain"))
		b.sub = shortCommand(r.Command)
	case counted:
		b.value = c.value(SeveritySuccess, th.Icons.Passed, tally[0])
		b.sub = strings.Join(append(tally[1:], r.Took), " · ")
	default:
		b.value = c.value(SeveritySuccess, th.Icons.Passed, words.Say("dashboard.box.run.passed_plain"))
		b.sub = shortCommand(r.Command)
	}
	b.sub = strings.Trim(b.sub, " ·")
	return b
}

// healthBox tells what the checks found.
func (c *canvas) healthBox(f Facts) box {
	th := c.th
	b := box{label: words.Say("dashboard.box.health")}
	switch {
	case f.CheckErr != "":
		b.value = c.value(SeverityFailure, th.Icons.Failed, words.Say("dashboard.box.health.error"))
	case !f.Checked:
		b.value = c.waiting(f)
	case f.Failed > 0:
		b.value = c.value(SeverityFailure, th.Icons.Failed, words.Count("dashboard.box.health.failed", f.Failed))
		b.sub = words.Count("dashboard.box.health.warnings", f.Warnings)
	case f.Warnings > 0:
		b.value = c.value(SeverityWarning, th.Icons.Warning, words.Count("dashboard.box.health.warnings", f.Warnings))
		b.sub = words.Count("dashboard.box.health.failed_checks", 0)
	default:
		b.value = c.value(SeveritySuccess, th.Icons.Passed, words.Say("dashboard.box.health.pass"))
		b.sub = words.Count("dashboard.box.health.passed", f.Passed)
	}
	return b
}

// releaseBox tells how many packages are ready to release and the mix of their bumps.
func (c *canvas) releaseBox(f Facts) box {
	th := c.th
	b := box{label: words.Say("dashboard.box.release")}
	switch r := f.Release; {
	case r != nil:
		b.value = c.value(SeveritySuccess, th.Icons.Passed, words.Count("dashboard.box.release.ready", r.Packages))
		var mix []string
		for _, n := range []struct {
			count int
			text  string
		}{{r.Major, words.Say("dashboard.box.release.major", "count", r.Major)}, {r.Minor, words.Say("dashboard.box.release.minor", "count", r.Minor)}, {r.Patch, words.Say("dashboard.box.release.patch", "count", r.Patch)}} {
			if n.count > 0 {
				mix = append(mix, n.text)
			}
		}
		b.sub = strings.Join(mix, " · ")
	case f.ReleaseLoading:
		b.value = c.waiting(f)
	default:
		b.value = c.value(SeverityMuted, th.Icons.Queued, words.Say("dashboard.box.release.none"))
	}
	return b
}
