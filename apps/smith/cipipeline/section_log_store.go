package cipipeline

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// sectionRow is what the summary says about one finished section.
type sectionRow struct {
	name     string
	projects int
	err      error
	took     time.Duration
}

// SectionLog gives each toolchain of one command (smith install) its own
// section in the log, in the order the toolchains run, so the output of npm,
// pip and go never mixes. A section opens with a rule that holds its name,
// then the toolchain's own lines follow, then one status line closes it. In a
// CI log that folds groups, the section is a group instead of a rule. The
// summary after the last section has one row per toolchain.
//
// A SectionLog is for one goroutine: toolchains run one after the other.
// Colour follows colourWanted and the symbols follow the terminal's icon set;
// nothing moves the cursor.
type SectionLog struct {
	out    io.Writer
	icons  theme.Icons
	colour bool
	now    func() time.Time

	// Remedy is the line printed after a failed section, starting with "→".
	Remedy string

	open  bool
	name  string
	count int
	began time.Time
	rows  []sectionRow
}

// defaultSectionRemedy is the way forward after a failed section when the
// other sections still run (smith install).
const defaultSectionRemedy = "→ fix the cause above, then run it again; the other toolchains are not affected"

// NewSectionLog starts a section log that writes to out. The icon set and the
// colour decision come from the process environment.
//
// Parameters: out receives the rules, status lines and summary.
// Returns: an empty log.
func NewSectionLog(out io.Writer) *SectionLog {
	return &SectionLog{out: out, icons: theme.Choose(os.Getenv, true).Icons, colour: colourOn(), now: time.Now, Remedy: defaultSectionRemedy}
}

// Begin opens the section of one toolchain. A section still open is closed as
// failed first, so a caller that forgets End never leaves a group open.
//
// Parameters: name is the toolchain (the plugin: node, python, go); projects
// is how many projects it covers, or -1 when that is not known.
func (l *SectionLog) Begin(name string, projects int) {
	if l.open {
		l.End(errors.New("the section ended without a result"))
	}
	l.open, l.name, l.count, l.began = true, name, projects, l.now()
	title := name
	if projects >= 0 {
		title = fmt.Sprintf("%s (%s)", name, plural(projects, "project"))
	}
	if inGroupedLog() {
		group(l.out, title)
		return
	}
	fmt.Fprintln(l.out, l.rule(title))
}

// rule is a banner line: the title inside a rule at most bannerWidth wide,
// or "== title ==" when the icon set is ASCII.
func (l *SectionLog) rule(title string) string {
	if l.icons.Passed == theme.ASCII.Passed {
		return "== " + title + " =="
	}
	rule := "══ " + title + " "
	return rule + strings.Repeat("═", max(bannerWidth-len([]rune(rule)), 3))
}

// End closes the open section with one status line: done and the time, or
// failed with the cause and, on the next line, the way forward. Without an
// open section it does nothing.
//
// Parameters: err is why the toolchain failed, or nil when it succeeded.
func (l *SectionLog) End(err error) {
	if !l.open {
		return
	}
	l.open = false
	took := l.now().Sub(l.began)
	l.rows = append(l.rows, sectionRow{name: l.name, projects: l.count, err: err, took: took})
	if inGroupedLog() {
		endGroup(l.out)
	}
	if err == nil {
		fmt.Fprintf(l.out, "%s %s done in %s\n", paint(l.icons.Passed, tintGreen, l.colour), l.name, formatDuration(took))
		return
	}
	fmt.Fprintf(l.out, "%s %s failed: %s\n", paint(l.icons.Failed, tintRed, l.colour), l.name, firstLine(err))
	remedy := l.Remedy
	if remedy == "" {
		remedy = defaultSectionRemedy
	}
	fmt.Fprintln(l.out, remedy)
}

// Summary prints what ran, one row per toolchain in the order they ran
// (toolchain, projects, status, time). With no section run it prints nothing.
//
// Returns: how many sections failed.
func (l *SectionLog) Summary() int {
	if l.open {
		l.End(errors.New("the section ended without a result"))
	}
	if len(l.rows) == 0 {
		return 0
	}
	t := resultTable{Headers: []string{"Toolchain", "Projects", "Status", "Time"}, Right: []bool{false, true, false, true}}
	failed := 0
	for _, r := range l.rows {
		status := l.icons.Passed + " done"
		row := tableRow{}
		if r.err != nil {
			status, row.Bad = l.icons.Failed+" failed", true
			failed++
		}
		count := "-"
		if r.projects >= 0 {
			count = fmt.Sprint(r.projects)
		}
		row.Cells = []string{r.name, count, status, formatDuration(r.took)}
		t.Rows = append(t.Rows, row)
	}
	fmt.Fprintln(l.out)
	for _, line := range t.render("done", l.colour) {
		fmt.Fprintln(l.out, line)
	}
	return failed
}

// firstLine is the error's first line, which is its cause; a long one is cut.
func firstLine(err error) string {
	msg := strings.TrimSpace(strings.SplitN(err.Error(), "\n", 2)[0])
	if r := []rune(msg); len(r) > 200 {
		msg = string(r[:199]) + "…"
	}
	return msg
}
