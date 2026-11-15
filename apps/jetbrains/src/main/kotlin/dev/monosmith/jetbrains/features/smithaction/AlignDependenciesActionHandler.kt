package dev.monosmith.jetbrains.features.smithaction

import dev.monosmith.jetbrains.features.dependencies.UpdateDependenciesUseCase

/** Aligns the versions the projects declare (`smith up --align`). */
class AlignDependenciesActionHandler : SmithActionHandler({ ide, root -> UpdateDependenciesUseCase.startAlign(ide, root) })
