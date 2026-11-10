package operation

import (
	"bufio"
	"bytes"
	"strings"
	"testing"
)

func TestTheFooterAfterCommitDoesNotSuggestCommitAgain(t *testing.T) {
	got := notCommittedText(true, "add app web", "add --commit")
	if strings.Contains(got, "--commit next time") || !strings.Contains(got, "smith commit add app web --only-operation") ||
		!strings.Contains(got, "git status") || !strings.HasPrefix(got, "Not committed.") {
		t.Errorf("footer: %q", got)
	}
	if plain := notCommittedText(false, "add app web", "add --commit"); !strings.Contains(plain, "add --commit next time") {
		t.Errorf("without the flag the footer suggests it: %q", plain)
	}
}

func TestACommitRefusedByADirtyTreeSaysWhatToDo(t *testing.T) {
	g := scratch(t, false)
	put(t, ".", "mine.txt", "my own work\n")
	out, errOut := execute(t, "core", "--commit")
	if !strings.Contains(errOut, "uncommitted changes") {
		t.Fatalf("stderr: %q", errOut)
	}
	if strings.Contains(out, "--commit next time") || !strings.Contains(out, "smith commit add lib core --only-operation") {
		t.Fatalf("footer:\n%s", out)
	}
	if got := commits(t, g); len(got) != 1 {
		t.Fatalf("nothing may be committed: %q", got)
	}
}

func TestTheDirtyTreeChoice(t *testing.T) {
	for answer, want := range map[string]dirtyChoice{
		"\n": chooseOperationFiles, "1\n": chooseOperationFiles, "": chooseOperationFiles,
		"2\n": chooseLeave, "3\n": chooseCancel, "x\n": chooseCancel,
	} {
		var out bytes.Buffer
		if got := askDirtyChoice(bufio.NewReader(strings.NewReader(answer)), &out, []string{"a", "b"}); got != want {
			t.Errorf("answer %q: got %q, want %q", answer, got, want)
		}
		if !strings.Contains(out.String(), "only the files this operation changed (default)") || !strings.Contains(out.String(), "stop here, leave everything uncommitted") {
			t.Errorf("menu: %q", out.String())
		}
	}
}
