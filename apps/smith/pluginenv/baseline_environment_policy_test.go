package pluginenv

import "testing"

func TestIsSmithSecret(t *testing.T) {
	for name, want := range map[string]bool{
		"SMITH_CACHE_KEY": true, "SMITH_RELEASE_TOKEN": true, "SMITH_SOMETHING_SECRET": true,
		"SMITH_PLUGIN": false, "SMITH_VERBOSE": false, "NPM_TOKEN": false, "smith_cache_key": true,
	} {
		if IsSmithSecret(name) != want {
			t.Errorf("IsSmithSecret(%q) = %v", name, !want)
		}
	}
}
