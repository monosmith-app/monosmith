package release

import (
	"path"

	"github.com/monosmith-app/monosmith/apps/smith/githubrelease"
)

func missing(r githubrelease.Release, assets []string) bool {
	have := map[string]bool{}
	for _, a := range r.Assets {
		have[a.Name] = true
	}
	for _, a := range assets {
		if !have[path.Base(a)] {
			return true
		}
	}
	return false
}
