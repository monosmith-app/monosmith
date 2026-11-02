package release

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// fakeMover answers each Move call: a package listed in late stays pending; one
// in broken fails.
type fakeMover struct {
	late, broken []string
	calls        []string
}

func (f *fakeMover) Move(project, plugin string, moves []protocol.TagMove) ([]string, []protocol.TagMove, []string, error) {
	var moved, notes []string
	var pending []protocol.TagMove
	for _, m := range moves {
		f.calls = append(f.calls, project+"/"+plugin+"/"+m.Package)
		switch {
		case slices.Contains(f.broken, m.Package):
			return moved, pending, notes, errors.New("npm refused")
		case slices.Contains(f.late, m.Package):
			pending = append(pending, m)
			notes = append(notes, "run: "+m.Command())
		default:
			moved = append(moved, m.Package+"@"+m.Version+" "+m.Tag)
		}
	}
	return moved, pending, notes, nil
}

func move(pkg string) protocol.TagMove {
	return protocol.TagMove{Package: pkg, Version: "2.0.0", Tag: "next", Awaits: []string{pkg + "-linux"}}
}

func TestTheLastPhaseMovesEveryTagAndKeepsTheLateOnesPending(t *testing.T) {
	smith := ReportResult{Item: Item{Project: "smith"}}
	smith.addPending("", move("monosmith"), move("@monosmith/cli"))
	sdk := ReportResult{Item: Item{Project: "sdk"}}
	sdk.addPending("node", move("@acme/sdk"))
	plain := ReportResult{Item: Item{Project: "docs"}}
	report := &Report{Results: []ReportResult{smith, plain, sdk}}
	if got := report.PendingTags(); len(got) != 3 {
		t.Fatalf("before the last phase every tag is pending, with its command: %v", got)
	}
	var said []string
	say := func(format string, args ...any) { said = append(said, fmt.Sprintf(format, args...)) }
	mover := &fakeMover{late: []string{"@acme/sdk"}}
	if err := finalizeTags(report, mover, say); err != nil {
		t.Fatal(err)
	}
	if want := []string{"smith//monosmith", "smith//@monosmith/cli", "sdk/node/@acme/sdk"}; !slices.Equal(mover.calls, want) {
		t.Errorf("calls %v, want %v", mover.calls, want)
	}
	if got := report.PendingTags(); !slices.Equal(got, []string{"npm dist-tag add @acme/sdk@2.0.0 next"}) {
		t.Errorf("only the late tag stays pending: %v", got)
	}
	text := strings.Join(said, "\n")
	for _, want := range []string{"smith: moved the dist-tag: monosmith@2.0.0 next", "sdk: note: run: npm dist-tag add @acme/sdk@2.0.0 next"} {
		if !strings.Contains(text, want) {
			t.Errorf("missing %q in\n%s", want, text)
		}
	}
	if len(report.Results[1].moves) != 0 || strings.Contains(text, "docs") {
		t.Error("a project with nothing pending is left alone")
	}
}

func TestTheLastPhaseDoesNothingWhenNothingIsPending(t *testing.T) {
	mover := &fakeMover{}
	var said []string
	say := func(format string, args ...any) { said = append(said, format) }
	if err := finalizeTags(&Report{Results: []ReportResult{{Item: Item{Project: "x"}}}}, mover, say); err != nil || len(mover.calls) != 0 || len(said) != 0 {
		t.Errorf("calls %v said %v err %v", mover.calls, said, err)
	}
}

func TestAFailedMoveStopsTheLastPhaseAndKeepsTheRestPending(t *testing.T) {
	smith := ReportResult{Item: Item{Project: "smith"}}
	smith.addPending("", move("monosmith"))
	smith.addPending("node", move("@acme/after"))
	report := &Report{Results: []ReportResult{smith}}
	mover := &fakeMover{broken: []string{"monosmith"}}
	err := finalizeTags(report, mover, func(string, ...any) {})
	if err == nil || !strings.Contains(err.Error(), "npm refused") || !strings.Contains(report.Error, "smith") {
		t.Fatalf("error %v, report error %q", err, report.Error)
	}
	for _, want := range []string{"npm dist-tag add monosmith@2.0.0 next", "npm dist-tag add @acme/after@2.0.0 next"} {
		if !slices.Contains(report.PendingTags(), want) {
			t.Errorf("the failed tag and the tag the phase never reached must stay pending: %v", report.PendingTags())
		}
	}
}
