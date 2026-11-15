package guide

import (
	"reflect"
	"strings"
	"testing"
)

func webAnswers(extra map[string][]string) map[string][]string {
	a := map[string][]string{"kind": {"app"}, "name": {"web"}, "lang": {"node"}, withKeyPrefix + "node": {"react"}}
	for k, v := range extra {
		a[k] = v
	}
	return a
}

func TestTheEnvironmentQuestionsAreAskedOnlyForAWebApp(t *testing.T) {
	offers, _ := ParseChoices(offersJSON)
	a := AddProjectAction(plainAdd(), offers, nil)
	var asked []string
	for _, s := range a.Steps {
		if strings.HasPrefix(s.Key, "environment") || s.Key == envFilesKey {
			if s.Skip == nil || !s.Skip(map[string][]string{"lang": {"go"}}) {
				t.Errorf("%s is asked for a Go project", s.Key)
			}
			if s.Skip(webAnswers(nil)) && s.Key == envsKey {
				t.Errorf("a React app is asked which environments")
			}
			asked = append(asked, s.Key)
		}
	}
	if len(asked) != 5 {
		t.Errorf("want 5 environment steps, got %v", asked)
	}
	if got := a.Argv(map[string][]string{"kind": {"lib"}, "name": {"core"}, "lang": {"go"}}); strings.Contains(strings.Join(got, " "), "--environments") {
		t.Errorf("a Go library has no environments: %v", got)
	}
}

func TestTheEnvironmentAnswersBecomeFlags(t *testing.T) {
	offers, _ := ParseChoices(offersJSON)
	a := AddProjectAction(plainAdd(), offers, nil)
	for name, tc := range map[string]struct {
		answers map[string][]string
		want    []string
	}{
		"the default four": {
			webAnswers(map[string][]string{envsKey: {envsDefault}, envDefaultKey: {"uat"}, envOutputKey: {"project"}, envFilesKey: {"example"}}),
			[]string{"--environments", "lcl,dev,uat,prd", "--default-environment", "uat", "--environment-output", "project", "--env-files", "example"},
		},
		"names of my own": {
			webAnswers(map[string][]string{envsKey: {envsCustom}, envNamesKey: {"qa, stage"}, envDefaultKey: {"qa"}, envOutputKey: {"workspace"}, envFilesKey: {"commit"}}),
			[]string{"--environments", "qa,stage", "--default-environment", "qa", "--environment-output", "workspace", "--env-files", "commit"},
		},
		"one build only": {
			webAnswers(map[string][]string{envsKey: {envsNone}}),
			[]string{"--environments", "none"},
		},
	} {
		got := a.Argv(tc.answers)
		if tail := got[len(got)-len(tc.want):]; !reflect.DeepEqual(tail, tc.want) {
			t.Errorf("%s: got %v", name, got)
		}
	}
}

func TestOwnNamesAreAskedWhichOneIsTheDefaultBuildWithTheUsualOneFirst(t *testing.T) {
	for name, tc := range map[string]struct {
		names string
		want  []string
	}{
		"prd wins when it is there": {"lcl, prd, qa", []string{"prd", "qa", "lcl"}},
		"else the last name":        {"qa, stage", []string{"stage", "qa"}},
	} {
		opts := envDefaultOptions(map[string][]string{envsKey: {envsCustom}, envNamesKey: {tc.names}})
		var got []string
		for _, o := range opts {
			got = append(got, o.Value)
		}
		if !reflect.DeepEqual(got, tc.want) || opts[0].Hint != "default" {
			t.Errorf("%s: got %v (first hint %q), want %v", name, got, opts[0].Hint, tc.want)
		}
	}
}

func TestTheEnvironmentNamesAreCheckedAsTheyAreTyped(t *testing.T) {
	var names Step
	for _, s := range webEnvironmentSteps(nil) {
		if s.Key == envNamesKey {
			names = s
		}
	}
	if msg := names.Check(nil, "qa, stage"); msg != "" {
		t.Errorf("lowercase words are fine: %q", msg)
	}
	if msg := names.Check(nil, "QA"); !strings.Contains(msg, "lowercase word") {
		t.Errorf("an upper case name is refused with the fix: %q", msg)
	}
}
