package doctor

import (
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func passed(source, check string) Finding {
	return Finding{Finding: protocol.Finding{Check: check, OK: true}, Source: source}
}

func TestCollapsePassedFoldsChecksThatDifferOnlyByProject(t *testing.T) {
	findings := []Finding{
		passed("node", "npm workspaces include packages/a"),
		passed("node", "npm workspaces include packages/b"),
		passed("node", "a declares the packages it imports"),
		passed("node", "b declares the packages it imports"),
		passed("smith", "the repository has commits"),
		passed("node", "npm workspaces include packages/c"),
		{Finding: protocol.Finding{Check: "go.work exists"}},
		passed("smith", "ab can be verified"),
	}
	names := []string{"a", "b", "c", "ab", "packages/a", "packages/b", "packages/c"}
	got := CollapsePassed(findings, names)
	want := []PassedLine{
		{"npm workspaces include each project (3 projects)", 3},
		{"each project declares the packages it imports (2 projects)", 2},
		{"the repository has commits", 1},
		{"ab can be verified", 1},
	}
	if len(got) != len(want) {
		t.Fatalf("got %+v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("line %d: %+v, want %+v", i, got[i], want[i])
		}
	}
}
