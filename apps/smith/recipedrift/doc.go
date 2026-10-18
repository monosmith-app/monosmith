// Package recipedrift compares each project made from a recipe with that
// recipe as it is now. A project records the recipe, its version and its
// answers when `smith add` creates it; Check resolves the workspace's current
// recipe with those answers and says, in the recipe's terms, what has
// changed: the version, the plugins and the plugins' settings. Plan turns the
// same comparison into the edit of smith.toml that `smith upgrade` makes, and
// Origins tells `smith explain` where each of a project's settings came from.
//
// See docs/reference/recipes.md.
package recipedrift
