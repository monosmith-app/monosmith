package toolrun

import "testing"

func TestCIPinFinding(t *testing.T) {
	for _, c := range []struct {
		have, ci      string
		found, ok, wn bool
	}{
		{"1.90.2", "1.90", true, true, false},
		{"1.76.0", "1.90", true, false, false},
		{"1.95.0", "1.90", true, false, true},
		{"v24.11.0", "24", true, true, false},
		{"22.3.0", "24", true, false, false},
		{"3.13.14", "3.12", true, false, true},
		{"1.90.0", "stable", false, false, false},
		{"", "1.90", false, false, false},
	} {
		f, found := CIPinFinding("Rust", c.have, c.ci)
		if found != c.found || f.OK != c.ok || f.Warning != c.wn {
			t.Errorf("%s vs %s: found=%v %+v", c.have, c.ci, found, f)
		}
	}
}
