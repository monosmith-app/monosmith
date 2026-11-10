package releasetable

import (
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

func TestStableOnlyIsValidated(t *testing.T) {
	cases := []struct {
		name string
		r    workspace.ProjectRelease
		ok   bool
	}{
		{"pypi set", workspace.ProjectRelease{PyPI: "acme", StableOnly: []string{"pypi"}}, true},
		{"npm cannot skip", workspace.ProjectRelease{NPM: []string{"acme"}, StableOnly: []string{"npm"}}, false},
		{"registry not set", workspace.ProjectRelease{NuGet: "acme", StableOnly: []string{"pypi"}}, false},
		{"twice", workspace.ProjectRelease{PyPI: "acme", StableOnly: []string{"pypi", "pypi"}}, false},
	}
	for _, c := range cases {
		if err := validateStableOnly("[projects.cli.release]", &c.r); (err == nil) != c.ok {
			t.Errorf("%s: %v", c.name, err)
		}
	}
}
