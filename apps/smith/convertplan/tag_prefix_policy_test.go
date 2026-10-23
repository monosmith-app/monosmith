package convertplan

import (
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/nxworkspace"
)

func TestTagPrefixFollowsTheNewestRelease(t *testing.T) {
	p := &nxworkspace.Project{Name: "@auto/atlassian"}
	const pattern = "{projectName}@{version}"
	cases := map[string]struct {
		tags []string
		want string
	}{
		// A project renamed along the way keeps one stray tag under smith's
		// name; the newer tags under the Nx name decide.
		"stray old tag":       {[]string{"atlassian@1.12.1", "@auto/atlassian@1.12.5", "@auto/atlassian@1.12.6"}, "@auto/atlassian@"},
		"only the Nx name":    {[]string{"@auto/atlassian@1.2.0"}, "@auto/atlassian@"},
		"smith's is newer":    {[]string{"@auto/atlassian@1.2.0", "atlassian@1.3.0"}, "atlassian@"},
		"only smith's":        {[]string{"atlassian@1.3.0"}, "atlassian@"},
		"equal versions":      {[]string{"@auto/atlassian@1.3.0", "atlassian@1.3.0"}, "atlassian@"},
		"no tags":             {nil, "atlassian@"},
		"not versions":        {[]string{"@auto/atlassian@latest", "atlassian@1.0.0"}, "atlassian@"},
		"a prerelease is old": {[]string{"@auto/atlassian@2.0.0-rc.1", "atlassian@2.0.0"}, "atlassian@"},
	}
	for name, c := range cases {
		got, err := TagPrefix(pattern, p, "atlassian@", c.tags)
		if err != nil || got != c.want {
			t.Errorf("%s: %q %v, want %q", name, got, err, c.want)
		}
	}
	if _, err := TagPrefix("v{version}-{projectName}", p, "atlassian@", nil); err == nil {
		t.Error("a pattern with something after {version} cannot become a prefix")
	}
}
