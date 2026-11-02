package dev.monosmith.jetbrains.settings

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.commitchanges.CommitChangesUseCase
import dev.monosmith.jetbrains.flowports.FlowPortsContract
import dev.monosmith.jetbrains.flowports.IdePortsClient
import dev.monosmith.jetbrains.planapply.PlanStepContract
import dev.monosmith.jetbrains.planapply.PlanThenApplyUseCase
import dev.monosmith.jetbrains.prompt.ChoiceRowContract
import dev.monosmith.jetbrains.smithrun.BackgroundFlowClient
import dev.monosmith.jetbrains.smithrun.FailureTextMapper
import java.nio.file.Path

/**
 * Edits one setting of smith.toml: the list comes from `smith config list --json` (with each setting's
 * value, description and the reason of its default), a setting of a few values is a list of those values, a
 * true-or-false setting is two answers, and any other is a text. The change is `smith config set` (or
 * `unset`), shown first. smith never commits here.
 */
object EditSettingsUseCase {
    private const val TITLE = "MonoSmith: settings"
    private const val CHANGE = "change"
    private const val RESET = "reset"

    /** Starts the flow in a background task. */
    fun start(
        ide: Project,
        root: Path,
    ) = BackgroundFlowClient.start(ide, TITLE) { run(IdePortsClient.of(ide, root)) { CommitChangesUseCase.start(ide, root, it) } }

    /** Starts the flow for one setting, such as `dependencies.one_version`, without the list: the questions start at its value. */
    fun startOne(
        ide: Project,
        root: Path,
        path: String,
    ) = BackgroundFlowClient.start(ide, TITLE) { run(IdePortsClient.of(ide, root), path) { CommitChangesUseCase.start(ide, root, it) } }

    /**
     * @param only the path of the one setting to edit, or null to pick from the list
     * @return true when a setting was changed
     */
    fun run(
        ports: FlowPortsContract,
        only: String? = null,
        commitLater: (String) -> Unit,
    ): Boolean {
        val listed = ports.smith.run(listOf("config", "list", "--json"))
        if (!listed.ok) {
            ports.notice.warn("No setting was changed. ${FailureTextMapper.describe(listed)}")
            return false
        }
        val settings =
            runCatching { SettingsJsonMapper.toSettings(listed.stdout) }.getOrElse {
                ports.notice.warn("smith printed its settings in a form this plugin does not read. Update the plugin.")
                return false
            }
        val picked = only ?: askSetting(ports, settings) ?: return false
        val path = picked
        val setting = settings.firstOrNull { it.path == path }
        if (setting == null) {
            ports.notice.warn("No setting was changed. This smith has no setting $path. Update smith.")
            return false
        }
        val words = SettingsJsonMapper.words(setting)
        val action =
            if (setting.set) {
                ports.prompt
                    .pickOne(
                        "$TITLE: ${setting.path}",
                        "${if (words.isEmpty()) "" else "$words\n"}It is set to ${SettingsJsonMapper.shown(setting)}.",
                        listOf(
                            ChoiceRowContract(CHANGE, "Change it", "(default)", "Write a new value in smith.toml."),
                            ChoiceRowContract(RESET, "Go back to the default", "", "Remove the line from smith.toml."),
                        ),
                    )?.name ?: return false
            } else {
                CHANGE
            }
        val args =
            if (action == RESET) {
                SettingsJsonMapper.unsetArguments(setting)
            } else {
                val value = askValue(ports, setting, words) ?: return false
                SettingsJsonMapper.setArguments(setting, value)
            }
        return PlanThenApplyUseCase.run(
            ports,
            PlanStepContract(
                title = "$TITLE: ${setting.path}?",
                planArgs = null,
                fixedPlan = "smith ${args.joinToString(" ")}",
                header = "${if (words.isEmpty()) "" else "$words\n\n"}Nothing was changed yet. smith edits smith.toml.",
                okText = if (action == RESET) "Reset it" else "Change it",
                offerCommit = false,
                apply = { args },
                done = "${setting.path} is ${if (action == RESET) "back to its default" else "changed"}",
            ),
            commitLater,
        )
    }

    private fun askValue(
        ports: FlowPortsContract,
        setting: SettingContract,
        words: String,
    ): String? =
        when {
            setting.choices.isNotEmpty() -> {
                ports.prompt
                    .pickOne(
                        "$TITLE: ${setting.path}",
                        "$words\nThe current value is first.",
                        SettingsJsonMapper.enumRows(setting),
                    )?.name
            }

            setting.type == "boolean" -> {
                ports.prompt
                    .pickOne(
                        "$TITLE: ${setting.path}",
                        "$words\nThe current value is first.",
                        SettingsJsonMapper.booleanRows(setting),
                    )?.name
            }

            else -> {
                ports.prompt.text(
                    "$TITLE: ${setting.path}",
                    "$words\n${if (setting.type == "array") "Values, separated by commas:" else "The new value:"}",
                    SettingsJsonMapper.shown(setting),
                ) { if (it.isEmpty()) "Type a value." else null }
            }
        }

    private fun askSetting(
        ports: FlowPortsContract,
        settings: List<SettingContract>,
    ): String? = ports.prompt.pickOne(TITLE, "Pick a setting. Type in the box to narrow the list.", SettingsJsonMapper.rows(settings))?.name
}
