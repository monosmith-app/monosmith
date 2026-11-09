package dev.monosmith.jetbrains.smithaction

import dev.monosmith.jetbrains.devservers.DevServersUseCase

/** Starts the dev servers of the projects that have a start task. */
class DevServersActionHandler : SmithActionHandler({ ide, root -> DevServersUseCase.start(ide, root) })
