package upgrade

import (
	"errors"
	"fmt"

	"github.com/monosmith-app/monosmith/apps/smith/recipe"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

// recipeChanges handles the recipes smith.toml names. A recipe loaded from a
// URL is pinned by content hash in smith.recipes.lock, and a changed hash is
// refused until the upgrade accepts it. With check it only reports the
// recipes whose content changed; otherwise it accepts them and says what
// changed. A recipe that is a path, or a workspace that names none, has
// nothing to accept.
func recipeChanges(loader *recipe.Loader, cfg workspace.Config, check bool) ([]string, error) {
	refs := workspace.RecipeRefs(cfg)
	if len(refs) == 0 {
		return nil, nil
	}
	if check {
		var out []string
		for _, ref := range refs {
			_, err := loader.Load(ref)
			var changed *recipe.ChangedError
			switch {
			case errors.As(err, &changed):
				out = append(out, fmt.Sprintf("%s: changed since it was pinned (sha256 %s, now %s)", changed.URL, short(changed.Locked), short(changed.Got)))
			case err != nil:
				return out, fmt.Errorf("recipe %s: %w", ref, err)
			}
		}
		return out, nil
	}
	changes, err := loader.AcceptAll()
	if err != nil {
		return nil, fmt.Errorf("accepting the pinned recipes: %w", err)
	}
	var out []string
	for _, c := range changes {
		if c.Changed() {
			out = append(out, fmt.Sprintf("%s: accepted (sha256 %s, was %s)", c.URL, short(c.New), short(c.Old)))
		}
	}
	return out, nil
}

func short(sum string) string {
	if len(sum) > 12 {
		return sum[:12]
	}
	if sum == "" {
		return "unpinned"
	}
	return sum
}
