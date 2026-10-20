package toolchainpin

import "testing"

func TestAPinIsAVersionPrefixOnDotBoundaries(t *testing.T) {
	for _, c := range []struct {
		pin, installed string
		want           bool
	}{
		{"22", "22.3.0", true}, {"22", "220.0.0", false}, {"22", "20.11.0", false},
		{"1.23", "1.23.4", true}, {"1.23", "1.230.0", false}, {"1.23.4", "1.23.4", true},
		{"1.23.4", "1.23.5", false}, {"v22", "22.1.0", true}, {"", "1.0.0", true},
	} {
		if got := Satisfies(c.pin, c.installed); got != c.want {
			t.Errorf("Satisfies(%q, %q) = %v", c.pin, c.installed, got)
		}
	}
}

func TestInstalledSplitsAProbedToolchain(t *testing.T) {
	if l, v := Installed("node 20.11.0"); l != "node" || v != "20.11.0" {
		t.Errorf("got %q %q", l, v)
	}
	if l, v := Installed(""); l != "" || v != "" {
		t.Errorf("an empty toolchain names nothing: %q %q", l, v)
	}
}
