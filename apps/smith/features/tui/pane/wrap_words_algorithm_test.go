package pane

import (
	"reflect"
	"testing"
)

func TestWrapWordsBreaksAtSpacesAndKeepsParagraphs(t *testing.T) {
	got := WrapWords("aaa bbb ccc\ndd", 7)
	want := []string{"aaa bbb", "ccc", "dd"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestWrapWordsCutsAWordLongerThanTheWidth(t *testing.T) {
	got := WrapWords("ab abcdefghij", 4)
	want := []string{"ab", "abcd", "efgh", "ij"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestWrapWordsWithNoRoomReturnsTheTextWhole(t *testing.T) {
	if got := WrapWords("a b", 0); !reflect.DeepEqual(got, []string{"a b"}) {
		t.Fatalf("got %q", got)
	}
}

func TestWrapHardCutsAtTheWidthIgnoringSpaces(t *testing.T) {
	got := WrapHard("ab cd ef", 3)
	want := []string{"ab ", "cd ", "ef"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %q, want %q", got, want)
	}
	if got := WrapHard("abc", 3); !reflect.DeepEqual(got, []string{"abc"}) {
		t.Fatalf("a fitting string must stay whole: %q", got)
	}
}
