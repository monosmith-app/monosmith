package releaseconfirm

import (
	"strings"
	"testing"
)

func TestAReleaseThatEndedWellListsTheTagsAndPointsToTheStatus(t *testing.T) {
	title, why, lines := Result(true, []string{"web@1.1.0", "kit@0.1.0"}, ready())
	out := strings.Join(lines, "\n")
	if title != "Released" || why == "" {
		t.Fatalf("title %q, why %q", title, why)
	}
	for _, want := range []string{"New tags:", "web@1.1.0", "kit@0.1.0", "What next: smith release --status"} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q:\n%s", want, out)
		}
	}
}

func TestTagsThatStayedLocalSayHowToPushThem(t *testing.T) {
	for _, f := range []Facts{{HasRemote: false}, {HasRemote: true, NoPush: true}} {
		_, _, lines := Result(true, []string{"web@1.1.0"}, f)
		if out := strings.Join(lines, "\n"); !strings.Contains(out, "git push --tags") {
			t.Fatalf("no push advice for %+v:\n%s", f, out)
		}
	}
}

func TestAReleaseThatStoppedListsNoTagAndSaysWhatToDo(t *testing.T) {
	title, why, lines := Result(false, []string{"web@1.1.0"}, ready())
	out := strings.Join(lines, "\n")
	if title != "The release stopped" || !strings.Contains(why, "run panel of Projects") {
		t.Fatalf("title %q, why %q", title, why)
	}
	if strings.Contains(out, "New tags") || !strings.Contains(out, "release again") {
		t.Fatalf("lines:\n%s", out)
	}
}
