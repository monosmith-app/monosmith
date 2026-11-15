package workspace

import (
	"strings"
	"testing"
)

func TestTagFormatIsReadWhenItNamesTheProject(t *testing.T) {
	cfg, err := Parse([]byte("[workspace]\nname = \"w\"\nscope = \"@acme\"\n[release]\ntag_format = \"{scope}/{name}@\"\n"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Release.TagFormat != "{scope}/{name}@" {
		t.Errorf("tag_format was not read: %q", cfg.Release.TagFormat)
	}
}

func TestTagFormatIsRefusedWhereItCannotWork(t *testing.T) {
	for name, tc := range map[string]struct{ toml, want string }{
		"no name":           {"[release]\ntag_format = \"v\"\n", "no {name}"},
		"scope not set":     {"[release]\ntag_format = \"{scope}/{name}@\"\n", "[workspace] scope is not set"},
		"unknown variable":  {"[workspace]\nscope = \"@a\"\n[release]\ntag_format = \"{name}-{version}\"\n", "other than {name} and {scope}"},
		"unbalanced braces": {"[release]\ntag_format = \"{name}@}\"\n", "other than {name} and {scope}"},
	} {
		head := "[workspace]\nname = \"w\"\n"
		if strings.HasPrefix(tc.toml, "[workspace]") {
			head = ""
			tc.toml = strings.Replace(tc.toml, "[workspace]\n", "[workspace]\nname = \"w\"\n", 1)
		}
		_, err := Parse([]byte(head + tc.toml))
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: err = %v, want %q", name, err, tc.want)
		}
	}
}
