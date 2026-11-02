package guide

import (
	"slices"
	"strings"
	"testing"
)

const functionOffersJSON = `[
 {"axis":"framework","name":"azure-functions","description":"a serverless Azure Functions app","language":"node","needs":["trigger","runtime","lint"]},
 {"axis":"framework","name":"gcp-functions","description":"a serverless Google Cloud (Cloud Run) function","language":"node","needs":["trigger","runtime","lint"]},
 {"axis":"framework","name":"react","description":"a React app","language":"node","needs":["lint"]},
 {"axis":"lint","name":"eslint","description":"lint","language":"node","default":true},
 {"axis":"trigger","name":"http-trigger","description":"an HTTP request","language":"node","suggested":true},
 {"axis":"trigger","name":"timer-trigger","description":"a schedule","language":"node","conflicts":["gcp-functions"]},
 {"axis":"runtime","name":"node24","description":"Node.js 24","language":"node","suggested":true},
 {"axis":"runtime","name":"node22","description":"Node.js 22","language":"node"},
 {"axis":"framework","name":"azure-functions","description":"a serverless Azure Functions app","language":"python","needs":["trigger","runtime"]},
 {"axis":"trigger","name":"http-trigger","description":"an HTTP request","language":"python","suggested":true},
 {"axis":"runtime","name":"python313","description":"Python 3.13","language":"python","suggested":true}
]`

func TestTheLanguageFilterFindsAFunctionHostByWhatItIsCalled(t *testing.T) {
	offers, err := ParseChoices(functionOffersJSON)
	if err != nil {
		t.Fatal(err)
	}
	form := AddProjectAction(plainAdd(), offers, nil)
	var lang Step
	for _, s := range form.Steps {
		if s.Key == "lang" {
			lang = s
		}
	}
	for _, term := range []string{"function", "functions", "serverless", "azure", "google", "cloud run", "gcp", "react"} {
		found := false
		for _, o := range lang.Options {
			found = found || ((o.Value == "node" || o.Value == "python") && MatchesFilter(o, term))
		}
		if !found {
			t.Errorf("the filter %q finds no language that offers function apps", term)
		}
	}
	for _, o := range lang.Options {
		if o.Value == "go" && MatchesFilter(o, "azure") {
			t.Error("go offers no function host")
		}
	}
}

func TestTheTriggerAndTheRuntimeAreAskedAfterAHostIsTicked(t *testing.T) {
	offers, _ := ParseChoices(functionOffersJSON)
	form := AddProjectAction(plainAdd(), offers, nil)
	steps := map[string]Step{}
	for _, s := range form.Steps {
		steps[s.Key] = s
	}
	multi := steps["with:node"]
	for _, o := range multi.Options {
		if o.Value == "http-trigger" || o.Value == "node24" {
			t.Errorf("%s is asked in its own question, not listed with the frameworks", o.Value)
		}
	}
	trigger, runtime := steps["with:node:trigger"], steps["with:node:runtime"]
	if trigger.Key == "" || runtime.Key == "" || steps["with:node:lint"].Key != "" {
		t.Fatalf("only the axes with no default are asked: %v", slices.Collect(func(yield func(string) bool) {
			for k := range steps {
				if !yield(k) {
					return
				}
			}
		}))
	}
	if trigger.Options[0].Value != "http-trigger" || !strings.Contains(trigger.Options[0].Hint, "(suggested)") || !strings.Contains(trigger.Options[1].Hint, "not with gcp-functions") {
		t.Errorf("the suggested option is first and says so, a conflict is named: %+v", trigger.Options)
	}
	if !trigger.Skip(map[string][]string{"lang": {"node"}, "with:node": {"react"}}) {
		t.Error("a form without a host does not ask the trigger")
	}
	if trigger.Skip(map[string][]string{"with:node": {"azure-functions"}}) {
		t.Error("a ticked host asks the trigger")
	}
	answers := map[string][]string{"lang": {"node"}, "name": {"api"}, "kind": {"app"}, "with:node": {"azure-functions"}, "with:node:trigger": {"http-trigger"}, "with:node:runtime": {"node24"}}
	if argv := strings.Join(form.Argv(answers), " "); !strings.Contains(argv, "--with azure-functions,node24,http-trigger") {
		t.Errorf("argv: %s", argv)
	}
}
