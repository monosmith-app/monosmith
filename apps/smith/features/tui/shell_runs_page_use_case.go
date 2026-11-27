package tui

import "github.com/monosmith-app/monosmith/libs/words/words"

// toggleRunErrors keeps only the failed tasks in the tree of the Runs view, or
// shows every task again. A selected task the filter hides gives way to the
// first failed one, so the log shown is always a row of the tree.
func (s *Shell) toggleRunErrors() {
	if s.run == nil {
		return
	}
	s.runErrors = !s.runErrors
	s.runTop = 0
	if s.runErrors {
		if t := s.run.Task(s.runSel); t != nil && s.runSel != "" && !containsID(s.runIDs(), s.runSel) {
			s.runSel = s.firstFailed()
		}
		if s.run.Counts().Failed == 0 {
			s.say(words.Say("runpage.tree.no_errors"))
		}
	}
	s.keepRunVisible()
}

// containsID is true when ids holds id.
func containsID(ids []string, id string) bool {
	for _, x := range ids {
		if x == id {
			return true
		}
	}
	return false
}

// wheelRuns scrolls the Runs view at the cell (x, y): the log when the pointer
// is over it (a narrow screen that shows the log included), else the tree.
func (s *Shell) wheelRuns(x, y, dir int) {
	if s.run == nil {
		return
	}
	in := s.mainInner()
	w, h := s.runRect()
	g := s.runsGeom(w, h)
	if g.onLog || (!g.narrow && x-in.X >= g.logX-1) {
		s.log = s.log.Scroll(dir*wheelStep, s.logTotal(), g.logRows)
		return
	}
	lines := len(s.treeLines())
	s.runTop = max(min(s.runTop+dir, lines-g.treeRows), 0)
}
