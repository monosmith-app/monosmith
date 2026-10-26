package pipeline

import (
	"errors"
	"strings"
	"testing"
)

func TestCheckPythonVersion(t *testing.T) {
	real := probe
	defer func() { probe = real }()
	var asked string
	for _, c := range []struct {
		out string
		err error
		ok  bool
		// warns: a newer Python or a venv with none warns; an older one fails.
		warns bool
	}{{"3.12\n", nil, true, false}, {"3.11\n", nil, false, false}, {"3.13\n", nil, false, true}, {"", errors.New("missing"), false, true}} {
		probe = func(_ string, argv []string) (string, error) { asked = argv[0]; return c.out, c.err }
		if f := CheckPythonVersion("/ws"); f.OK != c.ok || f.Warning != c.warns {
			t.Errorf("%q: %+v", c.out, f)
		}
	}
	if !strings.HasPrefix(asked, "/ws/.venv/") {
		t.Fatalf("asked %s, not the venv's python", asked)
	}
}
