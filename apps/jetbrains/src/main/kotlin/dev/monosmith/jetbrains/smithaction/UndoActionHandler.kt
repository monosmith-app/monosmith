package dev.monosmith.jetbrains.smithaction

import dev.monosmith.jetbrains.undochange.UndoChangeUseCase

/** `smith undo`: reverts the last operation smith committed, after its plan. */
class UndoActionHandler : SmithActionHandler({ ide, root -> UndoChangeUseCase.start(ide, root) })
