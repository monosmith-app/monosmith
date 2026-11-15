package doctorview

import "testing"

func TestAProblemOfManySentencesBecomesAList(t *testing.T) {
	got := SplitProblem("a:build writes x; b:build writes y;  c:build writes z ")
	if len(got) != 3 || got[0] != "a:build writes x" || got[2] != "c:build writes z" {
		t.Fatalf("%q", got)
	}
}

func TestAProblemWithoutSemicolonsIsOneItem(t *testing.T) {
	got := SplitProblem("the lockfile is missing: run install")
	if len(got) != 1 || got[0] != "the lockfile is missing: run install" {
		t.Fatalf("%q", got)
	}
}

func TestAnEmptyProblemHasNoItems(t *testing.T) {
	if got := SplitProblem(""); len(got) != 0 {
		t.Fatalf("%q", got)
	}
}
