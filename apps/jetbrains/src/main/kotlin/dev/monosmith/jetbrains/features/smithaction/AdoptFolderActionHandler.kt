package dev.monosmith.jetbrains.features.smithaction

import dev.monosmith.jetbrains.features.adoptfolder.AdoptFolderUseCase

/** Brings a repository in (`smith adopt --init`): the four questions, where to move a root project, the plan. */
class AdoptFolderActionHandler : FolderActionHandler({ ide, folder -> AdoptFolderUseCase.start(ide, folder) })
