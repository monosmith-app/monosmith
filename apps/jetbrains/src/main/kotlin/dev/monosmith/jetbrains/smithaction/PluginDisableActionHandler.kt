package dev.monosmith.jetbrains.smithaction

import dev.monosmith.jetbrains.pluginswitch.SwitchPluginUseCase

/** Turns a built-in plugin off; smith names the projects it hides, and a second yes forces it. */
class PluginDisableActionHandler : SmithActionHandler({ ide, root -> SwitchPluginUseCase.start(ide, root, on = false) })
