package commandtext

import (
	"errors"
	"reflect"
	"testing"
)

func TestWithoutDryRunLineDropsOnlyTheClosingLines(t *testing.T) {
	got := WithoutDryRunLine([]string{"a", DryRunLine(), "b", DryRunLine(), DryRunLine()})
	if !reflect.DeepEqual(got, []string{"a", DryRunLine(), "b"}) {
		t.Fatalf("got %q", got)
	}
}

func TestLinesTrimsTheBlankEnds(t *testing.T) {
	if got := Lines("\n\na\nb\n\n"); !reflect.DeepEqual(got, []string{"a", "b"}) {
		t.Fatalf("got %q", got)
	}
}

func TestFailurePrefersTheOutputThenTheError(t *testing.T) {
	if got := Failure("boom\n", errors.New("x")); got != "boom\n" {
		t.Fatalf("got %q", got)
	}
	if got := Failure("  ", errors.New("x")); got != "x" {
		t.Fatalf("got %q", got)
	}
	if got := Failure("", nil); got == "" {
		t.Fatal("a failure always says something")
	}
}

func TestFirstLineSkipsBlankLines(t *testing.T) {
	if got := FirstLine("\n  \n  hello \nworld"); got != "hello" {
		t.Fatalf("got %q", got)
	}
	if FirstLine("") != "" {
		t.Fatal("no text has no first line")
	}
}
