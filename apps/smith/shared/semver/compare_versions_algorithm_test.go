package semver

import "testing"

func TestCompareOrdersVersions(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"1.2.3", "1.2.3", 0},
		{"1.10.0", "1.9.0", 1},
		{"1.2", "1.2.1", -1},
		{"1.0.0-beta.1", "1.0.0", -1},
		{"1.0.0", "1.0.0-beta.1", 1},
		{"1.0.0+build5", "1.0.0", 0},
		{"1.0.0-alpha", "1.0.0-beta", -1},
	}
	for _, c := range cases {
		got := Compare(c.a, c.b)
		if (got < 0) != (c.want < 0) || (got > 0) != (c.want > 0) {
			t.Errorf("Compare(%q, %q) = %d, want sign of %d", c.a, c.b, got, c.want)
		}
	}
}
