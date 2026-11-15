// Package recipechoice answers a recipe's choices for `smith add`: the ones
// given with --choice, then, in a terminal, the rest one question at a time
// with the default pre-selected. With no terminal a missing answer is an
// error that names the choice. It only calls the recipe resolver; what a
// recipe resolves to is decided there.
package recipechoice
