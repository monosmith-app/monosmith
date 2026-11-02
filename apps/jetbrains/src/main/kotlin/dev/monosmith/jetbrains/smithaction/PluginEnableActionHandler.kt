package dev.monosmith.jetbrains.smithaction

import dev.monosmith.jetbrains.pluginswitch.SwitchPluginUseCase

/** Turns a built-in plugin on; smith names nothing it hides. */
class PluginEnableActionHandler : SmithActionHandler({ ide, root -> SwitchPluginUseCase.start(ide, root, on = true) })
