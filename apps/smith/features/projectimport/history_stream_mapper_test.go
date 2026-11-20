package projectimport

import (
	"strings"
	"testing"
)

func TestHistoryStreamPlacesEveryCommitUnderTheFolder(t *testing.T) {
	record := func(hash, parents, tree, msg string) string {
		return strings.Join([]string{hash, parents, tree, "Ann", "ann@example.com", "1700000000 +0100", "Bob", "bob@example.com", "1700000100 +0000", msg}, "\x00") + "\x01\n"
	}
	log := record("a1", "", emptyTree, "first\n") + record("b2", "a1", "t2", "second\n") + record("c3", "b2 x9", "t3", "merge\n")
	stream, n := HistoryStream(log, "refs/heads/smith-import", "apps/api")
	if n != 3 {
		t.Fatalf("commits = %d, want 3", n)
	}
	for _, want := range []string{
		"commit refs/heads/smith-import\nmark :1\nauthor Ann <ann@example.com> 1700000000 +0100\ncommitter Bob <bob@example.com> 1700000100 +0000\ndata 6\nfirst\n\ndeleteall\n\n",
		"mark :2\n", "from :1\n", "M 040000 t2 apps/api\n",
		"mark :3\n", "from :2\n", "M 040000 t3 apps/api\n",
	} {
		if !strings.Contains(stream, want) {
			t.Errorf("the stream lacks %q:\n%s", want, stream)
		}
	}
	if strings.Contains(stream, "merge :") {
		t.Error("a parent outside the history (x9) is left out")
	}
	if strings.Contains(stream, emptyTree) {
		t.Error("the empty tree is written as no file, not as a folder")
	}
}
