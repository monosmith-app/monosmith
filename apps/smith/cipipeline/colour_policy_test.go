package cipipeline

import (
	"regexp"
	"strings"
	"testing"
)

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
		if got := colourWanted(envOf(c.env), c.terminal); got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}

func TestStatusGlyphsOfBothSetsHaveTheirColour(t *testing.T) {
	for cell, want := range map[string]tint{
		"✓ passed": tintGreen, "[v] passed": tintGreen, "✗ failed": tintRed, "[x] failed": tintRed,
		"▲ problems": tintYellow, "↷ unchanged": tintDim, "○ none": tintDim, "42": tintNone, "✓": tintGreen, "✓passed": tintNone,
	} {
		if got := glyphTint(cell); got != want {
			t.Errorf("%q: got %v, want %v", cell, got, want)
		}
	}
}

var ansiPattern = regexp.MustCompile("\x1b\\[[0-9;]*m")

func TestColourNeverMovesAColumn(t *testing.T) {
	table := resultTable{
		Headers: []string{"Project", "Result", "Time"}, Right: []bool{false, false, true},
		Rows: []tableRow{
			{Cells: []string{"core", "✓ passed", "12s"}},
			{Cells: []string{"atlassian", "✗ failed", "3m 12s"}, Bad: true, Tints: map[int]tint{0: tintDim}},
		},
	}
	plain := strings.Join(table.render("passed", false), "\n")
	coloured := strings.Join(table.render("passed", true), "\n")
	if !strings.Contains(coloured, "\x1b[32m✓ passed\x1b[0m") || !strings.Contains(coloured, "\x1b[31m✗ failed\x1b[0m") || !strings.Contains(coloured, "\x1b[2matlassian\x1b[0m") {
		t.Errorf("not painted: %q", coloured)
	}
	if got := ansiPattern.ReplaceAllString(coloured, ""); got != plain {
		t.Errorf("colour changed the layout\n%s\nvs\n%s", got, plain)
	}
	if strings.Contains(plain, "\x1b") {
		t.Error("plain output carries ANSI")
	}
	if md := strings.Join(table.Markdown(), "\n"); strings.Contains(md, "\x1b") {
		t.Error("the job summary must stay plain")
	}
}

func TestPaintLeavesTextAloneWhenOffOrUntinted(t *testing.T) {
	if paint("x", tintGreen, false) != "x" || paint("x", tintNone, true) != "x" || paint("", tintGreen, true) != "" {
		t.Error("paint must only colour non-empty tinted text when on")
	}
}
