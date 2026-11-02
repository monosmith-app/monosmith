package composer

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/choiceselect"
	"github.com/monosmith-app/monosmith/apps/smith/compose"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func offer(plugin, language string, c protocol.Choice) choiceselect.Offer {
	c.Language = language
	return choiceselect.Offer{Plugin: plugin, Choice: c}
}

var scaffold = []protocol.Role{protocol.RoleScaffold}
var tool = []protocol.Role{protocol.RoleToolchain}

func catalogue() compose.Catalogue {
	return compose.Build(compose.Source{
		Languages: []compose.Language{{Name: "node", Title: "TypeScript / JavaScript"}, {Name: "python", Title: "Python"}},
		Plugins:   []compose.WorkspacePlugin{{Name: "slices", Layer: "architecture", Description: "structure plugin"}},
		Offers: []choiceselect.Offer{
			offer("vite", "node", protocol.Choice{Axis: "framework", Name: "react", Title: "React", Aliases: []string{"web"}, Description: "a React app", Roles: scaffold, Needs: []string{"lint", "test"}}),
			offer("electron", "node", protocol.Choice{Axis: "framework", Name: "electron", Title: "Electron", Layer: "shell", Aliases: []string{"desktop"}, Description: "a desktop app", Roles: scaffold, Needs: []string{"lint", "test"}}),
			offer("functions", "node", protocol.Choice{Axis: "framework", Name: "azure-functions", Title: "Azure Functions", Layer: "host", Aliases: []string{"serverless"}, Roles: scaffold, Needs: []string{"trigger", "lint", "test"}}),
			offer("functions", "node", protocol.Choice{Axis: "trigger", Name: "http-trigger", Title: "HTTP trigger", Layer: "host", Suggested: true, Description: "answers a request"}),
			offer("functions", "node", protocol.Choice{Axis: "trigger", Name: "timer-trigger", Title: "Timer trigger", Layer: "host", Description: "runs on a schedule"}),
			offer("eslint", "node", protocol.Choice{Axis: "lint", Name: "eslint", Title: "ESLint", Roles: tool, Default: true}),
			offer("vitest", "node", protocol.Choice{Axis: "test", Name: "vitest", Title: "Vitest", Roles: tool, Default: true}),
			offer("py", "python", protocol.Choice{Axis: "framework", Name: "fastapi", Title: "FastAPI", Roles: scaffold}),
		},
	})
}

func screen() Model { return New(catalogue()).Place(100, 40) }

func press(m Model, keys ...string) (Model, Outcome) {
	var out Outcome
	for _, k := range keys {
		var msg tea.KeyMsg
		switch k {
		case "enter":
			msg = tea.KeyMsg{Type: tea.KeyEnter}
		case "esc":
			msg = tea.KeyMsg{Type: tea.KeyEsc}
		case "down":
			msg = tea.KeyMsg{Type: tea.KeyDown}
		case "up":
			msg = tea.KeyMsg{Type: tea.KeyUp}
		case "tab":
			msg = tea.KeyMsg{Type: tea.KeyTab}
		case "ctrl+n":
			msg = tea.KeyMsg{Type: tea.KeyCtrlN}
		default:
			msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
		}
		m, out = m.Key(msg)
	}
	return m, out
}

func typed(m Model, text string) Model {
	for _, r := range text {
		m, _ = press(m, string(r))
	}
	return m
}

func text(m Model) string { return strings.Join(m.Lines(theme.Default()), "\n") }

func TestTypingAnAliasFindsTheItemAndEnterTicksIt(t *testing.T) {
	m := typed(screen(), "desktop")
	if len(m.hits) != 1 || m.hits[0].Name != "electron" {
		t.Fatalf("hits %+v", m.hits)
	}
	m, _ = press(m, "enter")
	if got := strings.Join(m.Ticked(), ","); got != "electron" {
		t.Fatalf("ticked %q", got)
	}
	if !m.Check().Complete || strings.Join(m.Check().Added, ",") != "eslint,vitest" {
		t.Fatalf("electron with the default lint and test is a project: %+v", m.Check())
	}
	for _, want := range []string{"Ready", "Added by default: eslint, vitest"} {
		if !strings.Contains(text(m), want) {
			t.Fatalf("the screen says %q:\n%s", want, text(m))
		}
	}
}

func TestAConflictIsShownAsSoonAsItIsTickedAndBlocksTheNextStep(t *testing.T) {
	m := typed(screen(), "react")
	m, _ = press(m, "enter")
	m, _ = press(m, "esc")
	m = typed(m, "electron")
	m, _ = press(m, "enter")
	if !strings.Contains(m.Check().Conflict, "react") || !strings.Contains(m.Check().Conflict, "electron") {
		t.Fatalf("the conflict names both: %+v", m.Check())
	}
	if !strings.Contains(text(m), "Conflict:") {
		t.Fatalf("the screen shows it:\n%s", text(m))
	}
	m, out := press(m, "ctrl+n")
	if m.stage != stBrowse || out.Plan != nil || !strings.Contains(m.note, "Fix the conflict first") {
		t.Fatalf("the next step is refused with the reason: stage %v note %q", m.stage, m.note)
	}
	m, _ = press(m, "enter")
	if m.Check().Conflict != "" {
		t.Fatalf("ticking it again removes it: %+v", m.Check())
	}
}

func TestAHostAsksItsTriggerAndTheSuggestedAnswerIsOnlyPreselected(t *testing.T) {
	m := typed(screen(), "serverless")
	m, _ = press(m, "enter")
	if m.stage != stQuestion || m.Check().Question == nil {
		t.Fatalf("a host that needs a trigger asks: stage %v", m.stage)
	}
	if got := strings.Join(m.Ticked(), ","); got != "azure-functions" {
		t.Fatalf("nothing was chosen for the person: %q", got)
	}
	if !strings.Contains(text(m), "(suggested)") || !strings.Contains(text(m), "does not choose this for you") {
		t.Fatalf("the question marks the suggestion and says who decides:\n%s", text(m))
	}
	m, _ = press(m, "enter")
	if m.stage != stBrowse || !m.Check().Complete {
		t.Fatalf("enter takes the preselected answer: stage %v %+v", m.stage, m.Check())
	}
	if got := strings.Join(m.Ticked(), ","); got != "azure-functions,http-trigger" {
		t.Fatalf("ticked %q", got)
	}
}

func TestEscLeavesAQuestionUnansweredAndNextAsksItAgain(t *testing.T) {
	m := typed(screen(), "serverless")
	m, _ = press(m, "enter", "esc")
	if m.stage != stBrowse || m.Check().Question == nil {
		t.Fatal("the question stays open")
	}
	m, _ = press(m, "ctrl+n")
	if m.stage != stQuestion {
		t.Fatalf("ctrl+n asks again: %v", m.stage)
	}
}

func TestAWorkspacePluginIsFoundButNotTicked(t *testing.T) {
	m := typed(screen(), "slices")
	m, _ = press(m, "enter")
	if len(m.Ticked()) != 0 || !strings.Contains(m.note, "smith plugin enable slices") {
		t.Fatalf("it says which command turns it on: ticked %v note %q", m.Ticked(), m.note)
	}
}

func TestTheLanguageNarrowsTheListToItsItems(t *testing.T) {
	m := typed(screen(), "typescript")
	m, _ = press(m, "enter")
	if m.lang == nil || m.lang.Name != "node" {
		t.Fatalf("the first row is the first language: %+v", m.lang)
	}
	for _, h := range m.hits {
		if h.Language == "python" && h.Kind != compose.KindLanguage {
			t.Fatalf("a python item is hidden for a node stack: %+v", h)
		}
	}
}

func TestTheNameStageValidatesAndAsksForThePlan(t *testing.T) {
	m := typed(screen(), "desktop")
	m, _ = press(m, "enter", "ctrl+n")
	if m.stage != stName {
		t.Fatalf("a complete selection goes on to the name: %v", m.stage)
	}
	m, out := press(m, "enter")
	if out.Plan != nil || m.note == "" {
		t.Fatalf("no name, no plan: %+v", out)
	}
	m = typed(m, "Bad Name")
	m, out = press(m, "enter")
	if out.Plan != nil {
		t.Fatal("a name with capitals or spaces is refused")
	}
	m.name.SetValue("shell")
	m, out = press(m, "enter")
	if out.Plan == nil {
		t.Fatalf("a good name asks for the plan: %q", m.note)
	}
	want := []string{"compose", "--json", "--kind", "app", "--name", "shell", "--lang", "node", "--with", "electron,eslint,vitest"}
	if strings.Join(out.Plan.Argv(), " ") != strings.Join(want, " ") {
		t.Fatalf("argv %v", out.Plan.Argv())
	}
}

func TestThePlanOffersRunCopyAndRecipe(t *testing.T) {
	m := typed(screen(), "desktop")
	m, _ = press(m, "enter", "ctrl+n")
	m.name.SetValue("shell")
	m, _ = press(m, "enter")
	plan := compose.Plan{
		Command: "smith add app shell --lang node --with electron,eslint,vitest",
		Argv:    []string{"smith", "add", "app", "shell", "--lang", "node", "--with", "electron,eslint,vitest"},
		Files:   []compose.PlanLine{{Action: "create", Path: "apps/shell/"}},
		Tools:   []compose.Tool{{Choice: "electron", Name: "Node.js", Install: "install Node"}},
		Notes:   []string{"eslint and vitest are added by default"},

		RecipeArgv: []string{"smith", "recipe", "new", "--name", "x"}, RecipeNote: "A recipe fixes the plugins.",
	}
	m = m.WithPlan(plan, nil)
	for _, want := range []string{"smith add app shell", "create apps/shell/", "It needs these tools", "run it now"} {
		if !strings.Contains(text(m), want) {
			t.Fatalf("the plan shows %q:\n%s", want, text(m))
		}
	}
	if _, out := press(m, "c"); out.Copy != plan.Command {
		t.Fatalf("copy %+v", out)
	}
	if _, out := press(m, "r"); !out.Done || strings.Join(out.Run, " ") != "add app shell --lang node --with electron,eslint,vitest" {
		t.Fatalf("run %+v", out)
	}
	if _, out := press(m, "s"); strings.Join(out.Run, " ") != "recipe new --name x" {
		t.Fatalf("recipe %+v", out)
	}
	failed := m.WithPlan(compose.Plan{}, errors.New("no workspace"))
	if !strings.Contains(text(failed), "could not be made: no workspace") {
		t.Fatalf("a failed plan says why:\n%s", text(failed))
	}
}

func TestEscClearsTheSearchThenLeavesTheStackViewThenCloses(t *testing.T) {
	m := typed(screen(), "react")
	m, out := press(m, "esc")
	if m.search.Value() != "" || out.Done {
		t.Fatal("the first esc clears the search")
	}
	m, _ = press(m, "tab")
	m, out = press(m, "esc")
	if m.stack || out.Done {
		t.Fatal("the second esc leaves the stack view")
	}
	if _, out = press(m, "esc"); !out.Done {
		t.Fatal("the third closes")
	}
}

func TestTheListShowsTheLayersWithWhatSitsOnThem(t *testing.T) {
	body := text(screen())
	for _, want := range []string{"Compose a stack", "Language", "Framework", "Desktop shell", "Cloud host"} {
		if !strings.Contains(body, want) {
			t.Fatalf("missing %q:\n%s", want, body)
		}
	}
}
