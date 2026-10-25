package pluginarchive

import "testing"

func TestPluginNamesAndVersions(t *testing.T) {
	for _, ok := range []string{"widget", "widget-lint", "a1"} {
		if err := ValidatePluginName(ok); err != nil {
			t.Errorf("%q: %v", ok, err)
		}
	}
	for _, bad := range []string{"", "..", "../x", "a/b", "Widget", "-x", "a b", "a.b"} {
		if err := ValidatePluginName(bad); err == nil {
			t.Errorf("%q accepted", bad)
		}
	}
	if ValidatePluginVersion("1.2.3-rc.1+build") != nil || ValidatePluginVersion("../1") == nil || ValidatePluginVersion("") == nil {
		t.Error("version validation is wrong")
	}
}
