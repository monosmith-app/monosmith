package fixwizard

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/doctorfix"
	"github.com/monosmith-app/monosmith/apps/smith/textdiff"
)

// realShow writes a fix the way doctorfix.ShowFix does: the description, a unified diff per file, the steps.
func realShow(p doctorfix.Planned) (string, error) {
	var b strings.Builder
	b.WriteString("● " + p.Description + "\n")
	for _, f := range p.Files {
		b.WriteString(textdiff.UnifiedDiff(f.Path, "one\ntwo\nthree\n", "one\nTWO\nthree\nfour\n"))
	}
	for i, s := range p.Steps {
		b.WriteString("  step " + string(rune('1'+i)) + ": $ " + strings.Join(s.Argv, " ") + "\n")
	}
	return b.String(), nil
}

func TestTheReviewListsFilesFirstAndShowsTheDiffOnDemand(t *testing.T) {
	w := New(safe(), realShow).Place(100, 30)
	text := strings.Join(w.bodyRows(), "\n")
	if !strings.Contains(text, ".gitignore  changed  +2 -1") || strings.Contains(text, "@@") {
		t.Fatalf("the first view is not one line per file:\n%s", text)
	}
	w, _ = press(t, w, "d")
	if text = strings.Join(w.bodyRows(), "\n"); !strings.Contains(text, "@@") || !strings.Contains(text, "+TWO") {
		t.Fatalf("d did not show the diff:\n%s", text)
	}
	w, _ = press(t, w, "d")
	if text = strings.Join(w.bodyRows(), "\n"); strings.Contains(text, "@@") {
		t.Fatalf("d again did not hide the diff:\n%s", text)
	}
}

func TestAReviewWithoutADiffOffersNoDiffKey(t *testing.T) {
	w := New(safe(), show).Place(100, 30)
	for _, b := range w.buttons() {
		if b.key == "d" {
			t.Fatal("a review with no diff offers the diff key")
		}
	}
}
