package runhistory

// StepSelection is the choice one step from sel in the run panel's list ids (dir
// -1 is up, 1 is down), kept inside the list. A choice not in the list counts as
// the first.
func StepSelection(ids []string, sel string, dir int) string {
	at := 0
	for i, id := range ids {
		if id == sel {
			at = i
		}
	}
	return ids[min(max(at+dir, 0), len(ids)-1)]
}

// StepRecent is the position of the chosen run in the list of recent runs after
// a step of dir, kept between 0 and total-1.
func StepRecent(at, total, dir int) int {
	return max(min(at+dir, total-1), 0)
}
