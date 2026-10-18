package recipestack

import (
	"sort"

	"github.com/monosmith-app/monosmith/apps/smith/recipe"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// ProjectSettings keeps the recipe's settings for the plugins the project is
// built with, and lists the plugins whose settings it leaves out.
func ProjectSettings(res recipe.Resolution, st protocol.Stack) (kept map[string]map[string]any, unused []string) {
	used := map[string]bool{st.Language: true}
	for _, name := range append(st.Plugins(), st.Scaffold) {
		used[name] = true
	}
	for plugin, table := range res.Settings {
		values, ok := table.(map[string]any)
		if !ok {
			continue
		}
		if !used[plugin] {
			unused = append(unused, plugin)
			continue
		}
		if kept == nil {
			kept = map[string]map[string]any{}
		}
		kept[plugin] = values
	}
	sort.Strings(unused)
	return kept, unused
}
