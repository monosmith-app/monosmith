package taskrun

import (
	"strings"
	"testing"
)

func TestMissingProgramMessage(t *testing.T) {
	yes := func(string) bool { return true }
	if got := MissingProgramMessage("dotnet", yes); !strings.Contains(got, "dot.net") || !strings.Contains(got, "smith tools install dotnet") {
		t.Fatalf("known tool: %s", got)
	}
	if got := MissingProgramMessage("weird", yes); strings.Contains(got, "Install it") {
		t.Fatalf("unexpected hint: %s", got)
	}
	got := MissingProgramMessage("gradle", func(string) bool { return false })
	if !strings.Contains(got, "gradle.org/install") || !strings.Contains(got, "smith cannot install it") || strings.Contains(got, "smith tools install") {
		t.Fatalf("unknown tool: %s", got)
	}
	if got := MissingProgramMessage("gradle", nil); strings.Contains(got, "smith tools install") {
		t.Fatalf("nil must not offer the command: %s", got)
	}
}
