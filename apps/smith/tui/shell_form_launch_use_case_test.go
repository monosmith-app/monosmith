package tui

import (
	"slices"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/clicontract"
	"github.com/monosmith-app/monosmith/apps/smith/doctor"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/doctorview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/guide"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// withCheckForms gives the lab the forms smith builds from the commands for
// smith boundaries and smith slices check, as tui_handler does.
func (l *lab) withCheckForms() *lab {
	sh := l.sh()
	sh.guide = sh.guide.
		WithAction(guide.FormFor(clicontract.CommandDescription{Name: "boundaries", Description: "Check which projects depend on which against your rules",
			Options: []clicontract.OptionDescription{{Name: "json", TakesValue: "none"}}}, []string{"boundaries"}, "boundaries")).
		WithAction(guide.FormFor(clicontract.CommandDescription{Name: "check", Description: "Check that files are role-named",
			Arguments: []clicontract.ArgumentDescription{{Name: "project directory", Required: true, Variadic: true}},
			Options:   []clicontract.OptionDescription{{Name: "lang", TakesValue: "required"}}}, []string{"slices", "check"}, "slices check"))
	l.m = sh
	return l
}

func boundaryResult() doctorview.Result {
	return doctorview.Result{Report: doctor.Report{Failed: 1, Passed: 1, Findings: []doctor.Finding{
		{Finding: protocol.Finding{Check: doctor.BoundaryRulesCheck, Problem: "web -> admin: ui must not use server", Area: "dependencies"}, Source: "smith"},
		{Finding: protocol.Finding{Check: "git is installed", OK: true}, Source: "smith"},
	}}}
}

func TestABoundaryFindingHasAButtonThatRunsSmithBoundariesNow(t *testing.T) {
	l := newLab(t, true, 140, 34).withCheckForms()
	l.keys("3")
	l.send(boundaryResult())
	l.has("Run this check now")
	l.runDetailAction("Run this check now")
	if got := l.lastStarted(); got != "boundaries" {
		t.Fatalf("the button started %q, not smith boundaries", got)
	}
}

func TestTheSliceCheckIsARowOfDoctorThatOpensItsForm(t *testing.T) {
	l := newLab(t, true, 140, 34).withCheckForms()
	l.keys("3")
	l.send(boundaryResult())
	l.has("Checks doctor does not run itself", "Vertical-slice layout")
	for range 10 {
		if l.sh().onDoctorExtra() {
			break
		}
		l.keys("down")
	}
	if !l.sh().onDoctorExtra() {
		t.Fatal("the keys never reach the slice check row")
	}
	l.has("Doctor does not run this check itself", "Run this check now")
	l.keys("enter")
	if l.sh().tab != commands.TabGuide || l.sh().guide.Title() != "Check that files are role-named" {
		t.Fatalf("Enter opened %q on %q", l.sh().guide.Title(), l.sh().tab)
	}
	// The form starts on the folders of the projects in the main language (go).
	l.has("apps/orders libs/kit")
}

func TestAFindingAboutAFileHasAWhyButtonThatOpensTheWhyFormForIt(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("3")
	for range 10 {
		if f, ok := l.sh().doctor.Selected(); ok && f.Check == "Docker registries of web agree" {
			break
		}
		l.keys("down")
	}
	l.has("Why is apps/web/Dockerfile here?")
	l.runDetailAction("Why is apps/web/Dockerfile here?")
	if l.sh().guide.Title() != "Why is this here?" {
		t.Fatalf("the button opened %q", l.sh().guide.Title())
	}
	l.has("apps/web/Dockerfile")
	l.keys("enter")
	if got := l.lastStarted(); got != "why apps/web/Dockerfile" {
		t.Fatalf("the form ran %q", got)
	}
}

func TestAFindingAboutNoFileHasNoWhyButton(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("3")
	if f, ok := l.sh().doctor.Selected(); !ok || f.Check != "go.work lists every module" {
		t.Fatalf("the cursor starts on %q", f.Check)
	}
	l.hasNot("Why is ")
}

func TestWOnATaskThatRanOpensWhyMissForItAndOnACachedOneSaysWhyNot(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.next = &scripted{lines: []string{"kit:build ▶ go build", "kit:build ✓ 1s", "web:build ✓ from cache"}}
	l.keys("b")
	if !l.runPanelHasKeys() {
		t.Fatal("the run panel does not have the keyboard")
	}
	l.keys("down") // web:build, from the cache: cached tasks are listed before passed ones
	if l.sh().runSel != "web:build" {
		t.Fatalf("selected %q", l.sh().runSel)
	}
	l.keys("w")
	l.has("came from the cache")
	if l.sh().tab == commands.TabGuide {
		t.Fatal("w opened a form for a task that did not run")
	}
	l.keys("down")
	l.keys("w")
	if l.sh().guide.Title() != "Why would a task run again?" {
		t.Fatalf("w opened %q", l.sh().guide.Title())
	}
	l.has("kit:build")
	l.keys("enter")
	if got := l.lastStarted(); got != "why-miss kit:build" {
		t.Fatalf("the form ran %q", got)
	}
}

func TestVerifyRunsSmithVerifyAndTheRunPanelListsItsTasks(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.next = &scripted{lines: []string{
		"== Verify ==",
		"kit:lint ▶ golangci-lint run",
		"kit:lint ✓ 1.2s",
		"web:test ▶ vitest run",
		"web:test × exited with code 1",
		"× 1/2 tasks succeeded (0 from cache) in 3.1s",
		"× Verify failed after 3.4s",
	}}
	l.keys("V")
	if got := l.lastStarted(); got != "verify" {
		t.Fatalf("V started %q", got)
	}
	l.has("kit:lint", "web:test")
	if c := l.sh().run.Counts(); c.Passed != 1 || c.Failed != 1 {
		t.Fatalf("%+v", c)
	}
}

func TestMakeMyOwnRecipeInAFormOpensTheWriteARecipeForm(t *testing.T) {
	l := newLab(t, true, 140, 34)
	sh := l.sh()
	sh.guide = sh.guide.WithAction(guide.Action{Title: "Add a thing", Steps: []guide.Step{{Key: "how", Prompt: "How?", Options: []guide.Option{
		{Label: "Own recipe", Value: "recipe", Wizard: guide.WizardRecipe},
	}}}})
	l.m = sh
	sh = l.sh()
	cmd := sh.perform(commands.Command{Kind: commands.Guided, Guide: "Add a thing"})
	l.m = sh
	l.drive(cmd)
	l.keys("enter")
	if !slices.ContainsFunc(l.captured, func(argv []string) bool { return strings.Join(argv, " ") == "recipe new --list-levels --json" }) {
		t.Fatalf("the recipe form did not read its levels: %v", l.captured)
	}
}
