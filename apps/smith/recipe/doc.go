// Package recipe is smith's recipes: a team's description of which plugin
// fills each level of a project, with what settings, written as data.
//
// A recipe is a TOML file (or an https URL) with a name, a version, at most
// one parent it extends, the choices it asks, the plugins it uses and their
// settings, and when blocks that apply only for certain answers. Load reads a
// recipe and its parents (pinning URLs by content hash), Resolve turns the
// chain and the answers into one set of plugins and settings, each with the
// recipe line it came from, and Validate checks the result with the same
// rules a stack is checked with.
//
// See docs/reference/recipes.md and docs/design/recipes-and-ci.md.
package recipe
