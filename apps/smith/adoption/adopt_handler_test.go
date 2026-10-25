package adoption

import (
	"bytes"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/doctor"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestPrintReportHasHeadersAndEachWayOut(t *testing.T) {
	r := Report{
		Root: "/r", Name: "acme", Scope: "@acme", Providers: []string{"github"}, Clean: true,
		Projects: []ProjectLine{{Name: "api", Root: "api", Language: "go", Kind: "app"}},
		Blockers: []Item{{Message: "web is published at 1.0.0 according to npm, but has no release tag", Blocks: BlocksRelease, Steps: []string{"git fetch --tags", "or tag the commit that published it"}, Step: "git fetch --tags"}},
	}
	var b bytes.Buffer
	printReport(&b, r)
	out := b.String()
	for _, want := range []string{"PROJECT", "LANGUAGE", "KIND", "FOLDER", "a Go main package is an app", "✖ web is published", "  → git fetch --tags\n", "  → or tag the commit"} {
		if !strings.Contains(out, want) {
			t.Errorf("report lacks %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "see the") {
		t.Errorf("a remedy must not point back at the message:\n%s", out)
	}
}

func TestPredictedFailureNamesTheStepThatClearsIt(t *testing.T) {
	for check, want := range map[string]string{
		"go.work exists":                  "--init runs go work init",
		"npm workspaces include web":      "--init adds the \"workspaces\"",
		"Node dependencies are installed": "--init --install",
		"something else":                  "do this first",
	} {
		item := predictedFailure(doctor.Finding{Finding: protocol.Finding{Check: check, Detail: "d", Remedy: "do this first\nand then this"}})
		if !strings.Contains(item.Step, want) {
			t.Errorf("%s: step %q, want %q", check, item.Step, want)
		}
		if !strings.HasPrefix(item.Message, "smith doctor will report: ") {
			t.Errorf("%s: message %q", check, item.Message)
		}
	}
}

func TestLevelFullSaysItIsConvertsAndNamesTheIssueWithoutTouchingTheRepository(t *testing.T) {
	cmd := Command("test")
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SilenceUsage, cmd.SilenceErrors = true, true
	cmd.SetArgs([]string{"--level", "full", "--init"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("--level full succeeded")
	}
	for _, want := range []string{"--level full", "smith convert --level full --dry-run", "conservative level", FullLevelIssue} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the problem lacks %q:\n%s", want, err)
		}
	}
	if out.Len() != 0 {
		t.Errorf("nothing is printed before the refusal, got %q", out.String())
	}
}

func TestLevelTakesConservativeAndFullOnly(t *testing.T) {
	cmd := Command("test")
	cmd.SetArgs([]string{"--level", "deep"})
	cmd.SilenceUsage, cmd.SilenceErrors = true, true
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "not one of conservative, full") {
		t.Fatalf("error %v", err)
	}
	if f := cmd.Flags().Lookup("level"); f == nil || f.DefValue != LevelConservative {
		t.Fatalf("the default level is %+v", f)
	}
}
