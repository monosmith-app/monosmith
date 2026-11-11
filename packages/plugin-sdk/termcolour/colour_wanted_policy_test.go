package termcolour

import "testing"

func envOf(m map[string]string) func(string) string { return func(k string) string { return m[k] } }

func TestColourFollowsTheCIThatRendersItAndNeverNoColourOrDumb(t *testing.T) {
	cases := []struct {
		name     string
		env      map[string]string
		terminal bool
		want     bool
	}{
		{"github actions", map[string]string{"GITHUB_ACTIONS": "true"}, false, true},
		{"azure pipelines", map[string]string{"TF_BUILD": "True"}, false, true},
		{"gitlab", map[string]string{"CI": "true", "GITLAB_CI": "true"}, false, true},
		{"a terminal", nil, true, true},
		{"piped, no CI", nil, false, false},
		{"another CI, piped", map[string]string{"CI": "true", "JENKINS_URL": "x"}, false, false},
		{"NO_COLOR beats GitHub", map[string]string{"GITHUB_ACTIONS": "true", "NO_COLOR": "1"}, true, false},
		{"dumb terminal", map[string]string{"TERM": "dumb"}, true, false},
		{"dumb beats Azure", map[string]string{"TF_BUILD": "True", "TERM": "dumb"}, false, false},
		{"forced for a sample", map[string]string{"FORCE_COLOR": "1"}, false, true},
		{"NO_COLOR beats forced", map[string]string{"FORCE_COLOR": "1", "NO_COLOR": "1"}, false, false},
	}
	for _, c := range cases {
		if got := Wanted(envOf(c.env), c.terminal); got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}

func TestToolWantedHonoursWhatSmithHandsOnAndNoColourFirst(t *testing.T) {
	cases := []struct {
		name     string
		env      map[string]string
		terminal bool
		want     bool
	}{
		{"handed on by the TUI", map[string]string{ToolColourEnv: "1"}, false, true},
		{"handed on as plain", map[string]string{ToolColourEnv: "0", "FORCE_COLOR": "1"}, true, false},
		{"NO_COLOR beats the hand-on", map[string]string{ToolColourEnv: "1", "NO_COLOR": "1"}, false, false},
		{"dumb beats the hand-on", map[string]string{ToolColourEnv: "1", "TERM": "dumb"}, false, false},
		{"no hand-on: a terminal", nil, true, true},
		{"no hand-on: piped", nil, false, false},
	}
	for _, c := range cases {
		if got := ToolWanted(envOf(c.env), c.terminal); got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
	if ToolEntry(true) != "SMITH_TOOL_COLOR=1" || ToolEntry(false) != "SMITH_TOOL_COLOR=0" {
		t.Error("ToolEntry spells the variable")
	}
}
