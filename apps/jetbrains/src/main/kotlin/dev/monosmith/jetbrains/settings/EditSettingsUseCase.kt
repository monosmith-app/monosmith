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
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDCHOICES_DEFAULT
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_SETTINGS_CHANGE_IT
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_SETTINGS_GO_BACK_TO_THE_DEFAULT
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_SETTINGS_IT_IS_SET_TO
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_SETTINGS_MONOSMITH_SETTINGS
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_SETTINGS_NOTHING_WAS_CHANGED_YET_SMITH
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_SETTINGS_NO_SETTING_WAS_CHANGED
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_SETTINGS_NO_SETTING_WAS_CHANGED_THIS
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_SETTINGS_PICK_A_SETTING_TYPE_IN
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_SETTINGS_REMOVE_THE_LINE_FROM_SMITH
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_SETTINGS_RESET_IT
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_SETTINGS_SMITH_PRINTED_ITS_SETTINGS_IN
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_SETTINGS_THE_CURRENT_VALUE_IS_FIRST
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_SETTINGS_TYPE_A_VALUE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_SETTINGS_WRITE_A_NEW_VALUE_IN
import dev.monosmith.jetbrains.words.SmithWordsConfig.say
import java.nio.file.Path

/**
 * Edits one setting of smith.toml: the list comes from `smith config list --json` (with each setting's
 * value, description and the reason of its default), a setting of a few values is a list of those values, a
 * true-or-false setting is two answers, and any other is a text. The change is `smith config set` (or
 * `unset`), shown first. smith never commits here.
 */
object EditSettingsUseCase {
    private val TITLE = say(JETBRAINS_SETTINGS_MONOSMITH_SETTINGS)
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
            ports.notice.warn(say(JETBRAINS_SETTINGS_NO_SETTING_WAS_CHANGED, FailureTextMapper.describe(listed).toString()))
            return false
        }
        val settings =
            runCatching { SettingsJsonMapper.toSettings(listed.stdout) }.getOrElse {
                ports.notice.warn(say(JETBRAINS_SETTINGS_SMITH_PRINTED_ITS_SETTINGS_IN))
                return false
            }
        val picked = only ?: askSetting(ports, settings) ?: return false
        val path = picked
        val setting = settings.firstOrNull { it.path == path }
        if (setting == null) {
            ports.notice.warn(say(JETBRAINS_SETTINGS_NO_SETTING_WAS_CHANGED_THIS, path.toString()))
            return false
        }
        val words = SettingsJsonMapper.words(setting)
        val action =
            if (setting.set) {
                ports.prompt
                    .pickOne(
                        "$TITLE: ${setting.path}",
                        say(
                            JETBRAINS_SETTINGS_IT_IS_SET_TO,
                            (if (words.isEmpty()) "" else "$words\n").toString(),
                            SettingsJsonMapper.shown(setting).toString(),
                        ),
                        listOf(
                            ChoiceRowContract(
                                CHANGE,
                                say(JETBRAINS_SETTINGS_CHANGE_IT),
                                say(JETBRAINS_ADDCHOICES_DEFAULT),
                                say(JETBRAINS_SETTINGS_WRITE_A_NEW_VALUE_IN),
                            ),
                            ChoiceRowContract(
                                RESET,
                                say(JETBRAINS_SETTINGS_GO_BACK_TO_THE_DEFAULT),
                                "",
                                say(JETBRAINS_SETTINGS_REMOVE_THE_LINE_FROM_SMITH),
                            ),
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
                header = say(JETBRAINS_SETTINGS_NOTHING_WAS_CHANGED_YET_SMITH, (if (words.isEmpty()) "" else "$words\n\n").toString()),
                okText = if (action == RESET) say(JETBRAINS_SETTINGS_RESET_IT) else say(JETBRAINS_SETTINGS_CHANGE_IT),
                offerCommit = true,
                apply = { args },
                done = "${setting.path} is ${if (action == RESET) "back to its default" else "changed"}",
                operation = "${if (action == RESET) "unset" else "set"} ${setting.path}",
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
                        say(JETBRAINS_SETTINGS_THE_CURRENT_VALUE_IS_FIRST, words.toString()),
                        SettingsJsonMapper.enumRows(setting),
                    )?.name
            }

            setting.type == "boolean" -> {
                ports.prompt
                    .pickOne(
                        "$TITLE: ${setting.path}",
                        say(JETBRAINS_SETTINGS_THE_CURRENT_VALUE_IS_FIRST, words.toString()),
                        SettingsJsonMapper.booleanRows(setting),
                    )?.name
            }

            else -> {
                ports.prompt.text(
                    "$TITLE: ${setting.path}",
                    "$words\n${if (setting.type == "array") "Values, separated by commas:" else "The new value:"}",
                    SettingsJsonMapper.shown(setting),
                ) { if (it.isEmpty()) say(JETBRAINS_SETTINGS_TYPE_A_VALUE) else null }
            }
        }

    private fun askSetting(
        ports: FlowPortsContract,
        settings: List<SettingContract>,
    ): String? = ports.prompt.pickOne(TITLE, say(JETBRAINS_SETTINGS_PICK_A_SETTING_TYPE_IN), SettingsJsonMapper.rows(settings))?.name
}
