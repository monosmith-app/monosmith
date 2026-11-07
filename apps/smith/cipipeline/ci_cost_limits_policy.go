package cipipeline

import (
	"slices"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

// costLimits are what a provider cannot express of [ci] native_on, triggers
// and audit, said in plain words when the workspace sets the key.
func costLimits(cfg workspace.Config, m Model) []string {
	var out []string
	for _, p := range []struct{ id, name string }{{"gitlab", "GitLab"}, {"circleci", "CircleCI"}, {"bitbucket", "Bitbucket"}} {
		if !slices.Contains(cfg.CI.Providers, p.id) {
			continue
		}
		if cfg.CI.Triggers != nil {
			out = append(out, p.name+": [ci] triggers does not apply; the pipeline keeps the triggers of "+p.name)
		}
		if cfg.CI.Audit != "" && cfg.CI.Audit != AuditAlways {
			out = append(out, p.name+": [ci] audit does not apply; the audit runs on every run")
		}
		if cfg.CI.PublishOn != "" {
			out = append(out, p.name+": [ci] publish_on does not apply; pack runs on "+cfg.Release.Branch+" only")
		}
		if cfg.CI.NativeOn != "" && p.id != "gitlab" && m.Native {
			out = append(out, p.name+": [ci] native_on does not apply; the native job runs after every successful ci job")
		}
	}
	if slices.Contains(cfg.CI.Providers, "azure") {
		if cfg.CI.Audit == AuditSchedule {
			out = append(out, "Azure: [ci] audit = \"schedule\" does not apply; the audit runs on every run (use \"main\", or add a schedule in Azure DevOps)")
		}
		if m.Native && m.NativeMode() == NativeOnRelease {
			out = append(out, "Azure: native_on = \"release\" runs the native job on every push to a release branch, because Azure has no output that says a release happened")
		}
	}
	return out
}
