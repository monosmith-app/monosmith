package dev.monosmith.jetbrains.features.smithaction

import dev.monosmith.jetbrains.features.devservers.DevServersUseCase

/** Starts the dev servers of the projects that have a start task. */
class DevServersActionHandler : SmithActionHandler({ ide, root -> DevServersUseCase.start(ide, root) })
