package cipipeline

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/release"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func releasedFixture() (release.Plan, release.Report) {
	core := release.Item{Project: "core", Current: "1.4.2", Next: "1.5.0", Tag: "core@1.5.0", Level: "minor",
		Changes: []release.Change{{Type: "fix", Description: "handle nil"}, {Type: "feat", Description: "add search"}}}
	logger := release.Item{Project: "logger", Current: "2.0.1", Next: "2.0.2", Tag: "logger@2.0.2", Level: "patch", Reason: "dependency core 1.5.0"}
	tool := release.Item{Project: "tool", Current: "0.1.0", Next: "1.0.0", Tag: "tool@1.0.0", Level: "forced", Forced: true, Reason: "forced by version(tool)[1.0.0]"}
	api := release.Item{Project: "api", Current: "1.0.0", Next: "2.0.0", Tag: "api@2.0.0", Level: "major",
		Changes: []release.Change{{Type: "feat", Description: "drop v1", Breaking: true}}}
	plan := release.Plan{Items: []release.Item{core, logger, tool, api}, Unchanged: []string{"a", "b"}}
	report := release.Report{Unchanged: plan.Unchanged, Results: []release.ReportResult{
		{Item: core, Status: "released", Published: true, Registry: "npm", Pushed: true, GitHubRelease: "https://example/r", TookMs: 12000},
		{Item: logger, Status: "released", Pushed: true, Publish: "published by its tag", TookMs: 3000},
		{Item: tool, Status: "released", Publish: "no registry"},
		{Item: api, Status: "failed", Error: "boom", TookMs: 1000},
	}}
	return plan, report
}

func TestTheReleaseIsOneTableWithFailuresFirstThenReleasedInOrder(t *testing.T) {
	plan, report := releasedFixture()
	table := releaseTable(plan, report, theme.Unicode)
	got := strings.Join(table.render("packages", false), "\n")
	for _, want := range []string{
		"Project  Was", "Now", "Bump", "Why", "Published", "Tag", "Time",
		"› 2.0.0  major  feat!: drop v1", "✗ failed",
		"› 1.5.0  minor  feat: add search (+1 more)", "✓ npm, GitHub Release", "core@1.5.0 (pushed)", "12s",
		"› 2.0.2  patch  dependency core 1.5.0", "✓ its tag",
		"› 1.0.0  major  version() directive", "✓ nothing (no registry)", "tool@1.0.0 (local)",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in\n%s", want, got)
		}
	}
	if table.Rows[0].Cells[0] != "api" || table.Rows[1].Cells[0] != "core" || table.Rows[3].Cells[0] != "tool" {
		t.Errorf("order: %v", table.Rows)
	}
}

func TestADependencyCausedBumpIsYellowAndNamesTheDependencyVersion(t *testing.T) {
	plan, report := releasedFixture()
	var row tableRow
	for _, r := range releaseTable(plan, report, theme.Unicode).Rows {
		if r.Cells[0] == "logger" {
			row = r
		}
	}
	if row.Cells[4] != "dependency core 1.5.0" || row.Tints[4] != tintYellow || row.Cells[3] != "patch" || row.Tints[3] != tintGreen {
		t.Errorf("row %+v", row)
	}
}

func TestAPackageTheReleaseNeverReachedIsListedAsNotReached(t *testing.T) {
	plan, _ := releasedFixture()
	got := strings.Join(releaseTable(plan, release.Report{Error: "preflight"}, theme.ASCII).render("packages", false), "\n")
	if strings.Count(got, "[ ] not reached") != 4 {
		t.Errorf("\n%s", got)
	}
}

func TestTheClosingCountsReleasedUnchangedAndFailed(t *testing.T) {
	_, report := releasedFixture()
	if c := releaseConclusion(report); c != "3 released (core 1.5.0, logger 2.0.2, tool 1.0.0), 2 unchanged, 1 failed" {
		t.Errorf("conclusion: %s", c)
	}
	report.Results[0].Status = "already-released"
	if c := releaseConclusion(report); !strings.Contains(c, ", 1 already released, ") {
		t.Errorf("conclusion: %s", c)
	}
}

func TestUnchangedProjectsAreAListWhenFewAndACountWhenMany(t *testing.T) {
	if got := unchangedLine([]string{"a", "b"}); got != "Unchanged (2): a, b" {
		t.Errorf("%q", got)
	}
	if got := unchangedLine(make([]string, 23)); got != "Unchanged: 23 projects have no releasing commit" {
		t.Errorf("%q", got)
	}
	if unchangedLine(nil) != "" {
		t.Error("nothing to say when nothing is unchanged")
	}
}

func TestTheIntroSaysWhatWillHappenAndTheOnePlanLine(t *testing.T) {
	plan, _ := releasedFixture()
	lines := releaseIntro(plan, "main", 23)
	if !strings.Contains(lines[0], "Releasing 4 projects of 23 from main") || lines[1] != "Plan: core 1.5.0, logger 2.0.2, tool 1.0.0, api 2.0.0" {
		t.Errorf("%v", lines)
	}
}

func TestThePublishedColumnSaysWhereEachPackageWent(t *testing.T) {
	cases := map[string]release.ReportResult{
		"1 image":               {Images: []protocol.PublishedImage{{Reference: "ghcr.io/o/api"}}},
		"its tag":               {Publish: "published by its tag"},
		"nothing (no registry)": {Publish: "no registry"},
		"nothing to publish":    {},
		"registry.npmjs.org":    {Published: true, Registry: "https://registry.npmjs.org/"},
		"-":                     {Status: "failed"},
	}
	for want, res := range cases {
		if got := publishedTo(res); got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}
