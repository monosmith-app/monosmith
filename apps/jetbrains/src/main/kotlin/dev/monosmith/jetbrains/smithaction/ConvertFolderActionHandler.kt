package dev.monosmith.jetbrains.smithaction

import dev.monosmith.jetbrains.convertfolder.ConvertFolderUseCase

/** Converts an Nx or Turborepo workspace (`smith convert`): the level, the choices, the plan with deletions. */
class ConvertFolderActionHandler : FolderActionHandler({ ide, folder -> ConvertFolderUseCase.start(ide, folder) })
