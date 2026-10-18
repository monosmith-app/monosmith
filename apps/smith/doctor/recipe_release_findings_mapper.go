package doctor

import (
	"fmt"
	"os"
	"sort"

	"github.com/monosmith-app/monosmith/apps/smith/recipereleases"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// latestRelease asks for the newest release of a recipe; tests replace it.
var latestRelease = func(loc recipereleases.Location) (string, error) {
	return recipereleases.Latest(nil, loc)
}

// recipeReleaseChecks reports a recipe the workspace pins by URL (a release
// tag of a recipes repository) that has a newer release. A recipe that is a
// path, a branch URL or any other URL is not pinned to a release and has
// nothing to report, and neither has one GitHub cannot be asked about
// (offline, rate limited): doctor says nothing rather than something wrong.
// SMITH_NO_UPDATE_CHECK turns the lookup off.
func recipeReleaseChecks(s *session.Session) []protocol.Finding {
	if os.Getenv("SMITH_NO_UPDATE_CHECK") != "" {
		return nil
	}
	refs := map[string]bool{}
	if s.Config.Workspace.Recipe != "" {
		refs[s.Config.Workspace.Recipe] = true
	}
	for _, ref := range s.Config.Recipes {
		refs[ref] = true
	}
	ordered := make([]string, 0, len(refs))
	for ref := range refs {
		ordered = append(ordered, ref)
	}
	sort.Strings(ordered)
	var out []protocol.Finding
	for _, ref := range ordered {
		loc, ok := recipereleases.Parse(ref)
		if !ok {
			continue
		}
		latest, err := latestRelease(loc)
		if err != nil {
			continue
		}
		f := protocol.Finding{Check: fmt.Sprintf("recipe %s %s is the newest release", loc.Name, loc.Version)}
		if recipereleases.Newer(latest, loc.Version) {
			f.Warning = true
			f.Detail = fmt.Sprintf("%s %s is released, and smith.toml pins %s", loc.Name, latest, loc.Version)
			f.Remedy = "point the recipe at the new release: " + loc.Pinned(latest) + "; then `smith upgrade` accepts its content and brings the projects up to date"
		} else {
			f.OK = true
		}
		out = append(out, f)
	}
	return out
}
