package releaseview

import (
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/release"
)

// maxCommits bounds the commits listed under one bump so a long history does
// not push the other projects off the screen.
const maxCommits = 12

// Model shows what the next release would do, project by project. Enter
// drills down to the commits behind the selected bump; releasing itself runs
// `smith release` in the run view, after a confirmation that shows the command.
type Model struct {
	Loaded bool
	// Waited is how long the plan has taken so far, set by the shell for drawing.
	Waited time.Duration
	// Projects is how many projects the plan covers, set by the shell for drawing (0 when unknown).
	Projects int
	Plan     release.Plan
	Blockers []string
	Warnings []string
	Err      error
	// Cursor is the project selected in the plan, Open is true while its
	// commits are shown.
	Cursor int
	Open   bool
}

// Result is what planning the release found: the shell receives it as a message.
type Result struct {
	Plan     release.Plan
	Blockers []string
	Warnings []string
	Err      error
}

// FromResult is the model for a planned release.
func FromResult(r Result) Model {
	return Model{Loaded: true, Plan: r.Plan, Blockers: r.Blockers, Warnings: r.Warnings, Err: r.Err}
}

// Update handles a key.
func (v Model) Update(key tea.KeyMsg) Model {
	n := len(v.Plan.Items)
	if !v.Loaded || v.Err != nil || len(v.Blockers) > 0 || n == 0 {
		return v
	}
	switch commands.Nav(key.String()) {
	case commands.NavUp:
		v.Cursor, v.Open = (v.Cursor-1+n)%n, false
	case commands.NavDown:
		v.Cursor, v.Open = (v.Cursor+1)%n, false
	case commands.NavHome:
		v.Cursor, v.Open = 0, false
	case commands.NavEnd:
		v.Cursor, v.Open = n-1, false
	case commands.NavConfirm:
		v.Open = !v.Open
	}
	return v
}

// SelectRow moves the cursor to item i and opens or closes its commits when it was already there.
func (v Model) SelectRow(i int) Model {
	if i < 0 || i >= len(v.Plan.Items) {
		return v
	}
	if v.Cursor == i {
		v.Open = !v.Open
	} else {
		v.Cursor, v.Open = i, false
	}
	return v
}

// Selected is the project under the cursor.
func (v Model) Selected() (release.Item, bool) {
	if v.Cursor < 0 || v.Cursor >= len(v.Plan.Items) {
		return release.Item{}, false
	}
	return v.Plan.Items[v.Cursor], true
}

// CanRelease reports whether a release can be started from the plan.
func (v Model) CanRelease() bool {
	return v.Loaded && v.Err == nil && len(v.Blockers) == 0 && len(v.Plan.Items) > 0
}
