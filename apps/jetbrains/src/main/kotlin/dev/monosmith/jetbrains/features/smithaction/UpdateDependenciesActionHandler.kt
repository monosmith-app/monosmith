package dev.monosmith.jetbrains.features.smithaction

import dev.monosmith.jetbrains.features.dependencies.UpdateDependenciesUseCase

/** Updates dependencies (`smith up`): grouped by risk, the plan, then the write. */
class UpdateDependenciesActionHandler : SmithActionHandler({ ide, root -> UpdateDependenciesUseCase.start(ide, root) })
