package doctorfix

import (
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func planned(desc, path string, ask bool, steps ...string) Planned {
	p := Planned{Fix: protocol.Fix{Description: desc, Safe: !ask, Files: []protocol.FileEdit{{Path: path, Content: desc}}}, Checks: []string{desc}, Ask: ask}
	for _, s := range steps {
		p.Steps = append(p.Steps, protocol.FixStep{Argv: []string{s}})
	}
	return p
}

func TestCombineSafeJoinsTheFixesThatNeedNoDecisionAndRunsEachStepOnce(t *testing.T) {
	c, ok := CombineSafe([]Planned{planned("a", "a.txt", false, "install"), planned("risky", "r.txt", true), planned("b", "b.txt", false, "install", "build")})
	if !ok || len(c.Files) != 2 || len(c.Steps) != 2 || len(c.Checks) != 2 || c.Ask || !c.Safe {
		t.Fatalf("%+v ok=%v", c, ok)
	}
	if c.Description != "2 fixes that need no decision" {
		t.Fatalf("description %q", c.Description)
	}
}

func TestCombineSafeKeepsASingleFixAsItIsAndLeavesOutOneThatWritesTheSameFile(t *testing.T) {
	one, ok := CombineSafe([]Planned{planned("a", "x", false), planned("b", "x", false)})
	if !ok || one.Description != "a" || len(one.Files) != 1 {
		t.Fatalf("the second writes the file the first does: %+v", one)
	}
	if _, ok := CombineSafe([]Planned{planned("risky", "r", true)}); ok {
		t.Fatal("a fix that asks was combined")
	}
	if _, ok := CombineSafe(nil); ok {
		t.Fatal("nothing combined into something")
	}
}
