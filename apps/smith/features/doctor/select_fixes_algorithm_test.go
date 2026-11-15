package doctor

import (
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestFixesSplitsSafeFromThoseThatAsk(t *testing.T) {
	work := &protocol.Fix{Description: "fix go.work", Safe: true, Files: []protocol.FileEdit{{Path: "go.work", Content: "x\n"}}}
	risky := &protocol.Fix{Description: "rewrite ci", Files: []protocol.FileEdit{{Path: "ci.yml", Content: "x\n"}}}
	asks := &protocol.Fix{Description: "pick", Safe: true, Choices: []protocol.FixChoice{{Question: "which?"}}}
	r := Report{Findings: []Finding{
		{Finding: protocol.Finding{Check: "a", Fix: work}},
		{Finding: protocol.Finding{Check: "b", Fix: work}},
		{Finding: protocol.Finding{Check: "c", Fix: risky}},
		{Finding: protocol.Finding{Check: "d", Fix: asks}},
		{Finding: protocol.Finding{Check: "e", OK: true, Fix: risky}},
	}}
	safe, manual := Fixes(r)
	if len(safe) != 1 || len(manual) != 2 {
		t.Fatalf("one fix per distinct change, a question is never safe: %d safe, %d manual", len(safe), len(manual))
	}
}

func TestOnlyNarrowsTheFixesToTheNamedChecksOrFixes(t *testing.T) {
	work := &protocol.Fix{Description: "fix go.work", Safe: true, Files: []protocol.FileEdit{{Path: "go.work", Content: "x\n"}}}
	ci := &protocol.Fix{Description: "regenerate ci", Safe: true, Files: []protocol.FileEdit{{Path: "ci.yml", Content: "y\n"}}}
	r := Report{Findings: []Finding{
		{Finding: protocol.Finding{Check: "go.work exists", Fix: work}},
		{Finding: protocol.Finding{Check: "CI pipelines match the workspace", Fix: ci}},
	}}
	if got := Plan(Only(r, nil)); len(got) != 2 {
		t.Fatalf("with no names every fix stays: %d", len(got))
	}
	got := Plan(Only(r, []string{"go.work exists"}))
	if len(got) != 1 || got[0].Description != "fix go.work" {
		t.Fatalf("only the named check's fix: %+v", got)
	}
	got = Plan(Only(r, []string{"regenerate ci"}))
	if len(got) != 1 || got[0].Checks[0] != "CI pipelines match the workspace" {
		t.Fatalf("a fix is named by its description too: %+v", got)
	}
	if got := Plan(Only(r, []string{"no such check"})); len(got) != 0 {
		t.Fatalf("an unknown name selects nothing: %+v", got)
	}
}
