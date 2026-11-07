package tui

import (
	"slices"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/release"
	"github.com/monosmith-app/monosmith/apps/smith/tui/button"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/dialog"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/releaseconfirm"
	"github.com/monosmith-app/monosmith/apps/smith/tui/releaseview"
	"github.com/monosmith-app/monosmith/libs/actions/catalog"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// releaseRun is what the shell remembers of the release it started: the
// packages, for the steps it draws, and where the tags go, for its result.
type releaseRun struct {
	items []release.Item
	facts releaseconfirm.Facts
}

// firstCommitMsg reports the end of the first release commit.
type firstCommitMsg struct{ err error }

// firstCommitArgv marks the confirmation of the first release commit: it runs
// git, not smith, so the shell tells it from a smith command line by its verb.
var firstCommitArgv = []string{"git", "commit", "--allow-empty", "-m", release.FirstReleaseMessage}

// isRelease is true for a command line that publishes: smith release without
// --dry-run (a preview), --status or --rehearse, which only read.
func isRelease(argv []string) bool {
	return len(argv) > 0 && argv[0] == "release" && catalog.Classify(argv) == catalog.Changes
}

// releasePlan is the plan a release command line will run: the plan of the
// Release view with the holds and bumps of the line applied, and the packages
// the line holds back.
func (s Shell) releasePlan(argv []string) (plan release.Plan, held []string) {
	plan = s.release.Plan
	if adjusted, _, err := release.Adjust(plan, adjustmentsIn(argv)); err == nil {
		for _, item := range plan.Items {
			if !slices.ContainsFunc(adjusted.Items, func(a release.Item) bool { return a.Project == item.Project }) {
				held = append(held, item.Project)
			}
		}
		plan = adjusted
	}
	return plan, held
}

// releaseFacts are what the checks before a release read: the doctor, the last
// test run and the repository.
func (s Shell) releaseFacts(argv []string) releaseconfirm.Facts {
	g := s.release.Git
	f := releaseconfirm.Facts{GitKnown: g.Known, Branch: g.Branch, Remote: g.Remote, HasRemote: g.HasRemote, Upstream: g.Upstream, Ahead: g.Ahead, NoPush: slices.Contains(argv, "--no-push")}
	if r := s.doctor.Report; r != nil {
		f.DoctorKnown, f.DoctorFindings = true, r.Failed+r.Warnings
	}
	f.Tests = s.lastTests()
	return f
}

// lastTests is how the newest finished test run of this session ended.
func (s Shell) lastTests() releaseconfirm.TestRun {
	for i := len(s.runs) - 1; i >= 0; i-- {
		r := s.runs[i]
		if req, ok := commands.ParseRun(r.Command); ok && r.Finished && !r.Cancelled && slices.Contains(req.Tasks, "test") {
			if r.OK() {
				return releaseconfirm.TestsPassed
			}
			return releaseconfirm.TestsFailed
		}
	}
	return releaseconfirm.TestsNone
}

// releaseDetails lists the checks, the packages and what the release does.
func (s Shell) releaseDetails(argv []string) []string {
	if !s.release.Loaded || len(s.release.Plan.Items) == 0 {
		return []string{words.Say("tui.confirm_details.the_plan_is_not_loaded_open", "shortcutOf", commands.ShortcutOf("view.release"))}
	}
	plan, held := s.releasePlan(argv)
	f := s.releaseFacts(argv)
	return releaseconfirm.Details(s.th, plan, held, releaseconfirm.Preflight(f), f)
}

// askReleaseConfirm shows the confirmation of a release: Cancel has the focus,
// so a hurried Enter does not publish, and the button says "Release anyway"
// when a check before the release warns.
func (s *Shell) askReleaseConfirm(argv []string) {
	_, why := commands.Describe(argv)
	title, label := words.Say("tui.release_confirm.title"), releaseconfirm.RunLabel(nil)
	if s.release.Loaded && !isForcedRelease(argv) {
		label = releaseconfirm.RunLabel(releaseconfirm.Preflight(s.releaseFacts(argv)))
	}
	c := dialog.NewConfirm(title, why, catalog.WithConfirmation(argv)).WithDetails(s.confirmDetails(argv)).WithRunLabel(label).FocusCancel().Within(s.dialogMargins()).Place(s.w, s.h)
	s.confirm = &c
}

// startRelease remembers the packages and the remote of the release that starts.
func (s *Shell) startRelease(argv []string) {
	plan, _ := s.releasePlan(argv)
	s.releasing = &releaseRun{facts: s.releaseFacts(argv)}
	if !isForcedRelease(argv) {
		s.releasing.items = plan.Items
	}
}

// releaseRunLines is the Release view's own picture of a release in progress or
// just ended: the command, the steps of each package with the spinner and, when
// it failed, the end of its output. room is how many lines it may take.
func (s Shell) releaseRunLines(w, room int) []string {
	if room < 3 {
		return nil
	}
	lines := []string{"", s.th.Title(words.Say("tui.release_run.title", "commandLine", s.run.CommandLine()))}
	var items []release.Item
	if s.releasing != nil {
		items = s.releasing.items
	}
	if s.running() {
		lines = append(lines, releaseconfirm.Steps(s.th, items, s.run.Lines, s.tick, false, false)...)
		lines = append(lines, pane.Wrap(s.th.Info(s.th.Spin(s.tick))+" "+words.Say("tui.release_run.working"), w, 2)...)
		return capTail(lines, room)
	}
	mark := words.Say("tui.release_preview.done", "passed", s.th.Success(s.th.Icons.Passed))
	if !s.run.OK() {
		mark = words.Say("tui.release_preview.failed", "failed", s.th.Failure(s.th.Icons.Failed))
	}
	lines = append(lines, mark+s.th.Muted(words.Say("tui.release_preview.full_output_on_the_run_screen", "shortcutOf", commands.ShortcutOf("view.run"))))
	lines = append(lines, releaseconfirm.Steps(s.th, items, s.run.Lines, s.tick, true, s.run.OK())...)
	if !s.run.OK() {
		var out []string
		for _, l := range s.run.Lines {
			out = append(out, pane.Wrap(pane.Strip(l), w-2, 2)...)
		}
		keep := max(room-len(lines), 1)
		if len(out) > keep {
			out = out[len(out)-keep:]
		}
		for _, l := range out {
			lines = append(lines, "  "+l)
		}
	}
	return capTail(lines, room)
}

// releaseEnded tells the result of a release that ended: the new tags, read from
// what it printed, and what to do next. It opens a notice, so the end of a
// release is never missed behind another tab.
func (s *Shell) releaseEnded() {
	var tags []string
	for _, l := range s.run.Lines {
		if _, tag, ok := strings.Cut(pane.Strip(l), ": tagged "); ok {
			tags = append(tags, strings.TrimSpace(tag))
		}
	}
	facts := releaseconfirm.Facts{HasRemote: true}
	if s.releasing != nil {
		facts = s.releasing.facts
	}
	title, why, lines := releaseconfirm.Result(s.run.OK(), tags, facts)
	n := dialog.NewNotice(title, why, lines).Within(s.dialogMargins()).Place(s.w, s.h)
	s.confirm = &n
}

// releaseWhy opens or closes the explanation of an empty release plan.
func (s *Shell) releaseWhy() {
	if s.release.Empty() {
		s.release = s.release.ToggleWhy()
	}
}

// askFirstCommit asks before the first release commit: one empty commit on the
// current branch, shown as the git command that makes it.
func (s *Shell) askFirstCommit() {
	if !s.release.FirstRelease() {
		s.say(words.Say("tui.release_reason.no_first"))
		return
	}
	if s.sources.FirstCommit == nil {
		s.say(words.Say("tui.release_first.no_source"))
		return
	}
	c := dialog.NewConfirm(words.Say("tui.release_first.title"), words.Say("tui.release_first.why"), firstCommitArgv).WithLine(release.FirstReleaseCommand).WithRunLabel(words.Say("tui.release_first.run")).FocusCancel().Within(s.dialogMargins()).Place(s.w, s.h)
	s.confirm = &c
}

// isFirstCommit is true for the confirmation that asks for the first release commit.
func isFirstCommit(argv []string) bool {
	return len(argv) > 0 && argv[0] == "git"
}

// firstCommit makes the commit that was confirmed, off the Update loop.
func (s *Shell) firstCommit() tea.Cmd {
	commit := s.sources.FirstCommit
	return func() tea.Msg { return firstCommitMsg{err: commit()} }
}

// firstCommitted tells how the first release commit ended and reads the plan
// again, which now sees the commit.
func (s *Shell) firstCommitted(msg firstCommitMsg) tea.Cmd {
	if msg.err != nil {
		n := dialog.NewNotice(words.Say("tui.release_first.title"), words.Say("tui.release_first.failed", "err", msg.err.Error()), nil).Within(s.dialogMargins()).Place(s.w, s.h)
		s.confirm = &n
		return nil
	}
	s.say(words.Say("tui.release_first.done"))
	s.release = releaseview.Model{}
	return s.loadRelease()
}

// releaseViewButton draws the button of a Release view action: it is muted while
// the action cannot run, and shows the hover and pressed looks otherwise.
func (s Shell) releaseViewButton(id, label string) string {
	state := s.buttonState(id, s.actionAt(s.mouseX, s.mouseY))
	if c, ok := commands.ByID(id); ok && c.Unavailable(s.state()) != "" {
		state = button.Disabled
	}
	return button.Render(s.th, words.Say(label), commands.ShortcutOf(id), state)
}
