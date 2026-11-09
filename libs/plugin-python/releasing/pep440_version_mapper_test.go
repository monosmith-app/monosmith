package releasing

import "testing"

func TestDistributionVersionSpellsAPrereleaseTheWayPEP440Does(t *testing.T) {
	for in, want := range map[string]string{
		"1.2.3":         "1.2.3",
		"2.0.0-beta.1":  "2.0.0b1",
		"2.0.0-alpha.3": "2.0.0a3",
		"2.0.0-rc.2":    "2.0.0rc2",
		"2.0.0-beta":    "2.0.0b0",
		"2.0.0-next.1":  "2.0.0-next.1",
		"2.0.0-beta.1x": "2.0.0-beta.1x",
	} {
		if got := DistributionVersion(in); got != want {
			t.Errorf("%s: got %s, want %s", in, got, want)
		}
	}
}

func TestADistributionIsRecognisedByItsVersionAndNeverByAPrefixOfIt(t *testing.T) {
	for name, want := range map[string]bool{
		"monosmith_plugin_sdk-2.0.0b1.tar.gz":            true,
		"monosmith_plugin_sdk-2.0.0b1-py3-none-any.whl":  true,
		"monosmith_plugin_sdk-2.0.0b10.tar.gz":           false,
		"monosmith_plugin_sdk-2.0.0b10-py3-none-any.whl": false,
		"monosmith_plugin_sdk-0.5.0.tar.gz":              false,
	} {
		if got := isDistributionOf(name, "2.0.0-beta.1"); got != want {
			t.Errorf("%s: got %v, want %v", name, got, want)
		}
	}
	if !isDistributionOf("pkg-1.2.3.tar.gz", "1.2.3") || isDistributionOf("pkg-1.2.30.tar.gz", "1.2.3") {
		t.Error("a stable version matches itself and not a longer one")
	}
}
