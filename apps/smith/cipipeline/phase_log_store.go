package cipipeline

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// bannerWidth keeps a banner readable in a narrow log viewer.
const bannerWidth = 60

// phaseTitles are how the phases are named to a reader.
var phaseTitles = map[string]string{
	"setup": "Setup", "audit": "Audit", "verify": "Verify", "pack": "Pack", "native": "Native build", "release": "Release",
}

// phaseLog is what a phase says to the pipeline log: a banner that says what
// it will do, the tools' own output folded in a collapsed group, then plain
// result lines and a closing line. Everything a reader needs stays visible; the
// wall of tool output does not. The group closes before the first line said
// after it and at the end of the phase, so a phase that fails early never
// leaves one open. Colour is written only where colourWanted allows it, never cursor movement.
type phaseLog struct {
	out   io.Writer
	phase string
	icons theme.Icons
	open  bool
	// conclusion is what the closing line says after the time ("7 steps ran").
	conclusion string
	// markdown is the phase's result for the CI's job summary.
	markdown []string
}

func newPhaseLog(out io.Writer, phase string) *phaseLog {
	return &phaseLog{out: out, phase: phase, icons: theme.Choose(os.Getenv, true).Icons}
}

func (l *phaseLog) title() string {
	if t, ok := phaseTitles[l.phase]; ok {
		return t
	}
	return l.phase
}

// banner marks the start of the phase and says, in plain words, what it will
// do; what follows is folded.
func (l *phaseLog) banner(lines ...string) {
	l.heading(lines...)
	l.fold()
}

// heading is a banner whose lines are followed by more visible lines (a plan),
// before anything is folded.
func (l *phaseLog) heading(lines ...string) {
	rule := "== " + l.title() + " =="
	if l.icons.Passed != theme.ASCII.Passed {
		rule = "══ " + l.title() + " "
		rule += strings.Repeat("═", max(bannerWidth-len([]rune(rule)), 3))
	}
	l.say("%s", rule)
	for _, line := range lines {
		l.say("%s", line)
	}
}

// fold folds what follows.
func (l *phaseLog) fold() {
	group(l.out, "smith ci "+l.phase+": output")
	l.open = true
}

// say prints a line outside the folded output.
func (l *phaseLog) say(format string, args ...any) {
	l.close()
	fmt.Fprintf(l.out, format+"\n", args...)
}

// close ends the folded output, if any.
func (l *phaseLog) close() {
	if l.open {
		endGroup(l.out)
		l.open = false
	}
}

// finish prints the closing line: the outcome, the time, and on a failure the
// way forward.
func (l *phaseLog) finish(err error, took time.Duration) {
	l.close()
	if err != nil {
		l.say("%s %s failed after %s", paint(l.icons.Failed, tintRed, colourOn()), l.title(), formatDuration(took))
		l.say("→ the output above shows the cause; repeat the phase with `smith ci %s`", l.phase)
		return
	}
	line := fmt.Sprintf("%s %s passed in %s", paint(l.icons.Passed, tintGreen, colourOn()), l.title(), formatDuration(took))
	if l.conclusion != "" {
		line += ": " + l.conclusion
	}
	l.say("%s", line)
}
