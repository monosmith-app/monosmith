package dev.monosmith.jetbrains.pluginswitch

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.commitchanges.CommitChangesUseCase
import dev.monosmith.jetbrains.flowports.FlowPortsContract
import dev.monosmith.jetbrains.flowports.IdePortsClient
import dev.monosmith.jetbrains.notice.NoticeButtonContract
import dev.monosmith.jetbrains.smithrun.BackgroundFlowClient
import dev.monosmith.jetbrains.smithrun.FailureTextMapper
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_MAKE_ONE_COMMIT_THAT_UNDO
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_PLANAPPLY_COMMIT_CHANGES
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_PLUGINMARKET_MONOSMITH_PLUGINS
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_PLUGINMARKET_SMITH_PRINTED_THE_PLUGINS_IN
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_PLUGINSWITCH_IS_COMMITTED_AS_ONE_CHANGE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_PLUGINSWITCH_IS_NOT_COMMITTED_REVIEW_IT
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_PLUGINSWITCH_NOTHING_WAS_CHANGED_THESE_PROJECTS
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_PLUGINSWITCH_NO_PLUGIN_WAS_CHANGED
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_PLUGINSWITCH_PICK_A_BUILT_IN_PLUGIN
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_PLUGINSWITCH_SMITH_COUNTS_THE_PROJECTS_FINDS
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_PLUGINSWITCH_SMITH_LISTS_NO_BUILT_IN
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_PLUGINSWITCH_SWITCHING_OFF_HIDES_PROJECTS
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_PLUGINSWITCH_THE_PROJECTS_FINDS_APPEAR_IN
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_PLUGINSWITCH_TURN_A_PLUGIN
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_PLUGINSWITCH_TURN_IT
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_PLUGINSWITCH_TURN_IT_OFF_ANYWAY
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_PLUGINSWITCH_TURN_OFF_ANYWAY
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_PLUGINSWITCH_WAS_NOT_TURNED
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_PLUGINSWITCH_WAS_NOT_TURNED_OFF
import dev.monosmith.jetbrains.words.SmithWordsConfig.say
import java.nio.file.Path

/**
 * Turns a built-in plugin of smith on or off in smith.toml (`smith plugin enable` and `disable`). The
 * plugin is picked from `smith plugin list --json`, and the command is shown before it runs. When smith
 * refuses to turn a plugin off because projects use it, the flow shows smith's list of those projects and
 * asks again; only a second yes runs `--force`. smith never commits here.
 */
object SwitchPluginUseCase {
    private val TITLE = say(JETBRAINS_PLUGINMARKET_MONOSMITH_PLUGINS)
    private val HIDES_PROJECTS = say(JETBRAINS_PLUGINSWITCH_SWITCHING_OFF_HIDES_PROJECTS)

    /** Starts the flow in a background task; [on] true turns a plugin on, false turns one off. */
    fun start(
        ide: Project,
        root: Path,
        on: Boolean,
    ) = BackgroundFlowClient.start(ide, TITLE) { run(IdePortsClient.of(ide, root), on) { CommitChangesUseCase.start(ide, root, it) } }

    /**
     * @param on true turns a plugin on, false turns one off
     * @return true when smith changed smith.toml
     */
    fun run(
        ports: FlowPortsContract,
        on: Boolean,
        commitLater: (String) -> Unit,
    ): Boolean {
        val listed = ports.smith.run(listOf("plugin", "list", "--json"))
        if (!listed.ok) {
            ports.notice.warn(say(JETBRAINS_PLUGINSWITCH_NO_PLUGIN_WAS_CHANGED, FailureTextMapper.describe(listed).toString()))
            return false
        }
        val plugins =
            runCatching { BuiltinPluginsMapper.toPlugins(listed.stdout) }.getOrElse {
                ports.notice.warn(say(JETBRAINS_PLUGINMARKET_SMITH_PRINTED_THE_PLUGINS_IN))
                return false
            }
        if (plugins.isEmpty()) {
            ports.notice.info(say(JETBRAINS_PLUGINSWITCH_SMITH_LISTS_NO_BUILT_IN))
            return false
        }
        val word = if (on) "on" else "off"
        val name =
            ports.prompt
                .pickOne(
                    say(JETBRAINS_PLUGINSWITCH_TURN_A_PLUGIN, TITLE.toString(), word.toString()),
                    say(JETBRAINS_PLUGINSWITCH_PICK_A_BUILT_IN_PLUGIN),
                    BuiltinPluginsMapper.rows(plugins),
                )?.name
                ?: return false
        val args = listOf("plugin", if (on) "enable" else "disable", name)
        val answer =
            ports.prompt.plan(
                "$TITLE: turn $name $word?",
                if (on) {
                    say(JETBRAINS_PLUGINSWITCH_THE_PROJECTS_FINDS_APPEAR_IN, name.toString())
                } else {
                    say(JETBRAINS_PLUGINSWITCH_SMITH_COUNTS_THE_PROJECTS_FINDS, name.toString())
                },
                "smith ${args.joinToString(" ")}",
                say(JETBRAINS_PLUGINSWITCH_TURN_IT, word.toString()),
                offerCommit = true,
            ) ?: return false
        val commit = if (answer.commit) listOf("--commit") else emptyList()
        val first = ports.smith.run(args + commit, show = true)
        if (first.ok) return done(ports, name, word, answer.commit, commitLater)
        if (on || !first.stderr.contains(HIDES_PROJECTS)) {
            ports.notice.warn(
                say(JETBRAINS_PLUGINSWITCH_WAS_NOT_TURNED, name.toString(), word.toString(), FailureTextMapper.describe(first).toString()),
            )
            return false
        }
        ports.prompt.plan(
            say(JETBRAINS_PLUGINSWITCH_TURN_OFF_ANYWAY, TITLE.toString(), name.toString()),
            say(JETBRAINS_PLUGINSWITCH_NOTHING_WAS_CHANGED_THESE_PROJECTS, name.toString()),
            first.stderr.trim(),
            say(JETBRAINS_PLUGINSWITCH_TURN_IT_OFF_ANYWAY),
            offerCommit = false,
        ) ?: return false
        val forced = ports.smith.run(args + "--force" + commit, show = true)
        if (!forced.ok) {
            ports.notice.warn(
                say(JETBRAINS_PLUGINSWITCH_WAS_NOT_TURNED_OFF, name.toString(), FailureTextMapper.describe(forced).toString()),
            )
            return false
        }
        return done(ports, name, word, answer.commit, commitLater)
    }

    private fun done(
        ports: FlowPortsContract,
        name: String,
        word: String,
        committed: Boolean,
        commitLater: (String) -> Unit,
    ): Boolean {
        ports.changed()
        if (committed) {
            ports.notice.info(say(JETBRAINS_PLUGINSWITCH_IS_COMMITTED_AS_ONE_CHANGE, name.toString(), word.toString()))
        } else {
            val button =
                NoticeButtonContract(say(JETBRAINS_PLANAPPLY_COMMIT_CHANGES), say(JETBRAINS_ADDPROJECT_MAKE_ONE_COMMIT_THAT_UNDO)) {
                    commitLater("turn $name $word")
                }
            ports.notice.info(say(JETBRAINS_PLUGINSWITCH_IS_NOT_COMMITTED_REVIEW_IT, name.toString(), word.toString()), listOf(button))
        }
        return true
    }
}
