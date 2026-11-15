package dev.monosmith.jetbrains.features.smithaction

import dev.monosmith.jetbrains.features.tauriapp.TauriAppUseCase

/** Makes the icons of a Tauri app from one PNG, after the plan. */
class TauriIconsActionHandler : SmithActionHandler({ ide, root -> TauriAppUseCase.startIcons(ide, root) })
