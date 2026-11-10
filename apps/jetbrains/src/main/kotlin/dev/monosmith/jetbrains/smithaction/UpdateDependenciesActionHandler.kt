package dev.monosmith.jetbrains.smithaction

import dev.monosmith.jetbrains.dependencies.UpdateDependenciesUseCase

/** Updates dependencies (`smith up`): grouped by risk, the plan, then the write. */
class UpdateDependenciesActionHandler : SmithActionHandler({ ide, root -> UpdateDependenciesUseCase.start(ide, root) })
