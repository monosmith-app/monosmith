package cipipeline

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// goStep is a step that runs a quick, real program (go is on the test PATH).
func goStep(name string, blocking bool, args ...string) protocol.Step {
	return protocol.Step{Name: name, Command: append([]string{"go"}, args...), Blocking: blocking}
}

func TestSetupRunsOneSectionPerPluginInOrder(t *testing.T) {
	noCI(t)
	var out bytes.Buffer
	l := testSections(&out, theme.Unicode, false)
	c := Contributions{
		Setup:         []protocol.Step{goStep("A1", true, "version"), goStep("A2", true, "version"), goStep("B1", true, "version")},
		SetupSections: []SetupSection{{Plugin: "node", Projects: 2, Steps: 2}, {Plugin: "go", Projects: 1, Steps: 1}},
	}
	outcomes, err := runSetupSections(context.Background(), t.TempDir(), c, l, &out)
	if err != nil || len(outcomes) != 3 {
		t.Fatalf("outcomes %d, err %v", len(outcomes), err)
	}
	got := out.String()
	last := -1
	for _, want := range []string{"══ node (2 projects)", "▶ A1", "▶ A2", "✓ node done", "══ go (1 project)", "▶ B1", "✓ go done"} {
		i := strings.Index(got, want)
		if i <= last {
			t.Fatalf("%q missing or out of order in:\n%s", want, got)
		}
		last = i
	}
}

func TestSetupFailureClosesItsSectionAndStopsTheLaterOnes(t *testing.T) {
	noCI(t)
	var out bytes.Buffer
	l := testSections(&out, theme.ASCII, false)
	c := Contributions{
		Setup:         []protocol.Step{goStep("Bad", true, "tool", "no-such-tool-for-smith"), goStep("Later", true, "version")},
		SetupSections: []SetupSection{{Plugin: "node", Projects: 1, Steps: 1}, {Plugin: "go", Projects: 1, Steps: 1}},
	}
	_, err := runSetupSections(context.Background(), t.TempDir(), c, l, &out)
	if err == nil {
		t.Fatal("a failed blocking step stops the phase")
	}
	got := out.String()
	if !strings.Contains(got, "== node (1 project) ==") || !strings.Contains(got, "[x] node failed: Bad (exit") || !strings.Contains(got, "→ the end of the step's output") || strings.Contains(got, "go (1 project)") {
		t.Fatalf("unexpected:\n%s", got)
	}
}

func TestSetupSectionsAreGroupsInCI(t *testing.T) {
	noCI(t)
	t.Setenv("GITHUB_ACTIONS", "true")
	var out bytes.Buffer
	l := testSections(&out, theme.Unicode, false)
	c := Contributions{Setup: []protocol.Step{goStep("A", true, "version"), goStep("B", true, "version")}, SetupSections: []SetupSection{{Plugin: "node", Projects: 1, Steps: 1}, {Plugin: "go", Projects: 1, Steps: 1}}}
	if _, err := runSetupSections(context.Background(), t.TempDir(), c, l, &out); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	if strings.Count(got, "::group::") != 2 || strings.Count(got, "::endgroup::") != 2 || strings.Contains(got, "══") {
		t.Fatalf("one group per section, no rule:\n%s", got)
	}
}

func TestSetupStepsNoPluginClaimedRunInTheLastSection(t *testing.T) {
	noCI(t)
	var out bytes.Buffer
	l := testSections(&out, theme.Unicode, false)
	c := Contributions{Setup: []protocol.Step{goStep("Loose", true, "version")}}
	if _, err := runSetupSections(context.Background(), t.TempDir(), c, l, &out); err != nil || !strings.Contains(out.String(), "══ setup ═") {
		t.Fatalf("%v\n%s", err, out.String())
	}
}
