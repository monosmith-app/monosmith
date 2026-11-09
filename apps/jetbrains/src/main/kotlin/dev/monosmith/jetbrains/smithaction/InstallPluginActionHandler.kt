package dev.monosmith.jetbrains.smithaction

import dev.monosmith.jetbrains.pluginmarket.InstallPluginUseCase

/** Searches the plugin marketplace and installs a plugin, after the trust question. */
class InstallPluginActionHandler : SmithActionHandler({ ide, root -> InstallPluginUseCase.start(ide, root) })
