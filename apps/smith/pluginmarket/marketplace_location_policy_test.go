package pluginmarket

import "testing"

func TestLocationPrefersTheEnvironmentThenTheConfigThenTheDefault(t *testing.T) {
	t.Setenv(LocationEnv, "")
	if got := Location(""); got != DefaultLocation {
		t.Fatal(got)
	}
	if got := Location(" https://a.example/index.json "); got != "https://a.example/index.json" {
		t.Fatal(got)
	}
	t.Setenv(LocationEnv, "https://b.example/index.json")
	if got := Location("https://a.example/index.json"); got != "https://b.example/index.json" {
		t.Fatal(got)
	}
}
