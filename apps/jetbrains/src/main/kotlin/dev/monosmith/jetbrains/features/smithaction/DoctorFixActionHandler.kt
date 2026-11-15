package dev.monosmith.jetbrains.features.smithaction

import dev.monosmith.jetbrains.features.health.HealthUseCase

/** `smith doctor --fix`: shows the safe fixes as diffs and applies them after a yes. */
class DoctorFixActionHandler : SmithActionHandler({ ide, root -> HealthUseCase.fixSafe(ide, root) })
