package tui

import (
	"slices"

	"github.com/monosmith-app/monosmith/apps/smith/features/doctor"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/button"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/doctorview"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/fixwizard"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/footer"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/layout"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// healthFixed is a fix of Health that was applied: the detail shows its third
// step, Check again, with the new state of the checks it clears.
type healthFixed struct {
	checks []string
	what   string
}

// inlineFix is true while the doctor fix wizard is drawn in the detail of
// Health instead of as a dialog: a fix of the workspace's checks, while Health
// shows its detail. The wizard of smith.toml and the deps repair stay dialogs.
func (s Shell) inlineFix() bool {
	return s.fix != nil && s.inWorkspace && s.repair == nil && s.depsRepair == nil &&
		s.tab == commands.TabDoctor && s.detailVisible()
}

// fixedShown is true while the detail of Health shows the step Check again of
// a fix that was applied.
func (s Shell) fixedShown() bool {
	return s.fixed != nil && s.fix == nil && s.tab == commands.TabDoctor && s.detailVisible()
}

// shownFixChecks are the checks of the fix the detail of Health shows (open or
// just applied), or nil while it shows none.
func (s Shell) shownFixChecks() []string {
	switch {
	case s.inlineFix():
		return s.fix.Fix.Checks
	case s.fixedShown():
		return s.fixed.checks
	}
	return nil
}

// fixFinding is the check a fix is about: the selected one when the fix clears
// it, else the first check of the report it clears. ok is false for none.
func (s Shell) fixFinding(checks []string) (doctor.Finding, bool) {
	if f, ok := s.doctor.Selected(); ok && slices.Contains(checks, f.Check) {
		return f, true
	}
	if s.doctor.Report != nil {
		for _, f := range s.doctor.Report.Findings {
			if slices.Contains(checks, f.Check) {
				return f, true
			}
		}
	}
	return doctor.Finding{}, false
}

// fixStep is the step the open wizard is on: the review while it asks or
// shows the changes, the apply while it runs or after it failed.
func fixStep(w fixwizard.Wizard) int {
	if p := w.Phase(); p == fixwizard.Applying || p == fixwizard.Failed {
		return doctorview.StepApply
	}
	return doctorview.StepReview
}

// fixHead is the top of the detail while a fix shows: the state of its check,
// what is wrong (at most three lines), how many checks the fix clears when it
// clears more than one, and the three steps with the current one lit.
func (s Shell) fixHead(checks []string, step, w int) []string {
	var out []string
	if f, ok := s.fixFinding(checks); ok {
		out = append(out, pane.Fit(" "+doctorview.StateLine(s.th, f), w), "")
		if f.Detail != "" {
			wrapped := pane.WrapWords(f.Detail, max(w-2, 1))
			if len(wrapped) > 3 {
				wrapped = append(wrapped[:2], pane.CutEllipsis(wrapped[2]+" "+wrapped[3], max(w-2, 1)))
			}
			for _, l := range wrapped {
				out = append(out, pane.Fit(" "+l, w))
			}
		}
	}
	if len(checks) > 1 {
		out = append(out, pane.Fit(" "+s.th.Muted(words.Say("healthfix.clears", "count", len(checks))), w))
	}
	if len(checks) == 1 && step == doctorview.StepReview {
		// The same fix in a terminal, so the command is learnt.
		for _, l := range pane.WrapWords(words.Say("tui.detail.or_in_a_terminal_smith_doctor", "check", checks[0]), max(w-2, 1)) {
			out = append(out, pane.Fit(" "+s.th.Muted(l), w))
		}
	}
	return append(out, "", pane.Fit(" "+doctorview.Steps(s.th, step), w), "")
}

// fixArea is where the inline wizard is drawn: the detail under the head, less
// the rows a toast keeps free. A detail too short for a head keeps all its
// rows for the wizard.
func (s Shell) fixArea() layout.Rect {
	r := s.detailRect()
	// The rows a toast keeps free at the bottom of the panes are no rows of the wizard.
	r.H -= s.toastRows(s.mainInner().H)
	head := len(s.fixHead(s.fix.Fix.Checks, fixStep(*s.fix), r.W))
	if r.H-head < 6 {
		return r
	}
	return layout.Rect{X: r.X, Y: r.Y + head, W: r.W, H: r.H - head}
}

// healthFixContent is the detail of Health while a fix shows in it: the head,
// then the wizard (the review, the progress or the failure), or the step
// Check again once the fix was applied.
func (s Shell) healthFixContent(w, h int) content {
	if s.inlineFix() {
		wiz := s.synced(*s.fix)
		var lines []string
		if r, area := s.detailRect(), s.fixArea(); area.Y > r.Y {
			// The head as fixArea measured it, so the wizard starts where its clicks land.
			lines = s.fixHead(s.fix.Fix.Checks, fixStep(wiz), r.W)
		}
		lines = append(lines, wiz.Lines(s.th, s.tick)...)
		return content{lines: pane.Block(lines, w, h), hits: map[int]hit{}}
	}
	lines := s.fixHead(s.fixed.checks, doctorview.StepCheck, w)
	f, found := s.fixFinding(s.fixed.checks)
	switch {
	case s.doctor.Loading || len(s.verifying) > 0:
		lines = append(lines, pane.Fit(" "+s.th.Accent(s.th.Spin(s.tick))+" "+words.Say("healthfix.checking"), w))
	case !found || f.OK:
		lines = append(lines, pane.Fit(" "+s.th.Success(s.th.Icons.Passed)+" "+words.Say("healthfix.passes"), w))
	default:
		lines = append(lines, pane.Fit(" "+doctorview.Mark(s.th, f)+" "+words.Say("healthfix.still"), w))
	}
	lines = append(lines, "", pane.Fit(" "+s.th.Muted(words.Say("healthfix.applied", "what", s.fixed.what)), w),
		pane.Fit(" "+s.th.Muted(notCommittedHint()), w), "")
	state := button.Active
	at := len(lines)
	if r := s.detailRect(); s.mouseY == r.Y+at && r.Contains(s.mouseX, s.mouseY) {
		state = button.Hover
	}
	lines = append(lines, pane.Fit(" "+button.Render(s.th, words.Say("tui.detail.check_again"), commands.ShortcutOf("view.reload"), state), w))
	return content{lines: pane.Block(lines, w, h), hits: map[int]hit{at: {kind: "action", id: "x:reload"}}}
}

// fixFooter is the footer while a fix shows in the detail of Health: the keys
// of the step it is on, since the wizard takes the keys.
func (s Shell) fixFooter() []footer.Placed {
	hint := func(key, label string) commands.Hint {
		return commands.Hint{Glyph: key, Label: words.Say(label), Key: key}
	}
	var hints []commands.Hint
	switch {
	case s.fix == nil:
		hints = []commands.Hint{hint(commands.ShortcutOf("view.reload"), "healthfix.footer.check"), hint("esc", "healthfix.footer.back")} // words:literal
	case s.fix.Phase() == fixwizard.Asking:
		hints = []commands.Hint{{Glyph: "↑/↓", Label: words.Say("healthfix.footer.choose")}, hint("enter", "healthfix.footer.pick"), hint("esc", "healthfix.footer.cancel")} // words:literal
	case s.fix.Phase() == fixwizard.Reviewing:
		hints = []commands.Hint{hint("enter", "healthfix.footer.apply"), {Glyph: "↑/↓", Label: words.Say("healthfix.footer.scroll")}, hint("esc", "healthfix.footer.not_now")} // words:literal
	case s.fix.Phase() == fixwizard.Applying:
		hints = []commands.Hint{hint("ctrl+c", "healthfix.footer.stop"), hint("l", "healthfix.footer.log")} // words:literal
	default:
		hints = []commands.Hint{hint("r", "healthfix.footer.retry"), hint("y", "healthfix.footer.copy"), hint("esc", "healthfix.footer.close")} // words:literal
	}
	return footer.Place(s.th, hints, nil, s.w)
}
