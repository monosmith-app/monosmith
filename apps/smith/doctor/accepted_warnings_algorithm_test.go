package doctor

import (
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func warn(check, project string) Finding {
	return Finding{Source: "smith", Finding: protocol.Finding{Check: check, Problem: check + " is off", Project: project, Warning: true}}
}

func TestAnAcceptedWarningPassesAndKeepsItsReason(t *testing.T) {
	out := applyAccepted([]Finding{warn("dist is tracked", "web")}, []workspace.AcceptedWarning{{Check: "dist is tracked", Reason: "the site is served from it"}})
	if len(out) != 1 || !out[0].OK || out[0].Warning || out[0].Detail != "Accepted: the site is served from it" {
		t.Fatalf("got %+v", out)
	}
}

func TestAnEntryWithAProjectAcceptsOnlyThatProject(t *testing.T) {
	out := applyAccepted([]Finding{warn("x", "web"), warn("x", "api")}, []workspace.AcceptedWarning{{Check: "x", Project: "web", Reason: "r"}})
	if !out[0].OK || out[1].OK || !out[1].Warning {
		t.Fatalf("got %+v", out)
	}
}

func TestAFailureCannotBeAccepted(t *testing.T) {
	failed := Finding{Source: "smith", Finding: protocol.Finding{Check: "x", Problem: "x is broken"}}
	out := applyAccepted([]Finding{failed}, []workspace.AcceptedWarning{{Check: "x", Reason: "r"}})
	if out[0].OK {
		t.Fatalf("a failure passed: %+v", out[0])
	}
	if len(out) != 2 || !out[1].Warning {
		t.Fatalf("the entry matched no warning, so it should warn: %+v", out)
	}
}

func TestAnEntryThatMatchesNothingWarns(t *testing.T) {
	out := applyAccepted(nil, []workspace.AcceptedWarning{{Check: "gone", Reason: "r"}})
	if len(out) != 1 || !out[0].Warning || out[0].Remedy == "" {
		t.Fatalf("got %+v", out)
	}
}
