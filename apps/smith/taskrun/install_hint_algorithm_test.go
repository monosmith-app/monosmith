package taskrun

import (
	"strings"
	"testing"
)

func TestMissingProgramMessage(t *testing.T) {
	if got := MissingProgramMessage("dotnet"); !strings.Contains(got, "dot.net") {
		t.Fatalf("no hint: %s", got)
	}
	if got := MissingProgramMessage("weird"); strings.Contains(got, "Install it") {
		t.Fatalf("unexpected hint: %s", got)
	}
}
