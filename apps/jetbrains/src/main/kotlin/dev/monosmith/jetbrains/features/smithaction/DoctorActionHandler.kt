package dev.monosmith.jetbrains.features.smithaction

import dev.monosmith.jetbrains.features.health.HealthUseCase

/** `smith doctor --json`: tells the summary and prints the checks that did not pass. */
class DoctorActionHandler : SmithActionHandler({ ide, root -> HealthUseCase.runDoctor(ide, root) })
