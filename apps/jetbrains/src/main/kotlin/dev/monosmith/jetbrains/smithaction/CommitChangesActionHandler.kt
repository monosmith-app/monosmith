package dev.monosmith.jetbrains.smithaction

import dev.monosmith.jetbrains.commitchanges.CommitChangesUseCase

/** Commit the changes of the last operation (or everything git shows), after the files are shown. */
class CommitChangesActionHandler : SmithActionHandler({ ide, root -> CommitChangesUseCase.start(ide, root) })
