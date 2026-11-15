// Package recipereleases finds out whether a recipe a workspace uses by URL
// has a newer release. A recipe in a recipes repository is released with a
// tag name@x.y.z and its file attached to the GitHub Release; a workspace
// pins a version by pointing its recipe URL at that tag. Parse reads such a
// URL, and Latest asks GitHub for the repository's releases of the recipe.
package recipereleases
