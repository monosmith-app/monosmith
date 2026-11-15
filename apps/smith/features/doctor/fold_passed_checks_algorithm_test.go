package doctor

import (
	"strings"
	"testing"
)

func TestFoldPassedKeepsShortListsAndCountsLongOnes(t *testing.T) {
	short := make([]PassedLine, foldAbove)
	for i := range short {
		short[i] = PassedLine{Text: "check", Count: 1}
	}
	if got := FoldPassed(short); len(got) != foldAbove {
		t.Fatalf("a short list must stay: %v", got)
	}
	long := append(short, PassedLine{Text: "npm workspaces include each project (5 projects)", Count: 5})
	got := FoldPassed(long)
	if len(got) != 1 || !strings.HasPrefix(got[0].Text, "13 checks passed") || !strings.Contains(got[0].Text, "--verbose") {
		t.Fatalf("a long list is one count line: %v", got)
	}
}
