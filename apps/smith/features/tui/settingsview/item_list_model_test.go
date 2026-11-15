package settingsview

import (
	"strings"
	"testing"
)

// placedForms are forms with the actions they present, as the shell passes them.
func placedForms() []Form {
	return []Form{
		{Title: "Add a project", Hint: "app, lib or package", Presents: "add"},
		{Title: "Manage plugins", Hint: "smith plugin set", Presents: "plugin set"},
		{Title: "Regenerate CI pipelines", Hint: "smith ci generate", Presents: "ci"},
		{Title: "Add a ready-made pipeline", Hint: "smith pipeline", Presents: "pipeline"},
		{Title: "Connect smith to a coding assistant", Hint: "smith mcp install", Presents: "mcp install"},
		{Title: "Install the agent files", Hint: "smith agent install", Presents: "agent install"},
	}
}

func openedWith(t *testing.T, forms []Form) Model {
	t.Helper()
	m := New(sample().backend(), forms)
	m, cmd := m.Open()
	m, _ = settle(m, Result{Cmd: cmd})
	return m
}

// formTitles are the titles of the form rows of a section, in order.
func formTitles(m Model, s Section) []string {
	var out []string
	for _, it := range m.items(s) {
		if it.kind == kindForm {
			out = append(out, it.key)
		}
	}
	return out
}

func TestTheAssistantsAndCISectionsListTheirForms(t *testing.T) {
	m := openedWith(t, placedForms())
	for s, want := range map[Section]string{
		Assistants: "Connect smith to a coding assistant;Install the agent files",
		CI:         "Regenerate CI pipelines;Add a ready-made pipeline",
		Plugins:    "Manage plugins",
	} {
		if got := strings.Join(formTitles(m, s), ";"); got != want {
			t.Errorf("%s lists %q, want %q", s.Title(), got, want)
		}
	}
	if got := len(formTitles(m, Forms)); got != len(placedForms()) {
		t.Errorf("the Forms section still lists every form: %d", got)
	}
	out := text(m.Show(Assistants), 100, 28)
	for _, want := range []string{"Assistants", "CI", "Connect smith to a coding assistant", "smith agent install"} {
		if !strings.Contains(out, want) {
			t.Errorf("the Assistants section lacks %q:\n%s", want, out)
		}
	}
}

func TestEnterOnAFormOfASectionOpensThatForm(t *testing.T) {
	m := openedWith(t, placedForms()).Show(CI)
	_, results := press(m, "down", "enter")
	if len(results) == 0 || results[len(results)-1].Form != "Add a ready-made pipeline" {
		t.Fatalf("%+v", results)
	}
	m = openedWith(t, placedForms()).Show(Plugins)
	m, _ = press(m, "end")
	_, results = press(m, "enter")
	if len(results) == 0 || results[0].Form != "Manage plugins" {
		t.Fatalf("the last row of Plugins opens Manage plugins: %+v", results)
	}
}

func TestASectionWhoseFormsTheWorkspaceLacksIsEmpty(t *testing.T) {
	m := openedWith(t, nil).Show(Assistants)
	if len(m.items(Assistants)) != 0 || len(m.items(CI)) != 0 {
		t.Fatal("no form, no row")
	}
	if out := text(m, 100, 28); !strings.Contains(out, "Assistants") {
		t.Fatalf("the section still shows:\n%s", out)
	}
}
