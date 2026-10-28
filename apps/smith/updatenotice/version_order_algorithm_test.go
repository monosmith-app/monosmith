package updatenotice

import "testing"

func TestNewer(t *testing.T) {
	cases := []struct {
		a, b string
		want bool
	}{
		{"1.73.0", "1.70.0", true},
		{"1.9.0", "1.10.0", false},
		{"v2.0.0", "1.99.99", true},
		{"1.0.0", "1.0.0", false},
		{"1.2.3-rc.1", "1.2.2", true},
		{"dev", "1.0.0", false},
		{"1.0.0", "", false},
	}
	for _, c := range cases {
		if got := Newer(c.a, c.b); got != c.want {
			t.Errorf("Newer(%q, %q) = %v, want %v", c.a, c.b, got, c.want)
		}
	}
}
