package dev.monosmith.jetbrains.words

import java.text.MessageFormat
import java.util.ResourceBundle

/**
 * Generated from libs/words/catalog by `go run ./wordgen` in libs/words. Do not edit: change the catalogue and run it again.
 */
object SmithWordsConfig {
    /** The base name of the resource bundle (messages.properties). */
    const val BUNDLE = "messages"

    /** The key action.MonoSmith.AddLibrary.description. */
    const val ACTION_MONOSMITH_ADDLIBRARY_DESCRIPTION = "action.MonoSmith.AddLibrary.description"

    /** The key action.MonoSmith.AddLibrary.text. */
    const val ACTION_MONOSMITH_ADDLIBRARY_TEXT = "action.MonoSmith.AddLibrary.text"

    /** The key action.MonoSmith.AddPipeline.description. */
    const val ACTION_MONOSMITH_ADDPIPELINE_DESCRIPTION = "action.MonoSmith.AddPipeline.description"

    /** The key action.MonoSmith.AddPipeline.text. */
    const val ACTION_MONOSMITH_ADDPIPELINE_TEXT = "action.MonoSmith.AddPipeline.text"

    /** The key action.MonoSmith.AddProject.description. */
    const val ACTION_MONOSMITH_ADDPROJECT_DESCRIPTION = "action.MonoSmith.AddProject.description"

    /** The key action.MonoSmith.AddProject.text. */
    const val ACTION_MONOSMITH_ADDPROJECT_TEXT = "action.MonoSmith.AddProject.text"

    /** The key action.MonoSmith.AdoptFolder.description. */
    const val ACTION_MONOSMITH_ADOPTFOLDER_DESCRIPTION = "action.MonoSmith.AdoptFolder.description"

    /** The key action.MonoSmith.AdoptFolder.text. */
    const val ACTION_MONOSMITH_ADOPTFOLDER_TEXT = "action.MonoSmith.AdoptFolder.text"

    /** The key action.MonoSmith.AgentFile.description. */
    const val ACTION_MONOSMITH_AGENTFILE_DESCRIPTION = "action.MonoSmith.AgentFile.description"

    /** The key action.MonoSmith.AgentFile.text. */
    const val ACTION_MONOSMITH_AGENTFILE_TEXT = "action.MonoSmith.AgentFile.text"

    /** The key action.MonoSmith.AlignDependencies.description. */
    const val ACTION_MONOSMITH_ALIGNDEPENDENCIES_DESCRIPTION = "action.MonoSmith.AlignDependencies.description"

    /** The key action.MonoSmith.AlignDependencies.text. */
    const val ACTION_MONOSMITH_ALIGNDEPENDENCIES_TEXT = "action.MonoSmith.AlignDependencies.text"

    /** The key action.MonoSmith.Bootstrap.description. */
    const val ACTION_MONOSMITH_BOOTSTRAP_DESCRIPTION = "action.MonoSmith.Bootstrap.description"

    /** The key action.MonoSmith.Bootstrap.text. */
    const val ACTION_MONOSMITH_BOOTSTRAP_TEXT = "action.MonoSmith.Bootstrap.text"

    /** The key action.MonoSmith.CacheClear.description. */
    const val ACTION_MONOSMITH_CACHECLEAR_DESCRIPTION = "action.MonoSmith.CacheClear.description"

    /** The key action.MonoSmith.CacheClear.text. */
    const val ACTION_MONOSMITH_CACHECLEAR_TEXT = "action.MonoSmith.CacheClear.text"

    /** The key action.MonoSmith.CachePrune.description. */
    const val ACTION_MONOSMITH_CACHEPRUNE_DESCRIPTION = "action.MonoSmith.CachePrune.description"

    /** The key action.MonoSmith.CachePrune.text. */
    const val ACTION_MONOSMITH_CACHEPRUNE_TEXT = "action.MonoSmith.CachePrune.text"

    /** The key action.MonoSmith.CacheRepair.description. */
    const val ACTION_MONOSMITH_CACHEREPAIR_DESCRIPTION = "action.MonoSmith.CacheRepair.description"

    /** The key action.MonoSmith.CacheRepair.text. */
    const val ACTION_MONOSMITH_CACHEREPAIR_TEXT = "action.MonoSmith.CacheRepair.text"

    /** The key action.MonoSmith.CacheStatus.description. */
    const val ACTION_MONOSMITH_CACHESTATUS_DESCRIPTION = "action.MonoSmith.CacheStatus.description"

    /** The key action.MonoSmith.CacheStatus.text. */
    const val ACTION_MONOSMITH_CACHESTATUS_TEXT = "action.MonoSmith.CacheStatus.text"

    /** The key action.MonoSmith.CacheVerify.description. */
    const val ACTION_MONOSMITH_CACHEVERIFY_DESCRIPTION = "action.MonoSmith.CacheVerify.description"

    /** The key action.MonoSmith.CacheVerify.text. */
    const val ACTION_MONOSMITH_CACHEVERIFY_TEXT = "action.MonoSmith.CacheVerify.text"

    /** The key action.MonoSmith.ChangeKind.description. */
    const val ACTION_MONOSMITH_CHANGEKIND_DESCRIPTION = "action.MonoSmith.ChangeKind.description"

    /** The key action.MonoSmith.ChangeKind.text. */
    const val ACTION_MONOSMITH_CHANGEKIND_TEXT = "action.MonoSmith.ChangeKind.text"

    /** The key action.MonoSmith.Commit.description. */
    const val ACTION_MONOSMITH_COMMIT_DESCRIPTION = "action.MonoSmith.Commit.description"

    /** The key action.MonoSmith.Commit.text. */
    const val ACTION_MONOSMITH_COMMIT_TEXT = "action.MonoSmith.Commit.text"

    /** The key action.MonoSmith.ComposeStack.description. */
    const val ACTION_MONOSMITH_COMPOSESTACK_DESCRIPTION = "action.MonoSmith.ComposeStack.description"

    /** The key action.MonoSmith.ComposeStack.text. */
    const val ACTION_MONOSMITH_COMPOSESTACK_TEXT = "action.MonoSmith.ComposeStack.text"

    /** The key action.MonoSmith.ConvertFolder.description. */
    const val ACTION_MONOSMITH_CONVERTFOLDER_DESCRIPTION = "action.MonoSmith.ConvertFolder.description"

    /** The key action.MonoSmith.ConvertFolder.text. */
    const val ACTION_MONOSMITH_CONVERTFOLDER_TEXT = "action.MonoSmith.ConvertFolder.text"

    /** The key action.MonoSmith.Copies.description. */
    const val ACTION_MONOSMITH_COPIES_DESCRIPTION = "action.MonoSmith.Copies.description"

    /** The key action.MonoSmith.Copies.text. */
    const val ACTION_MONOSMITH_COPIES_TEXT = "action.MonoSmith.Copies.text"

    /** The key action.MonoSmith.DeriveCheck.description. */
    const val ACTION_MONOSMITH_DERIVECHECK_DESCRIPTION = "action.MonoSmith.DeriveCheck.description"

    /** The key action.MonoSmith.DeriveCheck.text. */
    const val ACTION_MONOSMITH_DERIVECHECK_TEXT = "action.MonoSmith.DeriveCheck.text"

    /** The key action.MonoSmith.DeriveFiles.description. */
    const val ACTION_MONOSMITH_DERIVEFILES_DESCRIPTION = "action.MonoSmith.DeriveFiles.description"

    /** The key action.MonoSmith.DeriveFiles.text. */
    const val ACTION_MONOSMITH_DERIVEFILES_TEXT = "action.MonoSmith.DeriveFiles.text"

    /** The key action.MonoSmith.DerivePreview.description. */
    const val ACTION_MONOSMITH_DERIVEPREVIEW_DESCRIPTION = "action.MonoSmith.DerivePreview.description"

    /** The key action.MonoSmith.DerivePreview.text. */
    const val ACTION_MONOSMITH_DERIVEPREVIEW_TEXT = "action.MonoSmith.DerivePreview.text"

    /** The key action.MonoSmith.DevServers.description. */
    const val ACTION_MONOSMITH_DEVSERVERS_DESCRIPTION = "action.MonoSmith.DevServers.description"

    /** The key action.MonoSmith.DevServers.text. */
    const val ACTION_MONOSMITH_DEVSERVERS_TEXT = "action.MonoSmith.DevServers.text"

    /** The key action.MonoSmith.Divergences.description. */
    const val ACTION_MONOSMITH_DIVERGENCES_DESCRIPTION = "action.MonoSmith.Divergences.description"

    /** The key action.MonoSmith.Divergences.text. */
    const val ACTION_MONOSMITH_DIVERGENCES_TEXT = "action.MonoSmith.Divergences.text"

    /** The key action.MonoSmith.DockerDescription.description. */
    const val ACTION_MONOSMITH_DOCKERDESCRIPTION_DESCRIPTION = "action.MonoSmith.DockerDescription.description"

    /** The key action.MonoSmith.DockerDescription.text. */
    const val ACTION_MONOSMITH_DOCKERDESCRIPTION_TEXT = "action.MonoSmith.DockerDescription.text"

    /** The key action.MonoSmith.Doctor.description. */
    const val ACTION_MONOSMITH_DOCTOR_DESCRIPTION = "action.MonoSmith.Doctor.description"

    /** The key action.MonoSmith.Doctor.text. */
    const val ACTION_MONOSMITH_DOCTOR_TEXT = "action.MonoSmith.Doctor.text"

    /** The key action.MonoSmith.DoctorFix.description. */
    const val ACTION_MONOSMITH_DOCTORFIX_DESCRIPTION = "action.MonoSmith.DoctorFix.description"

    /** The key action.MonoSmith.DoctorFix.text. */
    const val ACTION_MONOSMITH_DOCTORFIX_TEXT = "action.MonoSmith.DoctorFix.text"

    /** The key action.MonoSmith.EditSettings.description. */
    const val ACTION_MONOSMITH_EDITSETTINGS_DESCRIPTION = "action.MonoSmith.EditSettings.description"

    /** The key action.MonoSmith.EditSettings.text. */
    const val ACTION_MONOSMITH_EDITSETTINGS_TEXT = "action.MonoSmith.EditSettings.text"

    /** The key action.MonoSmith.EnvironmentsAdd.description. */
    const val ACTION_MONOSMITH_ENVIRONMENTSADD_DESCRIPTION = "action.MonoSmith.EnvironmentsAdd.description"

    /** The key action.MonoSmith.EnvironmentsAdd.text. */
    const val ACTION_MONOSMITH_ENVIRONMENTSADD_TEXT = "action.MonoSmith.EnvironmentsAdd.text"

    /** The key action.MonoSmith.EnvironmentsList.description. */
    const val ACTION_MONOSMITH_ENVIRONMENTSLIST_DESCRIPTION = "action.MonoSmith.EnvironmentsList.description"

    /** The key action.MonoSmith.EnvironmentsList.text. */
    const val ACTION_MONOSMITH_ENVIRONMENTSLIST_TEXT = "action.MonoSmith.EnvironmentsList.text"

    /** The key action.MonoSmith.EnvironmentsRemove.description. */
    const val ACTION_MONOSMITH_ENVIRONMENTSREMOVE_DESCRIPTION = "action.MonoSmith.EnvironmentsRemove.description"

    /** The key action.MonoSmith.EnvironmentsRemove.text. */
    const val ACTION_MONOSMITH_ENVIRONMENTSREMOVE_TEXT = "action.MonoSmith.EnvironmentsRemove.text"

    /** The key action.MonoSmith.ExplainProject.description. */
    const val ACTION_MONOSMITH_EXPLAINPROJECT_DESCRIPTION = "action.MonoSmith.ExplainProject.description"

    /** The key action.MonoSmith.ExplainProject.text. */
    const val ACTION_MONOSMITH_EXPLAINPROJECT_TEXT = "action.MonoSmith.ExplainProject.text"

    /** The key action.MonoSmith.Info.description. */
    const val ACTION_MONOSMITH_INFO_DESCRIPTION = "action.MonoSmith.Info.description"

    /** The key action.MonoSmith.Info.text. */
    const val ACTION_MONOSMITH_INFO_TEXT = "action.MonoSmith.Info.text"

    /** The key action.MonoSmith.InstallAgent.description. */
    const val ACTION_MONOSMITH_INSTALLAGENT_DESCRIPTION = "action.MonoSmith.InstallAgent.description"

    /** The key action.MonoSmith.InstallAgent.text. */
    const val ACTION_MONOSMITH_INSTALLAGENT_TEXT = "action.MonoSmith.InstallAgent.text"

    /** The key action.MonoSmith.InstallPackage.description. */
    const val ACTION_MONOSMITH_INSTALLPACKAGE_DESCRIPTION = "action.MonoSmith.InstallPackage.description"

    /** The key action.MonoSmith.InstallPackage.text. */
    const val ACTION_MONOSMITH_INSTALLPACKAGE_TEXT = "action.MonoSmith.InstallPackage.text"

    /** The key action.MonoSmith.InstallPlugin.description. */
    const val ACTION_MONOSMITH_INSTALLPLUGIN_DESCRIPTION = "action.MonoSmith.InstallPlugin.description"

    /** The key action.MonoSmith.InstallPlugin.text. */
    const val ACTION_MONOSMITH_INSTALLPLUGIN_TEXT = "action.MonoSmith.InstallPlugin.text"

    /** The key action.MonoSmith.InstallTools.description. */
    const val ACTION_MONOSMITH_INSTALLTOOLS_DESCRIPTION = "action.MonoSmith.InstallTools.description"

    /** The key action.MonoSmith.InstallTools.text. */
    const val ACTION_MONOSMITH_INSTALLTOOLS_TEXT = "action.MonoSmith.InstallTools.text"

    /** The key action.MonoSmith.McpInstall.description. */
    const val ACTION_MONOSMITH_MCPINSTALL_DESCRIPTION = "action.MonoSmith.McpInstall.description"

    /** The key action.MonoSmith.McpInstall.text. */
    const val ACTION_MONOSMITH_MCPINSTALL_TEXT = "action.MonoSmith.McpInstall.text"

    /** The key action.MonoSmith.NewRecipe.description. */
    const val ACTION_MONOSMITH_NEWRECIPE_DESCRIPTION = "action.MonoSmith.NewRecipe.description"

    /** The key action.MonoSmith.NewRecipe.text. */
    const val ACTION_MONOSMITH_NEWRECIPE_TEXT = "action.MonoSmith.NewRecipe.text"

    /** The key action.MonoSmith.NewWorkspace.description. */
    const val ACTION_MONOSMITH_NEWWORKSPACE_DESCRIPTION = "action.MonoSmith.NewWorkspace.description"

    /** The key action.MonoSmith.NewWorkspace.text. */
    const val ACTION_MONOSMITH_NEWWORKSPACE_TEXT = "action.MonoSmith.NewWorkspace.text"

    /** The key action.MonoSmith.OneVersion.description. */
    const val ACTION_MONOSMITH_ONEVERSION_DESCRIPTION = "action.MonoSmith.OneVersion.description"

    /** The key action.MonoSmith.OneVersion.text. */
    const val ACTION_MONOSMITH_ONEVERSION_TEXT = "action.MonoSmith.OneVersion.text"

    /** The key action.MonoSmith.OpenReport.description. */
    const val ACTION_MONOSMITH_OPENREPORT_DESCRIPTION = "action.MonoSmith.OpenReport.description"

    /** The key action.MonoSmith.OpenReport.text. */
    const val ACTION_MONOSMITH_OPENREPORT_TEXT = "action.MonoSmith.OpenReport.text"

    /** The key action.MonoSmith.PluginDisable.description. */
    const val ACTION_MONOSMITH_PLUGINDISABLE_DESCRIPTION = "action.MonoSmith.PluginDisable.description"

    /** The key action.MonoSmith.PluginDisable.text. */
    const val ACTION_MONOSMITH_PLUGINDISABLE_TEXT = "action.MonoSmith.PluginDisable.text"

    /** The key action.MonoSmith.PluginEnable.description. */
    const val ACTION_MONOSMITH_PLUGINENABLE_DESCRIPTION = "action.MonoSmith.PluginEnable.description"

    /** The key action.MonoSmith.PluginEnable.text. */
    const val ACTION_MONOSMITH_PLUGINENABLE_TEXT = "action.MonoSmith.PluginEnable.text"

    /** The key action.MonoSmith.RegenerateCi.description. */
    const val ACTION_MONOSMITH_REGENERATECI_DESCRIPTION = "action.MonoSmith.RegenerateCi.description"

    /** The key action.MonoSmith.RegenerateCi.text. */
    const val ACTION_MONOSMITH_REGENERATECI_TEXT = "action.MonoSmith.RegenerateCi.text"

    /** The key action.MonoSmith.Release.description. */
    const val ACTION_MONOSMITH_RELEASE_DESCRIPTION = "action.MonoSmith.Release.description"

    /** The key action.MonoSmith.Release.text. */
    const val ACTION_MONOSMITH_RELEASE_TEXT = "action.MonoSmith.Release.text"

    /** The key action.MonoSmith.ReleasePreview.description. */
    const val ACTION_MONOSMITH_RELEASEPREVIEW_DESCRIPTION = "action.MonoSmith.ReleasePreview.description"

    /** The key action.MonoSmith.ReleasePreview.text. */
    const val ACTION_MONOSMITH_RELEASEPREVIEW_TEXT = "action.MonoSmith.ReleasePreview.text"

    /** The key action.MonoSmith.RemoveProject.description. */
    const val ACTION_MONOSMITH_REMOVEPROJECT_DESCRIPTION = "action.MonoSmith.RemoveProject.description"

    /** The key action.MonoSmith.RemoveProject.text. */
    const val ACTION_MONOSMITH_REMOVEPROJECT_TEXT = "action.MonoSmith.RemoveProject.text"

    /** The key action.MonoSmith.RunAffected.description. */
    const val ACTION_MONOSMITH_RUNAFFECTED_DESCRIPTION = "action.MonoSmith.RunAffected.description"

    /** The key action.MonoSmith.RunAffected.text. */
    const val ACTION_MONOSMITH_RUNAFFECTED_TEXT = "action.MonoSmith.RunAffected.text"

    /** The key action.MonoSmith.RunAll.description. */
    const val ACTION_MONOSMITH_RUNALL_DESCRIPTION = "action.MonoSmith.RunAll.description"

    /** The key action.MonoSmith.RunAll.text. */
    const val ACTION_MONOSMITH_RUNALL_TEXT = "action.MonoSmith.RunAll.text"

    /** The key action.MonoSmith.SwapChoice.description. */
    const val ACTION_MONOSMITH_SWAPCHOICE_DESCRIPTION = "action.MonoSmith.SwapChoice.description"

    /** The key action.MonoSmith.SwapChoice.text. */
    const val ACTION_MONOSMITH_SWAPCHOICE_TEXT = "action.MonoSmith.SwapChoice.text"

    /** The key action.MonoSmith.SyncVersions.description. */
    const val ACTION_MONOSMITH_SYNCVERSIONS_DESCRIPTION = "action.MonoSmith.SyncVersions.description"

    /** The key action.MonoSmith.SyncVersions.text. */
    const val ACTION_MONOSMITH_SYNCVERSIONS_TEXT = "action.MonoSmith.SyncVersions.text"

    /** The key action.MonoSmith.TabCompletion.description. */
    const val ACTION_MONOSMITH_TABCOMPLETION_DESCRIPTION = "action.MonoSmith.TabCompletion.description"

    /** The key action.MonoSmith.TabCompletion.text. */
    const val ACTION_MONOSMITH_TABCOMPLETION_TEXT = "action.MonoSmith.TabCompletion.text"

    /** The key action.MonoSmith.TauriAdd.description. */
    const val ACTION_MONOSMITH_TAURIADD_DESCRIPTION = "action.MonoSmith.TauriAdd.description"

    /** The key action.MonoSmith.TauriAdd.text. */
    const val ACTION_MONOSMITH_TAURIADD_TEXT = "action.MonoSmith.TauriAdd.text"

    /** The key action.MonoSmith.TauriIcons.description. */
    const val ACTION_MONOSMITH_TAURIICONS_DESCRIPTION = "action.MonoSmith.TauriIcons.description"

    /** The key action.MonoSmith.TauriIcons.text. */
    const val ACTION_MONOSMITH_TAURIICONS_TEXT = "action.MonoSmith.TauriIcons.text"

    /** The key action.MonoSmith.ToolsList.description. */
    const val ACTION_MONOSMITH_TOOLSLIST_DESCRIPTION = "action.MonoSmith.ToolsList.description"

    /** The key action.MonoSmith.ToolsList.text. */
    const val ACTION_MONOSMITH_TOOLSLIST_TEXT = "action.MonoSmith.ToolsList.text"

    /** The key action.MonoSmith.TryCommand.description. */
    const val ACTION_MONOSMITH_TRYCOMMAND_DESCRIPTION = "action.MonoSmith.TryCommand.description"

    /** The key action.MonoSmith.TryCommand.text. */
    const val ACTION_MONOSMITH_TRYCOMMAND_TEXT = "action.MonoSmith.TryCommand.text"

    /** The key action.MonoSmith.Undo.description. */
    const val ACTION_MONOSMITH_UNDO_DESCRIPTION = "action.MonoSmith.Undo.description"

    /** The key action.MonoSmith.Undo.text. */
    const val ACTION_MONOSMITH_UNDO_TEXT = "action.MonoSmith.Undo.text"

    /** The key action.MonoSmith.UnlinkLibrary.description. */
    const val ACTION_MONOSMITH_UNLINKLIBRARY_DESCRIPTION = "action.MonoSmith.UnlinkLibrary.description"

    /** The key action.MonoSmith.UnlinkLibrary.text. */
    const val ACTION_MONOSMITH_UNLINKLIBRARY_TEXT = "action.MonoSmith.UnlinkLibrary.text"

    /** The key action.MonoSmith.UpdateDependencies.description. */
    const val ACTION_MONOSMITH_UPDATEDEPENDENCIES_DESCRIPTION = "action.MonoSmith.UpdateDependencies.description"

    /** The key action.MonoSmith.UpdateDependencies.text. */
    const val ACTION_MONOSMITH_UPDATEDEPENDENCIES_TEXT = "action.MonoSmith.UpdateDependencies.text"

    /** The key action.MonoSmith.Why.description. */
    const val ACTION_MONOSMITH_WHY_DESCRIPTION = "action.MonoSmith.Why.description"

    /** The key action.MonoSmith.Why.text. */
    const val ACTION_MONOSMITH_WHY_TEXT = "action.MonoSmith.Why.text"

    /** The key action.MonoSmith.WhyAnything.description. */
    const val ACTION_MONOSMITH_WHYANYTHING_DESCRIPTION = "action.MonoSmith.WhyAnything.description"

    /** The key action.MonoSmith.WhyAnything.text. */
    const val ACTION_MONOSMITH_WHYANYTHING_TEXT = "action.MonoSmith.WhyAnything.text"

    /** The key action.MonoSmith.WhyMiss.description. */
    const val ACTION_MONOSMITH_WHYMISS_DESCRIPTION = "action.MonoSmith.WhyMiss.description"

    /** The key action.MonoSmith.WhyMiss.text. */
    const val ACTION_MONOSMITH_WHYMISS_TEXT = "action.MonoSmith.WhyMiss.text"

    /** The key adopt.agent.decline. */
    const val ADOPT_AGENT_DECLINE = "adopt.agent.decline"

    /** The key adopt.found.close. */
    const val ADOPT_FOUND_CLOSE = "adopt.found.close"

    /** The key adopt.found.continue. */
    const val ADOPT_FOUND_CONTINUE = "adopt.found.continue"

    /** The key adopt.review.convert. */
    const val ADOPT_REVIEW_CONVERT = "adopt.review.convert"

    /** The key commands.describe.bootstrap.title. */
    const val COMMANDS_DESCRIBE_BOOTSTRAP_TITLE = "commands.describe.bootstrap.title"

    /** The key commands.file.compose.title. */
    const val COMMANDS_FILE_COMPOSE_TITLE = "commands.file.compose.title"

    /** The key commands.file.new.title. */
    const val COMMANDS_FILE_NEW_TITLE = "commands.file.new.title"

    /** The key commands.group.release. */
    const val COMMANDS_GROUP_RELEASE = "commands.group.release"

    /** The key commands.group.start. */
    const val COMMANDS_GROUP_START = "commands.group.start"

    /** The key commands.proj.swap.title. */
    const val COMMANDS_PROJ_SWAP_TITLE = "commands.proj.swap.title"

    /** The key commands.view.deps.title. */
    const val COMMANDS_VIEW_DEPS_TITLE = "commands.view.deps.title"

    /** The key composer.plan.writes. */
    const val COMPOSER_PLAN_WRITES = "composer.plan.writes"

    /** The key dashboard.row.health. */
    const val DASHBOARD_ROW_HEALTH = "dashboard.row.health"

    /** The key dashboard.row.projects. */
    const val DASHBOARD_ROW_PROJECTS = "dashboard.row.projects"

    /** The key dialog.button.cancel. */
    const val DIALOG_BUTTON_CANCEL = "dialog.button.cancel"

    /** The key dialog.commit.label. */
    const val DIALOG_COMMIT_LABEL = "dialog.commit.label"

    /** The key fixwizard.style.apply. */
    const val FIXWIZARD_STYLE_APPLY = "fixwizard.style.apply"

    /** The key group.MonoSmith.Cache.text. */
    const val GROUP_MONOSMITH_CACHE_TEXT = "group.MonoSmith.Cache.text"

    /** The key group.MonoSmith.Derive.text. */
    const val GROUP_MONOSMITH_DERIVE_TEXT = "group.MonoSmith.Derive.text"

    /** The key group.MonoSmith.Environments.text. */
    const val GROUP_MONOSMITH_ENVIRONMENTS_TEXT = "group.MonoSmith.Environments.text"

    /** The key group.MonoSmith.Menu.text. */
    const val GROUP_MONOSMITH_MENU_TEXT = "group.MonoSmith.Menu.text"

    /** The key group.MonoSmith.Pipelines.text. */
    const val GROUP_MONOSMITH_PIPELINES_TEXT = "group.MonoSmith.Pipelines.text"

    /** The key group.MonoSmith.Plugins.text. */
    const val GROUP_MONOSMITH_PLUGINS_TEXT = "group.MonoSmith.Plugins.text"

    /** The key group.MonoSmith.Project.text. */
    const val GROUP_MONOSMITH_PROJECT_TEXT = "group.MonoSmith.Project.text"

    /** The key group.MonoSmith.Read.text. */
    const val GROUP_MONOSMITH_READ_TEXT = "group.MonoSmith.Read.text"

    /** The key guide.actions.which_project. */
    const val GUIDE_ACTIONS_WHICH_PROJECT = "guide.actions.which_project"

    /** The key guide.add_project_outcomes.decide_in_the_next_question. */
    const val GUIDE_ADD_PROJECT_OUTCOMES_DECIDE_IN_THE_NEXT_QUESTION = "guide.add_project_outcomes.decide_in_the_next_question"

    /** The key guide.add_project_outcomes.i_know_the_language_show_every. */
    const val GUIDE_ADD_PROJECT_OUTCOMES_I_KNOW_THE_LANGUAGE_SHOW_EVERY = "guide.add_project_outcomes.i_know_the_language_show_every"

    /** The key guide.add_project_picks.layer.framework. */
    const val GUIDE_ADD_PROJECT_PICKS_LAYER_FRAMEWORK = "guide.add_project_picks.layer.framework"

    /** The key guide.add_project_picks.layer.host. */
    const val GUIDE_ADD_PROJECT_PICKS_LAYER_HOST = "guide.add_project_picks.layer.host"

    /** The key guide.add_project_picks.layer.shell. */
    const val GUIDE_ADD_PROJECT_PICKS_LAYER_SHELL = "guide.add_project_picks.layer.shell"

    /** The key guide.add_project_picks.untick. */
    const val GUIDE_ADD_PROJECT_PICKS_UNTICK = "guide.add_project_picks.untick"

    /** The key guide.add_project_setup.customise. */
    const val GUIDE_ADD_PROJECT_SETUP_CUSTOMISE = "guide.add_project_setup.customise"

    /** The key guide.add_project_setup.start_with_the_defaults. */
    const val GUIDE_ADD_PROJECT_SETUP_START_WITH_THE_DEFAULTS = "guide.add_project_setup.start_with_the_defaults"

    /** The key guide.new_workspace_setup.question.agent.name. */
    const val GUIDE_NEW_WORKSPACE_SETUP_QUESTION_AGENT_NAME = "guide.new_workspace_setup.question.agent.name"

    /** The key guide.parity_forms.asks_for_the_app. */
    const val GUIDE_PARITY_FORMS_ASKS_FOR_THE_APP = "guide.parity_forms.asks_for_the_app"

    /** The key guide.parity_forms.claude_code. */
    const val GUIDE_PARITY_FORMS_CLAUDE_CODE = "guide.parity_forms.claude_code"

    /** The key guide.parity_forms.set_the_descriptions. */
    const val GUIDE_PARITY_FORMS_SET_THE_DESCRIPTIONS = "guide.parity_forms.set_the_descriptions"

    /** The key guide.parity_forms.vs_code. */
    const val GUIDE_PARITY_FORMS_VS_CODE = "guide.parity_forms.vs_code"

    /** The key guide.web_environments_steps.in_the_project. */
    const val GUIDE_WEB_ENVIRONMENTS_STEPS_IN_THE_PROJECT = "guide.web_environments_steps.in_the_project"

    /** The key guide.web_environments_steps.in_the_workspace. */
    const val GUIDE_WEB_ENVIRONMENTS_STEPS_IN_THE_WORKSPACE = "guide.web_environments_steps.in_the_workspace"

    /** The key jetbrains.addchoices.add-nothing. */
    const val JETBRAINS_ADDCHOICES_ADD_NOTHING = "jetbrains.addchoices.add-nothing"

    /** The key jetbrains.addchoices.built-from-the-stack. Values: {0} is stack. */
    const val JETBRAINS_ADDCHOICES_BUILT_FROM_THE_STACK = "jetbrains.addchoices.built-from-the-stack"

    /** The key jetbrains.addchoices.built-with-the-default-choices. Values: {0} is language. */
    const val JETBRAINS_ADDCHOICES_BUILT_WITH_THE_DEFAULT_CHOICES = "jetbrains.addchoices.built-with-the-default-choices"

    /** The key jetbrains.addchoices.choose. */
    const val JETBRAINS_ADDCHOICES_CHOOSE = "jetbrains.addchoices.choose"

    /** The key jetbrains.addchoices.default. */
    const val JETBRAINS_ADDCHOICES_DEFAULT = "jetbrains.addchoices.default"

    /** The key jetbrains.addchoices.default-nothing-extra. */
    const val JETBRAINS_ADDCHOICES_DEFAULT_NOTHING_EXTRA = "jetbrains.addchoices.default-nothing-extra"

    /** The key jetbrains.addchoices.go-back-to-the-questions. */
    const val JETBRAINS_ADDCHOICES_GO_BACK_TO_THE_QUESTIONS = "jetbrains.addchoices.go-back-to-the-questions"

    /** The key jetbrains.addchoices.leave-out-the-you-can. Values: {0} is layer. */
    const val JETBRAINS_ADDCHOICES_LEAVE_OUT_THE_YOU_CAN = "jetbrains.addchoices.leave-out-the-you-can"

    /** The key jetbrains.addchoices.monosmith-add-project. Values: {0} is layer. */
    const val JETBRAINS_ADDCHOICES_MONOSMITH_ADD_PROJECT = "jetbrains.addchoices.monosmith-add-project"

    /** The key jetbrains.addchoices.monosmith-add-project-check. */
    const val JETBRAINS_ADDCHOICES_MONOSMITH_ADD_PROJECT_CHECK = "jetbrains.addchoices.monosmith-add-project-check"

    /** The key jetbrains.addchoices.needs-an-answer-on-the. Values: {0} is needing, {1} is layer. */
    const val JETBRAINS_ADDCHOICES_NEEDS_AN_ANSWER_ON_THE = "jetbrains.addchoices.needs-an-answer-on-the"

    /** The key jetbrains.addchoices.none. */
    const val JETBRAINS_ADDCHOICES_NONE = "jetbrains.addchoices.none"

    /** The key jetbrains.addchoices.pick-again. */
    const val JETBRAINS_ADDCHOICES_PICK_AGAIN = "jetbrains.addchoices.pick-again"

    /** The key jetbrains.addchoices.pick-one-the-first-answer. Values: {0} is layer. */
    const val JETBRAINS_ADDCHOICES_PICK_ONE_THE_FIRST_ANSWER = "jetbrains.addchoices.pick-one-the-first-answer"

    /** The key jetbrains.addchoices.suggested. */
    const val JETBRAINS_ADDCHOICES_SUGGESTED = "jetbrains.addchoices.suggested"

    /** The key jetbrains.addchoices.tick-what-the-project-uses. */
    const val JETBRAINS_ADDCHOICES_TICK_WHAT_THE_PROJECT_USES = "jetbrains.addchoices.tick-what-the-project-uses"

    /** The key jetbrains.addchoices.your-choices. Values: {0} is list, {1} is value. */
    const val JETBRAINS_ADDCHOICES_YOUR_CHOICES = "jetbrains.addchoices.your-choices"

    /** The key jetbrains.addproject.a-check-of-smith-doctor. */
    const val JETBRAINS_ADDPROJECT_A_CHECK_OF_SMITH_DOCTOR = "jetbrains.addproject.a-check-of-smith-doctor"

    /** The key jetbrains.addproject.add-project. */
    const val JETBRAINS_ADDPROJECT_ADD_PROJECT = "jetbrains.addproject.add-project"

    /** The key jetbrains.addproject.add-this-library-to-an. */
    const val JETBRAINS_ADDPROJECT_ADD_THIS_LIBRARY_TO_AN = "jetbrains.addproject.add-this-library-to-an"

    /** The key jetbrains.addproject.an-environment-is-listed-twice. */
    const val JETBRAINS_ADDPROJECT_AN_ENVIRONMENT_IS_LISTED_TWICE = "jetbrains.addproject.an-environment-is-listed-twice"

    /** The key jetbrains.addproject.answer-the-questions-one-at. */
    const val JETBRAINS_ADDPROJECT_ANSWER_THE_QUESTIONS_ONE_AT = "jetbrains.addproject.answer-the-questions-one-at"

    /** The key jetbrains.addproject.app-lib-or-package. */
    const val JETBRAINS_ADDPROJECT_APP_LIB_OR_PACKAGE = "jetbrains.addproject.app-lib-or-package"

    /** The key jetbrains.addproject.built-by-smith-run-build. */
    const val JETBRAINS_ADDPROJECT_BUILT_BY_SMITH_RUN_BUILD = "jetbrains.addproject.built-by-smith-run-build"

    /** The key jetbrains.addproject.built-by-smith-run-build-2. Values: {0} is it. */
    const val JETBRAINS_ADDPROJECT_BUILT_BY_SMITH_RUN_BUILD_2 = "jetbrains.addproject.built-by-smith-run-build-2"

    /** The key jetbrains.addproject.cannot-be-an-environment-name. Values: {0} is name. */
    const val JETBRAINS_ADDPROJECT_CANNOT_BE_AN_ENVIRONMENT_NAME = "jetbrains.addproject.cannot-be-an-environment-name"

    /** The key jetbrains.addproject.choose-how-to-get-with. Values: {0} is names. */
    const val JETBRAINS_ADDPROJECT_CHOOSE_HOW_TO_GET_WITH = "jetbrains.addproject.choose-how-to-get-with"

    /** The key jetbrains.addproject.ci-and-releases-build-this. */
    const val JETBRAINS_ADDPROJECT_CI_AND_RELEASES_BUILD_THIS = "jetbrains.addproject.ci-and-releases-build-this"

    /** The key jetbrains.addproject.code-published-to-a-registry. */
    const val JETBRAINS_ADDPROJECT_CODE_PUBLISHED_TO_A_REGISTRY = "jetbrains.addproject.code-published-to-a-registry"

    /** The key jetbrains.addproject.code-the-other-projects-of. */
    const val JETBRAINS_ADDPROJECT_CODE_THE_OTHER_PROJECTS_OF = "jetbrains.addproject.code-the-other-projects-of"

    /** The key jetbrains.addproject.commit-anyway. */
    const val JETBRAINS_ADDPROJECT_COMMIT_ANYWAY = "jetbrains.addproject.commit-anyway"

    /** The key jetbrains.addproject.commit-now. */
    const val JETBRAINS_ADDPROJECT_COMMIT_NOW = "jetbrains.addproject.commit-now"

    /** The key jetbrains.addproject.commit-only-env-example. */
    const val JETBRAINS_ADDPROJECT_COMMIT_ONLY_ENV_EXAMPLE = "jetbrains.addproject.commit-only-env-example"

    /** The key jetbrains.addproject.commit-the-env-files. */
    const val JETBRAINS_ADDPROJECT_COMMIT_THE_ENV_FILES = "jetbrains.addproject.commit-the-env-files"

    /** The key jetbrains.addproject.commit-the-project-although-it. */
    const val JETBRAINS_ADDPROJECT_COMMIT_THE_PROJECT_ALTHOUGH_IT = "jetbrains.addproject.commit-the-project-although-it"

    /** The key jetbrains.addproject.committed-as-one-change-undo. */
    const val JETBRAINS_ADDPROJECT_COMMITTED_AS_ONE_CHANGE_UNDO = "jetbrains.addproject.committed-as-one-change-undo"

    /** The key jetbrains.addproject.continue-without-the-check-the. */
    const val JETBRAINS_ADDPROJECT_CONTINUE_WITHOUT_THE_CHECK_THE = "jetbrains.addproject.continue-without-the-check-the"

    /** The key jetbrains.addproject.copy-the-build-test-layout. */
    const val JETBRAINS_ADDPROJECT_COPY_THE_BUILD_TEST_LAYOUT = "jetbrains.addproject.copy-the-build-test-layout"

    /** The key jetbrains.addproject.dist-apps-app-name-at. */
    const val JETBRAINS_ADDPROJECT_DIST_APPS_APP_NAME_AT = "jetbrains.addproject.dist-apps-app-name-at"

    /** The key jetbrains.addproject.dist-name-inside-the-app. */
    const val JETBRAINS_ADDPROJECT_DIST_NAME_INSIDE_THE_APP = "jetbrains.addproject.dist-name-inside-the-app"

    /** The key jetbrains.addproject.each-machine-and-ci-job. */
    const val JETBRAINS_ADDPROJECT_EACH_MACHINE_AND_CI_JOB = "jetbrains.addproject.each-machine-and-ci-job"

    /** The key jetbrains.addproject.environments-of. Values: {0} is title, {1} is app. */
    const val JETBRAINS_ADDPROJECT_ENVIRONMENTS_OF = "jetbrains.addproject.environments-of"

    /** The key jetbrains.addproject.exists-already-pick-another-name. Values: {0} is name. */
    const val JETBRAINS_ADDPROJECT_EXISTS_ALREADY_PICK_ANOTHER_NAME = "jetbrains.addproject.exists-already-pick-another-name"

    /** The key jetbrains.addproject.follow-which-project. Values: {0} is title. */
    const val JETBRAINS_ADDPROJECT_FOLLOW_WHICH_PROJECT = "jetbrains.addproject.follow-which-project"

    /** The key jetbrains.addproject.how-do-you-want-to. Values: {0} is title. */
    const val JETBRAINS_ADDPROJECT_HOW_DO_YOU_WANT_TO = "jetbrains.addproject.how-do-you-want-to"

    /** The key jetbrains.addproject.install-page. */
    const val JETBRAINS_ADDPROJECT_INSTALL_PAGE = "jetbrains.addproject.install-page"

    /** The key jetbrains.addproject.install-the-missing-tools-with. */
    const val JETBRAINS_ADDPROJECT_INSTALL_THE_MISSING_TOOLS_WITH = "jetbrains.addproject.install-the-missing-tools-with"

    /** The key jetbrains.addproject.install-tools. */
    const val JETBRAINS_ADDPROJECT_INSTALL_TOOLS = "jetbrains.addproject.install-tools"

    /** The key jetbrains.addproject.is-added-and-the-ci. Values: {0} is name, {1} is value. */
    const val JETBRAINS_ADDPROJECT_IS_ADDED_AND_THE_CI = "jetbrains.addproject.is-added-and-the-ci"

    /** The key jetbrains.addproject.is-not-an-environment-name. Values: {0} is name. */
    const val JETBRAINS_ADDPROJECT_IS_NOT_AN_ENVIRONMENT_NAME = "jetbrains.addproject.is-not-an-environment-name"

    /** The key jetbrains.addproject.it-was-committed. */
    const val JETBRAINS_ADDPROJECT_IT_WAS_COMMITTED = "jetbrains.addproject.it-was-committed"

    /** The key jetbrains.addproject.java-jdk. */
    const val JETBRAINS_ADDPROJECT_JAVA_JDK = "jetbrains.addproject.java-jdk"

    /** The key jetbrains.addproject.keep-uncommitted. */
    const val JETBRAINS_ADDPROJECT_KEEP_UNCOMMITTED = "jetbrains.addproject.keep-uncommitted"

    /** The key jetbrains.addproject.leave-the-files-in-your. */
    const val JETBRAINS_ADDPROJECT_LEAVE_THE_FILES_IN_YOUR = "jetbrains.addproject.leave-the-files-in-your"

    /** The key jetbrains.addproject.make-one-commit-that-undo. */
    const val JETBRAINS_ADDPROJECT_MAKE_ONE_COMMIT_THAT_UNDO = "jetbrains.addproject.make-one-commit-that-undo"

    /** The key jetbrains.addproject.monosmith-add-project. */
    const val JETBRAINS_ADDPROJECT_MONOSMITH_ADD_PROJECT = "jetbrains.addproject.monosmith-add-project"

    /** The key jetbrains.addproject.name-at-least-one-environment. */
    const val JETBRAINS_ADDPROJECT_NAME_AT_LEAST_ONE_ENVIRONMENT = "jetbrains.addproject.name-at-least-one-environment"

    /** The key jetbrains.addproject.needs-tools-this-machine-lacks. Values: {0} is name, {1} is problems. */
    const val JETBRAINS_ADDPROJECT_NEEDS_TOOLS_THIS_MACHINE_LACKS = "jetbrains.addproject.needs-tools-this-machine-lacks"

    /** The key jetbrains.addproject.not-committed-review-it-with. */
    const val JETBRAINS_ADDPROJECT_NOT_COMMITTED_REVIEW_IT_WITH = "jetbrains.addproject.not-committed-review-it-with"

    /** The key jetbrains.addproject.nothing-was-added. Values: {0} is failureTextMapper. */
    const val JETBRAINS_ADDPROJECT_NOTHING_WAS_ADDED = "jetbrains.addproject.nothing-was-added"

    /** The key jetbrains.addproject.nothing-was-added-smith-printed. */
    const val JETBRAINS_ADDPROJECT_NOTHING_WAS_ADDED_SMITH_PRINTED = "jetbrains.addproject.nothing-was-added-smith-printed"

    /** The key jetbrains.addproject.nothing-was-changed-yet-by. Values: {0} is words. */
    const val JETBRAINS_ADDPROJECT_NOTHING_WAS_CHANGED_YET_BY = "jetbrains.addproject.nothing-was-changed-yet-by"

    /** The key jetbrains.addproject.nothing-was-committed-the-files. */
    const val JETBRAINS_ADDPROJECT_NOTHING_WAS_COMMITTED_THE_FILES = "jetbrains.addproject.nothing-was-committed-the-files"

    /** The key jetbrains.addproject.open-health. */
    const val JETBRAINS_ADDPROJECT_OPEN_HEALTH = "jetbrains.addproject.open-health"

    /** The key jetbrains.addproject.open-the-health-tab-of. */
    const val JETBRAINS_ADDPROJECT_OPEN_THE_HEALTH_TAB_OF = "jetbrains.addproject.open-the-health-tab-of"

    /** The key jetbrains.addproject.open-the-page-that-tells. Values: {0} is title. */
    const val JETBRAINS_ADDPROJECT_OPEN_THE_PAGE_THAT_TELLS = "jetbrains.addproject.open-the-page-that-tells"

    /** The key jetbrains.addproject.pick-the-language-what-each. */
    const val JETBRAINS_ADDPROJECT_PICK_THE_LANGUAGE_WHAT_EACH = "jetbrains.addproject.pick-the-language-what-each"

    /** The key jetbrains.addproject.pick-what-you-want-type. */
    const val JETBRAINS_ADDPROJECT_PICK_WHAT_YOU_WANT_TYPE = "jetbrains.addproject.pick-what-you-want-type"

    /** The key jetbrains.addproject.revert-the-commit-of-this. */
    const val JETBRAINS_ADDPROJECT_REVERT_THE_COMMIT_OF_THIS = "jetbrains.addproject.revert-the-commit-of-this"

    /** The key jetbrains.addproject.run-the-first-check-of. */
    const val JETBRAINS_ADDPROJECT_RUN_THE_FIRST_CHECK_OF = "jetbrains.addproject.run-the-first-check-of"

    /** The key jetbrains.addproject.search-frameworks-desktop-shells-and. */
    const val JETBRAINS_ADDPROJECT_SEARCH_FRAMEWORKS_DESKTOP_SHELLS_AND = "jetbrains.addproject.search-frameworks-desktop-shells-and"

    /** The key jetbrains.addproject.skip-the-check. */
    const val JETBRAINS_ADDPROJECT_SKIP_THE_CHECK = "jetbrains.addproject.skip-the-check"

    /** The key jetbrains.addproject.skip-the-shape-of-the. */
    const val JETBRAINS_ADDPROJECT_SKIP_THE_SHAPE_OF_THE = "jetbrains.addproject.skip-the-shape-of-the"

    /** The key jetbrains.addproject.smith-follows-the-siblings-that. */
    const val JETBRAINS_ADDPROJECT_SMITH_FOLLOWS_THE_SIBLINGS_THAT = "jetbrains.addproject.smith-follows-the-siblings-that"

    /** The key jetbrains.addproject.something-you-run-or-deploy. */
    const val JETBRAINS_ADDPROJECT_SOMETHING_YOU_RUN_OR_DEPLOY = "jetbrains.addproject.something-you-run-or-deploy"

    /** The key jetbrains.addproject.the-defaults-are-the-answers. */
    const val JETBRAINS_ADDPROJECT_THE_DEFAULTS_ARE_THE_ANSWERS = "jetbrains.addproject.the-defaults-are-the-answers"

    /** The key jetbrains.addproject.the-env-files. Values: {0} is title. */
    const val JETBRAINS_ADDPROJECT_THE_ENV_FILES = "jetbrains.addproject.the-env-files"

    /** The key jetbrains.addproject.the-first-answer-is-the. */
    const val JETBRAINS_ADDPROJECT_THE_FIRST_ANSWER_IS_THE = "jetbrains.addproject.the-first-answer-is-the"

    /** The key jetbrains.addproject.the-long-way-the-kind. */
    const val JETBRAINS_ADDPROJECT_THE_LONG_WAY_THE_KIND = "jetbrains.addproject.the-long-way-the-kind"

    /** The key jetbrains.addproject.the-name-of-the-new. Values: {0} is role. */
    const val JETBRAINS_ADDPROJECT_THE_NAME_OF_THE_NEW = "jetbrains.addproject.the-name-of-the-new"

    /** The key jetbrains.addproject.the-net-sdk. */
    const val JETBRAINS_ADDPROJECT_THE_NET_SDK = "jetbrains.addproject.the-net-sdk"

    /** The key jetbrains.addproject.the-new-project-copies-the. */
    const val JETBRAINS_ADDPROJECT_THE_NEW_PROJECT_COPIES_THE = "jetbrains.addproject.the-new-project-copies-the"

    /** The key jetbrains.addproject.the-project-was-created-but. Values: {0} is name, {1} is files, {2} is why, {3} is commit. */
    const val JETBRAINS_ADDPROJECT_THE_PROJECT_WAS_CREATED_BUT = "jetbrains.addproject.the-project-was-created-but"

    /** The key jetbrains.addproject.they-hold-vite-values-which. */
    const val JETBRAINS_ADDPROJECT_THEY_HOLD_VITE_VALUES_WHICH = "jetbrains.addproject.they-hold-vite-values-which"

    /** The key jetbrains.addproject.undo. */
    const val JETBRAINS_ADDPROJECT_UNDO = "jetbrains.addproject.undo"

    /** The key jetbrains.addproject.use-in-an-app. */
    const val JETBRAINS_ADDPROJECT_USE_IN_AN_APP = "jetbrains.addproject.use-in-an-app"

    /** The key jetbrains.addproject.use-lowercase-letters-digits-and. */
    const val JETBRAINS_ADDPROJECT_USE_LOWERCASE_LETTERS_DIGITS_AND = "jetbrains.addproject.use-lowercase-letters-digits-and"

    /** The key jetbrains.addproject.vite-writes-its-vite-values. */
    const val JETBRAINS_ADDPROJECT_VITE_WRITES_ITS_VITE_VALUES = "jetbrains.addproject.vite-writes-its-vite-values"

    /** The key jetbrains.addproject.was-created-and-could-not. Values: {0} is name, {1} is where, {2} is files, {3} is title. */
    const val JETBRAINS_ADDPROJECT_WAS_CREATED_AND_COULD_NOT = "jetbrains.addproject.was-created-and-could-not"

    /** The key jetbrains.addproject.what-are-you-building. Values: {0} is title. */
    const val JETBRAINS_ADDPROJECT_WHAT_ARE_YOU_BUILDING = "jetbrains.addproject.what-are-you-building"

    /** The key jetbrains.addproject.what-kind. Values: {0} is title. */
    const val JETBRAINS_ADDPROJECT_WHAT_KIND = "jetbrains.addproject.what-kind"

    /** The key jetbrains.addproject.what-most-projects-do-default. */
    const val JETBRAINS_ADDPROJECT_WHAT_MOST_PROJECTS_DO_DEFAULT = "jetbrains.addproject.what-most-projects-do-default"

    /** The key jetbrains.addproject.what-the-stack-suggests-your. */
    const val JETBRAINS_ADDPROJECT_WHAT_THE_STACK_SUGGESTS_YOUR = "jetbrains.addproject.what-the-stack-suggests-your"

    /** The key jetbrains.addproject.where-do-the-builds-go. Values: {0} is title. */
    const val JETBRAINS_ADDPROJECT_WHERE_DO_THE_BUILDS_GO = "jetbrains.addproject.where-do-the-builds-go"

    /** The key jetbrains.addproject.which-environment-is-build. Values: {0} is title. */
    const val JETBRAINS_ADDPROJECT_WHICH_ENVIRONMENT_IS_BUILD = "jetbrains.addproject.which-environment-is-build"

    /** The key jetbrains.addproject.which-language. Values: {0} is title. */
    const val JETBRAINS_ADDPROJECT_WHICH_LANGUAGE = "jetbrains.addproject.which-language"

    /** The key jetbrains.addproject.you-can-still-pick-a. */
    const val JETBRAINS_ADDPROJECT_YOU_CAN_STILL_PICK_A = "jetbrains.addproject.you-can-still-pick-a"

    /** The key jetbrains.addproject.your-pick-is-kept-the. */
    const val JETBRAINS_ADDPROJECT_YOUR_PICK_IS_KEPT_THE = "jetbrains.addproject.your-pick-is-kept-the"

    /** The key jetbrains.adoptfolder.a-folder-that-starts-with. */
    const val JETBRAINS_ADOPTFOLDER_A_FOLDER_THAT_STARTS_WITH = "jetbrains.adoptfolder.a-folder-that-starts-with"

    /** The key jetbrains.adoptfolder.add-a-gitattributes-with-text. */
    const val JETBRAINS_ADOPTFOLDER_ADD_A_GITATTRIBUTES_WITH_TEXT = "jetbrains.adoptfolder.add-a-gitattributes-with-text"

    /** The key jetbrains.adoptfolder.add-it. */
    const val JETBRAINS_ADOPTFOLDER_ADD_IT = "jetbrains.adoptfolder.add-it"

    /** The key jetbrains.adoptfolder.add-the-monosmith-agent-files. */
    const val JETBRAINS_ADOPTFOLDER_ADD_THE_MONOSMITH_AGENT_FILES = "jetbrains.adoptfolder.add-the-monosmith-agent-files"

    /** The key jetbrains.adoptfolder.add-them. */
    const val JETBRAINS_ADOPTFOLDER_ADD_THEM = "jetbrains.adoptfolder.add-them"

    /** The key jetbrains.adoptfolder.another-folder. */
    const val JETBRAINS_ADOPTFOLDER_ANOTHER_FOLDER = "jetbrains.adoptfolder.another-folder"

    /** The key jetbrains.adoptfolder.bring-it-in. */
    const val JETBRAINS_ADOPTFOLDER_BRING_IT_IN = "jetbrains.adoptfolder.bring-it-in"

    /** The key jetbrains.adoptfolder.is-at-the-root-smith. Values: {0} is manifest. */
    const val JETBRAINS_ADOPTFOLDER_IS_AT_THE_ROOT_SMITH = "jetbrains.adoptfolder.is-at-the-root-smith"

    /** The key jetbrains.adoptfolder.is-not-a-folder-smith. Values: {0} is folder, {1} is why. */
    const val JETBRAINS_ADOPTFOLDER_IS_NOT_A_FOLDER_SMITH = "jetbrains.adoptfolder.is-not-a-folder-smith"

    /** The key jetbrains.adoptfolder.it-is-the-repository-root. */
    const val JETBRAINS_ADOPTFOLDER_IT_IS_THE_REPOSITORY_ROOT = "jetbrains.adoptfolder.it-is-the-repository-root"

    /** The key jetbrains.adoptfolder.it-leaves-the-repository. */
    const val JETBRAINS_ADOPTFOLDER_IT_LEAVES_THE_REPOSITORY = "jetbrains.adoptfolder.it-leaves-the-repository"

    /** The key jetbrains.adoptfolder.it-must-be-relative-to. */
    const val JETBRAINS_ADOPTFOLDER_IT_MUST_BE_RELATIVE_TO = "jetbrains.adoptfolder.it-must-be-relative-to"

    /** The key jetbrains.adoptfolder.keep-mine. */
    const val JETBRAINS_ADOPTFOLDER_KEEP_MINE = "jetbrains.adoptfolder.keep-mine"

    /** The key jetbrains.adoptfolder.leave-line-endings-alone. */
    const val JETBRAINS_ADOPTFOLDER_LEAVE_LINE_ENDINGS_ALONE = "jetbrains.adoptfolder.leave-line-endings-alone"

    /** The key jetbrains.adoptfolder.line-endings. */
    const val JETBRAINS_ADOPTFOLDER_LINE_ENDINGS = "jetbrains.adoptfolder.line-endings"

    /** The key jetbrains.adoptfolder.monosmith-bring-this-folder-in. */
    const val JETBRAINS_ADOPTFOLDER_MONOSMITH_BRING_THIS_FOLDER_IN = "jetbrains.adoptfolder.monosmith-bring-this-folder-in"

    /** The key jetbrains.adoptfolder.must-be-lowercase-letters-and. Values: {0} is name. */
    const val JETBRAINS_ADOPTFOLDER_MUST_BE_LOWERCASE_LETTERS_AND = "jetbrains.adoptfolder.must-be-lowercase-letters-and"

    /** The key jetbrains.adoptfolder.never. */
    const val JETBRAINS_ADOPTFOLDER_NEVER = "jetbrains.adoptfolder.never"

    /** The key jetbrains.adoptfolder.next. */
    const val JETBRAINS_ADOPTFOLDER_NEXT = "jetbrains.adoptfolder.next"

    /** The key jetbrains.adoptfolder.no-leave-it. */
    const val JETBRAINS_ADOPTFOLDER_NO_LEAVE_IT = "jetbrains.adoptfolder.no-leave-it"

    /** The key jetbrains.adoptfolder.nothing-was-adopted. Values: {0} is failureTextMapper. */
    const val JETBRAINS_ADOPTFOLDER_NOTHING_WAS_ADOPTED = "jetbrains.adoptfolder.nothing-was-adopted"

    /** The key jetbrains.adoptfolder.nothing-was-adopted-2. Values: {0} is blockers. */
    const val JETBRAINS_ADOPTFOLDER_NOTHING_WAS_ADOPTED_2 = "jetbrains.adoptfolder.nothing-was-adopted-2"

    /** The key jetbrains.adoptfolder.nothing-was-changed-yet-this. */
    const val JETBRAINS_ADOPTFOLDER_NOTHING_WAS_CHANGED_YET_THIS = "jetbrains.adoptfolder.nothing-was-changed-yet-this"

    /** The key jetbrains.adoptfolder.records-agent-enabled-false-in. */
    const val JETBRAINS_ADOPTFOLDER_RECORDS_AGENT_ENABLED_FALSE_IN = "jetbrains.adoptfolder.records-agent-enabled-false-in"

    /** The key jetbrains.adoptfolder.replace-it. */
    const val JETBRAINS_ADOPTFOLDER_REPLACE_IT = "jetbrains.adoptfolder.replace-it"

    /** The key jetbrains.adoptfolder.smith-writes-its-own-yours. */
    const val JETBRAINS_ADOPTFOLDER_SMITH_WRITES_ITS_OWN_YOURS = "jetbrains.adoptfolder.smith-writes-its-own-yours"

    /** The key jetbrains.adoptfolder.smith-writes-no-pipeline-for. */
    const val JETBRAINS_ADOPTFOLDER_SMITH_WRITES_NO_PIPELINE_FOR = "jetbrains.adoptfolder.smith-writes-no-pipeline-for"

    /** The key jetbrains.adoptfolder.the-first-answer-is-the. Values: {0} is header. */
    const val JETBRAINS_ADOPTFOLDER_THE_FIRST_ANSWER_IS_THE = "jetbrains.adoptfolder.the-first-answer-is-the"

    /** The key jetbrains.adoptfolder.the-folder-from-the-repository. Values: {0} is name. */
    const val JETBRAINS_ADOPTFOLDER_THE_FOLDER_FROM_THE_REPOSITORY = "jetbrains.adoptfolder.the-folder-from-the-repository"

    /** The key jetbrains.adoptfolder.the-folder-is-a-monosmith. */
    const val JETBRAINS_ADOPTFOLDER_THE_FOLDER_IS_A_MONOSMITH = "jetbrains.adoptfolder.the-folder-is-a-monosmith"

    /** The key jetbrains.adoptfolder.the-name-in. Values: {0} is choice. */
    const val JETBRAINS_ADOPTFOLDER_THE_NAME_IN = "jetbrains.adoptfolder.the-name-in"

    /** The key jetbrains.adoptfolder.the-project-stays-at-the. */
    const val JETBRAINS_ADOPTFOLDER_THE_PROJECT_STAYS_AT_THE = "jetbrains.adoptfolder.the-project-stays-at-the"

    /** The key jetbrains.adoptfolder.tick-none-to-write-no. Values: {0} is prompt. */
    const val JETBRAINS_ADOPTFOLDER_TICK_NONE_TO_WRITE_NO = "jetbrains.adoptfolder.tick-none-to-write-no"

    /** The key jetbrains.adoptfolder.use-letters-digits-dot-hyphen. */
    const val JETBRAINS_ADOPTFOLDER_USE_LETTERS_DIGITS_DOT_HYPHEN = "jetbrains.adoptfolder.use-letters-digits-dot-hyphen"

    /** The key jetbrains.adoptfolder.where-does-go. Values: {0} is title, {1} is name. */
    const val JETBRAINS_ADOPTFOLDER_WHERE_DOES_GO = "jetbrains.adoptfolder.where-does-go"

    /** The key jetbrains.adoptfolder.writes-claude-agents-smith-md. */
    const val JETBRAINS_ADOPTFOLDER_WRITES_CLAUDE_AGENTS_SMITH_MD = "jetbrains.adoptfolder.writes-claude-agents-smith-md"

    /** The key jetbrains.adoptfolder.you-can-add-them-later. */
    const val JETBRAINS_ADOPTFOLDER_YOU_CAN_ADD_THEM_LATER = "jetbrains.adoptfolder.you-can-add-them-later"

    /** The key jetbrains.adoptfolder.your-pipelines. */
    const val JETBRAINS_ADOPTFOLDER_YOUR_PIPELINES = "jetbrains.adoptfolder.your-pipelines"

    /** The key jetbrains.agentfile.agents-md-is-written. */
    const val JETBRAINS_AGENTFILE_AGENTS_MD_IS_WRITTEN = "jetbrains.agentfile.agents-md-is-written"

    /** The key jetbrains.agentfile.monosmith-write-the-agent-file. */
    const val JETBRAINS_AGENTFILE_MONOSMITH_WRITE_THE_AGENT_FILE = "jetbrains.agentfile.monosmith-write-the-agent-file"

    /** The key jetbrains.agentfile.write-the-files. */
    const val JETBRAINS_AGENTFILE_WRITE_THE_FILES = "jetbrains.agentfile.write-the-files"

    /** The key jetbrains.agentfile.your-own-text-outside-smith. */
    const val JETBRAINS_AGENTFILE_YOUR_OWN_TEXT_OUTSIDE_SMITH = "jetbrains.agentfile.your-own-text-outside-smith"

    /** The key jetbrains.cachecontrol.delete-the-cache. */
    const val JETBRAINS_CACHECONTROL_DELETE_THE_CACHE = "jetbrains.cachecontrol.delete-the-cache"

    /** The key jetbrains.cachecontrol.it-deletes-corrupt-blobs-the. */
    const val JETBRAINS_CACHECONTROL_IT_DELETES_CORRUPT_BLOBS_THE = "jetbrains.cachecontrol.it-deletes-corrupt-blobs-the"

    /** The key jetbrains.cachecontrol.it-deletes-the-local-cache. */
    const val JETBRAINS_CACHECONTROL_IT_DELETES_THE_LOCAL_CACHE = "jetbrains.cachecontrol.it-deletes-the-local-cache"

    /** The key jetbrains.cachecontrol.it-removes-the-least-recently. */
    const val JETBRAINS_CACHECONTROL_IT_REMOVES_THE_LEAST_RECENTLY = "jetbrains.cachecontrol.it-removes-the-least-recently"

    /** The key jetbrains.cachecontrol.monosmith-cache. */
    const val JETBRAINS_CACHECONTROL_MONOSMITH_CACHE = "jetbrains.cachecontrol.monosmith-cache"

    /** The key jetbrains.cachecontrol.nothing-was-changed-yet. Values: {0} is consequence. */
    const val JETBRAINS_CACHECONTROL_NOTHING_WAS_CHANGED_YET = "jetbrains.cachecontrol.nothing-was-changed-yet"

    /** The key jetbrains.cachecontrol.remove-what-is-broken. */
    const val JETBRAINS_CACHECONTROL_REMOVE_WHAT_IS_BROKEN = "jetbrains.cachecontrol.remove-what-is-broken"

    /** The key jetbrains.cachecontrol.the-cache-is-trimmed. */
    const val JETBRAINS_CACHECONTROL_THE_CACHE_IS_TRIMMED = "jetbrains.cachecontrol.the-cache-is-trimmed"

    /** The key jetbrains.cachecontrol.the-local-cache-is-deleted. */
    const val JETBRAINS_CACHECONTROL_THE_LOCAL_CACHE_IS_DELETED = "jetbrains.cachecontrol.the-local-cache-is-deleted"

    /** The key jetbrains.cachecontrol.trim-the-cache. */
    const val JETBRAINS_CACHECONTROL_TRIM_THE_CACHE = "jetbrains.cachecontrol.trim-the-cache"

    /** The key jetbrains.cachecontrol.what-was-broken-is-removed. */
    const val JETBRAINS_CACHECONTROL_WHAT_WAS_BROKEN_IS_REMOVED = "jetbrains.cachecontrol.what-was-broken-is-removed"

    /** The key jetbrains.ciregenerate.monosmith-regenerate-the-ci-pipelines. */
    const val JETBRAINS_CIREGENERATE_MONOSMITH_REGENERATE_THE_CI_PIPELINES = "jetbrains.ciregenerate.monosmith-regenerate-the-ci-pipelines"

    /** The key jetbrains.ciregenerate.regenerate. */
    const val JETBRAINS_CIREGENERATE_REGENERATE = "jetbrains.ciregenerate.regenerate"

    /** The key jetbrains.ciregenerate.the-ci-pipelines-are-regenerated. */
    const val JETBRAINS_CIREGENERATE_THE_CI_PIPELINES_ARE_REGENERATED = "jetbrains.ciregenerate.the-ci-pipelines-are-regenerated"

    /** The key jetbrains.ciregenerate.the-plan-is-a-diff. */
    const val JETBRAINS_CIREGENERATE_THE_PLAN_IS_A_DIFF = "jetbrains.ciregenerate.the-plan-is-a-diff"

    /** The key jetbrains.commitchanges.commit-everything-git-shows. */
    const val JETBRAINS_COMMITCHANGES_COMMIT_EVERYTHING_GIT_SHOWS = "jetbrains.commitchanges.commit-everything-git-shows"

    /** The key jetbrains.commitchanges.commit-only-the-files-of. */
    const val JETBRAINS_COMMITCHANGES_COMMIT_ONLY_THE_FILES_OF = "jetbrains.commitchanges.commit-only-the-files-of"

    /** The key jetbrains.commitchanges.git-also-shows-other-changed. Values: {0} is more. */
    const val JETBRAINS_COMMITCHANGES_GIT_ALSO_SHOWS_OTHER_CHANGED = "jetbrains.commitchanges.git-also-shows-other-changed"

    /** The key jetbrains.commitchanges.monosmith-commit-the-changes. */
    const val JETBRAINS_COMMITCHANGES_MONOSMITH_COMMIT_THE_CHANGES = "jetbrains.commitchanges.monosmith-commit-the-changes"

    /** The key jetbrains.commitchanges.nothing-was-committed. Values: {0} is failureTextMapper. */
    const val JETBRAINS_COMMITCHANGES_NOTHING_WAS_COMMITTED = "jetbrains.commitchanges.nothing-was-committed"

    /** The key jetbrains.commitchanges.smith-makes-one-commit-that. Values: {0} is extras. */
    const val JETBRAINS_COMMITCHANGES_SMITH_MAKES_ONE_COMMIT_THAT = "jetbrains.commitchanges.smith-makes-one-commit-that"

    /** The key jetbrains.commitchanges.the-changes-are-committed-as. */
    const val JETBRAINS_COMMITCHANGES_THE_CHANGES_ARE_COMMITTED_AS = "jetbrains.commitchanges.the-changes-are-committed-as"

    /** The key jetbrains.composestack.2-of-3-name. Values: {0} is title. */
    const val JETBRAINS_COMPOSESTACK_2_OF_3_NAME = "jetbrains.composestack.2-of-3-name"

    /** The key jetbrains.composestack.3-of-3-plan. Values: {0} is title. */
    const val JETBRAINS_COMPOSESTACK_3_OF_3_PLAN = "jetbrains.composestack.3-of-3-plan"

    /** The key jetbrains.composestack.add-the-project. */
    const val JETBRAINS_COMPOSESTACK_ADD_THE_PROJECT = "jetbrains.composestack.add-the-project"

    /** The key jetbrains.composestack.cannot-be-used. Values: {0} is problem. */
    const val JETBRAINS_COMPOSESTACK_CANNOT_BE_USED = "jetbrains.composestack.cannot-be-used"

    /** The key jetbrains.composestack.conflict. Values: {0} is conflict. */
    const val JETBRAINS_COMPOSESTACK_CONFLICT = "jetbrains.composestack.conflict"

    /** The key jetbrains.composestack.install-the-tools-first. */
    const val JETBRAINS_COMPOSESTACK_INSTALL_THE_TOOLS_FIRST = "jetbrains.composestack.install-the-tools-first"

    /** The key jetbrains.composestack.is-added. Values: {0} is name, {1} is value. */
    const val JETBRAINS_COMPOSESTACK_IS_ADDED = "jetbrains.composestack.is-added"

    /** The key jetbrains.composestack.is-not-part-of-one. Values: {0} is title, {1} is route. */
    const val JETBRAINS_COMPOSESTACK_IS_NOT_PART_OF_ONE = "jetbrains.composestack.is-not-part-of-one"

    /** The key jetbrains.composestack.it-needs-these-tools-on. */
    const val JETBRAINS_COMPOSESTACK_IT_NEEDS_THESE_TOOLS_ON = "jetbrains.composestack.it-needs-these-tools-on"

    /** The key jetbrains.composestack.keep-it-as-a-recipe. */
    const val JETBRAINS_COMPOSESTACK_KEEP_IT_AS_A_RECIPE = "jetbrains.composestack.keep-it-as-a-recipe"

    /** The key jetbrains.composestack.monosmith-compose-a-stack. */
    const val JETBRAINS_COMPOSESTACK_MONOSMITH_COMPOSE_A_STACK = "jetbrains.composestack.monosmith-compose-a-stack"

    /** The key jetbrains.composestack.needs-an-answer-on-the. Values: {0} is choice, {1} is axis, {2} is about. */
    const val JETBRAINS_COMPOSESTACK_NEEDS_AN_ANSWER_ON_THE = "jetbrains.composestack.needs-an-answer-on-the"

    /** The key jetbrains.composestack.next-check. */
    const val JETBRAINS_COMPOSESTACK_NEXT_CHECK = "jetbrains.composestack.next-check"

    /** The key jetbrains.composestack.not-checked. */
    const val JETBRAINS_COMPOSESTACK_NOT_CHECKED = "jetbrains.composestack.not-checked"

    /** The key jetbrains.composestack.not-part-of-one-project. Values: {0} is route. */
    const val JETBRAINS_COMPOSESTACK_NOT_PART_OF_ONE_PROJECT = "jetbrains.composestack.not-part-of-one-project"

    /** The key jetbrains.composestack.nothing-is-selected-yet. */
    const val JETBRAINS_COMPOSESTACK_NOTHING_IS_SELECTED_YET = "jetbrains.composestack.nothing-is-selected-yet"

    /** The key jetbrains.composestack.nothing-is-written-yet-by. */
    const val JETBRAINS_COMPOSESTACK_NOTHING_IS_WRITTEN_YET_BY = "jetbrains.composestack.nothing-is-written-yet-by"

    /** The key jetbrains.composestack.nothing-was-added-the-command. Values: {0} is problem, {1} is command. */
    const val JETBRAINS_COMPOSESTACK_NOTHING_WAS_ADDED_THE_COMMAND = "jetbrains.composestack.nothing-was-added-the-command"

    /** The key jetbrains.composestack.nothing-was-written. Values: {0} is failureTextMapper. */
    const val JETBRAINS_COMPOSESTACK_NOTHING_WAS_WRITTEN = "jetbrains.composestack.nothing-was-written"

    /** The key jetbrains.composestack.run. Values: {0} is command. */
    const val JETBRAINS_COMPOSESTACK_RUN = "jetbrains.composestack.run"

    /** The key jetbrains.composestack.selected. Values: {0} is it. */
    const val JETBRAINS_COMPOSESTACK_SELECTED = "jetbrains.composestack.selected"

    /** The key jetbrains.composestack.smith-did-not-check-the. */
    const val JETBRAINS_COMPOSESTACK_SMITH_DID_NOT_CHECK_THE = "jetbrains.composestack.smith-did-not-check-the"

    /** The key jetbrains.composestack.smith-made-no-plan-for. */
    const val JETBRAINS_COMPOSESTACK_SMITH_MADE_NO_PLAN_FOR = "jetbrains.composestack.smith-made-no-plan-for"

    /** The key jetbrains.composestack.smith-printed-its-catalogue-in. */
    const val JETBRAINS_COMPOSESTACK_SMITH_PRINTED_ITS_CATALOGUE_IN = "jetbrains.composestack.smith-printed-its-catalogue-in"

    /** The key jetbrains.composestack.still-needs-a-choice-on. Values: {0} is choice, {1} is axis. */
    const val JETBRAINS_COMPOSESTACK_STILL_NEEDS_A_CHOICE_ON = "jetbrains.composestack.still-needs-a-choice-on"

    /** The key jetbrains.composestack.the-choice. Values: {0} is axis. */
    const val JETBRAINS_COMPOSESTACK_THE_CHOICE = "jetbrains.composestack.the-choice"

    /** The key jetbrains.composestack.the-name-of-the-new. Values: {0} is stackName. */
    const val JETBRAINS_COMPOSESTACK_THE_NAME_OF_THE_NEW = "jetbrains.composestack.the-name-of-the-new"

    /** The key jetbrains.composestack.the-recipe-is-written. */
    const val JETBRAINS_COMPOSESTACK_THE_RECIPE_IS_WRITTEN = "jetbrains.composestack.the-recipe-is-written"

    /** The key jetbrains.composestack.tick-at-least-one-choice. */
    const val JETBRAINS_COMPOSESTACK_TICK_AT_LEAST_ONE_CHOICE = "jetbrains.composestack.tick-at-least-one-choice"

    /** The key jetbrains.convertfolder.monosmith-convert-this-workspace. */
    const val JETBRAINS_CONVERTFOLDER_MONOSMITH_CONVERT_THIS_WORKSPACE = "jetbrains.convertfolder.monosmith-convert-this-workspace"

    /** The key jetbrains.convertfolder.nothing-was-changed-update-smith. Values: {0} is failureTextMapper. */
    const val JETBRAINS_CONVERTFOLDER_NOTHING_WAS_CHANGED_UPDATE_SMITH = "jetbrains.convertfolder.nothing-was-changed-update-smith"

    /** The key jetbrains.convertfolder.nothing-was-changed-yet-the. */
    const val JETBRAINS_CONVERTFOLDER_NOTHING_WAS_CHANGED_YET_THE = "jetbrains.convertfolder.nothing-was-changed-yet-the"

    /** The key jetbrains.convertfolder.smith-printed-the-questions-of. */
    const val JETBRAINS_CONVERTFOLDER_SMITH_PRINTED_THE_QUESTIONS_OF = "jetbrains.convertfolder.smith-printed-the-questions-of"

    /** The key jetbrains.convertfolder.the-first-answer-changes-the. */
    const val JETBRAINS_CONVERTFOLDER_THE_FIRST_ANSWER_CHANGES_THE = "jetbrains.convertfolder.the-first-answer-changes-the"

    /** The key jetbrains.convertfolder.the-workspace-is-converted. */
    const val JETBRAINS_CONVERTFOLDER_THE_WORKSPACE_IS_CONVERTED = "jetbrains.convertfolder.the-workspace-is-converted"

    /** The key jetbrains.dependencies.align. */
    const val JETBRAINS_DEPENDENCIES_ALIGN = "jetbrains.dependencies.align"

    /** The key jetbrains.dependencies.choose-the-packages. */
    const val JETBRAINS_DEPENDENCIES_CHOOSE_THE_PACKAGES = "jetbrains.dependencies.choose-the-packages"

    /** The key jetbrains.dependencies.dependencies-are-updated. Values: {0} is size. */
    const val JETBRAINS_DEPENDENCIES_DEPENDENCIES_ARE_UPDATED = "jetbrains.dependencies.dependencies-are-updated"

    /** The key jetbrains.dependencies.every-dependency-is-up-to. */
    const val JETBRAINS_DEPENDENCIES_EVERY_DEPENDENCY_IS_UP_TO = "jetbrains.dependencies.every-dependency-is-up-to"

    /** The key jetbrains.dependencies.every-patch-and-minor-update. Values: {0} is size. */
    const val JETBRAINS_DEPENDENCIES_EVERY_PATCH_AND_MINOR_UPDATE = "jetbrains.dependencies.every-patch-and-minor-update"

    /** The key jetbrains.dependencies.major-may-break-your-code. */
    const val JETBRAINS_DEPENDENCIES_MAJOR_MAY_BREAK_YOUR_CODE = "jetbrains.dependencies.major-may-break-your-code"

    /** The key jetbrains.dependencies.minor-new-features-no-breaking. */
    const val JETBRAINS_DEPENDENCIES_MINOR_NEW_FEATURES_NO_BREAKING = "jetbrains.dependencies.minor-new-features-no-breaking"

    /** The key jetbrains.dependencies.monosmith-align-dependency-versions. */
    const val JETBRAINS_DEPENDENCIES_MONOSMITH_ALIGN_DEPENDENCY_VERSIONS = "jetbrains.dependencies.monosmith-align-dependency-versions"

    /** The key jetbrains.dependencies.monosmith-align-the-versions. Values: {0} is rule. */
    const val JETBRAINS_DEPENDENCIES_MONOSMITH_ALIGN_THE_VERSIONS = "jetbrains.dependencies.monosmith-align-the-versions"

    /** The key jetbrains.dependencies.monosmith-update-dependencies. */
    const val JETBRAINS_DEPENDENCIES_MONOSMITH_UPDATE_DEPENDENCIES = "jetbrains.dependencies.monosmith-update-dependencies"

    /** The key jetbrains.dependencies.newer-tools. Values: {0} is tools. */
    const val JETBRAINS_DEPENDENCIES_NEWER_TOOLS = "jetbrains.dependencies.newer-tools"

    /** The key jetbrains.dependencies.nothing-was-changed-yet-every. */
    const val JETBRAINS_DEPENDENCIES_NOTHING_WAS_CHANGED_YET_EVERY = "jetbrains.dependencies.nothing-was-changed-yet-every"

    /** The key jetbrains.dependencies.nothing-was-changed-yet-smith. */
    const val JETBRAINS_DEPENDENCIES_NOTHING_WAS_CHANGED_YET_SMITH = "jetbrains.dependencies.nothing-was-changed-yet-smith"

    /** The key jetbrains.dependencies.nothing-was-updated. Values: {0} is message. */
    const val JETBRAINS_DEPENDENCIES_NOTHING_WAS_UPDATED = "jetbrains.dependencies.nothing-was-updated"

    /** The key jetbrains.dependencies.patch-fixes-only. */
    const val JETBRAINS_DEPENDENCIES_PATCH_FIXES_ONLY = "jetbrains.dependencies.patch-fixes-only"

    /** The key jetbrains.dependencies.pick-among-updates-grouped-by. Values: {0} is size. */
    const val JETBRAINS_DEPENDENCIES_PICK_AMONG_UPDATES_GROUPED_BY = "jetbrains.dependencies.pick-among-updates-grouped-by"

    /** The key jetbrains.dependencies.projects-on-a-newer-version. */
    const val JETBRAINS_DEPENDENCIES_PROJECTS_ON_A_NEWER_VERSION = "jetbrains.dependencies.projects-on-a-newer-version"

    /** The key jetbrains.dependencies.projects-on-an-older-version. */
    const val JETBRAINS_DEPENDENCIES_PROJECTS_ON_AN_OLDER_VERSION = "jetbrains.dependencies.projects-on-an-older-version"

    /** The key jetbrains.dependencies.the-risk-is-low-fixes. */
    const val JETBRAINS_DEPENDENCIES_THE_RISK_IS_LOW_FIXES = "jetbrains.dependencies.the-risk-is-low-fixes"

    /** The key jetbrains.dependencies.the-version-the-package-manager. */
    const val JETBRAINS_DEPENDENCIES_THE_VERSION_THE_PACKAGE_MANAGER = "jetbrains.dependencies.the-version-the-package-manager"

    /** The key jetbrains.dependencies.the-versions-are-aligned. */
    const val JETBRAINS_DEPENDENCIES_THE_VERSIONS_ARE_ALIGNED = "jetbrains.dependencies.the-versions-are-aligned"

    /** The key jetbrains.dependencies.tick-what-to-update-a. */
    const val JETBRAINS_DEPENDENCIES_TICK_WHAT_TO_UPDATE_A = "jetbrains.dependencies.tick-what-to-update-a"

    /** The key jetbrains.dependencies.update. */
    const val JETBRAINS_DEPENDENCIES_UPDATE = "jetbrains.dependencies.update"

    /** The key jetbrains.dependencies.when-projects-declare-a-dependency. */
    const val JETBRAINS_DEPENDENCIES_WHEN_PROJECTS_DECLARE_A_DEPENDENCY = "jetbrains.dependencies.when-projects-declare-a-dependency"

    /** The key jetbrains.derivefiles.a-plain-smith-derive-it. */
    const val JETBRAINS_DERIVEFILES_A_PLAIN_SMITH_DERIVE_IT = "jetbrains.derivefiles.a-plain-smith-derive-it"

    /** The key jetbrains.derivefiles.by-default-smith-leaves-the. Values: {0} is effects. */
    const val JETBRAINS_DERIVEFILES_BY_DEFAULT_SMITH_LEAVES_THE = "jetbrains.derivefiles.by-default-smith-leaves-the"

    /** The key jetbrains.derivefiles.by-default-smith-leaves-the-2. Values: {0} is effects. */
    const val JETBRAINS_DERIVEFILES_BY_DEFAULT_SMITH_LEAVES_THE_2 = "jetbrains.derivefiles.by-default-smith-leaves-the-2"

    /** The key jetbrains.derivefiles.derived-files-need-action-pick. Values: {0} is size. */
    const val JETBRAINS_DERIVEFILES_DERIVED_FILES_NEED_ACTION_PICK = "jetbrains.derivefiles.derived-files-need-action-pick"

    /** The key jetbrains.derivefiles.do-it. */
    const val JETBRAINS_DERIVEFILES_DO_IT = "jetbrains.derivefiles.do-it"

    /** The key jetbrains.derivefiles.eject-own-it. */
    const val JETBRAINS_DERIVEFILES_EJECT_OWN_IT = "jetbrains.derivefiles.eject-own-it"

    /** The key jetbrains.derivefiles.every-derived-file-is-current. */
    const val JETBRAINS_DERIVEFILES_EVERY_DERIVED_FILE_IS_CURRENT = "jetbrains.derivefiles.every-derived-file-is-current"

    /** The key jetbrains.derivefiles.every-file-that-is-stale. */
    const val JETBRAINS_DERIVEFILES_EVERY_FILE_THAT_IS_STALE = "jetbrains.derivefiles.every-file-that-is-stale"

    /** The key jetbrains.derivefiles.explain. */
    const val JETBRAINS_DERIVEFILES_EXPLAIN = "jetbrains.derivefiles.explain"

    /** The key jetbrains.derivefiles.is-done-commit-it-together. Values: {0} is path, {1} is action. */
    const val JETBRAINS_DERIVEFILES_IS_DONE_COMMIT_IT_TOGETHER = "jetbrains.derivefiles.is-done-commit-it-together"

    /** The key jetbrains.derivefiles.keep-my-edit. */
    const val JETBRAINS_DERIVEFILES_KEEP_MY_EDIT = "jetbrains.derivefiles.keep-my-edit"

    /** The key jetbrains.derivefiles.monosmith. Values: {0} is action, {1} is path. */
    const val JETBRAINS_DERIVEFILES_MONOSMITH = "jetbrains.derivefiles.monosmith"

    /** The key jetbrains.derivefiles.monosmith-check-the-derived-files. */
    const val JETBRAINS_DERIVEFILES_MONOSMITH_CHECK_THE_DERIVED_FILES = "jetbrains.derivefiles.monosmith-check-the-derived-files"

    /** The key jetbrains.derivefiles.monosmith-derived-files. */
    const val JETBRAINS_DERIVEFILES_MONOSMITH_DERIVED_FILES = "jetbrains.derivefiles.monosmith-derived-files"

    /** The key jetbrains.derivefiles.monosmith-preview-the-derived-files. */
    const val JETBRAINS_DERIVEFILES_MONOSMITH_PREVIEW_THE_DERIVED_FILES = "jetbrains.derivefiles.monosmith-preview-the-derived-files"

    /** The key jetbrains.derivefiles.monosmith-write-the-derived-files. */
    const val JETBRAINS_DERIVEFILES_MONOSMITH_WRITE_THE_DERIVED_FILES = "jetbrains.derivefiles.monosmith-write-the-derived-files"

    /** The key jetbrains.derivefiles.overwrite-with-the-derived-content. */
    const val JETBRAINS_DERIVEFILES_OVERWRITE_WITH_THE_DERIVED_CONTENT = "jetbrains.derivefiles.overwrite-with-the-derived-content"

    /** The key jetbrains.derivefiles.reclaim-give-it-back-to. */
    const val JETBRAINS_DERIVEFILES_RECLAIM_GIVE_IT_BACK_TO = "jetbrains.derivefiles.reclaim-give-it-back-to"

    /** The key jetbrains.derivefiles.say-who-owns-this-file. */
    const val JETBRAINS_DERIVEFILES_SAY_WHO_OWNS_THIS_FILE = "jetbrains.derivefiles.say-who-owns-this-file"

    /** The key jetbrains.derivefiles.smith-gives-the-file-back. */
    const val JETBRAINS_DERIVEFILES_SMITH_GIVES_THE_FILE_BACK = "jetbrains.derivefiles.smith-gives-the-file-back"

    /** The key jetbrains.derivefiles.smith-keeps-your-edit-as. */
    const val JETBRAINS_DERIVEFILES_SMITH_KEEPS_YOUR_EDIT_AS = "jetbrains.derivefiles.smith-keeps-your-edit-as"

    /** The key jetbrains.derivefiles.smith-overwrites-the-file-with. */
    const val JETBRAINS_DERIVEFILES_SMITH_OVERWRITES_THE_FILE_WITH = "jetbrains.derivefiles.smith-overwrites-the-file-with"

    /** The key jetbrains.derivefiles.smith-writes-every-derived-file. */
    const val JETBRAINS_DERIVEFILES_SMITH_WRITES_EVERY_DERIVED_FILE = "jetbrains.derivefiles.smith-writes-every-derived-file"

    /** The key jetbrains.derivefiles.the-derived-files-are-written. */
    const val JETBRAINS_DERIVEFILES_THE_DERIVED_FILES_ARE_WRITTEN = "jetbrains.derivefiles.the-derived-files-are-written"

    /** The key jetbrains.derivefiles.the-file-becomes-yours-smith. */
    const val JETBRAINS_DERIVEFILES_THE_FILE_BECOMES_YOURS_SMITH = "jetbrains.derivefiles.the-file-becomes-yours-smith"

    /** The key jetbrains.derivefiles.write-the-derived-content. */
    const val JETBRAINS_DERIVEFILES_WRITE_THE_DERIVED_CONTENT = "jetbrains.derivefiles.write-the-derived-content"

    /** The key jetbrains.devservers.monosmith-dev-servers. */
    const val JETBRAINS_DEVSERVERS_MONOSMITH_DEV_SERVERS = "jetbrains.devservers.monosmith-dev-servers"

    /** The key jetbrains.devservers.no-project-has-a-start. */
    const val JETBRAINS_DEVSERVERS_NO_PROJECT_HAS_A_START = "jetbrains.devservers.no-project-has-a-start"

    /** The key jetbrains.devservers.no-server-was-started. Values: {0} is message. */
    const val JETBRAINS_DEVSERVERS_NO_SERVER_WAS_STARTED = "jetbrains.devservers.no-server-was-started"

    /** The key jetbrains.devservers.runs-its-start-task. */
    const val JETBRAINS_DEVSERVERS_RUNS_ITS_START_TASK = "jetbrains.devservers.runs-its-start-task"

    /** The key jetbrains.devservers.smith-dev. Values: {0} is names. */
    const val JETBRAINS_DEVSERVERS_SMITH_DEV = "jetbrains.devservers.smith-dev"

    /** The key jetbrains.devservers.tick-the-projects-to-start. */
    const val JETBRAINS_DEVSERVERS_TICK_THE_PROJECTS_TO_START = "jetbrains.devservers.tick-the-projects-to-start"

    /** The key jetbrains.dockerhub.monosmith-docker-hub-descriptions. */
    const val JETBRAINS_DOCKERHUB_MONOSMITH_DOCKER_HUB_DESCRIPTIONS = "jetbrains.dockerhub.monosmith-docker-hub-descriptions"

    /** The key jetbrains.dockerhub.smith-calls-docker-hub-with. */
    const val JETBRAINS_DOCKERHUB_SMITH_CALLS_DOCKER_HUB_WITH = "jetbrains.dockerhub.smith-calls-docker-hub-with"

    /** The key jetbrains.dockerhub.the-docker-hub-descriptions-are. */
    const val JETBRAINS_DOCKERHUB_THE_DOCKER_HUB_DESCRIPTIONS_ARE = "jetbrains.dockerhub.the-docker-hub-descriptions-are"

    /** The key jetbrains.health.apply-the-safe-fixes. */
    const val JETBRAINS_HEALTH_APPLY_THE_SAFE_FIXES = "jetbrains.health.apply-the-safe-fixes"

    /** The key jetbrains.health.doctor-all-checks-passed. Values: {0} is passed. */
    const val JETBRAINS_HEALTH_DOCTOR_ALL_CHECKS_PASSED = "jetbrains.health.doctor-all-checks-passed"

    /** The key jetbrains.health.doctor-did-not-run. Values: {0} is message. */
    const val JETBRAINS_HEALTH_DOCTOR_DID_NOT_RUN = "jetbrains.health.doctor-did-not-run"

    /** The key jetbrains.health.doctor-failed-warnings-passed. Values: {0} is failed, {1} is warnings, {2} is passed. */
    const val JETBRAINS_HEALTH_DOCTOR_FAILED_WARNINGS_PASSED = "jetbrains.health.doctor-failed-warnings-passed"

    /** The key jetbrains.health.each-fix-is-shown-as. */
    const val JETBRAINS_HEALTH_EACH_FIX_IS_SHOWN_AS = "jetbrains.health.each-fix-is-shown-as"

    /** The key jetbrains.health.monosmith-apply-the-safe-fixes. */
    const val JETBRAINS_HEALTH_MONOSMITH_APPLY_THE_SAFE_FIXES = "jetbrains.health.monosmith-apply-the-safe-fixes"

    /** The key jetbrains.health.monosmith-apply-the-safe-fixes-2. */
    const val JETBRAINS_HEALTH_MONOSMITH_APPLY_THE_SAFE_FIXES_2 = "jetbrains.health.monosmith-apply-the-safe-fixes-2"

    /** The key jetbrains.health.monosmith-doctor. */
    const val JETBRAINS_HEALTH_MONOSMITH_DOCTOR = "jetbrains.health.monosmith-doctor"

    /** The key jetbrains.health.monosmith-fix. Values: {0} is check. */
    const val JETBRAINS_HEALTH_MONOSMITH_FIX = "jetbrains.health.monosmith-fix"

    /** The key jetbrains.health.monosmith-fix-2. Values: {0} is check. */
    const val JETBRAINS_HEALTH_MONOSMITH_FIX_2 = "jetbrains.health.monosmith-fix-2"

    /** The key jetbrains.health.nothing-was-changed. Values: {0} is failureTextMapper. */
    const val JETBRAINS_HEALTH_NOTHING_WAS_CHANGED = "jetbrains.health.nothing-was-changed"

    /** The key jetbrains.health.open-the-health-tab-each. */
    const val JETBRAINS_HEALTH_OPEN_THE_HEALTH_TAB_EACH = "jetbrains.health.open-the-health-tab-each"

    /** The key jetbrains.health.smith-has-no-fix-for. Values: {0} is remedy. */
    const val JETBRAINS_HEALTH_SMITH_HAS_NO_FIX_FOR = "jetbrains.health.smith-has-no-fix-for"

    /** The key jetbrains.health.smith-has-no-safe-fix. Values: {0} is failureTextMapper. */
    const val JETBRAINS_HEALTH_SMITH_HAS_NO_SAFE_FIX = "jetbrains.health.smith-has-no-safe-fix"

    /** The key jetbrains.health.the-fix-is-applied-another. */
    const val JETBRAINS_HEALTH_THE_FIX_IS_APPLIED_ANOTHER = "jetbrains.health.the-fix-is-applied-another"

    /** The key jetbrains.health.the-fix-is-applied-review. */
    const val JETBRAINS_HEALTH_THE_FIX_IS_APPLIED_REVIEW = "jetbrains.health.the-fix-is-applied-review"

    /** The key jetbrains.health.the-safe-fixes-are-applied. */
    const val JETBRAINS_HEALTH_THE_SAFE_FIXES_ARE_APPLIED = "jetbrains.health.the-safe-fixes-are-applied"

    /** The key jetbrains.health.this-is-what-smith-would. Values: {0} is description. */
    const val JETBRAINS_HEALTH_THIS_IS_WHAT_SMITH_WOULD = "jetbrains.health.this-is-what-smith-would"

    /** The key jetbrains.installdependency.a-package-from-the-registry. */
    const val JETBRAINS_INSTALLDEPENDENCY_A_PACKAGE_FROM_THE_REGISTRY = "jetbrains.installdependency.a-package-from-the-registry"

    /** The key jetbrains.installdependency.add-a-dependency. */
    const val JETBRAINS_INSTALLDEPENDENCY_ADD_A_DEPENDENCY = "jetbrains.installdependency.add-a-dependency"

    /** The key jetbrains.installdependency.downloaded-from-the-language. */
    const val JETBRAINS_INSTALLDEPENDENCY_DOWNLOADED_FROM_THE_LANGUAGE = "jetbrains.installdependency.downloaded-from-the-language"

    /** The key jetbrains.installdependency.follow-the-workspace. */
    const val JETBRAINS_INSTALLDEPENDENCY_FOLLOW_THE_WORKSPACE = "jetbrains.installdependency.follow-the-workspace"

    /** The key jetbrains.installdependency.libraries-may-use. Values: {0} is name. */
    const val JETBRAINS_INSTALLDEPENDENCY_LIBRARIES_MAY_USE = "jetbrains.installdependency.libraries-may-use"

    /** The key jetbrains.installdependency.linked-nothing-is-downloaded. */
    const val JETBRAINS_INSTALLDEPENDENCY_LINKED_NOTHING_IS_DOWNLOADED = "jetbrains.installdependency.linked-nothing-is-downloaded"

    /** The key jetbrains.installdependency.monosmith-add-a-dependency. */
    const val JETBRAINS_INSTALLDEPENDENCY_MONOSMITH_ADD_A_DEPENDENCY = "jetbrains.installdependency.monosmith-add-a-dependency"

    /** The key jetbrains.installdependency.no-library-of-this-workspace. Values: {0} is name. */
    const val JETBRAINS_INSTALLDEPENDENCY_NO_LIBRARY_OF_THIS_WORKSPACE = "jetbrains.installdependency.no-library-of-this-workspace"

    /** The key jetbrains.installdependency.nothing-was-added. Values: {0} is message. */
    const val JETBRAINS_INSTALLDEPENDENCY_NOTHING_WAS_ADDED = "jetbrains.installdependency.nothing-was-added"

    /** The key jetbrains.installdependency.package-names-separated-by. */
    const val JETBRAINS_INSTALLDEPENDENCY_PACKAGE_NAMES_SEPARATED_BY = "jetbrains.installdependency.package-names-separated-by"

    /** The key jetbrains.installdependency.smith-runs-this-command-it. */
    const val JETBRAINS_INSTALLDEPENDENCY_SMITH_RUNS_THIS_COMMAND_IT = "jetbrains.installdependency.smith-runs-this-command-it"

    /** The key jetbrains.installdependency.the-dependency-is-added. */
    const val JETBRAINS_INSTALLDEPENDENCY_THE_DEPENDENCY_IS_ADDED = "jetbrains.installdependency.the-dependency-is-added"

    /** The key jetbrains.installdependency.the-dependency-is-added-not. */
    const val JETBRAINS_INSTALLDEPENDENCY_THE_DEPENDENCY_IS_ADDED_NOT = "jetbrains.installdependency.the-dependency-is-added-not"

    /** The key jetbrains.installdependency.type-at-least-one-package. */
    const val JETBRAINS_INSTALLDEPENDENCY_TYPE_AT_LEAST_ONE_PACKAGE = "jetbrains.installdependency.type-at-least-one-package"

    /** The key jetbrains.installdependency.uses-the-code-of-another. Values: {0} is project. */
    const val JETBRAINS_INSTALLDEPENDENCY_USES_THE_CODE_OF_ANOTHER = "jetbrains.installdependency.uses-the-code-of-another"

    /** The key jetbrains.installdependency.what-do-you-add-to. Values: {0} is title, {1} is project. */
    const val JETBRAINS_INSTALLDEPENDENCY_WHAT_DO_YOU_ADD_TO = "jetbrains.installdependency.what-do-you-add-to"

    /** The key jetbrains.installdependency.which-library. Values: {0} is title. */
    const val JETBRAINS_INSTALLDEPENDENCY_WHICH_LIBRARY = "jetbrains.installdependency.which-library"

    /** The key jetbrains.installdependency.which-version-of-does-ask. Values: {0} is title, {1} is library, {2} is project. */
    const val JETBRAINS_INSTALLDEPENDENCY_WHICH_VERSION_OF_DOES_ASK = "jetbrains.installdependency.which-version-of-does-ask"

    /** The key jetbrains.installtools.install-yourself-then-restart-the. Values: {0} is title. */
    const val JETBRAINS_INSTALLTOOLS_INSTALL_YOURSELF_THEN_RESTART_THE = "jetbrains.installtools.install-yourself-then-restart-the"

    /** The key jetbrains.installtools.is-installed. Values: {0} is title. */
    const val JETBRAINS_INSTALLTOOLS_IS_INSTALLED = "jetbrains.installtools.is-installed"

    /** The key jetbrains.installtools.is-without-it-the-first. Values: {0} is title, {1} is state, {2} is why, {3} is consequence. */
    const val JETBRAINS_INSTALLTOOLS_IS_WITHOUT_IT_THE_FIRST = "jetbrains.installtools.is-without-it-the-first"

    /** The key jetbrains.installtools.it-needs-administrator-rights-and. */
    const val JETBRAINS_INSTALLTOOLS_IT_NEEDS_ADMINISTRATOR_RIGHTS_AND = "jetbrains.installtools.it-needs-administrator-rights-and"

    /** The key jetbrains.installtools.monosmith-get. Values: {0} is title. */
    const val JETBRAINS_INSTALLTOOLS_MONOSMITH_GET = "jetbrains.installtools.monosmith-get"

    /** The key jetbrains.installtools.monosmith-install-the-tools. */
    const val JETBRAINS_INSTALLTOOLS_MONOSMITH_INSTALL_THE_TOOLS = "jetbrains.installtools.monosmith-install-the-tools"

    /** The key jetbrains.installtools.monosmith-install-yourself. Values: {0} is title. */
    const val JETBRAINS_INSTALLTOOLS_MONOSMITH_INSTALL_YOURSELF = "jetbrains.installtools.monosmith-install-yourself"

    /** The key jetbrains.installtools.monosmith-needs-administrator-rights. Values: {0} is title. */
    const val JETBRAINS_INSTALLTOOLS_MONOSMITH_NEEDS_ADMINISTRATOR_RIGHTS = "jetbrains.installtools.monosmith-needs-administrator-rights"

    /** The key jetbrains.installtools.needs-administrator-rights-smith. Values: {0} is command. */
    const val JETBRAINS_INSTALLTOOLS_NEEDS_ADMINISTRATOR_RIGHTS_SMITH = "jetbrains.installtools.needs-administrator-rights-smith"

    /** The key jetbrains.installtools.nothing-was-installed. Values: {0} is failureTextMapper. */
    const val JETBRAINS_INSTALLTOOLS_NOTHING_WAS_INSTALLED = "jetbrains.installtools.nothing-was-installed"

    /** The key jetbrains.installtools.opened-install-then-restart-the. Values: {0} is download, {1} is title. */
    const val JETBRAINS_INSTALLTOOLS_OPENED_INSTALL_THEN_RESTART_THE = "jetbrains.installtools.opened-install-then-restart-the"

    /** The key jetbrains.installtools.run-one-of-these-then. */
    const val JETBRAINS_INSTALLTOOLS_RUN_ONE_OF_THESE_THEN = "jetbrains.installtools.run-one-of-these-then"

    /** The key jetbrains.installtools.run-this-yourself-then-restart. Values: {0} is command. */
    const val JETBRAINS_INSTALLTOOLS_RUN_THIS_YOURSELF_THEN_RESTART = "jetbrains.installtools.run-this-yourself-then-restart"

    /** The key jetbrains.installtools.runs. Values: {0} is command. */
    const val JETBRAINS_INSTALLTOOLS_RUNS = "jetbrains.installtools.runs"

    /** The key jetbrains.installtools.skip. Values: {0} is title. */
    const val JETBRAINS_INSTALLTOOLS_SKIP = "jetbrains.installtools.skip"

    /** The key jetbrains.installtools.skipped. Values: {0} is title, {1} is consequence. */
    const val JETBRAINS_INSTALLTOOLS_SKIPPED = "jetbrains.installtools.skipped"

    /** The key jetbrains.installtools.smith-runs-the-command-and. */
    const val JETBRAINS_INSTALLTOOLS_SMITH_RUNS_THE_COMMAND_AND = "jetbrains.installtools.smith-runs-the-command-and"

    /** The key jetbrains.installtools.smith-shows-the-command. Values: {0} is command. */
    const val JETBRAINS_INSTALLTOOLS_SMITH_SHOWS_THE_COMMAND = "jetbrains.installtools.smith-shows-the-command"

    /** The key jetbrains.installtools.smith-tools-install-printed-something. */
    const val JETBRAINS_INSTALLTOOLS_SMITH_TOOLS_INSTALL_PRINTED_SOMETHING = "jetbrains.installtools.smith-tools-install-printed-something"

    /** The key jetbrains.installtools.the-install-of-did-not. Values: {0} is title, {1} is failureTextMapper. */
    const val JETBRAINS_INSTALLTOOLS_THE_INSTALL_OF_DID_NOT = "jetbrains.installtools.the-install-of-did-not"

    /** The key jetbrains.mcpinstall.monosmith-connect-an-assistant. */
    const val JETBRAINS_MCPINSTALL_MONOSMITH_CONNECT_AN_ASSISTANT = "jetbrains.mcpinstall.monosmith-connect-an-assistant"

    /** The key jetbrains.mcpinstall.nothing-was-written-yet. */
    const val JETBRAINS_MCPINSTALL_NOTHING_WAS_WRITTEN_YET = "jetbrains.mcpinstall.nothing-was-written-yet"

    /** The key jetbrains.mcpinstall.smith-is-connected. */
    const val JETBRAINS_MCPINSTALL_SMITH_IS_CONNECTED = "jetbrains.mcpinstall.smith-is-connected"

    /** The key jetbrains.mcpinstall.smith-it-writes-the-file. Values: {0} is args. */
    const val JETBRAINS_MCPINSTALL_SMITH_IT_WRITES_THE_FILE = "jetbrains.mcpinstall.smith-it-writes-the-file"

    /** The key jetbrains.mcpinstall.tick-the-assistants-to-connect. */
    const val JETBRAINS_MCPINSTALL_TICK_THE_ASSISTANTS_TO_CONNECT = "jetbrains.mcpinstall.tick-the-assistants-to-connect"

    /** The key jetbrains.mcpinstall.writes-mcp-json. */
    const val JETBRAINS_MCPINSTALL_WRITES_MCP_JSON = "jetbrains.mcpinstall.writes-mcp-json"

    /** The key jetbrains.mcpinstall.writes-vscode-mcp-json. */
    const val JETBRAINS_MCPINSTALL_WRITES_VSCODE_MCP_JSON = "jetbrains.mcpinstall.writes-vscode-mcp-json"

    /** The key jetbrains.newrecipe.is-written. Values: {0} is recipeLevelsMapper. */
    const val JETBRAINS_NEWRECIPE_IS_WRITTEN = "jetbrains.newrecipe.is-written"

    /** The key jetbrains.newrecipe.leave-it-to-the-users. */
    const val JETBRAINS_NEWRECIPE_LEAVE_IT_TO_THE_USERS = "jetbrains.newrecipe.leave-it-to-the-users"

    /** The key jetbrains.newrecipe.monosmith-new-recipe. */
    const val JETBRAINS_NEWRECIPE_MONOSMITH_NEW_RECIPE = "jetbrains.newrecipe.monosmith-new-recipe"

    /** The key jetbrains.newrecipe.no-recipe-was-written. Values: {0} is failureTextMapper. */
    const val JETBRAINS_NEWRECIPE_NO_RECIPE_WAS_WRITTEN = "jetbrains.newrecipe.no-recipe-was-written"

    /** The key jetbrains.newrecipe.pick-the-plugin-the-recipe. Values: {0} is level. */
    const val JETBRAINS_NEWRECIPE_PICK_THE_PLUGIN_THE_RECIPE = "jetbrains.newrecipe.pick-the-plugin-the-recipe"

    /** The key jetbrains.newrecipe.smith-knows-no-plugin-a. */
    const val JETBRAINS_NEWRECIPE_SMITH_KNOWS_NO_PLUGIN_A = "jetbrains.newrecipe.smith-knows-no-plugin-a"

    /** The key jetbrains.newrecipe.smith-printed-its-levels-in. */
    const val JETBRAINS_NEWRECIPE_SMITH_PRINTED_ITS_LEVELS_IN = "jetbrains.newrecipe.smith-printed-its-levels-in"

    /** The key jetbrains.newrecipe.smith-runs-this-command-and. Values: {0} is recipeLevelsMapper. */
    const val JETBRAINS_NEWRECIPE_SMITH_RUNS_THIS_COMMAND_AND = "jetbrains.newrecipe.smith-runs-this-command-and"

    /** The key jetbrains.newrecipe.the-first-version-of-the. */
    const val JETBRAINS_NEWRECIPE_THE_FIRST_VERSION_OF_THE = "jetbrains.newrecipe.the-first-version-of-the"

    /** The key jetbrains.newrecipe.the-name-of-the-recipe. */
    const val JETBRAINS_NEWRECIPE_THE_NAME_OF_THE_RECIPE = "jetbrains.newrecipe.the-name-of-the-recipe"

    /** The key jetbrains.newrecipe.the-recipe-fixes-no-plugin. Values: {0} is level. */
    const val JETBRAINS_NEWRECIPE_THE_RECIPE_FIXES_NO_PLUGIN = "jetbrains.newrecipe.the-recipe-fixes-no-plugin"

    /** The key jetbrains.newrecipe.the-recipe-this-one-extends. */
    const val JETBRAINS_NEWRECIPE_THE_RECIPE_THIS_ONE_EXTENDS = "jetbrains.newrecipe.the-recipe-this-one-extends"

    /** The key jetbrains.newrecipe.tick-the-plugins-the-recipe. Values: {0} is level. */
    const val JETBRAINS_NEWRECIPE_TICK_THE_PLUGINS_THE_RECIPE = "jetbrains.newrecipe.tick-the-plugins-the-recipe"

    /** The key jetbrains.newrecipe.use-lowercase-words-joined-by. */
    const val JETBRAINS_NEWRECIPE_USE_LOWERCASE_WORDS_JOINED_BY = "jetbrains.newrecipe.use-lowercase-words-joined-by"

    /** The key jetbrains.newrecipe.use-major-minor-patch-for. */
    const val JETBRAINS_NEWRECIPE_USE_MAJOR_MINOR_PATCH_FOR = "jetbrains.newrecipe.use-major-minor-patch-for"

    /** The key jetbrains.newrecipe.write-the-recipe. */
    const val JETBRAINS_NEWRECIPE_WRITE_THE_RECIPE = "jetbrains.newrecipe.write-the-recipe"

    /** The key jetbrains.newworkspace.agent-files-for-claude-code. Values: {0} is value. */
    const val JETBRAINS_NEWWORKSPACE_AGENT_FILES_FOR_CLAUDE_CODE = "jetbrains.newworkspace.agent-files-for-claude-code"

    /** The key jetbrains.newworkspace.answer-each-question-of-smith. */
    const val JETBRAINS_NEWWORKSPACE_ANSWER_EACH_QUESTION_OF_SMITH = "jetbrains.newworkspace.answer-each-question-of-smith"

    /** The key jetbrains.newworkspace.ci. Values: {0} is ci. */
    const val JETBRAINS_NEWWORKSPACE_CI = "jetbrains.newworkspace.ci"

    /** The key jetbrains.newworkspace.code-layout-smith-checks. Values: {0} is architecture. */
    const val JETBRAINS_NEWWORKSPACE_CODE_LAYOUT_SMITH_CHECKS = "jetbrains.newworkspace.code-layout-smith-checks"

    /** The key jetbrains.newworkspace.create-it. Values: {0} is title. */
    const val JETBRAINS_NEWWORKSPACE_CREATE_IT = "jetbrains.newworkspace.create-it"

    /** The key jetbrains.newworkspace.create-the-workspace. */
    const val JETBRAINS_NEWWORKSPACE_CREATE_THE_WORKSPACE = "jetbrains.newworkspace.create-the-workspace"

    /** The key jetbrains.newworkspace.dev-container. Values: {0} is value. */
    const val JETBRAINS_NEWWORKSPACE_DEV_CONTAINER = "jetbrains.newworkspace.dev-container"

    /** The key jetbrains.newworkspace.first-commit. Values: {0} is title. */
    const val JETBRAINS_NEWWORKSPACE_FIRST_COMMIT = "jetbrains.newworkspace.first-commit"

    /** The key jetbrains.newworkspace.first-commit-of-smith-s. Values: {0} is value. */
    const val JETBRAINS_NEWWORKSPACE_FIRST_COMMIT_OF_SMITH_S = "jetbrains.newworkspace.first-commit-of-smith-s"

    /** The key jetbrains.newworkspace.is-created. Values: {0} is name. */
    const val JETBRAINS_NEWWORKSPACE_IS_CREATED = "jetbrains.newworkspace.is-created"

    /** The key jetbrains.newworkspace.main-language. Values: {0} is newQuestionsMapper. */
    const val JETBRAINS_NEWWORKSPACE_MAIN_LANGUAGE = "jetbrains.newworkspace.main-language"

    /** The key jetbrains.newworkspace.make-the-first-commit-of. */
    const val JETBRAINS_NEWWORKSPACE_MAKE_THE_FIRST_COMMIT_OF = "jetbrains.newworkspace.make-the-first-commit-of"

    /** The key jetbrains.newworkspace.monosmith-new-workspace. */
    const val JETBRAINS_NEWWORKSPACE_MONOSMITH_NEW_WORKSPACE = "jetbrains.newworkspace.monosmith-new-workspace"

    /** The key jetbrains.newworkspace.name. Values: {0} is name, {1} is value. */
    const val JETBRAINS_NEWWORKSPACE_NAME = "jetbrains.newworkspace.name"

    /** The key jetbrains.newworkspace.no. */
    const val JETBRAINS_NEWWORKSPACE_NO = "jetbrains.newworkspace.no"

    /** The key jetbrains.newworkspace.nothing-is-written-before-you. */
    const val JETBRAINS_NEWWORKSPACE_NOTHING_IS_WRITTEN_BEFORE_YOU = "jetbrains.newworkspace.nothing-is-written-before-you"

    /** The key jetbrains.newworkspace.nothing-is-written-yet-these. */
    const val JETBRAINS_NEWWORKSPACE_NOTHING_IS_WRITTEN_YET_THESE = "jetbrains.newworkspace.nothing-is-written-yet-these"

    /** The key jetbrains.newworkspace.nothing-was-created. Values: {0} is failureTextMapper. */
    const val JETBRAINS_NEWWORKSPACE_NOTHING_WAS_CREATED = "jetbrains.newworkspace.nothing-was-created"

    /** The key jetbrains.newworkspace.nothing-was-created-smith-did. Values: {0} is failureTextMapper. */
    const val JETBRAINS_NEWWORKSPACE_NOTHING_WAS_CREATED_SMITH_DID = "jetbrains.newworkspace.nothing-was-created-smith-did"

    /** The key jetbrains.newworkspace.smith-makes-the-first-commit. */
    const val JETBRAINS_NEWWORKSPACE_SMITH_MAKES_THE_FIRST_COMMIT = "jetbrains.newworkspace.smith-makes-the-first-commit"

    /** The key jetbrains.newworkspace.tests-and-lint-for-node. Values: {0} is newQuestionsMapper. */
    const val JETBRAINS_NEWWORKSPACE_TESTS_AND_LINT_FOR_NODE = "jetbrains.newworkspace.tests-and-lint-for-node"

    /** The key jetbrains.newworkspace.the-default-is-ticked. Values: {0} is prompt. */
    const val JETBRAINS_NEWWORKSPACE_THE_DEFAULT_IS_TICKED = "jetbrains.newworkspace.the-default-is-ticked"

    /** The key jetbrains.newworkspace.the-files-stay-uncommitted. */
    const val JETBRAINS_NEWWORKSPACE_THE_FILES_STAY_UNCOMMITTED = "jetbrains.newworkspace.the-files-stay-uncommitted"

    /** The key jetbrains.newworkspace.the-first-answer-is-the. Values: {0} is prompt. */
    const val JETBRAINS_NEWWORKSPACE_THE_FIRST_ANSWER_IS_THE = "jetbrains.newworkspace.the-first-answer-is-the"

    /** The key jetbrains.newworkspace.the-name-of-the-workspace. Values: {0} is folderName. */
    const val JETBRAINS_NEWWORKSPACE_THE_NAME_OF_THE_WORKSPACE = "jetbrains.newworkspace.the-name-of-the-workspace"

    /** The key jetbrains.newworkspace.yes. */
    const val JETBRAINS_NEWWORKSPACE_YES = "jetbrains.newworkspace.yes"

    /** The key jetbrains.pipelinetemplate.monosmith-add-a-pipeline. */
    const val JETBRAINS_PIPELINETEMPLATE_MONOSMITH_ADD_A_PIPELINE = "jetbrains.pipelinetemplate.monosmith-add-a-pipeline"

    /** The key jetbrains.pipelinetemplate.name-the-app. */
    const val JETBRAINS_PIPELINETEMPLATE_NAME_THE_APP = "jetbrains.pipelinetemplate.name-the-app"

    /** The key jetbrains.pipelinetemplate.no-pipeline-was-added. Values: {0} is failureTextMapper. */
    const val JETBRAINS_PIPELINETEMPLATE_NO_PIPELINE_WAS_ADDED = "jetbrains.pipelinetemplate.no-pipeline-was-added"

    /** The key jetbrains.pipelinetemplate.pick-a-pipeline-it-is. */
    const val JETBRAINS_PIPELINETEMPLATE_PICK_A_PIPELINE_IT_IS = "jetbrains.pipelinetemplate.pick-a-pipeline-it-is"

    /** The key jetbrains.pipelinetemplate.smith-listed-no-pipeline-update. */
    const val JETBRAINS_PIPELINETEMPLATE_SMITH_LISTED_NO_PIPELINE_UPDATE = "jetbrains.pipelinetemplate.smith-listed-no-pipeline-update"

    /** The key jetbrains.pipelinetemplate.the-app-this-pipeline-builds. */
    const val JETBRAINS_PIPELINETEMPLATE_THE_APP_THIS_PIPELINE_BUILDS = "jetbrains.pipelinetemplate.the-app-this-pipeline-builds"

    /** The key jetbrains.pipelinetemplate.the-file-is-yours-from. */
    const val JETBRAINS_PIPELINETEMPLATE_THE_FILE_IS_YOURS_FROM = "jetbrains.pipelinetemplate.the-file-is-yours-from"

    /** The key jetbrains.pipelinetemplate.the-pipeline-is-written. Values: {0} is name. */
    const val JETBRAINS_PIPELINETEMPLATE_THE_PIPELINE_IS_WRITTEN = "jetbrains.pipelinetemplate.the-pipeline-is-written"

    /** The key jetbrains.pipelinetemplate.write-the-pipeline. */
    const val JETBRAINS_PIPELINETEMPLATE_WRITE_THE_PIPELINE = "jetbrains.pipelinetemplate.write-the-pipeline"

    /** The key jetbrains.planapply.commit-changes. */
    const val JETBRAINS_PLANAPPLY_COMMIT_CHANGES = "jetbrains.planapply.commit-changes"

    /** The key jetbrains.planapply.nothing-is-committed. */
    const val JETBRAINS_PLANAPPLY_NOTHING_IS_COMMITTED = "jetbrains.planapply.nothing-is-committed"

    /** The key jetbrains.pluginmarket.also-pin-the-sha-256. */
    const val JETBRAINS_PLUGINMARKET_ALSO_PIN_THE_SHA_256 = "jetbrains.pluginmarket.also-pin-the-sha-256"

    /** The key jetbrains.pluginmarket.also-trust-it-in-this. */
    const val JETBRAINS_PLUGINMARKET_ALSO_TRUST_IT_IN_THIS = "jetbrains.pluginmarket.also-trust-it-in-this"

    /** The key jetbrains.pluginmarket.is-a-program-that-runs. Values: {0} is name, {1} is newest, {2} is value. */
    const val JETBRAINS_PLUGINMARKET_IS_A_PROGRAM_THAT_RUNS = "jetbrains.pluginmarket.is-a-program-that-runs"

    /** The key jetbrains.pluginmarket.is-installed. Values: {0} is name, {1} is newest. */
    const val JETBRAINS_PLUGINMARKET_IS_INSTALLED = "jetbrains.pluginmarket.is-installed"

    /** The key jetbrains.pluginmarket.monosmith-install. Values: {0} is name. */
    const val JETBRAINS_PLUGINMARKET_MONOSMITH_INSTALL = "jetbrains.pluginmarket.monosmith-install"

    /** The key jetbrains.pluginmarket.monosmith-plugins. */
    const val JETBRAINS_PLUGINMARKET_MONOSMITH_PLUGINS = "jetbrains.pluginmarket.monosmith-plugins"

    /** The key jetbrains.pluginmarket.no-plugin-matches. Values: {0} is term. */
    const val JETBRAINS_PLUGINMARKET_NO_PLUGIN_MATCHES = "jetbrains.pluginmarket.no-plugin-matches"

    /** The key jetbrains.pluginmarket.no-plugin-was-installed. Values: {0} is failureTextMapper. */
    const val JETBRAINS_PLUGINMARKET_NO_PLUGIN_WAS_INSTALLED = "jetbrains.pluginmarket.no-plugin-was-installed"

    /** The key jetbrains.pluginmarket.pick-a-plugin-the-newest. */
    const val JETBRAINS_PLUGINMARKET_PICK_A_PLUGIN_THE_NEWEST = "jetbrains.pluginmarket.pick-a-plugin-the-newest"

    /** The key jetbrains.pluginmarket.smith-printed-the-plugins-in. */
    const val JETBRAINS_PLUGINMARKET_SMITH_PRINTED_THE_PLUGINS_IN = "jetbrains.pluginmarket.smith-printed-the-plugins-in"

    /** The key jetbrains.pluginmarket.was-not-installed. Values: {0} is name, {1} is failureTextMapper. */
    const val JETBRAINS_PLUGINMARKET_WAS_NOT_INSTALLED = "jetbrains.pluginmarket.was-not-installed"

    /** The key jetbrains.pluginmarket.words-to-search-for-empty. */
    const val JETBRAINS_PLUGINMARKET_WORDS_TO_SEARCH_FOR_EMPTY = "jetbrains.pluginmarket.words-to-search-for-empty"

    /** The key jetbrains.pluginswitch.is-committed-as-one-change. Values: {0} is name, {1} is word. */
    const val JETBRAINS_PLUGINSWITCH_IS_COMMITTED_AS_ONE_CHANGE = "jetbrains.pluginswitch.is-committed-as-one-change"

    /** The key jetbrains.pluginswitch.is-not-committed-review-it. Values: {0} is name, {1} is word. */
    const val JETBRAINS_PLUGINSWITCH_IS_NOT_COMMITTED_REVIEW_IT = "jetbrains.pluginswitch.is-not-committed-review-it"

    /** The key jetbrains.pluginswitch.no-plugin-was-changed. Values: {0} is failureTextMapper. */
    const val JETBRAINS_PLUGINSWITCH_NO_PLUGIN_WAS_CHANGED = "jetbrains.pluginswitch.no-plugin-was-changed"

    /** The key jetbrains.pluginswitch.nothing-was-changed-these-projects. Values: {0} is name. */
    const val JETBRAINS_PLUGINSWITCH_NOTHING_WAS_CHANGED_THESE_PROJECTS = "jetbrains.pluginswitch.nothing-was-changed-these-projects"

    /** The key jetbrains.pluginswitch.pick-a-built-in-plugin. */
    const val JETBRAINS_PLUGINSWITCH_PICK_A_BUILT_IN_PLUGIN = "jetbrains.pluginswitch.pick-a-built-in-plugin"

    /** The key jetbrains.pluginswitch.smith-counts-the-projects-finds. Values: {0} is name. */
    const val JETBRAINS_PLUGINSWITCH_SMITH_COUNTS_THE_PROJECTS_FINDS = "jetbrains.pluginswitch.smith-counts-the-projects-finds"

    /** The key jetbrains.pluginswitch.smith-lists-no-built-in. */
    const val JETBRAINS_PLUGINSWITCH_SMITH_LISTS_NO_BUILT_IN = "jetbrains.pluginswitch.smith-lists-no-built-in"

    /** The key jetbrains.pluginswitch.switching-off-hides-projects. */
    const val JETBRAINS_PLUGINSWITCH_SWITCHING_OFF_HIDES_PROJECTS = "jetbrains.pluginswitch.switching-off-hides-projects"

    /** The key jetbrains.pluginswitch.the-projects-finds-appear-in. Values: {0} is name. */
    const val JETBRAINS_PLUGINSWITCH_THE_PROJECTS_FINDS_APPEAR_IN = "jetbrains.pluginswitch.the-projects-finds-appear-in"

    /** The key jetbrains.pluginswitch.turn-a-plugin. Values: {0} is title, {1} is word. */
    const val JETBRAINS_PLUGINSWITCH_TURN_A_PLUGIN = "jetbrains.pluginswitch.turn-a-plugin"

    /** The key jetbrains.pluginswitch.turn-it. Values: {0} is word. */
    const val JETBRAINS_PLUGINSWITCH_TURN_IT = "jetbrains.pluginswitch.turn-it"

    /** The key jetbrains.pluginswitch.turn-it-off-anyway. */
    const val JETBRAINS_PLUGINSWITCH_TURN_IT_OFF_ANYWAY = "jetbrains.pluginswitch.turn-it-off-anyway"

    /** The key jetbrains.pluginswitch.turn-off-anyway. Values: {0} is title, {1} is name. */
    const val JETBRAINS_PLUGINSWITCH_TURN_OFF_ANYWAY = "jetbrains.pluginswitch.turn-off-anyway"

    /** The key jetbrains.pluginswitch.was-not-turned. Values: {0} is name, {1} is word, {2} is failureTextMapper. */
    const val JETBRAINS_PLUGINSWITCH_WAS_NOT_TURNED = "jetbrains.pluginswitch.was-not-turned"

    /** The key jetbrains.pluginswitch.was-not-turned-off. Values: {0} is name, {1} is failureTextMapper. */
    const val JETBRAINS_PLUGINSWITCH_WAS_NOT_TURNED_OFF = "jetbrains.pluginswitch.was-not-turned-off"

    /** The key jetbrains.projectactions.an-app. */
    const val JETBRAINS_PROJECTACTIONS_AN_APP = "jetbrains.projectactions.an-app"

    /** The key jetbrains.projectactions.depends-on-no-library-of. Values: {0} is name. */
    const val JETBRAINS_PROJECTACTIONS_DEPENDS_ON_NO_LIBRARY_OF = "jetbrains.projectactions.depends-on-no-library-of"

    /** The key jetbrains.projectactions.is-now. Values: {0} is name, {1} is projectActionArgumentsAlgorithm. */
    const val JETBRAINS_PROJECTACTIONS_IS_NOW = "jetbrains.projectactions.is-now"

    /** The key jetbrains.projectactions.is-now-what-should-it. Values: {0} is name, {1} is projectActionArgumentsAlgorithm. */
    const val JETBRAINS_PROJECTACTIONS_IS_NOW_WHAT_SHOULD_IT = "jetbrains.projectactions.is-now-what-should-it"

    /** The key jetbrains.projectactions.is-removed. Values: {0} is name. */
    const val JETBRAINS_PROJECTACTIONS_IS_REMOVED = "jetbrains.projectactions.is-removed"

    /** The key jetbrains.projectactions.is-unlinked-from. Values: {0} is library, {1} is name. */
    const val JETBRAINS_PROJECTACTIONS_IS_UNLINKED_FROM = "jetbrains.projectactions.is-unlinked-from"

    /** The key jetbrains.projectactions.keep-a-copy-moves-the. */
    const val JETBRAINS_PROJECTACTIONS_KEEP_A_COPY_MOVES_THE = "jetbrains.projectactions.keep-a-copy-moves-the"

    /** The key jetbrains.projectactions.make-it. Values: {0} is projectActionArgumentsAlgorithm. */
    const val JETBRAINS_PROJECTACTIONS_MAKE_IT = "jetbrains.projectactions.make-it"

    /** The key jetbrains.projectactions.make-it-an-app-a. */
    const val JETBRAINS_PROJECTACTIONS_MAKE_IT_AN_APP_A = "jetbrains.projectactions.make-it-an-app-a"

    /** The key jetbrains.projectactions.monosmith-change-the-kind-of. */
    const val JETBRAINS_PROJECTACTIONS_MONOSMITH_CHANGE_THE_KIND_OF = "jetbrains.projectactions.monosmith-change-the-kind-of"

    /** The key jetbrains.projectactions.monosmith-make. Values: {0} is name, {1} is projectActionArgumentsAlgorithm. */
    const val JETBRAINS_PROJECTACTIONS_MONOSMITH_MAKE = "jetbrains.projectactions.monosmith-make"

    /** The key jetbrains.projectactions.monosmith-make-a. Values: {0} is name. */
    const val JETBRAINS_PROJECTACTIONS_MONOSMITH_MAKE_A = "jetbrains.projectactions.monosmith-make-a"

    /** The key jetbrains.projectactions.monosmith-remove. Values: {0} is name. */
    const val JETBRAINS_PROJECTACTIONS_MONOSMITH_REMOVE = "jetbrains.projectactions.monosmith-remove"

    /** The key jetbrains.projectactions.monosmith-remove-a-project. */
    const val JETBRAINS_PROJECTACTIONS_MONOSMITH_REMOVE_A_PROJECT = "jetbrains.projectactions.monosmith-remove-a-project"

    /** The key jetbrains.projectactions.monosmith-unlink-a-library. */
    const val JETBRAINS_PROJECTACTIONS_MONOSMITH_UNLINK_A_LIBRARY = "jetbrains.projectactions.monosmith-unlink-a-library"

    /** The key jetbrains.projectactions.monosmith-unlink-a-library-from. Values: {0} is name. */
    const val JETBRAINS_PROJECTACTIONS_MONOSMITH_UNLINK_A_LIBRARY_FROM = "jetbrains.projectactions.monosmith-unlink-a-library-from"

    /** The key jetbrains.projectactions.monosmith-unlink-from. Values: {0} is library, {1} is name. */
    const val JETBRAINS_PROJECTACTIONS_MONOSMITH_UNLINK_FROM = "jetbrains.projectactions.monosmith-unlink-from"

    /** The key jetbrains.projectactions.nothing-was-changed. Values: {0} is listed. */
    const val JETBRAINS_PROJECTACTIONS_NOTHING_WAS_CHANGED = "jetbrains.projectactions.nothing-was-changed"

    /** The key jetbrains.projectactions.nothing-was-changed-yet-smith. */
    const val JETBRAINS_PROJECTACTIONS_NOTHING_WAS_CHANGED_YET_SMITH = "jetbrains.projectactions.nothing-was-changed-yet-smith"

    /** The key jetbrains.projectactions.remove-a-project. */
    const val JETBRAINS_PROJECTACTIONS_REMOVE_A_PROJECT = "jetbrains.projectactions.remove-a-project"

    /** The key jetbrains.projectactions.remove-it-and-delete-the. */
    const val JETBRAINS_PROJECTACTIONS_REMOVE_IT_AND_DELETE_THE = "jetbrains.projectactions.remove-it-and-delete-the"

    /** The key jetbrains.projectactions.remove-it-and-keep-a. */
    const val JETBRAINS_PROJECTACTIONS_REMOVE_IT_AND_KEEP_A = "jetbrains.projectactions.remove-it-and-keep-a"

    /** The key jetbrains.projectactions.stops-depending-on-the-library. Values: {0} is name, {1} is library. */
    const val JETBRAINS_PROJECTACTIONS_STOPS_DEPENDING_ON_THE_LIBRARY = "jetbrains.projectactions.stops-depending-on-the-library"

    /** The key jetbrains.projectactions.unlink-a-library. */
    const val JETBRAINS_PROJECTACTIONS_UNLINK_A_LIBRARY = "jetbrains.projectactions.unlink-a-library"

    /** The key jetbrains.projectactions.unlink-it. */
    const val JETBRAINS_PROJECTACTIONS_UNLINK_IT = "jetbrains.projectactions.unlink-it"

    /** The key jetbrains.projectactions.which-library-should-the-project. */
    const val JETBRAINS_PROJECTACTIONS_WHICH_LIBRARY_SHOULD_THE_PROJECT = "jetbrains.projectactions.which-library-should-the-project"

    /** The key jetbrains.readcommands.monosmith. Values: {0} is what. */
    const val JETBRAINS_READCOMMANDS_MONOSMITH = "jetbrains.readcommands.monosmith"

    /** The key jetbrains.readcommands.monosmith-why-would-a-task. */
    const val JETBRAINS_READCOMMANDS_MONOSMITH_WHY_WOULD_A_TASK = "jetbrains.readcommands.monosmith-why-would-a-task"

    /** The key jetbrains.readcommands.monosmith-why-would-a-task-2. */
    const val JETBRAINS_READCOMMANDS_MONOSMITH_WHY_WOULD_A_TASK_2 = "jetbrains.readcommands.monosmith-why-would-a-task-2"

    /** The key jetbrains.readcommands.smith-could-not-show. Values: {0} is what, {1} is failureTextMapper. */
    const val JETBRAINS_READCOMMANDS_SMITH_COULD_NOT_SHOW = "jetbrains.readcommands.smith-could-not-show"

    /** The key jetbrains.readcommands.the-project-and-the-task. */
    const val JETBRAINS_READCOMMANDS_THE_PROJECT_AND_THE_TASK = "jetbrains.readcommands.the-project-and-the-task"

    /** The key jetbrains.readcommands.why-would-run-again. Values: {0} is target. */
    const val JETBRAINS_READCOMMANDS_WHY_WOULD_RUN_AGAIN = "jetbrains.readcommands.why-would-run-again"

    /** The key jetbrains.readcommands.write-project-task-for-example. */
    const val JETBRAINS_READCOMMANDS_WRITE_PROJECT_TASK_FOR_EXAMPLE = "jetbrains.readcommands.write-project-task-for-example"

    /** The key jetbrains.releaseflow.a-breaking-change-1-2. */
    const val JETBRAINS_RELEASEFLOW_A_BREAKING_CHANGE_1_2 = "jetbrains.releaseflow.a-breaking-change-1-2"

    /** The key jetbrains.releaseflow.a-fix-1-2-3. */
    const val JETBRAINS_RELEASEFLOW_A_FIX_1_2_3 = "jetbrains.releaseflow.a-fix-1-2-3"

    /** The key jetbrains.releaseflow.a-new-feature-1-2. */
    const val JETBRAINS_RELEASEFLOW_A_NEW_FEATURE_1_2 = "jetbrains.releaseflow.a-new-feature-1-2"

    /** The key jetbrains.releaseflow.a-package-releases-when-a. */
    const val JETBRAINS_RELEASEFLOW_A_PACKAGE_RELEASES_WHEN_A = "jetbrains.releaseflow.a-package-releases-when-a"

    /** The key jetbrains.releaseflow.change-a-bump. Values: {0} is title. */
    const val JETBRAINS_RELEASEFLOW_CHANGE_A_BUMP = "jetbrains.releaseflow.change-a-bump"

    /** The key jetbrains.releaseflow.change-its-bump-a-request. */
    const val JETBRAINS_RELEASEFLOW_CHANGE_ITS_BUMP_A_REQUEST = "jetbrains.releaseflow.change-its-bump-a-request"

    /** The key jetbrains.releaseflow.changed-bumps. Values: {0} is bump. */
    const val JETBRAINS_RELEASEFLOW_CHANGED_BUMPS = "jetbrains.releaseflow.changed-bumps"

    /** The key jetbrains.releaseflow.channel-its-versions-are-prereleases. Values: {0} is channel. */
    const val JETBRAINS_RELEASEFLOW_CHANNEL_ITS_VERSIONS_ARE_PRERELEASES = "jetbrains.releaseflow.channel-its-versions-are-prereleases"

    /** The key jetbrains.releaseflow.channel-stable. */
    const val JETBRAINS_RELEASEFLOW_CHANNEL_STABLE = "jetbrains.releaseflow.channel-stable"

    /** The key jetbrains.releaseflow.choose-a-package-to-release. */
    const val JETBRAINS_RELEASEFLOW_CHOOSE_A_PACKAGE_TO_RELEASE = "jetbrains.releaseflow.choose-a-package-to-release"

    /** The key jetbrains.releaseflow.held-back. Values: {0} is hold. */
    const val JETBRAINS_RELEASEFLOW_HELD_BACK = "jetbrains.releaseflow.held-back"

    /** The key jetbrains.releaseflow.hold-it-back-it-does. */
    const val JETBRAINS_RELEASEFLOW_HOLD_IT_BACK_IT_DOES = "jetbrains.releaseflow.hold-it-back-it-does"

    /** The key jetbrains.releaseflow.hold-packages-back. Values: {0} is title. */
    const val JETBRAINS_RELEASEFLOW_HOLD_PACKAGES_BACK = "jetbrains.releaseflow.hold-packages-back"

    /** The key jetbrains.releaseflow.keep-the-bumps-as-they. */
    const val JETBRAINS_RELEASEFLOW_KEEP_THE_BUMPS_AS_THEY = "jetbrains.releaseflow.keep-the-bumps-as-they"

    /** The key jetbrains.releaseflow.monosmith-release. */
    const val JETBRAINS_RELEASEFLOW_MONOSMITH_RELEASE = "jetbrains.releaseflow.monosmith-release"

    /** The key jetbrains.releaseflow.monosmith-release-preview. */
    const val JETBRAINS_RELEASEFLOW_MONOSMITH_RELEASE_PREVIEW = "jetbrains.releaseflow.monosmith-release-preview"

    /** The key jetbrains.releaseflow.no-commit-releases. Values: {0} is unchanged. */
    const val JETBRAINS_RELEASEFLOW_NO_COMMIT_RELEASES = "jetbrains.releaseflow.no-commit-releases"

    /** The key jetbrains.releaseflow.no-commit-releases-anything-now. */
    const val JETBRAINS_RELEASEFLOW_NO_COMMIT_RELEASES_ANYTHING_NOW = "jetbrains.releaseflow.no-commit-releases-anything-now"

    /** The key jetbrains.releaseflow.no-commit-releases-it. */
    const val JETBRAINS_RELEASEFLOW_NO_COMMIT_RELEASES_IT = "jetbrains.releaseflow.no-commit-releases-it"

    /** The key jetbrains.releaseflow.no-commit-releases-it-release. Values: {0} is it. */
    const val JETBRAINS_RELEASEFLOW_NO_COMMIT_RELEASES_IT_RELEASE = "jetbrains.releaseflow.no-commit-releases-it-release"

    /** The key jetbrains.releaseflow.no-commit-releases-it-the. Values: {0} is names. */
    const val JETBRAINS_RELEASEFLOW_NO_COMMIT_RELEASES_IT_THE = "jetbrains.releaseflow.no-commit-releases-it-the"

    /** The key jetbrains.releaseflow.no-commit-releases-these-packages. */
    const val JETBRAINS_RELEASEFLOW_NO_COMMIT_RELEASES_THESE_PACKAGES = "jetbrains.releaseflow.no-commit-releases-these-packages"

    /** The key jetbrains.releaseflow.nothing-is-published-yet-next. */
    const val JETBRAINS_RELEASEFLOW_NOTHING_IS_PUBLISHED_YET_NEXT = "jetbrains.releaseflow.nothing-is-published-yet-next"

    /** The key jetbrains.releaseflow.nothing-to-release. */
    const val JETBRAINS_RELEASEFLOW_NOTHING_TO_RELEASE = "jetbrains.releaseflow.nothing-to-release"

    /** The key jetbrains.releaseflow.nothing-to-release-a-package. */
    const val JETBRAINS_RELEASEFLOW_NOTHING_TO_RELEASE_A_PACKAGE = "jetbrains.releaseflow.nothing-to-release-a-package"

    /** The key jetbrains.releaseflow.nothing-was-released. Values: {0} is failureTextMapper. */
    const val JETBRAINS_RELEASEFLOW_NOTHING_WAS_RELEASED = "jetbrains.releaseflow.nothing-was-released"

    /** The key jetbrains.releaseflow.open-preview. */
    const val JETBRAINS_RELEASEFLOW_OPEN_PREVIEW = "jetbrains.releaseflow.open-preview"

    /** The key jetbrains.releaseflow.pick-a-package-to-change. */
    const val JETBRAINS_RELEASEFLOW_PICK_A_PACKAGE_TO_CHANGE = "jetbrains.releaseflow.pick-a-package-to-change"

    /** The key jetbrains.releaseflow.release-a-package-without-a. Values: {0} is title. */
    const val JETBRAINS_RELEASEFLOW_RELEASE_A_PACKAGE_WITHOUT_A = "jetbrains.releaseflow.release-a-package-without-a"

    /** The key jetbrains.releaseflow.release-it-anyway-the-same. */
    const val JETBRAINS_RELEASEFLOW_RELEASE_IT_ANYWAY_THE_SAME = "jetbrains.releaseflow.release-it-anyway-the-same"

    /** The key jetbrains.releaseflow.release-now. Values: {0} is title. */
    const val JETBRAINS_RELEASEFLOW_RELEASE_NOW = "jetbrains.releaseflow.release-now"

    /** The key jetbrains.releaseflow.release-now-2. */
    const val JETBRAINS_RELEASEFLOW_RELEASE_NOW_2 = "jetbrains.releaseflow.release-now-2"

    /** The key jetbrains.releaseflow.show-the-release-preview-it. */
    const val JETBRAINS_RELEASEFLOW_SHOW_THE_RELEASE_PREVIEW_IT = "jetbrains.releaseflow.show-the-release-preview-it"

    /** The key jetbrains.releaseflow.smith-printed-the-release-plan. */
    const val JETBRAINS_RELEASEFLOW_SMITH_PRINTED_THE_RELEASE_PLAN = "jetbrains.releaseflow.smith-printed-the-release-plan"

    /** The key jetbrains.releaseflow.the-commits-decide-the-levels. */
    const val JETBRAINS_RELEASEFLOW_THE_COMMITS_DECIDE_THE_LEVELS = "jetbrains.releaseflow.the-commits-decide-the-levels"

    /** The key jetbrains.releaseflow.the-plan-stopped. Values: {0} is error. */
    const val JETBRAINS_RELEASEFLOW_THE_PLAN_STOPPED = "jetbrains.releaseflow.the-plan-stopped"

    /** The key jetbrains.releaseflow.the-release-is-done. */
    const val JETBRAINS_RELEASEFLOW_THE_RELEASE_IS_DONE = "jetbrains.releaseflow.the-release-is-done"

    /** The key jetbrains.releaseflow.this-cannot-be-undone-smith. */
    const val JETBRAINS_RELEASEFLOW_THIS_CANNOT_BE_UNDONE_SMITH = "jetbrains.releaseflow.this-cannot-be-undone-smith"

    /** The key jetbrains.releaseflow.this-is-what-a-release. Values: {0} is title. */
    const val JETBRAINS_RELEASEFLOW_THIS_IS_WHAT_A_RELEASE = "jetbrains.releaseflow.this-is-what-a-release"

    /** The key jetbrains.releaseflow.tick-the-packages-that-must. */
    const val JETBRAINS_RELEASEFLOW_TICK_THE_PACKAGES_THAT_MUST = "jetbrains.releaseflow.tick-the-packages-that-must"

    /** The key jetbrains.releaseflow.which-level-for. Values: {0} is title, {1} is name. */
    const val JETBRAINS_RELEASEFLOW_WHICH_LEVEL_FOR = "jetbrains.releaseflow.which-level-for"

    /** The key jetbrains.releaseflow.without-a-commit. */
    const val JETBRAINS_RELEASEFLOW_WITHOUT_A_COMMIT = "jetbrains.releaseflow.without-a-commit"

    /** The key jetbrains.runacross.did-not-finish. Values: {0} is task, {1} is failureTextMapper. */
    const val JETBRAINS_RUNACROSS_DID_NOT_FINISH = "jetbrains.runacross.did-not-finish"

    /** The key jetbrains.runacross.is-done. Values: {0} is task. */
    const val JETBRAINS_RUNACROSS_IS_DONE = "jetbrains.runacross.is-done"

    /** The key jetbrains.runacross.monosmith-run-affected. */
    const val JETBRAINS_RUNACROSS_MONOSMITH_RUN_AFFECTED = "jetbrains.runacross.monosmith-run-affected"

    /** The key jetbrains.runacross.monosmith-run-in-all-projects. */
    const val JETBRAINS_RUNACROSS_MONOSMITH_RUN_IN_ALL_PROJECTS = "jetbrains.runacross.monosmith-run-in-all-projects"

    /** The key jetbrains.runacross.no-project-has-a-task. */
    const val JETBRAINS_RUNACROSS_NO_PROJECT_HAS_A_TASK = "jetbrains.runacross.no-project-has-a-task"

    /** The key jetbrains.runacross.no-project-is-affected-by. */
    const val JETBRAINS_RUNACROSS_NO_PROJECT_IS_AFFECTED_BY = "jetbrains.runacross.no-project-is-affected-by"

    /** The key jetbrains.runacross.nothing-ran. Values: {0} is message. */
    const val JETBRAINS_RUNACROSS_NOTHING_RAN = "jetbrains.runacross.nothing-ran"

    /** The key jetbrains.runacross.nothing-ran-2. Values: {0} is failureTextMapper. */
    const val JETBRAINS_RUNACROSS_NOTHING_RAN_2 = "jetbrains.runacross.nothing-ran-2"

    /** The key jetbrains.runacross.pick-the-task-to-run. */
    const val JETBRAINS_RUNACROSS_PICK_THE_TASK_TO_RUN = "jetbrains.runacross.pick-the-task-to-run"

    /** The key jetbrains.runacross.pick-the-task-to-run-2. */
    const val JETBRAINS_RUNACROSS_PICK_THE_TASK_TO_RUN_2 = "jetbrains.runacross.pick-the-task-to-run-2"

    /** The key jetbrains.runacross.smith-printed-the-affected-projects. */
    const val JETBRAINS_RUNACROSS_SMITH_PRINTED_THE_AFFECTED_PROJECTS = "jetbrains.runacross.smith-printed-the-affected-projects"

    /** The key jetbrains.settings.change-it. */
    const val JETBRAINS_SETTINGS_CHANGE_IT = "jetbrains.settings.change-it"

    /** The key jetbrains.settings.go-back-to-the-default. */
    const val JETBRAINS_SETTINGS_GO_BACK_TO_THE_DEFAULT = "jetbrains.settings.go-back-to-the-default"

    /** The key jetbrains.settings.it-is-set-to. Values: {0} is value, {1} is settingsJsonMapper. */
    const val JETBRAINS_SETTINGS_IT_IS_SET_TO = "jetbrains.settings.it-is-set-to"

    /** The key jetbrains.settings.monosmith-settings. */
    const val JETBRAINS_SETTINGS_MONOSMITH_SETTINGS = "jetbrains.settings.monosmith-settings"

    /** The key jetbrains.settings.no-setting-was-changed. Values: {0} is failureTextMapper. */
    const val JETBRAINS_SETTINGS_NO_SETTING_WAS_CHANGED = "jetbrains.settings.no-setting-was-changed"

    /** The key jetbrains.settings.no-setting-was-changed-this. Values: {0} is path. */
    const val JETBRAINS_SETTINGS_NO_SETTING_WAS_CHANGED_THIS = "jetbrains.settings.no-setting-was-changed-this"

    /** The key jetbrains.settings.nothing-was-changed-yet-smith. Values: {0} is value. */
    const val JETBRAINS_SETTINGS_NOTHING_WAS_CHANGED_YET_SMITH = "jetbrains.settings.nothing-was-changed-yet-smith"

    /** The key jetbrains.settings.pick-a-setting-type-in. */
    const val JETBRAINS_SETTINGS_PICK_A_SETTING_TYPE_IN = "jetbrains.settings.pick-a-setting-type-in"

    /** The key jetbrains.settings.remove-the-line-from-smith. */
    const val JETBRAINS_SETTINGS_REMOVE_THE_LINE_FROM_SMITH = "jetbrains.settings.remove-the-line-from-smith"

    /** The key jetbrains.settings.reset-it. */
    const val JETBRAINS_SETTINGS_RESET_IT = "jetbrains.settings.reset-it"

    /** The key jetbrains.settings.smith-printed-its-settings-in. */
    const val JETBRAINS_SETTINGS_SMITH_PRINTED_ITS_SETTINGS_IN = "jetbrains.settings.smith-printed-its-settings-in"

    /** The key jetbrains.settings.the-current-value-is-first. Values: {0} is words. */
    const val JETBRAINS_SETTINGS_THE_CURRENT_VALUE_IS_FIRST = "jetbrains.settings.the-current-value-is-first"

    /** The key jetbrains.settings.type-a-value. */
    const val JETBRAINS_SETTINGS_TYPE_A_VALUE = "jetbrains.settings.type-a-value"

    /** The key jetbrains.settings.why-the-default. Values: {0} is reason. */
    const val JETBRAINS_SETTINGS_WHY_THE_DEFAULT = "jetbrains.settings.why-the-default"

    /** The key jetbrains.settings.write-a-new-value-in. */
    const val JETBRAINS_SETTINGS_WRITE_A_NEW_VALUE_IN = "jetbrains.settings.write-a-new-value-in"

    /** The key jetbrains.smithaction.is-outside-the-monosmith-workspace. Values: {0} is name, {1} is root. */
    const val JETBRAINS_SMITHACTION_IS_OUTSIDE_THE_MONOSMITH_WORKSPACE = "jetbrains.smithaction.is-outside-the-monosmith-workspace"

    /** The key jetbrains.smithaction.monosmith. */
    const val JETBRAINS_SMITHACTION_MONOSMITH = "jetbrains.smithaction.monosmith"

    /** The key jetbrains.smithaction.the-cache-check. */
    const val JETBRAINS_SMITHACTION_THE_CACHE_CHECK = "jetbrains.smithaction.the-cache-check"

    /** The key jetbrains.smithaction.the-cache-status. */
    const val JETBRAINS_SMITHACTION_THE_CACHE_STATUS = "jetbrains.smithaction.the-cache-status"

    /** The key jetbrains.smithaction.the-copies-between-projects. */
    const val JETBRAINS_SMITHACTION_THE_COPIES_BETWEEN_PROJECTS = "jetbrains.smithaction.the-copies-between-projects"

    /** The key jetbrains.smithaction.the-smith-info. */
    const val JETBRAINS_SMITHACTION_THE_SMITH_INFO = "jetbrains.smithaction.the-smith-info"

    /** The key jetbrains.smithaction.the-tools. */
    const val JETBRAINS_SMITHACTION_THE_TOOLS = "jetbrains.smithaction.the-tools"

    /** The key jetbrains.smithaction.there-is-no-smith-toml. */
    const val JETBRAINS_SMITHACTION_THERE_IS_NO_SMITH_TOML = "jetbrains.smithaction.there-is-no-smith-toml"

    /** The key jetbrains.smithrun.nothing-was-changed. Values: {0} is value. */
    const val JETBRAINS_SMITHRUN_NOTHING_WAS_CHANGED = "jetbrains.smithrun.nothing-was-changed"

    /** The key jetbrains.smithrun.smith-stopped-with-exit-code. Values: {0} is code. */
    const val JETBRAINS_SMITHRUN_SMITH_STOPPED_WITH_EXIT_CODE = "jetbrains.smithrun.smith-stopped-with-exit-code"

    /** The key jetbrains.smithversion.copy-command. */
    const val JETBRAINS_SMITHVERSION_COPY_COMMAND = "jetbrains.smithversion.copy-command"

    /** The key jetbrains.smithversion.copy-to-the-clipboard. Values: {0} is installcommand. */
    const val JETBRAINS_SMITHVERSION_COPY_TO_THE_CLIPBOARD = "jetbrains.smithversion.copy-to-the-clipboard"

    /** The key jetbrains.smithversion.monosmith-needs-smith-to-list. Values: {0} is problem, {1} is installcommand. */
    const val JETBRAINS_SMITHVERSION_MONOSMITH_NEEDS_SMITH_TO_LIST = "jetbrains.smithversion.monosmith-needs-smith-to-list"

    /** The key jetbrains.smithversion.smith-answered-with-an-error. Values: {0} is stderr. */
    const val JETBRAINS_SMITHVERSION_SMITH_ANSWERED_WITH_AN_ERROR = "jetbrains.smithversion.smith-answered-with-an-error"

    /** The key jetbrains.smithversion.smith-is-not-installed-here. */
    const val JETBRAINS_SMITHVERSION_SMITH_IS_NOT_INSTALLED_HERE = "jetbrains.smithversion.smith-is-not-installed-here"

    /** The key jetbrains.swapchoice.has-no-choice-to-swap. Values: {0} is name. */
    const val JETBRAINS_SWAPCHOICE_HAS_NO_CHOICE_TO_SWAP = "jetbrains.swapchoice.has-no-choice-to-swap"

    /** The key jetbrains.swapchoice.monosmith-swap-a-choice. */
    const val JETBRAINS_SWAPCHOICE_MONOSMITH_SWAP_A_CHOICE = "jetbrains.swapchoice.monosmith-swap-a-choice"

    /** The key jetbrains.swapchoice.monosmith-swap-a-choice-of. Values: {0} is name. */
    const val JETBRAINS_SWAPCHOICE_MONOSMITH_SWAP_A_CHOICE_OF = "jetbrains.swapchoice.monosmith-swap-a-choice-of"

    /** The key jetbrains.swapchoice.monosmith-swap-the-of. Values: {0} is axis, {1} is name. */
    const val JETBRAINS_SWAPCHOICE_MONOSMITH_SWAP_THE_OF = "jetbrains.swapchoice.monosmith-swap-the-of"

    /** The key jetbrains.swapchoice.monosmith-swap-the-of-to. Values: {0} is axis, {1} is name, {2} is choice. */
    const val JETBRAINS_SWAPCHOICE_MONOSMITH_SWAP_THE_OF_TO = "jetbrains.swapchoice.monosmith-swap-the-of-to"

    /** The key jetbrains.swapchoice.nothing-was-changed. Values: {0} is message. */
    const val JETBRAINS_SWAPCHOICE_NOTHING_WAS_CHANGED = "jetbrains.swapchoice.nothing-was-changed"

    /** The key jetbrains.swapchoice.nothing-was-changed-yet-by. */
    const val JETBRAINS_SWAPCHOICE_NOTHING_WAS_CHANGED_YET_BY = "jetbrains.swapchoice.nothing-was-changed-yet-by"

    /** The key jetbrains.swapchoice.one-answer. */
    const val JETBRAINS_SWAPCHOICE_ONE_ANSWER = "jetbrains.swapchoice.one-answer"

    /** The key jetbrains.swapchoice.pick-the-answer-gets-the. Values: {0} is name. */
    const val JETBRAINS_SWAPCHOICE_PICK_THE_ANSWER_GETS_THE = "jetbrains.swapchoice.pick-the-answer-gets-the"

    /** The key jetbrains.swapchoice.several-can-be-used-at. */
    const val JETBRAINS_SWAPCHOICE_SEVERAL_CAN_BE_USED_AT = "jetbrains.swapchoice.several-can-be-used-at"

    /** The key jetbrains.swapchoice.smith-printed-its-choices-in. */
    const val JETBRAINS_SWAPCHOICE_SMITH_PRINTED_ITS_CHOICES_IN = "jetbrains.swapchoice.smith-printed-its-choices-in"

    /** The key jetbrains.swapchoice.swap. */
    const val JETBRAINS_SWAPCHOICE_SWAP = "jetbrains.swapchoice.swap"

    /** The key jetbrains.swapchoice.swap-a-choice. */
    const val JETBRAINS_SWAPCHOICE_SWAP_A_CHOICE = "jetbrains.swapchoice.swap-a-choice"

    /** The key jetbrains.swapchoice.the-of-is-now. Values: {0} is axis, {1} is name, {2} is choice. */
    const val JETBRAINS_SWAPCHOICE_THE_OF_IS_NOW = "jetbrains.swapchoice.the-of-is-now"

    /** The key jetbrains.swapchoice.which-decision-do-you-want. */
    const val JETBRAINS_SWAPCHOICE_WHICH_DECISION_DO_YOU_WANT = "jetbrains.swapchoice.which-decision-do-you-want"

    /** The key jetbrains.syncversions.monosmith-sync-dependency-versions. */
    const val JETBRAINS_SYNCVERSIONS_MONOSMITH_SYNC_DEPENDENCY_VERSIONS = "jetbrains.syncversions.monosmith-sync-dependency-versions"

    /** The key jetbrains.syncversions.monosmith-sync-dependency-versions-2. Values: {0} is rule. */
    const val JETBRAINS_SYNCVERSIONS_MONOSMITH_SYNC_DEPENDENCY_VERSIONS_2 = "jetbrains.syncversions.monosmith-sync-dependency-versions-2"

    /** The key jetbrains.syncversions.monosmith-which-version-wins. */
    const val JETBRAINS_SYNCVERSIONS_MONOSMITH_WHICH_VERSION_WINS = "jetbrains.syncversions.monosmith-which-version-wins"

    /** The key jetbrains.syncversions.nothing-was-changed-yet-this. */
    const val JETBRAINS_SYNCVERSIONS_NOTHING_WAS_CHANGED_YET_THIS = "jetbrains.syncversions.nothing-was-changed-yet-this"

    /** The key jetbrains.syncversions.the-versions-are-in-sync. */
    const val JETBRAINS_SYNCVERSIONS_THE_VERSIONS_ARE_IN_SYNC = "jetbrains.syncversions.the-versions-are-in-sync"

    /** The key jetbrains.syncversions.when-projects-declare-a-dependency. */
    const val JETBRAINS_SYNCVERSIONS_WHEN_PROJECTS_DECLARE_A_DEPENDENCY = "jetbrains.syncversions.when-projects-declare-a-dependency"

    /** The key jetbrains.syncversions.write-it. */
    const val JETBRAINS_SYNCVERSIONS_WRITE_IT = "jetbrains.syncversions.write-it"

    /** The key jetbrains.tabcompletion.dont-ask-again. */
    const val JETBRAINS_TABCOMPLETION_DONT_ASK_AGAIN = "jetbrains.tabcompletion.dont-ask-again"

    /** The key jetbrains.tabcompletion.not-now. */
    const val JETBRAINS_TABCOMPLETION_NOT_NOW = "jetbrains.tabcompletion.not-now"

    /** The key jetbrains.tabcompletion.offer. */
    const val JETBRAINS_TABCOMPLETION_OFFER = "jetbrains.tabcompletion.offer"

    /** The key jetbrains.tabcompletion.set-up. */
    const val JETBRAINS_TABCOMPLETION_SET_UP = "jetbrains.tabcompletion.set-up"

    /** The key jetbrains.tauriapp.add-tauri. Values: {0} is title. */
    const val JETBRAINS_TAURIAPP_ADD_TAURI = "jetbrains.tauriapp.add-tauri"

    /** The key jetbrains.tauriapp.add-tauri-2. */
    const val JETBRAINS_TAURIAPP_ADD_TAURI_2 = "jetbrains.tauriapp.add-tauri-2"

    /** The key jetbrains.tauriapp.add-tauri-to. Values: {0} is title, {1} is app. */
    const val JETBRAINS_TAURIAPP_ADD_TAURI_TO = "jetbrains.tauriapp.add-tauri-to"

    /** The key jetbrains.tauriapp.is-a-tauri-app. Values: {0} is app. */
    const val JETBRAINS_TAURIAPP_IS_A_TAURI_APP = "jetbrains.tauriapp.is-a-tauri-app"

    /** The key jetbrains.tauriapp.make-a-tauri-app. Values: {0} is title, {1} is app. */
    const val JETBRAINS_TAURIAPP_MAKE_A_TAURI_APP = "jetbrains.tauriapp.make-a-tauri-app"

    /** The key jetbrains.tauriapp.make-the-icons. Values: {0} is title. */
    const val JETBRAINS_TAURIAPP_MAKE_THE_ICONS = "jetbrains.tauriapp.make-the-icons"

    /** The key jetbrains.tauriapp.make-the-icons-2. */
    const val JETBRAINS_TAURIAPP_MAKE_THE_ICONS_2 = "jetbrains.tauriapp.make-the-icons-2"

    /** The key jetbrains.tauriapp.monosmith-tauri. */
    const val JETBRAINS_TAURIAPP_MONOSMITH_TAURI = "jetbrains.tauriapp.monosmith-tauri"

    /** The key jetbrains.tauriapp.name-a-png-file. */
    const val JETBRAINS_TAURIAPP_NAME_A_PNG_FILE = "jetbrains.tauriapp.name-a-png-file"

    /** The key jetbrains.tauriapp.smith-overwrites-the-icons-in. */
    const val JETBRAINS_TAURIAPP_SMITH_OVERWRITES_THE_ICONS_IN = "jetbrains.tauriapp.smith-overwrites-the-icons-in"

    /** The key jetbrains.tauriapp.smith-writes-src-tauri-and. */
    const val JETBRAINS_TAURIAPP_SMITH_WRITES_SRC_TAURI_AND = "jetbrains.tauriapp.smith-writes-src-tauri-and"

    /** The key jetbrains.tauriapp.the-app-only-when-more. */
    const val JETBRAINS_TAURIAPP_THE_APP_ONLY_WHEN_MORE = "jetbrains.tauriapp.the-app-only-when-more"

    /** The key jetbrains.tauriapp.the-bundle-identifier-like-com. */
    const val JETBRAINS_TAURIAPP_THE_BUNDLE_IDENTIFIER_LIKE_COM = "jetbrains.tauriapp.the-bundle-identifier-like-com"

    /** The key jetbrains.tauriapp.the-icons. Values: {0} is title. */
    const val JETBRAINS_TAURIAPP_THE_ICONS = "jetbrains.tauriapp.the-icons"

    /** The key jetbrains.tauriapp.the-icons-are-made. */
    const val JETBRAINS_TAURIAPP_THE_ICONS_ARE_MADE = "jetbrains.tauriapp.the-icons-are-made"

    /** The key jetbrains.tauriapp.the-node-app-to-make. */
    const val JETBRAINS_TAURIAPP_THE_NODE_APP_TO_MAKE = "jetbrains.tauriapp.the-node-app-to-make"

    /** The key jetbrains.tauriapp.the-square-png-the-icons. */
    const val JETBRAINS_TAURIAPP_THE_SQUARE_PNG_THE_ICONS = "jetbrains.tauriapp.the-square-png-the-icons"

    /** The key jetbrains.toolwindow.add-library. */
    const val JETBRAINS_TOOLWINDOW_ADD_LIBRARY = "jetbrains.toolwindow.add-library"

    /** The key jetbrains.toolwindow.add-your-first-project. */
    const val JETBRAINS_TOOLWINDOW_ADD_YOUR_FIRST_PROJECT = "jetbrains.toolwindow.add-your-first-project"

    /** The key jetbrains.toolwindow.affected-only. */
    const val JETBRAINS_TOOLWINDOW_AFFECTED_ONLY = "jetbrains.toolwindow.affected-only"

    /** The key jetbrains.toolwindow.align-versions. */
    const val JETBRAINS_TOOLWINDOW_ALIGN_VERSIONS = "jetbrains.toolwindow.align-versions"

    /** The key jetbrains.toolwindow.fix-this-check. Values: {0} is description. */
    const val JETBRAINS_TOOLWINDOW_FIX_THIS_CHECK = "jetbrains.toolwindow.fix-this-check"

    /** The key jetbrains.toolwindow.install-package. */
    const val JETBRAINS_TOOLWINDOW_INSTALL_PACKAGE = "jetbrains.toolwindow.install-package"

    /** The key jetbrains.toolwindow.make-app-library-or-package. */
    const val JETBRAINS_TOOLWINDOW_MAKE_APP_LIBRARY_OR_PACKAGE = "jetbrains.toolwindow.make-app-library-or-package"

    /** The key jetbrains.toolwindow.new-project. */
    const val JETBRAINS_TOOLWINDOW_NEW_PROJECT = "jetbrains.toolwindow.new-project"

    /** The key jetbrains.toolwindow.no-project-matches-the-filter. */
    const val JETBRAINS_TOOLWINDOW_NO_PROJECT_MATCHES_THE_FILTER = "jetbrains.toolwindow.no-project-matches-the-filter"

    /** The key jetbrains.toolwindow.no-smith-toml-in-this. */
    const val JETBRAINS_TOOLWINDOW_NO_SMITH_TOML_IN_THIS = "jetbrains.toolwindow.no-smith-toml-in-this"

    /** The key jetbrains.toolwindow.read-the-projects-the-updates. */
    const val JETBRAINS_TOOLWINDOW_READ_THE_PROJECTS_THE_UPDATES = "jetbrains.toolwindow.read-the-projects-the-updates"

    /** The key jetbrains.toolwindow.reading-the-release-plan. */
    const val JETBRAINS_TOOLWINDOW_READING_THE_RELEASE_PLAN = "jetbrains.toolwindow.reading-the-release-plan"

    /** The key jetbrains.toolwindow.reading-the-updates. */
    const val JETBRAINS_TOOLWINDOW_READING_THE_UPDATES = "jetbrains.toolwindow.reading-the-updates"

    /** The key jetbrains.toolwindow.refresh. */
    const val JETBRAINS_TOOLWINDOW_REFRESH = "jetbrains.toolwindow.refresh"

    /** The key jetbrains.toolwindow.release-hold-bump-force. */
    const val JETBRAINS_TOOLWINDOW_RELEASE_HOLD_BUMP_FORCE = "jetbrains.toolwindow.release-hold-bump-force"

    /** The key jetbrains.toolwindow.run-task. */
    const val JETBRAINS_TOOLWINDOW_RUN_TASK = "jetbrains.toolwindow.run-task"

    /** The key jetbrains.toolwindow.running-doctor. */
    const val JETBRAINS_TOOLWINDOW_RUNNING_DOCTOR = "jetbrains.toolwindow.running-doctor"

    /** The key jetbrains.toolwindow.smith-failed. Values: {0} is message. */
    const val JETBRAINS_TOOLWINDOW_SMITH_FAILED = "jetbrains.toolwindow.smith-failed"

    /** The key jetbrains.toolwindow.tools-with-a-newer-version. Values: {0} is size. */
    const val JETBRAINS_TOOLWINDOW_TOOLS_WITH_A_NEWER_VERSION = "jetbrains.toolwindow.tools-with-a-newer-version"

    /** The key jetbrains.toolwindow.unlink-library. */
    const val JETBRAINS_TOOLWINDOW_UNLINK_LIBRARY = "jetbrains.toolwindow.unlink-library"

    /** The key jetbrains.toolwindow.update-dependencies. */
    const val JETBRAINS_TOOLWINDOW_UPDATE_DEPENDENCIES = "jetbrains.toolwindow.update-dependencies"

    /** The key jetbrains.trycommand.files-it-changed. */
    const val JETBRAINS_TRYCOMMAND_FILES_IT_CHANGED = "jetbrains.trycommand.files-it-changed"

    /** The key jetbrains.trycommand.it-failed-in-the-copy. */
    const val JETBRAINS_TRYCOMMAND_IT_FAILED_IN_THE_COPY = "jetbrains.trycommand.it-failed-in-the-copy"

    /** The key jetbrains.trycommand.it-passed-in-a-copy. */
    const val JETBRAINS_TRYCOMMAND_IT_PASSED_IN_A_COPY = "jetbrains.trycommand.it-passed-in-a-copy"

    /** The key jetbrains.trycommand.monosmith-try-a-command. */
    const val JETBRAINS_TRYCOMMAND_MONOSMITH_TRY_A_COMMAND = "jetbrains.trycommand.monosmith-try-a-command"

    /** The key jetbrains.trycommand.monosmith-try-a-command-in. */
    const val JETBRAINS_TRYCOMMAND_MONOSMITH_TRY_A_COMMAND_IN = "jetbrains.trycommand.monosmith-try-a-command-in"

    /** The key jetbrains.trycommand.monosmith-try-smith. Values: {0} is command. */
    const val JETBRAINS_TRYCOMMAND_MONOSMITH_TRY_SMITH = "jetbrains.trycommand.monosmith-try-smith"

    /** The key jetbrains.trycommand.none-were-asked-for. */
    const val JETBRAINS_TRYCOMMAND_NONE_WERE_ASKED_FOR = "jetbrains.trycommand.none-were-asked-for"

    /** The key jetbrains.trycommand.nothing-was-tried. Values: {0} is failureTextMapper. */
    const val JETBRAINS_TRYCOMMAND_NOTHING_WAS_TRIED = "jetbrains.trycommand.nothing-was-tried"

    /** The key jetbrains.trycommand.run-it-for-real. */
    const val JETBRAINS_TRYCOMMAND_RUN_IT_FOR_REAL = "jetbrains.trycommand.run-it-for-real"

    /** The key jetbrains.trycommand.smith-printed-no-result. */
    const val JETBRAINS_TRYCOMMAND_SMITH_PRINTED_NO_RESULT = "jetbrains.trycommand.smith-printed-no-result"

    /** The key jetbrains.trycommand.something-failed-in-the-copy. */
    const val JETBRAINS_TRYCOMMAND_SOMETHING_FAILED_IN_THE_COPY = "jetbrains.trycommand.something-failed-in-the-copy"

    /** The key jetbrains.trycommand.the-command-and-every-check. */
    const val JETBRAINS_TRYCOMMAND_THE_COMMAND_AND_EVERY_CHECK = "jetbrains.trycommand.the-command-and-every-check"

    /** The key jetbrains.trycommand.the-command-ran-for-real. */
    const val JETBRAINS_TRYCOMMAND_THE_COMMAND_RAN_FOR_REAL = "jetbrains.trycommand.the-command-ran-for-real"

    /** The key jetbrains.trycommand.the-smith-command-to-try. */
    const val JETBRAINS_TRYCOMMAND_THE_SMITH_COMMAND_TO_TRY = "jetbrains.trycommand.the-smith-command-to-try"

    /** The key jetbrains.trycommand.try-smith. Values: {0} is command. */
    const val JETBRAINS_TRYCOMMAND_TRY_SMITH = "jetbrains.trycommand.try-smith"

    /** The key jetbrains.trycommand.type-the-command-to-try. */
    const val JETBRAINS_TRYCOMMAND_TYPE_THE_COMMAND_TO_TRY = "jetbrains.trycommand.type-the-command-to-try"

    /** The key jetbrains.undochange.monosmith-undo-the-last-change. */
    const val JETBRAINS_UNDOCHANGE_MONOSMITH_UNDO_THE_LAST_CHANGE = "jetbrains.undochange.monosmith-undo-the-last-change"

    /** The key jetbrains.undochange.monosmith-undo-the-last-change-2. */
    const val JETBRAINS_UNDOCHANGE_MONOSMITH_UNDO_THE_LAST_CHANGE_2 = "jetbrains.undochange.monosmith-undo-the-last-change-2"

    /** The key jetbrains.undochange.nothing-was-undone. Values: {0} is failureTextMapper. */
    const val JETBRAINS_UNDOCHANGE_NOTHING_WAS_UNDONE = "jetbrains.undochange.nothing-was-undone"

    /** The key jetbrains.undochange.smith-reverts-the-last-change. */
    const val JETBRAINS_UNDOCHANGE_SMITH_REVERTS_THE_LAST_CHANGE = "jetbrains.undochange.smith-reverts-the-last-change"

    /** The key jetbrains.undochange.the-last-change-is-undone. */
    const val JETBRAINS_UNDOCHANGE_THE_LAST_CHANGE_IS_UNDONE = "jetbrains.undochange.the-last-change-is-undone"

    /** The key jetbrains.workspace.monosmith. Values: {0} is title. */
    const val JETBRAINS_WORKSPACE_MONOSMITH = "jetbrains.workspace.monosmith"

    /** The key jetbrains.workspace.smith-projects-exited-with. Values: {0} is exitCode. */
    const val JETBRAINS_WORKSPACE_SMITH_PROJECTS_EXITED_WITH = "jetbrains.workspace.smith-projects-exited-with"

    /** The key jetbrains.workspacereport.monosmith-report. */
    const val JETBRAINS_WORKSPACEREPORT_MONOSMITH_REPORT = "jetbrains.workspacereport.monosmith-report"

    /** The key jetbrains.workspacereport.the-report-was-not-written. Values: {0} is failureTextMapper. */
    const val JETBRAINS_WORKSPACEREPORT_THE_REPORT_WAS_NOT_WRITTEN = "jetbrains.workspacereport.the-report-was-not-written"

    /** The key jetbrains.workspacetools.a-file-a-folder-a. */
    const val JETBRAINS_WORKSPACETOOLS_A_FILE_A_FOLDER_A = "jetbrains.workspacetools.a-file-a-folder-a"

    /** The key jetbrains.workspacetools.environments-of-a-web-app. Values: {0} is verb. */
    const val JETBRAINS_WORKSPACETOOLS_ENVIRONMENTS_OF_A_WEB_APP = "jetbrains.workspacetools.environments-of-a-web-app"

    /** The key jetbrains.workspacetools.explain-a-project. */
    const val JETBRAINS_WORKSPACETOOLS_EXPLAIN_A_PROJECT = "jetbrains.workspacetools.explain-a-project"

    /** The key jetbrains.workspacetools.monosmith-divergences. */
    const val JETBRAINS_WORKSPACETOOLS_MONOSMITH_DIVERGENCES = "jetbrains.workspacetools.monosmith-divergences"

    /** The key jetbrains.workspacetools.monosmith-environments. */
    const val JETBRAINS_WORKSPACETOOLS_MONOSMITH_ENVIRONMENTS = "jetbrains.workspacetools.monosmith-environments"

    /** The key jetbrains.workspacetools.monosmith-environments-of. Values: {0} is verb, {1} is name. */
    const val JETBRAINS_WORKSPACETOOLS_MONOSMITH_ENVIRONMENTS_OF = "jetbrains.workspacetools.monosmith-environments-of"

    /** The key jetbrains.workspacetools.monosmith-explain-a-project. */
    const val JETBRAINS_WORKSPACETOOLS_MONOSMITH_EXPLAIN_A_PROJECT = "jetbrains.workspacetools.monosmith-explain-a-project"

    /** The key jetbrains.workspacetools.monosmith-install-the-agent. */
    const val JETBRAINS_WORKSPACETOOLS_MONOSMITH_INSTALL_THE_AGENT = "jetbrains.workspacetools.monosmith-install-the-agent"

    /** The key jetbrains.workspacetools.monosmith-install-the-mcp-server. */
    const val JETBRAINS_WORKSPACETOOLS_MONOSMITH_INSTALL_THE_MCP_SERVER = "jetbrains.workspacetools.monosmith-install-the-mcp-server"

    /** The key jetbrains.workspacetools.monosmith-on. Values: {0} is verb, {1} is names, {2} is name. */
    const val JETBRAINS_WORKSPACETOOLS_MONOSMITH_ON = "jetbrains.workspacetools.monosmith-on"

    /** The key jetbrains.workspacetools.monosmith-why. */
    const val JETBRAINS_WORKSPACETOOLS_MONOSMITH_WHY = "jetbrains.workspacetools.monosmith-why"

    /** The key jetbrains.workspacetools.monosmith-why-is-this-here. */
    const val JETBRAINS_WORKSPACETOOLS_MONOSMITH_WHY_IS_THIS_HERE = "jetbrains.workspacetools.monosmith-why-is-this-here"

    /** The key jetbrains.workspacetools.monosmith-write-the-launcher. */
    const val JETBRAINS_WORKSPACETOOLS_MONOSMITH_WRITE_THE_LAUNCHER = "jetbrains.workspacetools.monosmith-write-the-launcher"

    /** The key jetbrains.workspacetools.monosmith-write-the-smith-launcher. */
    const val JETBRAINS_WORKSPACETOOLS_MONOSMITH_WRITE_THE_SMITH_LAUNCHER = "jetbrains.workspacetools.monosmith-write-the-smith-launcher"

    /** The key jetbrains.workspacetools.name-at-least-one-environment. */
    const val JETBRAINS_WORKSPACETOOLS_NAME_AT_LEAST_ONE_ENVIRONMENT = "jetbrains.workspacetools.name-at-least-one-environment"

    /** The key jetbrains.workspacetools.nothing-was-explained. Values: {0} is message. */
    const val JETBRAINS_WORKSPACETOOLS_NOTHING_WAS_EXPLAINED = "jetbrains.workspacetools.nothing-was-explained"

    /** The key jetbrains.workspacetools.smith-writes-mcp-json-and. */
    const val JETBRAINS_WORKSPACETOOLS_SMITH_WRITES_MCP_JSON_AND = "jetbrains.workspacetools.smith-writes-mcp-json-and"

    /** The key jetbrains.workspacetools.the-agent-files-are-written. */
    const val JETBRAINS_WORKSPACETOOLS_THE_AGENT_FILES_ARE_WRITTEN = "jetbrains.workspacetools.the-agent-files-are-written"

    /** The key jetbrains.workspacetools.the-environments-of-changed. Values: {0} is name. */
    const val JETBRAINS_WORKSPACETOOLS_THE_ENVIRONMENTS_OF_CHANGED = "jetbrains.workspacetools.the-environments-of-changed"

    /** The key jetbrains.workspacetools.the-environments-to-lowercase-words. Values: {0} is verb. */
    const val JETBRAINS_WORKSPACETOOLS_THE_ENVIRONMENTS_TO_LOWERCASE_WORDS = "jetbrains.workspacetools.the-environments-to-lowercase-words"

    /** The key jetbrains.workspacetools.the-files-are-meant-to. */
    const val JETBRAINS_WORKSPACETOOLS_THE_FILES_ARE_MEANT_TO = "jetbrains.workspacetools.the-files-are-meant-to"

    /** The key jetbrains.workspacetools.the-files-it-creates-the. */
    const val JETBRAINS_WORKSPACETOOLS_THE_FILES_IT_CREATES_THE = "jetbrains.workspacetools.the-files-it-creates-the"

    /** The key jetbrains.workspacetools.the-files-scripts-and-smith. */
    const val JETBRAINS_WORKSPACETOOLS_THE_FILES_SCRIPTS_AND_SMITH = "jetbrains.workspacetools.the-files-scripts-and-smith"

    /** The key jetbrains.workspacetools.the-launcher-is-written. */
    const val JETBRAINS_WORKSPACETOOLS_THE_LAUNCHER_IS_WRITTEN = "jetbrains.workspacetools.the-launcher-is-written"

    /** The key jetbrains.workspacetools.type-what-you-want-explained. */
    const val JETBRAINS_WORKSPACETOOLS_TYPE_WHAT_YOU_WANT_EXPLAINED = "jetbrains.workspacetools.type-what-you-want-explained"

    /** The key packageprompt.range.any.label. */
    const val PACKAGEPROMPT_RANGE_ANY_LABEL = "packageprompt.range.any.label"

    /** The key packageprompt.range.any.meaning. */
    const val PACKAGEPROMPT_RANGE_ANY_MEANING = "packageprompt.range.any.meaning"

    /** The key packageprompt.range.follow.meaning. */
    const val PACKAGEPROMPT_RANGE_FOLLOW_MEANING = "packageprompt.range.follow.meaning"

    /** The key packageprompt.source.workspace. */
    const val PACKAGEPROMPT_SOURCE_WORKSPACE = "packageprompt.source.workspace"

    /** The key resultdialog.check.label. */
    const val RESULTDIALOG_CHECK_LABEL = "resultdialog.check.label"

    /** The key resultdialog.version.highest.label. */
    const val RESULTDIALOG_VERSION_HIGHEST_LABEL = "resultdialog.version.highest.label"

    /** The key resultdialog.version.installed.label. */
    const val RESULTDIALOG_VERSION_INSTALLED_LABEL = "resultdialog.version.installed.label"

    /** The key resultdialog.version.lowest.label. */
    const val RESULTDIALOG_VERSION_LOWEST_LABEL = "resultdialog.version.lowest.label"

    /** The key settingsview.group.install. */
    const val SETTINGSVIEW_GROUP_INSTALL = "settingsview.group.install"

    /** The key settingsview.reading. */
    const val SETTINGSVIEW_READING = "settingsview.reading"

    /** The key toolsdialog.answer.myself. */
    const val TOOLSDIALOG_ANSWER_MYSELF = "toolsdialog.answer.myself"

    /** The key toolsdialog.answer.smith. */
    const val TOOLSDIALOG_ANSWER_SMITH = "toolsdialog.answer.smith"

    /** The key toolsdialog.done.none. */
    const val TOOLSDIALOG_DONE_NONE = "toolsdialog.done.none"

    /**
     * The text of a key in the default locale, with the values in {0}, {1}...
     *
     * @param key one of the constants above.
     * @param values the value of each placeholder, in the order the constant lists.
     * @return the text; the key itself when the bundle lacks it.
     */
    fun say(
        key: String,
        vararg values: Any,
    ): String {
        val bundle = ResourceBundle.getBundle(BUNDLE)
        if (!bundle.containsKey(key)) return key
        val text = bundle.getString(key)
        return if (values.isEmpty()) text else MessageFormat.format(text, *values)
    }
}
