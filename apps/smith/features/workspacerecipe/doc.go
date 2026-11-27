// Package workspacerecipe reads the recipe a workspace already follows:
// `smith recipe new --from-workspace`.
//
// A workspace says which plugins its projects use without a recipe: each
// project has a language plugin and a stack (smith.toml [projects.<name>]
// stack, its manifest, or [default_stacks]), and smith.toml turns a code
// layout on with a [conventions.<architecture>] table. The package gathers
// those facts (Load, which opens the workspace) and turns them, for one
// language, into the levels a recipe fixes (Read, a pure mapper): the
// language, the scaffold, the toolchain, a task slot such as lint or test for
// a toolchain plugin that fills exactly one, the structure and the deploy
// plugins. The result is the level=plugin values `smith recipe new --fixed`
// takes, so the recipe is checked and written the same way as one written by
// hand.
//
// See docs/manual/recipe-new.md.
package workspacerecipe
