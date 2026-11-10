package dev.monosmith.jetbrains.smithaction

import dev.monosmith.jetbrains.runacross.RunAcrossUseCase

/** Runs one task in the projects the changes affect (`smith run <task> --affected`). */
class RunAffectedActionHandler : SmithActionHandler({ ide, root -> RunAcrossUseCase.start(ide, root, affected = true) })
