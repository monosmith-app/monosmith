package pluginmarket

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseNameTellsNamesFromSources(t *testing.T) {
	t.Chdir(t.TempDir())
	cases := []struct {
		arg, name, version string
		ok                 bool
	}{
		{"widget", "widget", "", true},
		{"widget@1.2.0", "widget", "1.2.0", true},
		{"widget@v1.2.0", "widget", "v1.2.0", true},
		{"widget@", "", "", false},
		{"./widget", "", "", false},
		{"dist/widget-1.0.0.smith-plugin.zip", "", "", false},
		{"github:acme/widget@v1", "", "", false},
		{"https://example.com/widget.zip", "", "", false},
		{"C:\\plugins\\widget", "", "", false},
		{"Widget", "", "", false},
		{"", "", "", false},
	}
	for _, c := range cases {
		name, version, ok := ParseName(c.arg)
		if name != c.name || version != c.version || ok != c.ok {
			t.Errorf("%q: got %q %q %v", c.arg, name, version, ok)
		}
	}
}

func TestParseNameLeavesAnExistingPathAlone(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	if err := os.Mkdir(filepath.Join(dir, "widget"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, _, ok := ParseName("widget"); ok {
		t.Fatal("a directory called widget is a path, not a marketplace name")
	}
}
