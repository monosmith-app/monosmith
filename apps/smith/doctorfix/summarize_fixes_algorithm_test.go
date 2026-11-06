package doctorfix

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestHeadlineCountsFixesByHowTheyAreApplied(t *testing.T) {
	ask := Planned{Fix: protocol.Fix{Description: "b"}, Ask: true}
	safe := Planned{Fix: protocol.Fix{Description: "a"}}
	for _, tc := range []struct {
		plan []Planned
		want string
	}{
		{nil, ""},
		{[]Planned{safe}, "1 fix available: all is applied by --yes"},
		{[]Planned{safe, safe}, "2 fixes available: all are applied by --yes"},
		{[]Planned{safe, safe, ask}, "3 fixes available: 2 applied by --yes, 1 needs a choice"},
		{[]Planned{ask, ask}, "2 fixes available: 0 applied by --yes, 2 need a choice"},
	} {
		if got := Headline(tc.plan); got != tc.want {
			t.Errorf("Headline = %q, want %q", got, tc.want)
		}
	}
}

func TestReminderRepeatsTheCountAndNamesTheSafeCommand(t *testing.T) {
	ask := Planned{Fix: protocol.Fix{Description: "b"}, Ask: true}
	if got := Reminder([]Planned{{Fix: protocol.Fix{Description: "a"}}, ask}); !strings.Contains(got, "2 fixes available") || !strings.Contains(got, "smith doctor --fix --yes") {
		t.Errorf("%q", got)
	}
	if got := Reminder([]Planned{ask}); strings.Contains(got, "smith doctor --fix --yes") || got == "" {
		t.Errorf("only an asking fix: no safe command: %q", got)
	}
}

func TestSummarySaysWhichFixesYesAppliesAndWhichNeedAChoice(t *testing.T) {
	if Summary(nil) != "" {
		t.Fatal("an empty plan has nothing to summarise")
	}
	got := Summary([]Planned{
		{Fix: protocol.Fix{Description: "regenerate go.work"}},
		{Fix: protocol.Fix{Description: "tell the Cargo workspace about tools/x"}, Ask: true},
	})
	for _, want := range []string{"2 fixes available: 1 applied by --yes, 1 needs a choice", "--yes applies the 1 safe fix(es)", `"regenerate go.work"`, "1 need(s) your decision", `"tell the Cargo workspace about tools/x"`, "--choose N"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in\n%s", want, got)
		}
	}
	if strings.Contains(Summary([]Planned{{Fix: protocol.Fix{Description: "a"}}}), "need(s) your decision") {
		t.Error("no asking fix, no decision line")
	}
}
