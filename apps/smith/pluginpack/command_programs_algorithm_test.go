package pluginpack

import "testing"

func TestProgramPattern(t *testing.T) {
	re, ok := programPattern("./bin/smith-plugin-widget-{os}-{arch}{exe}")
	if !ok {
		t.Fatal("not a file of the plugin")
	}
	for name, want := range map[string]bool{
		"bin/smith-plugin-widget-linux-amd64":       true,
		"bin/smith-plugin-widget-windows-arm64.exe": true,
		"bin/smith-plugin-widget-linux-amd64.sig":   false,
		"bin/other-linux-amd64":                     false,
		"x/bin/smith-plugin-widget-linux-amd64":     false,
	} {
		if re.MatchString(name) != want {
			t.Errorf("%s: want %v", name, want)
		}
	}
	for _, not := range []string{"node", "../up/prog", "/abs/prog"} {
		if _, ok := programPattern(not); ok {
			t.Errorf("%q is not shipped", not)
		}
	}
	if re, ok := programPattern("./plugin.py"); !ok || !re.MatchString("plugin.py") || re.MatchString("pluginXpy") {
		t.Error("a literal name must match literally")
	}
}
