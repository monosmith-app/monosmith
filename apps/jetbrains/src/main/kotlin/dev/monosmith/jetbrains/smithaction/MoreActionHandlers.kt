package dev.monosmith.jetbrains.smithaction

import dev.monosmith.jetbrains.adoptfolder.AdoptFolderUseCase
import dev.monosmith.jetbrains.convertfolder.ConvertFolderUseCase
import dev.monosmith.jetbrains.dependencies.UpdateDependenciesUseCase
import dev.monosmith.jetbrains.devservers.DevServersUseCase
import dev.monosmith.jetbrains.pluginmarket.InstallPluginUseCase
import dev.monosmith.jetbrains.runacross.RunAcrossUseCase
import dev.monosmith.jetbrains.settings.EditSettingsUseCase
import dev.monosmith.jetbrains.workspacereport.WriteReportUseCase

/** Runs one task in every project that has it. */
class RunAllActionHandler : SmithActionHandler({ ide, root -> RunAcrossUseCase.start(ide, root, affected = false) })

/** Runs one task in the projects the changes affect (`smith run <task> --affected`). */
class RunAffectedActionHandler : SmithActionHandler({ ide, root -> RunAcrossUseCase.start(ide, root, affected = true) })

/** Updates dependencies (`smith up`): grouped by risk, the plan, then the write. */
class UpdateDependenciesActionHandler : SmithActionHandler({ ide, root -> UpdateDependenciesUseCase.start(ide, root) })

/** Aligns the versions the projects declare (`smith up --align`). */
class AlignDependenciesActionHandler : SmithActionHandler({ ide, root -> UpdateDependenciesUseCase.startAlign(ide, root) })

/** Writes the report page of the workspace and opens it. */
class OpenReportActionHandler : SmithActionHandler({ ide, root -> WriteReportUseCase.start(ide, root) })

/** Searches the plugin marketplace and installs a plugin, after the trust question. */
class InstallPluginActionHandler : SmithActionHandler({ ide, root -> InstallPluginUseCase.start(ide, root) })

/** Edits a setting of smith.toml with its reason and the plan of `smith config set`. */
class EditSettingsActionHandler : SmithActionHandler({ ide, root -> EditSettingsUseCase.start(ide, root) })

/** Starts the dev servers of the projects that have a start task. */
class DevServersActionHandler : SmithActionHandler({ ide, root -> DevServersUseCase.start(ide, root) })

/** Converts an Nx or Turborepo workspace (`smith convert`): the level, the choices, the plan with deletions. */
class ConvertFolderActionHandler : FolderActionHandler({ ide, folder -> ConvertFolderUseCase.start(ide, folder) })

/** Brings a repository in (`smith adopt --init`): the four questions, where to move a root project, the plan. */
class AdoptFolderActionHandler : FolderActionHandler({ ide, folder -> AdoptFolderUseCase.start(ide, folder) })
