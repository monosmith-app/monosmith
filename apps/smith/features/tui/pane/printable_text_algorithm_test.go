package pane

import (
	"strings"
	"testing"
)

func TestPrintableDropsTheCarriageReturnOfACRLFLine(t *testing.T) {
	if got := Printable("- const a = 1;\r"); got != "- const a = 1;" {
		t.Fatalf("got %q", got)
	}
}

func TestPrintableExpandsTabsToSpaces(t *testing.T) {
	if got := Printable("\tname: 'x'"); got != "    name: 'x'" {
		t.Fatalf("got %q", got)
	}
}

func TestPrintableKeepsColourCodesAndPlainText(t *testing.T) {
	for _, s := range []string{"plain — text …", "\x1b[31mred\x1b[m"} {
		if got := Printable(s); got != s {
			t.Fatalf("changed %q to %q", s, got)
		}
	}
}

func TestFitCountsAControlLadenLineByWhatIsDrawn(t *testing.T) {
	for _, line := range []string{"\tindented\r", "a\rb\tc\r"} {
		got := Fit(line, 12)
		if Width(got) != 12 || strings.ContainsAny(got, "\r\t") {
			t.Fatalf("Fit(%q) = %q (%d cells)", line, got, Width(got))
		}
	}
}

func TestWrapOfACRLFLineLeavesNoCarriageReturn(t *testing.T) {
	for _, part := range Wrap("one two three four five six seven\r", 10, 2) {
		if strings.ContainsAny(part, "\r\t") || Width(part) > 10 {
			t.Fatalf("part %q (%d cells)", part, Width(part))
		}
	}
}
