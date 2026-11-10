package dialect

import (
	"strings"
	"testing"
)

func TestSummaryListsWhatSiblingsAgreeOn(t *testing.T) {
	got := Summary([]map[string]string{lib("jest"), lib("jest"), lib("jest")}, "lib")
	if len(got) == 0 || !strings.Contains(got[len(got)-1], "(3 of 3 libs)") {
		t.Fatalf("summary %v", got)
	}
	if got := Summary([]map[string]string{lib("jest")}, "lib"); len(got) != 0 {
		t.Fatalf("one sibling is no habit: %v", got)
	}
}
