package dev.monosmith.jetbrains.features.settings

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.features.commitchanges.CommitChangesUseCase
import dev.monosmith.jetbrains.features.flowports.FlowPortsContract
import dev.monosmith.jetbrains.features.flowports.IdePortsClient
import dev.monosmith.jetbrains.features.smithrun.BackgroundFlowClient
import dev.monosmith.jetbrains.features.smithrun.FailureTextMapper
import dev.monosmith.jetbrains.integrations.smithcommand.SmithFlagsConfig
import dev.monosmith.jetbrains.shared.prompt.ChoiceRowContract
import dev.monosmith.jetbrains.shared.words.SmithWordsConfig.JETBRAINS_ADDCHOICES_DEFAULT
import dev.monosmith.jetbrains.shared.words.SmithWordsConfig.JETBRAINS_PROJECTSETTINGS_COMES_FROM
import dev.monosmith.jetbrains.shared.words.SmithWordsConfig.JETBRAINS_PROJECTSETTINGS_PICK
import dev.monosmith.jetbrains.shared.words.SmithWordsConfig.JETBRAINS_PROJECTSETTINGS_READ_ONLY_FROM
import dev.monosmith.jetbrains.shared.words.SmithWordsConfig.JETBRAINS_PROJECTSETTINGS_TITLE
import dev.monosmith.jetbrains.shared.words.SmithWordsConfig.JETBRAINS_SETTINGS_NO_SETTING_WAS_CHANGED
import dev.monosmith.jetbrains.shared.words.SmithWordsConfig.JETBRAINS_SETTINGS_SMITH_PRINTED_ITS_SETTINGS_IN
import dev.monosmith.jetbrains.shared.words.SmithWordsConfig.say
import java.nio.file.Path

/**
 * The settings of one project (`[projects.<name>]` in smith.toml): the list comes from
 * `smith config list -p <name> --json`. The values the project's manifest owns come first, read-only, and say
 * which file to change them in; a setting is changed with the same questions as a workspace setting
 * ([EditSettingsUseCase.edit]), and going back to the default removes its line.
 */
object ProjectSettingsUseCase {
    private const val FACT = "fact:"

    /** Starts the flow for [project] in a background task. */
    fun start(
        ide: Project,
        root: Path,
        project: String,
    ) = BackgroundFlowClient.start(ide, say(JETBRAINS_PROJECTSETTINGS_TITLE, project)) {
        run(IdePortsClient.of(ide, root), project) { CommitChangesUseCase.start(ide, root, it) }
    }

    /**
     * @param project the name of the project
     * @return true when a setting was changed
     */
    fun run(
        ports: FlowPortsContract,
        project: String,
        commitLater: (String) -> Unit,
    ): Boolean {
        val listed = ports.smith.run(listOf("config", "list", SmithFlagsConfig.PROJECT, project, "--json"))
        if (!listed.ok) {
            ports.notice.warn(say(JETBRAINS_SETTINGS_NO_SETTING_WAS_CHANGED, FailureTextMapper.describe(listed).toString()))
            return false
        }
        val view =
            runCatching { SettingsJsonMapper.toProjectView(listed.stdout) }.getOrElse {
                ports.notice.warn(say(JETBRAINS_SETTINGS_SMITH_PRINTED_ITS_SETTINGS_IN))
                return false
            }
        val picked =
            ports.prompt
                .pickOne(say(JETBRAINS_PROJECTSETTINGS_TITLE, project), say(JETBRAINS_PROJECTSETTINGS_PICK, view.manifest), rows(view))
                ?.name ?: return false
        if (picked.startsWith(FACT)) {
            val fact = view.facts.first { FACT + it.key == picked }
            ports.notice.info(say(JETBRAINS_PROJECTSETTINGS_COMES_FROM, fact.key, fact.source))
            return false
        }
        val setting = view.settings.firstOrNull { it.path == picked } ?: return false
        return EditSettingsUseCase.edit(ports, setting, commitLater)
    }

    /** The lines of "which setting?": the read-only values with their file, then each setting with its value or `default`. */
    fun rows(view: ProjectViewContract): List<ChoiceRowContract> =
        view.facts.map {
            ChoiceRowContract(FACT + it.key, it.key, "= ${it.value}", say(JETBRAINS_PROJECTSETTINGS_READ_ONLY_FROM, it.source))
        } +
            view.settings.map {
                ChoiceRowContract(
                    it.path,
                    it.path.substringAfterLast('.'),
                    if (it.set) "= ${SettingsJsonMapper.shown(it)}" else say(JETBRAINS_ADDCHOICES_DEFAULT),
                    it.description,
                )
            }
}
