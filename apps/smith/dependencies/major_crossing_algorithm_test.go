package dependencies

import (
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestFindDriftHoldsBackMajorCrossings(t *testing.T) {
	declared := protocol.DeclaredResult{
		Declarations: []protocol.Declaration{decl("web", "zod", "^4.0.0", true), decl("api", "zod", "^3.22.0", true), decl("cli", "zod", "^4.1.0", true)},
		Installed:    map[string]string{"zod": "4.1.5"},
	}
	drift := FindDrift("node", declared, nil)
	if len(drift) != 1 || len(drift[0].Fixable) != 2 || len(drift[0].Crossing) != 1 || drift[0].Crossing[0].Project != "api" {
		t.Fatalf("drift %+v", drift)
	}
}
