package pluginusage

import "testing"

func TestWarningNamesTheProjectsAThePluginFinds(t *testing.T) {
	used := map[string][]string{"node": {"a", "b", "c", "d", "e", "f", "g"}, "go": {"x"}}
	cases := []struct {
		names []string
		want  string
	}{
		{[]string{"python"}, ""},
		{[]string{"go"}, "go finds 1 project (x)"},
		{[]string{"node", "go"}, "node finds 7 projects (a, b, c, d, e, and 2 more); go finds 1 project (x)"},
	}
	for _, c := range cases {
		if got := Warning(used, c.names); got != c.want {
			t.Errorf("%v: got %q, want %q", c.names, got, c.want)
		}
	}
}
