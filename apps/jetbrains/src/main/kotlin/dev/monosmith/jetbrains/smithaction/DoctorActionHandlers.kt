package dev.monosmith.jetbrains.smithaction

import dev.monosmith.jetbrains.health.HealthUseCase
import dev.monosmith.jetbrains.workspacetools.WorkspaceToolsUseCase

/** `smith doctor --json`: tells the summary and prints the checks that did not pass. */
class DoctorActionHandler : SmithActionHandler({ ide, root -> HealthUseCase.runDoctor(ide, root) })

/** `smith doctor --fix`: shows the safe fixes as diffs and applies them after a yes. */
class DoctorFixActionHandler : SmithActionHandler({ ide, root -> HealthUseCase.fixSafe(ide, root) })

/** Shows the divergences from the defaults (override, ejected, drift). */
class DivergencesActionHandler : SmithActionHandler({ ide, root -> WorkspaceToolsUseCase.divergences(ide, root) })
