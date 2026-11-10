package workspace

import "testing"

func TestGitHubDeploymentsAreOnUnlessTheWorkspaceTurnsThemOff(t *testing.T) {
	off, on := false, true
	for name, c := range map[string]struct {
		section ReleaseSection
		want    bool
	}{"unset": {ReleaseSection{}, true}, "true": {ReleaseSection{GitHubDeployments: &on}, true}, "false": {ReleaseSection{GitHubDeployments: &off}, false}} {
		if got := c.section.DeploymentsOn(); got != c.want {
			t.Errorf("%s: %v", name, got)
		}
	}
}
