package dev.monosmith.jetbrains.smithaction

import dev.monosmith.jetbrains.tauriapp.TauriAppUseCase

/** Makes a Node app a Tauri app, after the plan; asks the bundle identifier. */
class TauriAddActionHandler : SmithActionHandler({ ide, root -> TauriAppUseCase.startAdd(ide, root) })
