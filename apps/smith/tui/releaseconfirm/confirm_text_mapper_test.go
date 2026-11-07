package releaseconfirm

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/release"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

var plain = theme.New(theme.Choice{Tokens: theme.Plain, Icons: theme.Unicode})

func twoPackages() release.Plan {
	return release.Plan{Items: []release.Item{
		{Project: "web", Current: "1.0.0", Next: "1.1.0", Tag: "web@1.1.0"},
		{Project: "kit", Next: "0.1.0", Tag: "kit@0.1.0"},
	}}
}

func TestTheDetailsListTheChecksThePackagesWithTheirTagsAndTheEffects(t *testing.T) {
	f := ready()
	f.Tests = TestsFailed
	out := strings.Join(Details(plain, twoPackages(), []string{"sdk"}, Preflight(f), f), "\n")
	for _, want := range []string{
		"Before you release:", "✓ Doctor: no findings", "▲ Tests: the last test run failed", "A warning does not stop the release. You decide.",
		"Packages and versions to publish:", "web  1.0.0 → 1.1.0  tag web@1.1.0", "kit  new → 0.1.0  tag kit@0.1.0", "Held back, not released: sdk",
		"What will happen:", "Commit: no.", "Tag: 2 tags are created.", "Push: the tags go to origin.", "Publish: yes.",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q:\n%s", want, out)
		}
	}
}

func TestTheDetailsOfAReadyWorkspaceHaveNoWarningLine(t *testing.T) {
	out := strings.Join(Details(plain, twoPackages(), nil, Preflight(ready()), ready()), "\n")
	if strings.Contains(out, "warning") || strings.Contains(out, "▲") {
		t.Fatalf("a warning on a ready workspace:\n%s", out)
	}
}

func TestThePushLineSaysWhereTheTagsGo(t *testing.T) {
	cases := []struct {
		f    Facts
		want string
	}{
		{Facts{NoPush: true, HasRemote: true, Remote: "origin"}, "Push: no. The command line has --no-push."},
		{Facts{}, "Push: no. There is no remote"},
		{Facts{HasRemote: true, Remote: "upstream"}, "Push: the tags go to upstream."},
	}
	for _, c := range cases {
		if got := pushLine(c.f); !strings.Contains(got, c.want) {
			t.Errorf("pushLine(%+v) = %q, want %q", c.f, got, c.want)
		}
	}
}

func TestALongPlanIsCutToACount(t *testing.T) {
	var plan release.Plan
	for i := 0; i < maxRows+3; i++ {
		plan.Items = append(plan.Items, release.Item{Project: "p", Next: "1.0.0", Tag: "t"})
	}
	out := strings.Join(Details(plain, plan, nil, nil, Facts{}), "\n")
	if !strings.Contains(out, "and 3 more") {
		t.Fatalf("no count of the rest:\n%s", out)
	}
}

func TestAChannelIsNamedBeforeTheChecks(t *testing.T) {
	plan := twoPackages()
	plan.Channel = "next"
	lines := Details(plain, plan, nil, nil, Facts{})
	if !strings.Contains(lines[0], "next") {
		t.Fatalf("the first line is %q", lines[0])
	}
}
