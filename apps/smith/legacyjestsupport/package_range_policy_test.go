package legacyjesthelpers

import "testing"

func TestPackageRangeForFollowsTheReleaseLineOfSmith(t *testing.T) {
	for smith, want := range map[string]string{
		"1.123.0":       "^0.2.0",
		"v1.4.2":        "^0.2.0",
		"0.9.0":         "^0.2.0",
		"2.0.0-beta.1":  "^2.0.0-beta.1",
		"2.0.0-beta.7":  "^2.0.0-beta.1",
		"v2.1.0-rc.3":   "^2.1.0-rc.1",
		"2.0.0":         "^2.0.0",
		"3.4.5":         "^3.0.0",
		"dev":           "^0.2.0",
		"":              "^0.2.0",
		"not-a-version": "^0.2.0",
	} {
		if got := PackageRangeFor(smith); got != want {
			t.Errorf("smith %q: range %q, want %q", smith, got, want)
		}
	}
}
