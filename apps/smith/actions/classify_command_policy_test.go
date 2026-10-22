package actions

import (
	"reflect"
	"testing"
)

func TestClassify(t *testing.T) {
	cases := []struct {
		argv []string
		want Safety
	}{
		{[]string{"run", "build", "-p", "kit"}, Reads},
		{[]string{"test"}, Reads},
		{[]string{"doctor"}, Reads},
		{[]string{"doctor", "--fix"}, Changes},
		{[]string{"release", "--dry-run"}, Reads},
		{[]string{"release"}, Changes},
		{[]string{"install"}, Changes},
		{[]string{"add", "app", "web", "--lang", "node"}, Changes},
		{[]string{"remove", "web"}, Changes},
		{[]string{"upgrade"}, Changes},
		{[]string{"upgrade", "--dry-run"}, Reads},
		{[]string{"ci", "generate"}, Changes},
		{[]string{"config", "set", "ci.providers", "github"}, Changes},
		{[]string{"plugin", "set", "go"}, Changes},
		{[]string{"plugin", "list"}, Reads},
		{[]string{"plugin", "enable", "go"}, Changes},
		{[]string{"plugin", "disable", "go"}, Changes},
		{[]string{"deps", "up", "zod"}, Changes},
		{[]string{"deps", "up", "--align"}, Changes},
		{[]string{"doctor", "--fix", "--only", "go.work exists"}, Changes},
		{[]string{"recipe", "check", "team.recipe.toml"}, Reads},
		{[]string{"recipe", "new", "team.recipe.toml"}, NeedsTerminal},
		{[]string{"recipe", "new", "team.recipe.toml", "--from-answers", "a.toml"}, Changes},
		{[]string{"graph", "--focus", "web"}, Reads},
		{[]string{"graph"}, Reads},
		{[]string{"graph", "--format", "tui"}, NeedsTerminal},
		{[]string{"mcp"}, NeedsTerminal},
		{[]string{"serve"}, NeedsTerminal},
		{[]string{"dev"}, NeedsTerminal},
		{[]string{"adopt"}, Reads},
		{[]string{"adopt", "--json"}, Reads},
		{[]string{"adopt", "--tags"}, Changes},
		{[]string{"adopt", "--init"}, Changes},
		{[]string{"convert", "--dry-run"}, Reads},
		{[]string{"convert"}, Changes},
		{[]string{"up", "--json"}, Reads},
		{[]string{"up", "--yes"}, Changes},
		{[]string{"sync", "--check"}, Reads},
		{[]string{"version"}, Reads},
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
		{[]string{"doctor", "--fix", "--only", "go.work exists"}, []string{"doctor", "--fix", "--only", "go.work exists", "--yes"}},
		{[]string{"release", "--yes"}, []string{"release", "--yes"}},
		{[]string{"install"}, []string{"install"}},
	}
	for _, c := range cases {
		if got := WithConfirmation(c.in); !reflect.DeepEqual(got, c.want) {
			t.Errorf("%v = %v, want %v", c.in, got, c.want)
		}
	}
}
