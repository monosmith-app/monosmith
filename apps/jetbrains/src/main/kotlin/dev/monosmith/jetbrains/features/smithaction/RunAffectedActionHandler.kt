package dev.monosmith.jetbrains.features.smithaction

import dev.monosmith.jetbrains.features.runacross.RunAcrossUseCase

/** Runs one task in the projects the changes affect (`smith run <task> --affected`). */
class RunAffectedActionHandler : SmithActionHandler({ ide, root -> RunAcrossUseCase.start(ide, root, affected = true) })
