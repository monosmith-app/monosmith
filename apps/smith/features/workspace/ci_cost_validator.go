package workspace

import (
	"fmt"
	"slices"
	"strings"
)

// validateCICost checks the [ci] keys that decide how many CI minutes a run
// costs: native_on, triggers, audit and audit_cron. An empty string or an
// absent list is the default and is accepted; an empty `triggers = []` is not,
// because a pipeline with no trigger never runs.
func validateCICost(ci CISection) error {
	switch ci.NativeOn {
	case "", "release", "main", "always":
	default:
		return fmt.Errorf("%s: [ci] native_on must be \"release\", \"main\" or \"always\", not %q", FileName, ci.NativeOn)
	}
	switch ci.PublishOn {
	case "", "release", "channels":
	default:
		return fmt.Errorf("%s: [ci] publish_on must be \"release\" or \"channels\", not %q", FileName, ci.PublishOn)
	}
	switch ci.Audit {
	case "", "always", "main", "schedule":
	default:
		return fmt.Errorf("%s: [ci] audit must be \"always\", \"main\" or \"schedule\", not %q", FileName, ci.Audit)
	}
	if ci.Triggers != nil && len(ci.Triggers) == 0 {
		return fmt.Errorf("%s: [ci] triggers must not be empty: use \"push\", \"pull_request\" or both (omit the key for both)", FileName)
	}
	for i, t := range ci.Triggers {
		if t != "push" && t != "pull_request" {
			return fmt.Errorf("%s: [ci] triggers has unknown trigger %q (use \"push\" or \"pull_request\")", FileName, t)
		}
		if slices.Contains(ci.Triggers[:i], t) {
			return fmt.Errorf("%s: [ci] triggers lists %q twice", FileName, t)
		}
	}
	if ci.AuditCron != "" && len(strings.Fields(ci.AuditCron)) != 5 {
		return fmt.Errorf("%s: [ci] audit_cron must be a cron schedule of five fields (minute hour day month weekday), not %q", FileName, ci.AuditCron)
	}
	if strings.ContainsAny(ci.AuditCron, "\"'\\n\r") {
		return fmt.Errorf("%s: [ci] audit_cron must not hold quotes, backslashes or line breaks", FileName)
	}
	return nil
}
