package doctorview

// Replace is the model for a check that finished while the page was already
// showing one (after a fix, doctor checks again): the new findings, with the
// grouping, the passed checks and the cursor where the user left them.
func (v Model) Replace(r Result) Model {
	next := FromResult(r)
	if r.Err != nil || v.Report == nil {
		return next
	}
	next.Grouped, next.ShowPassed = v.Grouped, v.ShowPassed
	next.Cursor = min(v.Cursor, max(next.rows()-1, 0))
	return next
}
