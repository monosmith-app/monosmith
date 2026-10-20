package commands

import (
	"reflect"
	"testing"
)

func TestClassify(t *testing.T) {
	cases := []struct {
		argv []string
		want Class
	}{
		{[]string{"run", "build", "-p", "kit"}, Free},
		{[]string{"test"}, Free},
		{[]string{"doctor"}, Free},
		{[]string{"doctor", "--fix"}, Changes},
		{[]string{"release", "--dry-run"}, Free},
		{[]string{"release"}, Changes},
		{[]string{"install"}, Changes},
		{[]string{"add", "app", "web", "--lang", "node"}, Changes},
		{[]string{"remove", "web"}, Changes},
		{[]string{"upgrade"}, Changes},
		{[]string{"upgrade", "--dry-run"}, Free},
		{[]string{"ci", "generate"}, Changes},
		{[]string{"config", "set", "ci.providers", "github"}, Changes},
		{[]string{"plugin", "set", "go"}, Changes},
		{[]string{"plugin", "list"}, Free},
		{[]string{"plugin", "enable", "go"}, Changes},
		{[]string{"plugin", "disable", "go"}, Changes},
		{[]string{"deps", "up", "zod"}, Changes},
		{[]string{"deps", "up", "--align"}, Changes},
		{[]string{"graph", "--focus", "web"}, Free},
		{[]string{"graph"}, Free},
		{[]string{"graph", "--format", "tui"}, NeedsTerminal},
		{[]string{"mcp"}, NeedsTerminal},
		{[]string{"up"}, NeedsTerminal},
		{[]string{"new", "acme"}, Changes},
		{[]string{"something-new-and-unknown"}, Changes},
		{nil, NeedsTerminal},
	}
	for _, c := range cases {
		if got := Classify(c.argv); got != c.want {
			t.Errorf("%v = %v, want %v", c.argv, got, c.want)
		}
	}
}

func TestWithConfirmationAddsYesToCommandsThatWouldAskOnATerminal(t *testing.T) {
	cases := []struct{ in, want []string }{
		{[]string{"release"}, []string{"release", "--yes"}},
		{[]string{"remove", "web"}, []string{"remove", "web", "--yes"}},
		{[]string{"doctor", "--fix"}, []string{"doctor", "--fix", "--yes"}},
		{[]string{"release", "--yes"}, []string{"release", "--yes"}},
		{[]string{"install"}, []string{"install"}},
	}
	for _, c := range cases {
		if got := WithConfirmation(c.in); !reflect.DeepEqual(got, c.want) {
			t.Errorf("%v = %v, want %v", c.in, got, c.want)
		}
	}
}
