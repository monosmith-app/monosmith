package dev.monosmith.jetbrains.smithaction

import dev.monosmith.jetbrains.newrecipe.NewRecipeUseCase

/** Write a recipe that fixes plugins: one question per level, from `smith recipe new --list-levels --json`. */
class NewRecipeActionHandler : SmithActionHandler({ ide, root -> NewRecipeUseCase.start(ide, root) })
