package releaseconfirm

import "github.com/monosmith-app/monosmith/libs/words/words"

// Preflight lists what the release depends on and was not made by it: the
// doctor, the tests and the branch. A failing item is a warning, not a stop.
//
// Parameters: f is what the shell knows.
// Returns: one check for the doctor, one for the tests and one for the branch.
func Preflight(f Facts) []Check {
	return []Check{doctorCheck(f), testsCheck(f), branchCheck(f)}
}

// Warns reports whether any check is a warning.
func Warns(checks []Check) bool {
	for _, c := range checks {
		if c.Warn {
			return true
		}
	}
	return false
}

func doctorCheck(f Facts) Check {
	switch {
	case !f.DoctorKnown:
		return Check{Text: words.Say("tui.release_confirm.doctor_unknown"), Warn: true}
	case f.DoctorFindings > 0:
		return Check{Text: words.Count("tui.release_confirm.doctor_findings", f.DoctorFindings), Warn: true}
	}
	return Check{Text: words.Say("tui.release_confirm.doctor_clean")}
}

func testsCheck(f Facts) Check {
	switch f.Tests {
	case TestsPassed:
		return Check{Text: words.Say("tui.release_confirm.tests_passed")}
	case TestsFailed:
		return Check{Text: words.Say("tui.release_confirm.tests_failed"), Warn: true}
	}
	return Check{Text: words.Say("tui.release_confirm.tests_none"), Warn: true}
}

func branchCheck(f Facts) Check {
	switch {
	case !f.GitKnown || f.Branch == "":
		return Check{Text: words.Say("tui.release_confirm.branch_unknown"), Warn: true}
	case !f.HasRemote:
		return Check{Text: words.Say("tui.release_confirm.branch_no_remote", "branch", f.Branch), Warn: true}
	case !f.Upstream:
		return Check{Text: words.Say("tui.release_confirm.branch_no_upstream", "branch", f.Branch), Warn: true}
	case f.Ahead > 0:
		return Check{Text: words.Count("tui.release_confirm.branch_ahead", f.Ahead, "branch", f.Branch), Warn: true}
	}
	return Check{Text: words.Say("tui.release_confirm.branch_pushed", "branch", f.Branch)}
}
