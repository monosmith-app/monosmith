package doctor

// StillFailing is how a fix verifies itself: after the fix ran, doctor checks
// again, and the checks the fix said it clears are looked up in the new report.
//
// Parameters: report is the doctor report made after the fix; checks are the
// names of the checks the fix clears (Planned.Checks).
// Returns: the findings of those checks that still do not pass, with the
// detail and the remedy the new run gives, so a person reads the real
// remaining problem. It is empty when the fix cleared every check.
func StillFailing(report Report, checks []string) []Finding {
	var out []Finding
	for _, f := range report.Findings {
		if f.OK {
			continue
		}
		for _, c := range checks {
			if f.Check == c {
				out = append(out, f)
				break
			}
		}
	}
	return out
}

// StillFailingText words one finding that survived its fix, for a toast or a
// line: the check, what was found, and what to do.
func StillFailingText(f Finding) string {
	text := "Still not fixed: " + f.Check
	if f.Detail != "" {
		text += ". " + f.Detail
	}
	if f.Remedy != "" {
		text += " → " + f.Remedy
	}
	return text
}
