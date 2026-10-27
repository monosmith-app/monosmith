package cipipeline

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/release"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestTheReleaseDecisionIsATableOfWasNowAndWhy(t *testing.T) {
	plan := release.Plan{Items: []release.Item{
		{Project: "core", Current: "1.4.2", Next: "1.5.0", Changes: []release.Change{{Type: "fix", Description: "handle nil"}, {Type: "feat", Description: "add search"}}},
		{Project: "logger", Next: "0.1.0", Reason: "dependency core releases"},
	}}
	got := strings.Join(planTable(plan).Text("packages"), "\n")
	for _, want := range []string{"Package  Was    Now    Why", "core     1.4.2  1.5.0  feat: add search (+1 more)", "logger   new    0.1.0  dependency core releases"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in\n%s", want, got)
		}
	}
}

func TestTheReleaseResultSaysWhereEachPackageWentAndItsTag(t *testing.T) {
	report := release.Report{
		Unchanged: []string{"a", "b"},
		Results: []release.ReportResult{
			{Item: release.Item{Project: "core", Next: "1.5.0", Tag: "core@1.5.0"}, Status: "released", Published: true, Registry: "npm", Pushed: true, GitHubRelease: "https://github.com/o/r/releases/tag/core@1.5.0"},
			{Item: release.Item{Project: "tool", Next: "0.2.0", Tag: "tool@0.2.0"}, Status: "released", Pushed: true, Publish: "published by its tag"},
			{Item: release.Item{Project: "docs", Next: "1.0.0", Tag: "docs@1.0.0"}, Status: "released", Publish: "no registry"},
			{Item: release.Item{Project: "api", Next: "2.0.0", Tag: "api@2.0.0"}, Status: "released", Images: []protocol.PublishedImage{{Reference: "ghcr.io/o/api"}}},
		},
	}
	got := strings.Join(releaseResultTable(report, theme.Unicode).Text("packages"), "\n")
	for _, want := range []string{"core", "✓ released", "npm, GitHub Release", "core@1.5.0 (pushed)", "its tag", "1 image", "api@2.0.0 (local)", "nothing (no registry)"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in\n%s", want, got)
		}
	}
	if c := releaseConclusion(report); c != "4 projects released (core 1.5.0, tool 0.2.0, docs 1.0.0, api 2.0.0), 2 projects unchanged" {
		t.Errorf("conclusion: %s", c)
	}
}
