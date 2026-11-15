package convertplan

import (
	"reflect"
	"testing"
)

func TestChooseTagFormatWritesTheWorkspaceFormatWhenMostTagsFollowIt(t *testing.T) {
	format, exceptions := chooseTagFormat("@acme", []tagChoice{
		{project: "web", prefix: "@acme/web@"},
		{project: "api", prefix: "@acme/api@"},
		{project: "old", prefix: "old@"},
		{project: "new", prefix: "new@", fresh: true},
	})
	if format != ScopedTagFormat {
		t.Fatalf("format %q", format)
	}
	if want := map[string]string{"old": "old@"}; !reflect.DeepEqual(exceptions, want) {
		t.Errorf("exceptions %v, want %v", exceptions, want)
	}
}

func TestChooseTagFormatWritesNothingWhenItDoesNotPayOff(t *testing.T) {
	cases := map[string]struct {
		scope   string
		choices []tagChoice
	}{
		"no scope":        {"", []tagChoice{{project: "web", prefix: "/web@"}}},
		"no scoped tag":   {"@acme", []tagChoice{{project: "web", prefix: "web@"}}},
		"mostly excepted": {"@acme", []tagChoice{{project: "web", prefix: "@acme/web@"}, {project: "a", prefix: "a@"}, {project: "b", prefix: "b@"}}},
		"only fresh":      {"@acme", []tagChoice{{project: "web", prefix: "web@", fresh: true}}},
		"one follower":    {"@acme", []tagChoice{{project: "web", prefix: "@acme/web@"}}},
	}
	for name, c := range cases {
		if format, exceptions := chooseTagFormat(c.scope, c.choices); format != "" || exceptions != nil {
			t.Errorf("%s: format %q, exceptions %v", name, format, exceptions)
		}
	}
}
