package commands

import "testing"

func TestNeedsOwnTerminal(t *testing.T) {
	cases := map[string]struct {
		argv []string
		want bool
	}{
		"start of an app":   {[]string{"run", "start", "-p", "web"}, true},
		"start, no project": {[]string{"run", "start"}, true},
		"start and build":   {[]string{"run", "start", "build", "-p", "web"}, false},
		"a build":           {[]string{"run", "build", "-p", "web"}, false},
		"the build key":     {[]string{"build", "start"}, false},
		"another command":   {[]string{"doctor"}, false},
		"nothing":           {nil, false},
	}
	for name, c := range cases {
		if got := NeedsOwnTerminal(c.argv); got != c.want {
			t.Errorf("%s: got %v, want %v", name, got, c.want)
		}
	}
}
