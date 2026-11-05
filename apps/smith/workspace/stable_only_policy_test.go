package workspace

import "testing"

func TestStableOnlySkipsOnlyPrereleasesOfTheNamedRegistries(t *testing.T) {
	r := &ProjectRelease{PyPI: "acme", NuGet: "acme", StableOnly: []string{"pypi"}}
	cases := []struct {
		registry, version string
		want              bool
	}{
		{"pypi", "2.0.0-beta.47", true},
		{"pypi", "2.0.0", false},
		{"nuget", "2.0.0-beta.47", false},
	}
	for _, c := range cases {
		if got := r.SkipsPrerelease(c.registry, c.version); got != c.want {
			t.Errorf("%s %s: got %v, want %v", c.registry, c.version, got, c.want)
		}
	}
	if (*ProjectRelease)(nil).SkipsPrerelease("pypi", "1.0.0-rc.1") || (&ProjectRelease{PyPI: "acme"}).SkipsPrerelease("pypi", "1.0.0-rc.1") {
		t.Error("an unset stable_only skips nothing")
	}
}

func TestStableOnlyIsValidated(t *testing.T) {
	cases := []struct {
		name string
		r    ProjectRelease
		ok   bool
	}{
		{"pypi set", ProjectRelease{PyPI: "acme", StableOnly: []string{"pypi"}}, true},
		{"npm cannot skip", ProjectRelease{NPM: []string{"acme"}, StableOnly: []string{"npm"}}, false},
		{"registry not set", ProjectRelease{NuGet: "acme", StableOnly: []string{"pypi"}}, false},
		{"twice", ProjectRelease{PyPI: "acme", StableOnly: []string{"pypi", "pypi"}}, false},
	}
	for _, c := range cases {
		if err := validateStableOnly("[projects.cli.release]", &c.r); (err == nil) != c.ok {
			t.Errorf("%s: %v", c.name, err)
		}
	}
}
