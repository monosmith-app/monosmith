package guide

import (
	"strings"
	"testing"
)

const outcomeOffersJSON = `[
 {"axis":"framework","name":"react","title":"React","description":"a React app","language":"node","aliases":["ui"]},
 {"axis":"framework","name":"vue","title":"Vue","description":"a Vue app","language":"node"},
 {"axis":"framework","name":"tauri","title":"Tauri","description":"a small desktop app","language":"node","layer":"shell","aliases":["desktop"]},
 {"axis":"framework","name":"electron","title":"Electron","description":"a desktop app that bundles Node","language":"node","layer":"shell","aliases":["desktop"]},
 {"axis":"framework","name":"azure-functions","title":"Azure Functions","description":"an Azure Functions app","language":"node","layer":"host","needs":["trigger"]},
 {"axis":"framework","name":"azure-functions","title":"Azure Functions","description":"an Azure Functions app","language":"python","layer":"host","needs":["trigger"]},
 {"axis":"trigger","name":"http-trigger","description":"an HTTP request","language":"node","layer":"host","suggested":true},
 {"axis":"trigger","name":"http-trigger","description":"an HTTP request","language":"python","layer":"host","suggested":true}
]`

func outcomeForm(t *testing.T, existing ...string) Model {
	t.Helper()
	var add Action
	for _, a := range Actions([]string{"go", "node", "python"}, nil, nil) {
		if a.Title == "Add a project" {
			add = a
		}
	}
	offers, err := ParseChoices(outcomeOffersJSON)
	if err != nil {
		t.Fatal(err)
	}
	m, ok := NewModel([]Action{AddProjectAction(add, offers, existing)}).Pick("Add a project")
	if !ok {
		t.Fatal("no form")
	}
	return m
}

func TestTheAddFormAsksWhatYouAreBuildingBeforeAnyLanguage(t *testing.T) {
	view := outcomeForm(t).View()
	for _, want := range []string{"What are you building?", "A website or web app", "An API or backend service", "A function app (Azure, AWS, Google)",
		"A desktop app (Tauri, Electron)", "A library for my other projects", "A package I publish", "I know the language", "try react, azure function, electron, tauri, api", "Not sure? Pick the closest one"} {
		if !strings.Contains(view, want) {
			t.Errorf("the first question lacks %q:\n%s", want, view)
		}
	}
	if strings.Contains(view, "AppleScript") || strings.Contains(view, "Which language") {
		t.Fatalf("no language list before the outcome:\n%s", view)
	}
}

func TestTheOutcomeListIsFoundByTheWordsPeopleUse(t *testing.T) {
	step := outcomeForm(t).current()
	cases := map[string]string{
		"react": "web", "vue": "web", "website": "web", "azure function": "function", "lambda": "", "electron": "desktop",
		"tauri": "desktop", "desktop": "desktop", "api": "api", "serverless": "function", "npm": "package", "shared": "library",
	}
	for typed, want := range cases {
		var got []string
		for _, o := range step.Options {
			if MatchesFilter(o, typed) {
				got = append(got, o.Value)
			}
		}
		if want == "" {
			continue // a word no plugin offers here finds nothing but the long way, if anything
		}
		found := false
		for _, g := range got {
			found = found || g == want
		}
		if !found {
			t.Errorf("%q finds %v, want %s among them", typed, got, want)
		}
	}
}

func TestEachOutcomeLineShowsTheTypicalThingsAndWhatItMeans(t *testing.T) {
	for _, o := range outcomeForm(t).current().Options {
		if o.Value == "web" && (!strings.Contains(o.Hint, "React") || !strings.Contains(o.Hint, "browser")) {
			t.Errorf("a website lacks its typical frameworks or its consequence: %q", o.Hint)
		}
		if o.Value == "desktop" && (!strings.Contains(o.Hint, "Tauri") || !strings.Contains(o.Hint, "Electron")) {
			t.Errorf("a desktop app lacks its shells: %q", o.Hint)
		}
	}
}

func TestAWebsiteHasOnlyNodeSoNoLanguageIsAskedAndTheFrameworkStartsTicked(t *testing.T) {
	m := press(outcomeForm(t), "enter") // A website
	if !strings.Contains(m.View(), "Which framework?") {
		t.Fatalf("the framework is asked, not the language:\n%s", m.View())
	}
	m = press(m, "down", "enter") // Vue
	m = press(m, "s", "i", "t", "e", "enter")
	if !strings.Contains(m.View(), "[x] framework: vue") {
		t.Fatalf("the pick starts ticked in the choices:\n%s", m.View())
	}
	m = press(m, "enter")
	if got := strings.Join(m.Argv, " "); got != "add app site --lang node --with vue" {
		t.Fatalf("argv %q", got)
	}
}

func TestTypingTauriStartsTheDesktopShellOnTauri(t *testing.T) {
	m := press(outcomeForm(t), "e", "l", "e", "c", "t", "r", "o", "n", "enter")
	if !strings.Contains(m.View(), "Which desktop shell?") || !strings.Contains(m.View(), "› Electron") {
		t.Fatalf("what was typed starts the cursor on Electron:\n%s", m.View())
	}
	m = press(m, "enter", "d", "e", "s", "k", "enter", "enter")
	if got := strings.Join(m.Argv, " "); got != "add app desk --lang node --with electron" {
		t.Fatalf("argv %q", got)
	}
}

func TestAFunctionAppAsksItsLanguageThenItsCloudAndItsTrigger(t *testing.T) {
	m := press(outcomeForm(t), "a", "z", "u", "r", "e", "enter") // A function app
	if !strings.Contains(m.View(), "Which language is it written in?") || strings.Contains(m.View(), "Go") {
		t.Fatalf("only the languages with a host are asked:\n%s", m.View())
	}
	m = press(m, "down", "enter") // Python
	if !strings.Contains(m.View(), "Which cloud?") || strings.Contains(m.View(), "Decide in the next question") {
		t.Fatalf("a function app needs its host:\n%s", m.View())
	}
	m = press(m, "enter", "f", "n", "enter", "enter") // Azure Functions, name fn, keep the pick
	if !strings.Contains(m.View(), "Which trigger?") {
		t.Fatalf("the trigger is still asked:\n%s", m.View())
	}
	m = press(m, "enter")
	if got := strings.Join(m.Argv, " "); got != "add app fn --lang python --with azure-functions,http-trigger" {
		t.Fatalf("argv %q", got)
	}
}

func TestALibraryAnswersItsKindAndAsksTheLanguage(t *testing.T) {
	m := press(outcomeForm(t), "l", "i", "b", "enter")
	if !strings.Contains(m.View(), "Which language is it written in?") {
		t.Fatalf("a library may be written in any language:\n%s", m.View())
	}
	m = press(m, "enter", "c", "o", "r", "e", "enter", "enter")
	if got := strings.Join(m.Argv, " "); !strings.HasPrefix(got, "add lib core --lang ") {
		t.Fatalf("argv %q", got)
	}
}

func TestTheLongWayStaysReachable(t *testing.T) {
	m := press(outcomeForm(t), "k", "n", "o", "w", "enter")
	if !strings.Contains(m.View(), "What are you adding?") {
		t.Fatalf("the long way asks the kind first:\n%s", m.View())
	}
	m = press(m, "enter") // an app
	if !strings.Contains(m.View(), "Which language is it written in?") || !strings.Contains(m.View(), "Python") {
		t.Fatalf("then every language:\n%s", m.View())
	}
}

func TestThePreselectedReasonIsShownOnlyWhereThePreselectedLanguageIsOffered(t *testing.T) {
	// Go is the main language here, but a website has only Node.
	m := outcomeForm(t).WithDefault("lang", "go").WithDefaultWhy("lang", "Preselected: 3 of the 4 projects here use Go.")
	m, _ = m.Pick("Add a project")
	m = press(m, "l", "i", "b", "enter")
	if !strings.Contains(m.View(), "Preselected: 3 of the 4") || !strings.Contains(m.View(), "› Go") {
		t.Fatalf("a library may be Go, so the reason shows:\n%s", m.View())
	}
	m, _ = m.Pick("Add a project")
	m = press(m, "e", "l", "e", "c", "enter", "enter")
	if strings.Contains(m.View(), "Preselected") {
		t.Fatalf("the reason is for a language that is not on this list:\n%s", m.View())
	}
}
