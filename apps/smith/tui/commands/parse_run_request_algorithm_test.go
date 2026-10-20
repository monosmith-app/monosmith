package commands

import (
	"reflect"
	"testing"
)

func TestParseRunReadsTasksProjectsAndFlags(t *testing.T) {
	cases := []struct {
		argv []string
		want RunRequest
		ok   bool
	}{
		{[]string{"run", "lint", "test", "--affected"}, RunRequest{Tasks: []string{"lint", "test"}, Affected: true}, true},
		{[]string{"run", "build", "-p", "kit,web", "--no-cache"}, RunRequest{Tasks: []string{"build"}, Projects: []string{"kit", "web"}}, true},
		{[]string{"test", "web", "--parallel", "2"}, RunRequest{Tasks: []string{"test"}, Projects: []string{"web"}}, true},
		{[]string{"run", "build", "--projects=kind:app", "--base", "origin/main"}, RunRequest{Tasks: []string{"build"}, Projects: []string{"kind:app"}, Base: "origin/main"}, true},
		{[]string{"doctor"}, RunRequest{}, false},
		{[]string{"run"}, RunRequest{}, false},
		{nil, RunRequest{}, false},
	}
	for _, c := range cases {
		got, ok := ParseRun(c.argv)
		if ok != c.ok || (ok && !reflect.DeepEqual(got, c.want)) {
			t.Errorf("%v = %+v %v, want %+v %v", c.argv, got, ok, c.want, c.ok)
		}
	}
}

func TestRunArgv(t *testing.T) {
	if got := RunArgv("test", "kit"); !reflect.DeepEqual(got, []string{"run", "test", "-p", "kit"}) {
		t.Fatal(got)
	}
	if got := RunArgv("test", ""); !reflect.DeepEqual(got, []string{"run", "test"}) {
		t.Fatal(got)
	}
}
