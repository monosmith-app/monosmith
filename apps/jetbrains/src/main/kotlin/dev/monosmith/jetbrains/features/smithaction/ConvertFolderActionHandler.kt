package dev.monosmith.jetbrains.features.smithaction

import dev.monosmith.jetbrains.features.convertfolder.ConvertFolderUseCase

/** Converts an Nx or Turborepo workspace (`smith convert`): the level, the choices, the plan with deletions. */
class ConvertFolderActionHandler : FolderActionHandler({ ide, folder -> ConvertFolderUseCase.start(ide, folder) })
