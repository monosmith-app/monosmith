package cipipeline

import (
	"regexp"
	"strings"
	"testing"
)

func TestStatusGlyphsOfBothSetsHaveTheirColour(t *testing.T) {
	for cell, want := range map[string]tint{
		"✓ passed": tintGreen, "[v] passed": tintGreen, "× failed": tintRed, "[x] failed": tintRed,
		"▲ problems": tintYellow, "» unchanged": tintDim, "○ none": tintDim, "42": tintNone, "✓": tintGreen, "✓passed": tintNone,
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
			{Cells: []string{"atlassian", "× failed", "3m 12s"}, Bad: true, Tints: map[int]tint{0: tintDim}},
		},
	}
	plain := strings.Join(table.render("passed", false), "\n")
	coloured := strings.Join(table.render("passed", true), "\n")
	if !strings.Contains(coloured, "\x1b[32m✓ passed\x1b[0m") || !strings.Contains(coloured, "\x1b[31m× failed\x1b[0m") || !strings.Contains(coloured, "\x1b[2matlassian\x1b[0m") {
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
