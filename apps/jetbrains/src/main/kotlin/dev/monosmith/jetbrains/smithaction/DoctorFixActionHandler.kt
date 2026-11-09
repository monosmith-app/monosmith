package dev.monosmith.jetbrains.smithaction

import dev.monosmith.jetbrains.health.HealthUseCase

/** `smith doctor --fix`: shows the safe fixes as diffs and applies them after a yes. */
class DoctorFixActionHandler : SmithActionHandler({ ide, root -> HealthUseCase.fixSafe(ide, root) })
