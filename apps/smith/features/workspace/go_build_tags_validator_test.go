package workspace

import (
	"strings"
	"testing"
)

func TestGoBuildTagsMustBeLists(t *testing.T) {
	cases := []struct {
		name     string
		settings map[string]any
		want     string
	}{
		{"valid", map[string]any{"tags": []any{"production", "gtk3"}, "tags_linux": []any{"webkit2_41"}}, ""},
		{"no tags", map[string]any{"other": 1}, ""},
		{"a string", map[string]any{"tags": "production"}, "must be a list"},
		{"a comma", map[string]any{"tags_darwin": []any{"a,b"}}, "not a build tag"},
		{"a number", map[string]any{"tags_windows": []any{3}}, "not a build tag"},
	}
	for _, c := range cases {
		err := validateGoBuildTags(map[string]ProjectOverride{"mvd": {Settings: map[string]map[string]any{"go": c.settings}}})
		switch {
		case c.want == "" && err != nil:
			t.Errorf("%s: %v", c.name, err)
		case c.want != "" && (err == nil || !strings.Contains(err.Error(), c.want) || !strings.Contains(err.Error(), "[projects.mvd.settings.go]")):
			t.Errorf("%s: %v", c.name, err)
		}
	}
}
