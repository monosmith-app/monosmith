package doctor

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestStillFailingFindsTheChecksTheFixDidNotClear(t *testing.T) {
	report := Report{Findings: []Finding{
		{Finding: protocol.Finding{Check: "golangci-lint is 2.15.0, as in CI", Detail: "found /old/golangci-lint (1.64.8)", Remedy: "move go/bin earlier in PATH"}},
		{Finding: protocol.Finding{Check: "Go is installed", OK: true}},
		{Finding: protocol.Finding{Check: "another failing check"}},
	}}
	got := StillFailing(report, []string{"golangci-lint is 2.15.0, as in CI", "Go is installed"})
	if len(got) != 1 || got[0].Check != "golangci-lint is 2.15.0, as in CI" {
		t.Fatalf("only the named check that still fails: %+v", got)
	}
	text := StillFailingText(got[0])
	for _, want := range []string{"Still not fixed: golangci-lint is 2.15.0", "found /old/golangci-lint (1.64.8)", "→ move go/bin earlier in PATH"} {
		if !strings.Contains(text, want) {
			t.Errorf("text lacks %q: %s", want, text)
		}
	}
}

func TestStillFailingIsEmptyWhenTheFixWorked(t *testing.T) {
	report := Report{Findings: []Finding{{Finding: protocol.Finding{Check: "x", OK: true}}}}
	if got := StillFailing(report, []string{"x"}); len(got) != 0 {
		t.Fatalf("got %+v", got)
	}
}
