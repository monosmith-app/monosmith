// Package pluginrecipe is the recipe language plugin for smith: a recipe file
// (name.recipe.toml) in a folder of its own is a project, versioned from the
// conventional commits that touch it, released with a tag name@x.y.z and
// attached to the GitHub Release, so a recipe URL can pin a version. Its code
// lives in the slice packages below; plugin is the entry point.
package pluginrecipe
