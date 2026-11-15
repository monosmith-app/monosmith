// Package recipecatalog reads the recipes of a workspace without resolving
// them: ListRecipes finds the recipe files, and Describe says what a loaded
// recipe offers (its choices, the roles it fills, the plugins it configures)
// before any choice is answered. Resolution and validation stay in recipe.
package recipecatalog
