package commands

import (
	"slices"
	"strings"
	"testing"
)

func TestFixWithDoctorOpensTheWizardInsteadOfRunningACommandLine(t *testing.T) {
	c, ok := ByID("ws.fix")
	if !ok || c.Kind != OpenFixWizard || len(c.Argv) != 0 || c.Confirm {
		t.Fatalf("ws.fix is %+v: it must open the wizard, with no non-interactive doctor --fix behind it", c)
	}
	// The key is the Doctor view's; the palette lists the command everywhere.
	if _, _, ok := ByKey("F", State{View: TabDoctor, InWorkspace: true}); !ok {
		t.Fatal("F does nothing in Doctor")
	}
	if got, _, ok := ByKey("F", State{View: TabProjects, InWorkspace: true}); ok && got.ID == "ws.fix" {
		t.Fatal("F opens the fix wizard outside Doctor")
	}
	if !c.Visible(State{View: TabHome, InWorkspace: true}) {
		t.Fatal("the palette does not list Fix with doctor on Home")
	}
}

func TestTheFixKeysAreInTheHelpAndFooterOfDoctorWithTheirReasonsWhenRefused(t *testing.T) {
	st := State{View: TabDoctor, InWorkspace: true, Reasons: map[string]string{"doctor.fix": "this check has no fix"}}
	view, _ := Hints(st)
	var labels []string
	for _, h := range view {
		labels = append(labels, h.Label)
	}
	if slices.Contains(labels, "fix") {
		t.Fatalf("a refused fix is in the footer: %v", labels)
	}
	if !slices.Contains(labels, "fix all") {
		t.Fatalf("fix all is not in the footer: %v", labels)
	}
	help := ""
	for _, r := range HelpRows(State{View: TabDoctor, InWorkspace: true}, "Doctor") {
		help += r.Title + "|" + r.Shortcut + "\n"
	}
	for _, want := range []string{"Fix this check: the wizard asks, shows what changes, then applies|f", "Fix with doctor…|F"} {
		if !strings.Contains(help, want) {
			t.Fatalf("the help lacks %q:\n%s", want, help)
		}
	}
}
