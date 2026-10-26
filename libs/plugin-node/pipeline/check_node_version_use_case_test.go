package pipeline

import (
	"errors"
	"testing"
)

func TestCheckNodeVersion(t *testing.T) {
	real := probe
	defer func() { probe = real }()
	for _, c := range []struct {
		out       string
		err       error
		ok, warns bool
	}{
		{"v24.11.0\n", nil, true, false},
		{"v25.1.0\n", nil, false, true},  // newer than CI warns
		{"v22.3.0\n", nil, false, false}, // older than CI fails
		{"", errors.New("not found"), false, false},
	} {
		probe = func(string, []string) (string, error) { return c.out, c.err }
		f := CheckNodeVersion(".", "24")
		if f.OK != c.ok || f.Warning != c.warns {
			t.Errorf("%q: %+v", c.out, f)
		}
	}
}
