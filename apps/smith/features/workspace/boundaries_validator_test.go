package workspace

import (
	"strings"
	"testing"
)

func TestValidateBoundaries(t *testing.T) {
	cases := map[string]struct {
		cfg  BoundariesSection
		want string
	}{
		"nothing set is fine": {BoundariesSection{}, ""},
		"a complete rule and exception": {BoundariesSection{
			Rules:      []BoundaryRule{{From: "tag:ui", Deny: []string{"tag:server"}}},
			Exceptions: []BoundaryException{{From: "web", To: "api", Reason: "migrating"}},
		}, ""},
		"a rule without from":           {BoundariesSection{Rules: []BoundaryRule{{Deny: []string{"x"}}}}, "has no from"},
		"a rule that says nothing":      {BoundariesSection{Rules: []BoundaryRule{{From: "kind:app"}}}, "neither allow nor deny"},
		"an exception without a reason": {BoundariesSection{Exceptions: []BoundaryException{{From: "a", To: "b"}}}, "needs a reason"},
		"an exception without a target": {BoundariesSection{Exceptions: []BoundaryException{{From: "a", Reason: "x"}}}, "needs from and to"},
	}
	for name, c := range cases {
		err := validateBoundaries(&Config{Boundaries: c.cfg})
		switch {
		case c.want == "" && err != nil:
			t.Errorf("%s: unexpected error %v", name, err)
		case c.want != "" && (err == nil || !strings.Contains(err.Error(), c.want)):
			t.Errorf("%s: error %v, want it to contain %q", name, err, c.want)
		}
	}
}
