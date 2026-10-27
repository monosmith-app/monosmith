package cipipeline

import (
	"fmt"
	goruntime "runtime"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// smithLine says which smith runs the pipeline, so the install step's result
// is visible in the log of the first command.
func smithLine(version string) string {
	return fmt.Sprintf("smith %s on %s/%s", version, goruntime.GOOS, goruntime.GOARCH)
}

// toolchainNames lists the toolchains as "node 22, go 1.25".
func toolchainNames(toolchains []protocol.Toolchain) []string {
	names := make([]string, 0, len(toolchains))
	for _, t := range toolchains {
		switch t.Version {
		case "", "latest", "system":
			names = append(names, t.Name)
		default:
			names = append(names, t.Name+" "+t.Version)
		}
	}
	return names
}

// setupIntro says what setup is about to do and on what.
func setupIntro(s *session.Session, c Contributions, steps []protocol.Step) []string {
	projects := plural(len(s.Graph.Projects), "project")
	if len(c.Toolchains) == 0 {
		return []string{fmt.Sprintf("Preparing %s: %s.", projects, plural(len(steps), "step"))}
	}
	return []string{fmt.Sprintf("Installing dependencies for %s (%s) and preparing %s: %s.",
		plural(len(c.Toolchains), "toolchain"), strings.Join(toolchainNames(c.Toolchains), ", "), projects, plural(len(steps), "step"))}
}

// auditIntro says what the audit is about to check and by which rules.
func auditIntro(s *session.Session, steps []protocol.Step) []string {
	lines := []string{fmt.Sprintf("Auditing %s with %s.",
		plural(len(s.Graph.Projects), "project"), plural(len(steps), "check"))}
	ci := s.Config.CI
	rules := "smith's default policies"
	if len(ci.AuditPolicies) > 0 {
		rules = "policies " + strings.Join(ci.AuditPolicies, ", ")
	}
	if len(ci.AuditWarn) > 0 {
		rules += "; warning only for " + strings.Join(ci.AuditWarn, ", ")
	}
	return append(lines, "Rules: "+rules+" ([ci] audit_policies).")
}
