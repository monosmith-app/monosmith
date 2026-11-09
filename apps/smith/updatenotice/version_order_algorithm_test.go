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
		{"2.0.0-beta.23", "2.0.0-beta.22", true},
		{"2.0.0-beta.10", "2.0.0-beta.9", true},
		{"2.0.0-beta.9", "2.0.0-beta.10", false},
		{"2.0.0", "2.0.0-beta.22", true},
		{"2.0.0-beta.22", "2.0.0", false},
		{"2.0.0-beta.22", "2.0.0-beta.22", false},
		{"dev", "1.0.0", false},
		{"1.0.0", "", false},
	}
	for _, c := range cases {
		if got := Newer(c.a, c.b); got != c.want {
			t.Errorf("Newer(%q, %q) = %v, want %v", c.a, c.b, got, c.want)
		}
	}
}

func TestChannelOf(t *testing.T) {
	for v, want := range map[string]string{"1.73.0": "latest", "2.0.0-beta.22": "next", "v2.1.0-rc.1": "next", "dev": "latest"} {
		if got := ChannelOf(v); got != want {
			t.Errorf("ChannelOf(%q) = %q, want %q", v, got, want)
		}
	}
}
