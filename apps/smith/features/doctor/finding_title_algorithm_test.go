package doctor

import (
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestTitleNamesTheProblemOfAFailedFinding(t *testing.T) {
	f := Finding{Finding: protocol.Finding{Check: "golangci-lint is 2.15.0, as in CI", Problem: "golangci-lint is 2.14.0, CI uses 2.15.0"}}
	if got := f.Title(); got != f.Problem {
		t.Fatalf("Title() = %q, want the problem %q", got, f.Problem)
	}
}

func TestTitleKeepsTheRuleForAPassedFindingAndForOneWithoutAProblem(t *testing.T) {
	passed := Finding{Finding: protocol.Finding{Check: "Gradle is installed", OK: true, Problem: "stale"}}
	if got := passed.Title(); got != passed.Check {
		t.Fatalf("passed Title() = %q, want the rule %q", got, passed.Check)
	}
	bare := Finding{Finding: protocol.Finding{Check: "Gradle is installed"}}
	if got := bare.Title(); got != bare.Check {
		t.Fatalf("Title() without a problem = %q, want the rule %q", got, bare.Check)
	}
}
