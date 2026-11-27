package doctorview

// Replace is the model for a check that finished while the page was already
// showing one (after a fix, doctor checks again): the new findings, with the
// grouping, the passed checks and the cursor where the user left them.
func (v Model) Replace(r Result) Model {
	next := FromResult(r)
	next.Extras, next.ByState = v.Extras, v.ByState
	if r.Err != nil || v.Report == nil {
		next.Cursor = next.firstFinding()
		return next
	}
	next.Grouped, next.ShowPassed, next.Folds, next.All = v.Grouped, v.ShowPassed, v.Folds, v.All
	next.HideFailed, next.HideWarnings = v.HideFailed, v.HideWarnings
	// A cursor on the findings stays on them: fewer findings must not move it
	// onto a check doctor does not run, where Enter would start that check.
	last := next.rows() - 1
	if _, onExtra := v.SelectedExtra(); !onExtra {
		last -= len(next.Extras)
	}
	next.Cursor = min(v.Cursor, max(last, 0))
	return next
}
