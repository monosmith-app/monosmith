package doctor

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/derive"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestDeriveFindingsNameTheStateTheOwnerAndTheChoice(t *testing.T) {
	file := protocol.DerivedFile{Path: "web/tsconfig.lib.json", Renderer: "node.tsconfig.references"}
	plan := derive.Plan{Items: []derive.Item{
		{File: file, Status: derive.StatusCurrent, Owner: "smith"},
		{File: file, Status: derive.StatusDrift, Owner: "smith", Detail: "edited by hand"},
		{File: protocol.DerivedFile{Path: "a.json", Renderer: "r"}, Status: derive.StatusStale, Owner: "smith", Detail: "inputs changed"},
	}}
	got := deriveFindings(plan)
	if len(got) != 2 {
		t.Fatalf("findings: %+v", got)
	}
	if got[0].OK || !strings.Contains(got[0].Detail, "owner: smith") || !strings.Contains(got[0].Remedy, "--replace") {
		t.Errorf("drift finding: %+v", got[0])
	}
	if got[1].Remedy != "run smith derive" {
		t.Errorf("stale finding: %+v", got[1])
	}
}

func TestDeriveFindingsTellAboutFilesThatDivergeFromTheirSiblingsWithoutFailing(t *testing.T) {
	plan := derive.Plan{Diverging: []derive.Divergence{{Plugin: "node", DivergentFile: protocol.DivergentFile{Path: "a/jest.config.cts", Detail: "diverges from siblings: x"}}}}
	got := deriveFindings(plan)
	last := got[len(got)-1]
	if len(got) != 2 || !last.OK || !strings.Contains(last.Detail, "a/jest.config.cts") {
		t.Fatalf("%+v", got)
	}
}

func TestDeriveFindingsPassWhenNothingNeedsAction(t *testing.T) {
	got := deriveFindings(derive.Plan{})
	if len(got) != 1 || !got[0].OK {
		t.Fatalf("%+v", got)
	}
}
