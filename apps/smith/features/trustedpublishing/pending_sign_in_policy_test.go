package trustedpublishing

import (
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/features/workspace"
)

func settings(pending []Option) []string {
	out := make([]string, len(pending))
	for i, o := range pending {
		out[i] = o.Setting
	}
	return out
}

func TestPendingOffersOnlyUndecidedRegistriesThePipelinesUse(t *testing.T) {
	all := []string{"NUGET_API_KEY", "GEM_HOST_API_KEY", "VSCE_PAT", "PYPI_TOKEN"}
	github := workspace.Config{CI: workspace.CISection{Providers: []string{"github"}}}
	if got := settings(Pending(github, all)); len(got) != 3 {
		t.Errorf("got %v, want the three options", got)
	}
	if got := Pending(github, []string{"PYPI_TOKEN"}); len(got) != 0 {
		t.Errorf("a workspace without the registries got %v", settings(got))
	}
	decided := github
	decided.Release.NuGetTrustedPublishing = new(false)
	decided.Release.VSCEEntra = new(true)
	if got := settings(Pending(decided, all)); len(got) != 1 || got[0] != "rubygems_trusted_publishing" {
		t.Errorf("got %v, want only the undecided RubyGems option", got)
	}
}

func TestPendingOffersNothingOutsideGitHub(t *testing.T) {
	for _, providers := range [][]string{{"azure"}, {"github", "gitlab"}, nil} {
		cfg := workspace.Config{CI: workspace.CISection{Providers: providers}}
		if got := Pending(cfg, []string{"NUGET_API_KEY"}); len(got) != 0 {
			t.Errorf("providers %v got %v", providers, settings(got))
		}
	}
}
