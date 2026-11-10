package doctorview

import (
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/doctor"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestTheBoundaryChecksRunAgainAsSmithBoundaries(t *testing.T) {
	for _, vs := range [][]string{{doctor.BoundaryRulesCheck}, {doctor.BoundaryAppCheck}} {
		if got := RunNowAction(doctor.Finding{Finding: protocol.Finding{Check: vs[0]}}); got != "boundaries" {
			t.Errorf("%q: got %q", vs[0], got)
		}
	}
	if got := RunNowAction(doctor.Finding{Finding: protocol.Finding{Check: "git is installed"}}); got != "" {
		t.Fatalf("a check with no command of its own: got %q", got)
	}
}
