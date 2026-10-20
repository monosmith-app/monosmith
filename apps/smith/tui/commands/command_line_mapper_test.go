package commands

import "testing"

func TestLineQuotesWhatAShellWouldSplit(t *testing.T) {
	cases := []struct {
		argv []string
		want string
	}{
		{[]string{"run", "build", "-p", "kit"}, "smith run build -p kit"},
		{[]string{"doctor", "--fix", "--only", "go.work exists"}, "smith doctor --fix --only 'go.work exists'"},
		{[]string{"add", "app", "web", "--choice", "language=node"}, "smith add app web --choice language=node"},
		{[]string{"x", "it's"}, `smith x 'it'\''s'`},
		{[]string{"x", ""}, "smith x ''"},
		{nil, "smith"},
	}
	for _, c := range cases {
		if got := Line(c.argv); got != c.want {
			t.Errorf("%q = %q, want %q", c.argv, got, c.want)
		}
	}
}
