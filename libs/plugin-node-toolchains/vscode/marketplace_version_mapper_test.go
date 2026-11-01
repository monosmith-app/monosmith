package vscode

import "testing"

func TestAPrereleaseGetsAnOddMinorAndItsCounterAsThePatch(t *testing.T) {
	for in, want := range map[string]string{
		"1.131.1":       "1.131.1",
		"2.0.0-beta.1":  "2.1.1",
		"2.0.0-beta.12": "2.1.12",
		"2.3.0-rc.2":    "2.7.2",
	} {
		got, err := MarketplaceVersion(in)
		if err != nil || got != want {
			t.Errorf("%s: got %q (%v), want %q", in, got, err, want)
		}
	}
	for _, in := range []string{"2.0.0-beta", "2.0.0-beta.x", "2.0-beta.1"} {
		if got, err := MarketplaceVersion(in); err == nil {
			t.Errorf("%s: got %q, want an error", in, got)
		}
	}
}
