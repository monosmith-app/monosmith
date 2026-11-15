package guide

import (
	"strings"
	"testing"
)

const convertPlanText = `Reading nx.json…
Checking that the converted workspace loads…
{
  "source": "nx",
  "outputChoices": [
    {"project": "app", "task": "build", "declared": ["dist", "dist-dev"], "detected": ["dist"]}
  ]
}`

func convertOutputsMenu(t *testing.T) Model {
	t.Helper()
	o, err := ParseConvertOutputs(convertPlanText)
	if err != nil {
		t.Fatal(err)
	}
	m := NewModel(Actions([]string{"node"}, []string{"node"}, []string{"node"}))
	base, ok := m.ActionTitled(ConvertTitle())
	if !ok {
		t.Fatal("the convert form is in the menu")
	}
	m, ok = m.WithAction(ConvertOutputsAction(base, o)).Pick(ConvertTitle())
	if !ok {
		t.Fatal("the convert form opens")
	}
	return m
}

func TestParseConvertOutputsReadsThePlanAfterTheStageLines(t *testing.T) {
	o, err := ParseConvertOutputs(convertPlanText)
	if err != nil || o.Source != "nx" || len(o.Choices) != 1 || o.Choices[0].Task != "build" {
		t.Fatalf("%+v %v", o, err)
	}
	if _, err := ParseConvertOutputs("Error: not a git repository"); err == nil {
		t.Fatal("a text without a plan is an error")
	}
}

func TestTheConvertFormAsksTheOutputsWithTheSuggestionFirst(t *testing.T) {
	// Conservative, then the outputs question on its suggestion, then no commit.
	m := keys(convertOutputsMenu(t), "enter").(Model)
	view := m.View()
	for _, want := range []string{"Which outputs of app:build", "dist, what the build writes", "dist, dist-dev, as Nx declared them"} {
		if !strings.Contains(view, want) {
			t.Fatalf("missing %q:\n%s", want, view)
		}
	}
	if strings.Index(view, "what the build writes") > strings.Index(view, "as Nx declared them") {
		t.Fatalf("the suggestion is not first:\n%s", view)
	}
	done := keys(m, "enter", "enter").(Model)
	if got := strings.Join(done.Argv, " "); got != "convert --outputs app:build=detected" {
		t.Fatalf("argv %q", got)
	}
	done = keys(m, "down", "enter", "enter").(Model)
	if got := strings.Join(done.Argv, " "); got != "convert --outputs app:build=declared" {
		t.Fatalf("argv %q", got)
	}
}

func TestTheConvertFormWithoutOutputChoicesIsThePlainForm(t *testing.T) {
	base, _ := NewModel(Actions([]string{"node"}, nil, nil)).ActionTitled(ConvertTitle())
	if a := ConvertOutputsAction(base, ConvertOutputs{}); len(a.Steps) != len(base.Steps) {
		t.Fatalf("steps %d, want %d", len(a.Steps), len(base.Steps))
	}
}
