package cipipeline

import (
	"fmt"
	"strings"
)

// githubTriggers are the `on:` events of the workflow, from [ci] triggers, and
// the audit's schedule: the push and pull request triggers on the trigger
// branches, then a cron when the audit runs on a schedule.
func githubTriggers(m Model) string {
	var b strings.Builder
	branches := strings.Join(m.TriggerBranches(), ", ")
	if m.TriggersOn("push") {
		fmt.Fprintf(&b, "  push:\n    branches: [%s]\n", branches)
	}
	if m.TriggersOn("pull_request") {
		fmt.Fprintf(&b, "  pull_request:\n    branches: [%s]\n", branches)
	}
	if m.AuditMode() == AuditSchedule {
		fmt.Fprintf(&b, "  schedule:\n    - cron: \"%s\"\n", m.AuditSchedule())
	}
	return b.String()
}

// githubCancelInProgress is the concurrency expression: a newer run cancels an
// older one on the same ref, except on a release branch, where a release run
// must finish.
func githubCancelInProgress(m Model) string {
	return "${{ " + m.branchCondition(
		func(branch string) string { return "github.ref != 'refs/heads/" + branch + "'" },
		func(c []string) string { return strings.Join(c, " && ") }) + " }}"
}

// githubAuditStep is the audit step: always, or only where [ci] audit says.
func githubAuditStep(m Model) string {
	step := "      - name: Audit\n"
	push := "(" + githubOnReleaseBranches(m) + " && github.event_name == 'push')"
	switch m.AuditMode() {
	case AuditMain:
		step += "        if: " + push + "\n"
	case AuditSchedule:
		step += "        if: github.event_name == 'schedule' || " + push + "\n"
	}
	return step + "        run: smith ci audit\n"
}

// githubVerifyStep is the verify step; a scheduled run only audits.
func githubVerifyStep(m Model) string {
	step := "      - name: Verify\n"
	if m.AuditMode() == AuditSchedule {
		step += "        if: github.event_name != 'schedule'\n"
	}
	return step + "        run: smith ci verify\n"
}

// githubNativeCondition is the native job's `if`, from [ci] native_on. A job
// that needs the ci job and uses always() also runs when that job failed after
// the release step, so a release that exists still gets its builds.
func githubNativeCondition(m Model) string {
	released := fmt.Sprintf("(%s && %s && needs.ci.outputs.released == 'true')", githubReleaseEvent(m), githubOnReleaseBranches(m))
	switch m.NativeMode() {
	case NativeOnAlways:
		skip := ""
		if m.AuditMode() == AuditSchedule {
			skip = " && github.event_name != 'schedule'"
		}
		return fmt.Sprintf("${{ always()%s && (needs.ci.result == 'success' || (%s && needs.ci.outputs.released == 'true')) }}", skip, githubReleaseEvent(m))
	case NativeOnMain:
		intoRelease := m.branchCondition(
			func(branch string) string { return "github.base_ref == '" + branch + "'" },
			func(c []string) string { return "(" + strings.Join(c, " || ") + ")" })
		return fmt.Sprintf("${{ always() && ((needs.ci.result == 'success' && ((github.event_name == 'push' && %s) || (github.event_name == 'pull_request' && %s))) || %s) }}",
			githubOnReleaseBranches(m), intoRelease, released)
	}
	return "${{ always() && " + released + " }}"
}

// azureOnReleaseBranches is the Azure condition that holds on a push to the
// release branch or to a release channel's branch.
func azureOnReleaseBranches(m Model) string {
	return fmt.Sprintf("and(succeeded(), in(variables['Build.Reason'], 'IndividualCI', 'BatchedCI'), %s)", m.branchCondition(
		func(branch string) string {
			return fmt.Sprintf("eq(variables['Build.SourceBranch'], 'refs/heads/%s')", branch)
		},
		func(c []string) string { return "or(" + strings.Join(c, ", ") + ")" }))
}

// azureTriggers are the trigger and pr blocks: `none` for a dropped trigger.
func azureTriggers(m Model) string {
	branches := strings.Join(m.TriggerBranches(), ", ")
	var b strings.Builder
	for _, t := range []struct{ event, key string }{{"push", "trigger"}, {"pull_request", "pr"}} {
		if m.TriggersOn(t.event) {
			fmt.Fprintf(&b, "%s:\n  branches:\n    include: [%s]\n\n", t.key, branches)
		} else {
			fmt.Fprintf(&b, "%s: none\n\n", t.key)
		}
	}
	return b.String()
}

// azureNativeCondition is the native job's condition line, from [ci]
// native_on, or empty when the job runs after every successful ci job. Azure
// has no output that says a release happened, so "release" is a push to a
// release branch.
func azureNativeCondition(m Model) string {
	switch m.NativeMode() {
	case NativeOnAlways:
		return ""
	case NativeOnMain:
		intoRelease := m.branchCondition(
			func(branch string) string {
				return fmt.Sprintf("in(variables['System.PullRequest.TargetBranch'], '%[1]s', 'refs/heads/%[1]s')", branch)
			},
			func(c []string) string { return "or(" + strings.Join(c, ", ") + ")" })
		return fmt.Sprintf("    condition: or(%s, and(succeeded(), eq(variables['Build.Reason'], 'PullRequest'), %s))\n", azureOnReleaseBranches(m), intoRelease)
	}
	return "    condition: " + azureOnReleaseBranches(m) + "\n"
}

// azureAuditStep is the audit step: always, or on a release branch push for
// audit = "main". The definition smith writes has no schedule, so "schedule"
// audits on every run (Limits says so).
func azureAuditStep(m Model) string {
	step := "  - bash: smith ci audit\n    displayName: Audit\n"
	if m.AuditMode() == AuditMain {
		step += "    condition: " + azureOnReleaseBranches(m) + "\n"
	}
	return step
}

// githubOnPublishBranches is the condition that holds on the branches that
// pack, attest and build the release files (see Model.PublishBranches).
func githubOnPublishBranches(m Model) string {
	return m.publishCondition(func(branch string) string { return "github.ref_name == '" + branch + "'" }, func(c []string) string { return "(" + strings.Join(c, " || ") + ")" })
}

// azureOnPublishBranches is the Azure condition for a push to a publish branch.
func azureOnPublishBranches(m Model) string {
	return fmt.Sprintf("and(succeeded(), in(variables['Build.Reason'], 'IndividualCI', 'BatchedCI'), %s)", m.publishCondition(
		func(branch string) string {
			return fmt.Sprintf("eq(variables['Build.SourceBranch'], 'refs/heads/%s')", branch)
		},
		func(c []string) string { return "or(" + strings.Join(c, ", ") + ")" }))
}
