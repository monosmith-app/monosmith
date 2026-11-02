package dev.monosmith.jetbrains.smithaction

import dev.monosmith.jetbrains.settings.EditSettingsUseCase

/** Edits `dependencies.one_version` through the settings editor, with its reason and the plan. */
class OneVersionActionHandler : SmithActionHandler({ ide, root -> EditSettingsUseCase.startOne(ide, root, "dependencies.one_version") })
