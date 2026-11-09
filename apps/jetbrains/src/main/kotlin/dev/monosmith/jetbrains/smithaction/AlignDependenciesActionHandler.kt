package dev.monosmith.jetbrains.smithaction

import dev.monosmith.jetbrains.dependencies.UpdateDependenciesUseCase

/** Aligns the versions the projects declare (`smith up --align`). */
class AlignDependenciesActionHandler : SmithActionHandler({ ide, root -> UpdateDependenciesUseCase.startAlign(ide, root) })
