package problem

import (
	"errors"
	"testing"
)

func TestErrorPrintsProblemThenEachRemedy(t *testing.T) {
	got := New("no name", "smith new acme", "smith adopt").Error()
	want := "✖ no name\n  → smith new acme\n  → smith adopt"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestErrorUnwrapsItsCause(t *testing.T) {
	sentinel := errors.New("x")
	e := &Error{Problem: "p", Cause: sentinel}
	if !errors.Is(e, sentinel) {
		t.Fatal("errors.Is does not see the cause")
	}
}
