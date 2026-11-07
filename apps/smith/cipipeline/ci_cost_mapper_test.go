package cipipeline

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

func nativeModel() Model {
	m := model()
	m.Native, m.NativeRunners = true, DefaultNativeRunners
	return m
}

func mustContain(t *testing.T, name, out string, wants ...string) {
	t.Helper()
	for _, want := range wants {
		if !strings.Contains(out, want) {
			t.Errorf("%s lacks %q in\n%s", name, want, out)
		}
	}
}

func TestNativeOnDecidesWhenTheNativeJobRuns(t *testing.T) {
	release := "    if: ${{ always() && (github.event_name == 'push' && github.ref_name == 'main' && needs.ci.outputs.released == 'true') }}\n"
	m := nativeModel()
	mustContain(t, "default", RenderGitHub(m), release)
	m.NativeOn = NativeOnRelease
	mustContain(t, "release", RenderGitHub(m), release)
	m.NativeOn = NativeOnMain
	mustContain(t, "main", RenderGitHub(m),
		"    if: ${{ always() && ((needs.ci.result == 'success' && ((github.event_name == 'push' && github.ref_name == 'main') || (github.event_name == 'pull_request' && github.base_ref == 'main'))) || (github.event_name == 'push' && github.ref_name == 'main' && needs.ci.outputs.released == 'true')) }}\n")
	m.NativeOn = NativeOnAlways
	mustContain(t, "always", RenderGitHub(m),
		"    if: ${{ always() && (needs.ci.result == 'success' || (github.event_name == 'push' && needs.ci.outputs.released == 'true')) }}\n")
}

func TestNativeOnInAzure(t *testing.T) {
	m := nativeModel()
	on := "and(succeeded(), in(variables['Build.Reason'], 'IndividualCI', 'BatchedCI'), eq(variables['Build.SourceBranch'], 'refs/heads/main'))"
	mustContain(t, "release", RenderAzure(m), "  - job: native\n    dependsOn: ci\n    condition: "+on+"\n")
	m.NativeOn = NativeOnMain
	mustContain(t, "main", RenderAzure(m), "    condition: or("+on+", and(succeeded(), eq(variables['Build.Reason'], 'PullRequest'), in(variables['System.PullRequest.TargetBranch'], 'main', 'refs/heads/main')))\n")
	m.NativeOn = NativeOnAlways
	mustContain(t, "always", RenderAzure(m), "  - job: native\n    dependsOn: ci\n    strategy:")
}

func TestConcurrencyNeverCancelsAReleaseBranch(t *testing.T) {
	m := model()
	mustContain(t, "plain", RenderGitHub(m), "  cancel-in-progress: ${{ github.ref != 'refs/heads/main' }}\n")
	m.ChannelBranches = []string{"next"}
	mustContain(t, "channel", RenderGitHub(m), "  cancel-in-progress: ${{ github.ref != 'refs/heads/main' && github.ref != 'refs/heads/next' }}\n")
}

func TestTriggersChooseTheEvents(t *testing.T) {
	m := model()
	mustContain(t, "default", RenderGitHub(m), "on:\n  push:\n    branches: [main]\n  pull_request:\n    branches: [main]\n")
	m.Triggers = []string{"pull_request"}
	gh := RenderGitHub(m)
	mustContain(t, "pull_request", gh, "on:\n  pull_request:\n    branches: [main]\n")
	if strings.Contains(gh, "  push:\n") {
		t.Errorf("pull_request only has no push trigger:\n%s", gh)
	}
	mustContain(t, "azure pull_request", RenderAzure(m), "trigger: none\n\npr:\n  branches:\n    include: [main]\n")
	m.Triggers = []string{"push"}
	gh = RenderGitHub(m)
	mustContain(t, "push", gh, "on:\n  push:\n    branches: [main]\n")
	if strings.Contains(gh, "  pull_request:\n    branches") {
		t.Errorf("push only has no pull_request trigger:\n%s", gh)
	}
	mustContain(t, "azure push", RenderAzure(m), "trigger:\n  branches:\n    include: [main]\n\npr: none\n")
}

func TestAuditDecidesWhenTheAuditStepRuns(t *testing.T) {
	m := model()
	mustContain(t, "always", RenderGitHub(m), "      - name: Audit\n        run: smith ci audit\n")
	m.Audit = AuditMain
	gh := RenderGitHub(m)
	mustContain(t, "main", gh, "      - name: Audit\n        if: (github.ref_name == 'main' && github.event_name == 'push')\n        run: smith ci audit\n", "      - name: Verify\n        run: smith ci verify\n")
	mustContain(t, "azure main", RenderAzure(m), "    displayName: Audit\n    condition: and(succeeded(), in(variables['Build.Reason']")
	m.Audit = AuditSchedule
	gh = RenderGitHub(m)
	mustContain(t, "schedule", gh,
		"  schedule:\n    - cron: \"0 6 * * 1\"\n",
		"        if: github.event_name == 'schedule' || (github.ref_name == 'main' && github.event_name == 'push')\n        run: smith ci audit\n",
		"      - name: Verify\n        if: github.event_name != 'schedule'\n        run: smith ci verify\n")
	m.AuditCron = "30 2 * * 0"
	mustContain(t, "cron", RenderGitHub(m), "    - cron: \"30 2 * * 0\"\n")
	if strings.Contains(RenderAzure(m), "schedule") {
		t.Error("Azure writes no schedule")
	}
}

func TestAScheduledAuditDoesNotStartTheNativeJobs(t *testing.T) {
	m := nativeModel()
	m.Audit, m.NativeOn = AuditSchedule, NativeOnAlways
	mustContain(t, "always", RenderGitHub(m), "if: ${{ always() && github.event_name != 'schedule' && (needs.ci.result == 'success'")
}

func TestPublishOnChannelsPacksAndBuildsAssetsOnTheChannelBranchToo(t *testing.T) {
	m := model()
	m.Pack, m.Attest = true, true
	m.AssetPlatforms = []string{"linux-x64"}
	m.ChannelBranches = []string{"next"}
	gh := RenderGitHub(m)
	mustContain(t, "release", gh, "      - name: Pack\n        if: github.event_name == 'push' && github.ref_name == 'main'\n", "needs.ci.outputs.released == 'true' && github.event_name == 'push' && github.ref_name == 'main')")
	m.PublishOn = "channels"
	gh = RenderGitHub(m)
	both := "(github.ref_name == 'main' || github.ref_name == 'next')"
	mustContain(t, "channels", gh, "      - name: Pack\n        if: github.event_name == 'push' && "+both+"\n", "&& github.event_name == 'push' && "+both+") ||", "if: github.event_name == 'push' && "+both+" && hashFiles(")
	mustContain(t, "azure", RenderAzure(m), "or(eq(variables['Build.SourceBranch'], 'refs/heads/main'), eq(variables['Build.SourceBranch'], 'refs/heads/next'))")
}

func TestLimitsSayWhatOtherProvidersDoNotDo(t *testing.T) {
	cfg := workspace.Config{}
	cfg.Release.Branch = "main"
	cfg.CI.Providers = []string{"gitlab", "circleci", "azure"}
	cfg.CI.Triggers, cfg.CI.Audit, cfg.CI.NativeOn, cfg.CI.PublishOn = []string{"push"}, AuditSchedule, "main", "channels"
	got := strings.Join(Limits(cfg, nativeModel()), "\n")
	mustContain(t, "limits", got, "GitLab: [ci] triggers does not apply", "CircleCI: [ci] audit does not apply", "CircleCI: [ci] native_on does not apply", "GitLab: [ci] publish_on does not apply", "Azure: [ci] audit = \"schedule\" does not apply")
	cfg.CI.Triggers, cfg.CI.Audit, cfg.CI.NativeOn, cfg.CI.PublishOn = nil, "", "", ""
	if strings.Contains(strings.Join(Limits(cfg, Model{}), "\n"), "does not apply") {
		t.Error("an unset key needs no note")
	}
}
