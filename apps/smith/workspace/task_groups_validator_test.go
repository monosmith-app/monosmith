package workspace

import (
	"strings"
	"testing"
)

func TestValidateTaskGroups(t *testing.T) {
	build := map[string]TaskOverride{"build": {Outputs: []string{"dist"}}}
	cases := map[string]struct {
		groups []TaskGroup
		want   string
	}{
		"none is fine":              {nil, ""},
		"a complete group":          {[]TaskGroup{{Select: []string{"lang:node", "kind:lib", "tag:x", "libs/*"}, Tasks: build}}, ""},
		"no select":                 {[]TaskGroup{{Tasks: build}}, "has no select"},
		"an empty selector":         {[]TaskGroup{{Select: []string{" "}, Tasks: build}}, "empty selector"},
		"an unknown kind":           {[]TaskGroup{{Select: []string{"kind:library"}, Tasks: build}}, "kind must be"},
		"a language without a name": {[]TaskGroup{{Select: []string{"lang:"}, Tasks: build}}, "nothing after lang:"},
		"an exclusion":              {[]TaskGroup{{Select: []string{"!web"}, Tasks: build}}, "leave out the !"},
		"no tasks":                  {[]TaskGroup{{Select: []string{"kind:lib"}}}, "sets no task"},
	}
	for name, c := range cases {
		err := validateTaskGroups(&Config{TaskGroups: c.groups})
		switch {
		case c.want == "" && err != nil:
			t.Errorf("%s: unexpected error %v", name, err)
		case c.want != "" && (err == nil || !strings.Contains(err.Error(), c.want)):
			t.Errorf("%s: error %v, want it to contain %q", name, err, c.want)
		}
	}
}
