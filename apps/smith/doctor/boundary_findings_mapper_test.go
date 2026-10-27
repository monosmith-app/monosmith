package doctor

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/boundaries"
)

func TestBoundaryFindingsPassWhenNothingBreaksARule(t *testing.T) {
	fs := BoundaryFindings(nil)
	if len(fs) != 1 || !fs[0].OK {
		t.Fatalf("findings %+v", fs)
	}
}

func TestBoundaryFindingsNameTheEdgeTheRuleAndTheHint(t *testing.T) {
	fs := BoundaryFindings([]boundaries.Violation{
		{From: "web", To: "admin", Rule: "an app is deployed", Hint: "remove web's import of admin"},
		{From: "ui", To: "srv", Rule: "no server code", Hint: "x", Warning: true},
	})
	if len(fs) != 2 || fs[0].OK || fs[0].Warning || !fs[1].Warning {
		t.Fatalf("findings %+v", fs)
	}
	if !strings.Contains(fs[0].Detail, "web -> admin: an app is deployed") || !strings.Contains(fs[0].Remedy, "remove web's import of admin") {
		t.Fatalf("finding %+v", fs[0])
	}
	if !strings.Contains(fs[1].Remedy, "kinds = true") {
		t.Fatalf("warning remedy %q", fs[1].Remedy)
	}
}
