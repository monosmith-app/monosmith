package guide

import (
	"slices"
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
	m = press(m, "down", "down", "enter") // Vue (the cursor starts on "Decide in the next question")
	m = press(m, "s", "i", "t", "e", "enter")
	if !strings.Contains(m.View(), "[x] framework: vue") {
		t.Fatalf("the pick starts ticked in the choices:\n%s", m.View())
	}
	m = press(m, "enter")                            // the choices
	m = press(m, "enter", "enter", "enter", "enter") // the environments take their defaults
	if got := strings.Join(m.Argv, " "); got != "add app site --lang node --with vue --environments lcl,dev,uat,prd --default-environment prd --environment-output workspace --env-files commit" {
		t.Fatalf("argv %q", got)
	}
}

// frameworkForm is the add form of a node workspace that offers the four web
// frameworks.
func frameworkForm(t *testing.T) Model {
	t.Helper()
	offers, err := ParseChoices(`[
 {"axis":"framework","name":"html","title":"Plain page","description":"a page","language":"node"},
 {"axis":"framework","name":"react","title":"React","description":"a React app","language":"node"},
 {"axis":"framework","name":"svelte","title":"Svelte","description":"a Svelte app","language":"node"},
 {"axis":"framework","name":"vue","title":"Vue","description":"a Vue app","language":"node"},
 {"axis":"test","name":"vitest","description":"tests","language":"node","default":true}]`)
	if err != nil {
		t.Fatal(err)
	}
	var add Action
	for _, a := range Actions([]string{"node"}, nil, nil) {
		if a.Title == "Add a project" {
			add = a
		}
	}
	m, _ := NewModel([]Action{AddProjectAction(add, offers, nil)}).Pick("Add a project")
	return m
}

// TestAFrameworkThePersonPickedIsTickedAndInTheCommand follows the exact keys
// of a first-time user: a website, a framework from the list, a name, Enter
// through the choices and the environments. The pick must be ticked and the
// command must carry it (F-1 of the beta.27 review).
func TestAFrameworkThePersonPickedIsTickedAndInTheCommand(t *testing.T) {
	for name, tc := range map[string]struct {
		downs int
		with  string
	}{"Plain page": {0, "html"}, "React": {1, "react"}, "Svelte": {2, "svelte"}, "Vue": {3, "vue"}} {
		m := press(frameworkForm(t), "enter") // A website
		m = press(m, "down")                  // the first row is "Decide in the next question"
		m = press(m, strings.Split(strings.Repeat("down ", tc.downs), " ")[:tc.downs]...)
		m = press(m, "enter")                     // the framework
		m = press(m, "s", "i", "t", "e", "enter") // the name
		ticked := false
		for i, o := range m.current().Options {
			ticked = ticked || (m.marked[i] && o.Value == tc.with)
		}
		if !ticked {
			t.Fatalf("%s: the pick is not ticked in the choices:\n%s", name, m.View())
		}
		m = press(m, "enter", "enter") // the choices, then "Use these" for the environments
		if got := strings.Join(m.Argv, " "); !strings.Contains(got, "--with "+tc.with) {
			t.Fatalf("%s: the command lacks the pick: %q", name, got)
		}
	}
}

// tickedValues are the values ticked in the choices question on screen.
func tickedValues(m Model) []string {
	var out []string
	for i, o := range m.current().Options {
		if m.marked[i] {
			out = append(out, o.Value)
		}
	}
	return out
}

// TestTheChoicesListOpensOnTheTickedPickOnASmallScreen makes the tick visible
// on a short terminal: the list starts at the framework that was picked.
func TestTheChoicesListOpensOnTheTickedPickOnASmallScreen(t *testing.T) {
	m := press(frameworkForm(t), "enter", "down", "enter", "s", "i", "t", "e", "enter") // Plain page
	if got := tickedValues(m); len(got) != 1 || got[0] != "html" {
		t.Fatalf("html is not ticked: %v", got)
	}
	if m.current().Options[m.cursor].Value != "html" {
		t.Fatalf("the cursor is on %q, so the tick may be out of sight", m.current().Options[m.cursor].Value)
	}
}

// TestAnAnswerChangedAfterGoingBackReplacesTheOldOneEverywhere follows the keys
// of G-1: React, name web, Enter to the plan, Esc back to the framework, Svelte.
func TestAnAnswerChangedAfterGoingBackReplacesTheOldOneEverywhere(t *testing.T) {
	m := press(frameworkForm(t), "enter")         // A website
	m = press(m, "down", "down", "enter")         // React (the first row is "Decide in the next question")
	m = press(m, "w", "e", "b", "enter", "enter") // the name, the choices
	m = press(m, "enter")                         // "Use these": the plan
	if !m.Done() || !strings.Contains(strings.Join(m.Argv, " "), "--with react") {
		t.Fatalf("the first plan is not React: %v", m.Argv)
	}
	m = m.Reopen()
	m = press(m, "esc", "esc", "esc") // the environments, the choices, the name: back at the framework
	if !strings.Contains(m.View(), "› React") {
		t.Fatalf("the framework question starts on the earlier answer:\n%s", m.View())
	}
	m = press(m, "down", "enter") // Svelte
	if !strings.Contains(m.View(), "type to replace it") {
		t.Fatalf("the name does not say that typing replaces it:\n%s", m.View())
	}
	m = press(m, "w", "e", "b") // typing replaces: no "webweb"
	if v := m.View(); !strings.Contains(v, "> web") || strings.Contains(v, "webweb") {
		t.Fatalf("typing appended to the earlier name:\n%s", v)
	}
	m = press(m, "enter")
	if got := tickedValues(m); len(got) != 1 || got[0] != "svelte" {
		t.Fatalf("the choices still hold the old pick: %v", got)
	}
	m = press(m, "enter", "enter")
	if got := strings.Join(m.Argv, " "); !strings.Contains(got, "--with svelte") || strings.Contains(got, "react") || !strings.Contains(got, " web ") {
		t.Fatalf("the second plan: %q", got)
	}
}

// TestOtherTicksSurviveAChangedFramework keeps what the person ticked besides.
func TestOtherTicksSurviveAChangedFramework(t *testing.T) {
	m := press(frameworkForm(t), "enter", "down", "down", "enter", "s", "i", "t", "e", "enter")
	// the choices: react is ticked; also tick vitest (listed after the frameworks)
	for i, o := range m.current().Options {
		if o.Value == "vitest" {
			m.marked[i] = true
		}
	}
	m = press(m, "enter", "enter") // the plan
	m = m.Reopen()
	m = press(m, "esc", "esc", "esc", "down", "enter") // Svelte, the name
	m = press(m, "enter")                              // the choices
	got := tickedValues(m)
	if !slices.Contains(got, "svelte") || slices.Contains(got, "react") {
		t.Fatalf("the framework was not replaced: %v", got)
	}
}

// TestEnterOnTheEnvironmentsUsesTheDefaultsAndTheCommandSaysThem keeps the one
// question short: Use these ends it, with prd as the build and the files kept.
func TestEnterOnTheEnvironmentsUsesTheDefaultsAndTheCommandSaysThem(t *testing.T) {
	m := press(frameworkForm(t), "enter", "down", "down", "enter", "s", "i", "t", "e", "enter", "enter")
	if view := m.View(); !strings.Contains(view, "Use these: lcl, dev, uat, prd") || !strings.Contains(view, "Change the names…") || !strings.Contains(view, "Change the build default, the output or the .env files…") {
		t.Fatalf("the question lacks its options:\n%s", view)
	}
	m = press(m, "enter")
	want := "--environments lcl,dev,uat,prd --default-environment prd --environment-output workspace --env-files commit"
	if got := strings.Join(m.Argv, " "); !strings.HasSuffix(got, want) {
		t.Fatalf("argv %q", got)
	}
}

// TestChangingTheDetailsAsksTheOutputAndTheFiles opens the detailed questions.
func TestChangingTheDetailsAsksTheOutputAndTheFiles(t *testing.T) {
	m := press(frameworkForm(t), "enter", "down", "down", "enter", "s", "i", "t", "e", "enter", "enter", "down", "down", "enter")
	for _, want := range []string{"Which environment is", "Where do the builds go?", "Keep them in git?"} {
		if !strings.Contains(m.View(), want) {
			t.Fatalf("want %q on screen:\n%s", want, m.View())
		}
		m = press(m, "enter")
	}
}

func TestDecidingInTheNextQuestionTicksNothingInTheChoices(t *testing.T) {
	m := press(outcomeForm(t), "enter") // A website
	if !strings.Contains(m.View(), "› Decide in the next question") {
		t.Fatalf("Enter on the first screen must pick nothing:\n%s", m.View())
	}
	m = press(m, "enter")                     // Decide in the next question: where the cursor starts
	m = press(m, "s", "i", "t", "e", "enter") // the name
	if view := m.View(); strings.Contains(view, "[x]") || !strings.Contains(view, "[ ] framework: react") {
		t.Fatalf("deciding later must leave the list empty:\n%s", view)
	}
}

func TestDecidingLaterWithTheWholeFrameworkListTicksNothing(t *testing.T) {
	offers, err := ParseChoices(`[
 {"axis":"framework","name":"html","title":"Plain page","description":"a page","language":"node"},
 {"axis":"framework","name":"react","title":"React","description":"a React app","language":"node"},
 {"axis":"framework","name":"svelte","title":"Svelte","description":"a Svelte app","language":"node"},
 {"axis":"framework","name":"vue","title":"Vue","description":"a Vue app","language":"node"},
 {"axis":"test","name":"vitest","description":"tests","language":"node","default":true}]`)
	if err != nil {
		t.Fatal(err)
	}
	var add Action
	for _, a := range Actions([]string{"node"}, nil, nil) {
		if a.Title == "Add a project" {
			add = a
		}
	}
	m, _ := NewModel([]Action{AddProjectAction(add, offers, nil)}).Pick("Add a project")
	m = press(m, "enter", "enter") // the cursor starts on "Decide in the next question"
	if got := m.Answers()[pickKey]; len(got) != 1 || got[0] != pickNone {
		t.Fatalf("the answer is %v", got)
	}
	m = press(m, "s", "i", "t", "e", "enter")
	if view := m.View(); strings.Contains(view, "[x] framework") {
		t.Fatalf("deciding later must tick no framework:\n%s", view)
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
