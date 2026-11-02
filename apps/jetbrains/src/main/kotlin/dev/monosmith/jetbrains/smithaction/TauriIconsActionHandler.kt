package dev.monosmith.jetbrains.smithaction

import dev.monosmith.jetbrains.tauriapp.TauriAppUseCase

/** Makes the icons of a Tauri app from one PNG, after the plan. */
class TauriIconsActionHandler : SmithActionHandler({ ide, root -> TauriAppUseCase.startIcons(ide, root) })
