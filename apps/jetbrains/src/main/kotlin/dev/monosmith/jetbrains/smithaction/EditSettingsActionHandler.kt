package dev.monosmith.jetbrains.smithaction

import dev.monosmith.jetbrains.settings.EditSettingsUseCase

/** Edits a setting of smith.toml with its reason and the plan of `smith config set`. */
class EditSettingsActionHandler : SmithActionHandler({ ide, root -> EditSettingsUseCase.start(ide, root) })
