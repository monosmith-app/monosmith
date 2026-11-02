package doctorfix

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestSummarySaysWhichFixesYesAppliesAndWhichNeedAChoice(t *testing.T) {
	if Summary(nil) != "" {
		t.Fatal("an empty plan has nothing to summarise")
	}
	got := Summary([]Planned{
		{Fix: protocol.Fix{Description: "regenerate go.work"}},
		{Fix: protocol.Fix{Description: "tell the Cargo workspace about tools/x"}, Ask: true},
	})
	for _, want := range []string{"2 fix(es) available", "--yes applies the 1 safe fix(es)", `"regenerate go.work"`, "1 need(s) your decision", `"tell the Cargo workspace about tools/x"`, "--choose N"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in\n%s", want, got)
		}
	}
	if strings.Contains(Summary([]Planned{{Fix: protocol.Fix{Description: "a"}}}), "need(s) your decision") {
		t.Error("no asking fix, no decision line")
	}
}
