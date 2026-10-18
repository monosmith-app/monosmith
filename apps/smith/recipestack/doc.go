// Package recipestack turns a resolved recipe into what smith already
// understands: a stack (the plugins a project is built with) declared in
// smith.toml as [stacks.<recipe>-<answers>], the recipe's settings as
// [projects.<name>.settings.<plugin>], and the recipe the project came from as
// [projects.<name>.recipe]. A project created from a recipe is then an
// ordinary project of that stack: the manifest records the stack's name, the
// graph applies its plugins, and no later command needs the recipe or the
// network to run.
package recipestack
