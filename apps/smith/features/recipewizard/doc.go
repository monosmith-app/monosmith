// Package recipewizard is `smith recipe new`: it writes a recipe file by
// asking for it.
//
// What the wizard learns is a Plan: the recipe's name, version and parent, the
// choices it asks, and for each level (language, scaffold, a task slot like
// lint, structure, deploy) either a plugin that is fixed or a plugin per
// answer of a choice. Asking builds a Plan from a person's answers, reading a
// file builds the same Plan (smith recipe new --from-answers), and both go
// the same way after that: the Plan is rendered as TOML, the result is
// validated with the resolver for every combination of answers, and only a
// recipe that resolves is written.
//
// While asking, a plugin is offered for a level only when it fits the plugins
// chosen so far, in every branch of the choices: a Node linter is not offered
// next to Python.
//
// See docs/reference/recipes.md.
package recipewizard
