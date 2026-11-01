package workspace

import (
	"strings"
	"testing"
)

func TestValidateGroups(t *testing.T) {
	build := map[string]TaskOverride{"build": {Outputs: []string{"dist"}}}
	cases := map[string]struct {
		groups []Group
		want   string
	}{
		"none is fine":               {nil, ""},
		"a complete group":           {[]Group{{Select: []string{"lang:node", "kind:lib", "tag:x", "libs/*"}, Tasks: build}}, ""},
		"no select":                  {[]Group{{Tasks: build}}, "has no select"},
		"an empty selector":          {[]Group{{Select: []string{" "}, Tasks: build}}, "empty selector"},
		"an unknown kind":            {[]Group{{Select: []string{"kind:library"}, Tasks: build}}, "kind must be"},
		"a language without a name":  {[]Group{{Select: []string{"lang:"}, Tasks: build}}, "nothing after lang:"},
		"an exclusion":               {[]Group{{Select: []string{"!web"}, Tasks: build}}, "leave out the !"},
		"nothing set":                {[]Group{{Select: []string{"kind:lib"}}}, "sets nothing"},
		"tags only":                  {[]Group{{Select: []string{"packages/*"}, Tags: []string{"shared"}}}, ""},
		"an empty tag":               {[]Group{{Select: []string{"packages/*"}, Tags: []string{" "}}}, "empty tag"},
		"a release without a name":   {[]Group{{Select: []string{"kind:app"}, Release: &ProjectRelease{Platforms: []string{"linux/amd64"}}}}, ""},
		"a release naming a package": {[]Group{{Select: []string{"kind:app"}, Release: &ProjectRelease{NPM: []string{"tool"}}}}, "cannot name the package"},
	}
	for name, c := range cases {
		err := validateGroups(&Config{Groups: c.groups})
		switch {
		case c.want == "" && err != nil:
			t.Errorf("%s: unexpected error %v", name, err)
		case c.want != "" && (err == nil || !strings.Contains(err.Error(), c.want)):
			t.Errorf("%s: error %v, want it to contain %q", name, err, c.want)
		}
	}
}
