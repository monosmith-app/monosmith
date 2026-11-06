package dev.monosmith.jetbrains.pluginmarket

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.flowports.FlowPortsContract
import dev.monosmith.jetbrains.flowports.IdePortsClient
import dev.monosmith.jetbrains.smithcommand.TaskCommandAlgorithm
import dev.monosmith.jetbrains.smithrun.BackgroundFlowClient
import dev.monosmith.jetbrains.smithrun.FailureTextMapper
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_PLUGINMARKET_ALSO_PIN_THE_SHA_256
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_PLUGINMARKET_ALSO_TRUST_IT_IN_THIS
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_PLUGINMARKET_IS_INSTALLED
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_PLUGINMARKET_MONOSMITH_INSTALL
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_PLUGINMARKET_MONOSMITH_PLUGINS
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_PLUGINMARKET_NO_PLUGIN_MATCHES
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_PLUGINMARKET_NO_PLUGIN_WAS_INSTALLED
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_PLUGINMARKET_PICK_A_PLUGIN_THE_NEWEST
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_PLUGINMARKET_SMITH_PRINTED_THE_PLUGINS_IN
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_PLUGINMARKET_WAS_NOT_INSTALLED
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_PLUGINMARKET_WORDS_TO_SEARCH_FOR_EMPTY
import dev.monosmith.jetbrains.words.SmithWordsConfig.SETTINGSVIEW_GROUP_INSTALL
import dev.monosmith.jetbrains.words.SmithWordsConfig.say
import java.nio.file.Path

/**
 * Searches the plugin marketplace (`smith plugin search --json`) and installs the plugin the person picks.
 * The trust question comes before the install: the two boxes (trust it in this workspace, pin its program's
 * SHA-256) are unticked, so a plugin is never trusted by default.
 */
object InstallPluginUseCase {
    private val TITLE = say(JETBRAINS_PLUGINMARKET_MONOSMITH_PLUGINS)

    /** Starts the flow in a background task. */
    fun start(
        ide: Project,
        root: Path,
    ) = BackgroundFlowClient.start(ide, TITLE) { run(IdePortsClient.of(ide, root)) }

    /**
     * @return true when a plugin was installed
     */
    fun run(ports: FlowPortsContract): Boolean {
        val term = ports.prompt.text(TITLE, say(JETBRAINS_PLUGINMARKET_WORDS_TO_SEARCH_FOR_EMPTY)) ?: return false
        val found = ports.smith.run(PluginSearchMapper.searchArguments(term))
        if (!found.ok) {
            ports.notice.warn(say(JETBRAINS_PLUGINMARKET_NO_PLUGIN_WAS_INSTALLED, FailureTextMapper.describe(found).toString()))
            return false
        }
        val plugins =
            runCatching { PluginSearchMapper.toPlugins(found.stdout) }.getOrElse {
                ports.notice.warn(say(JETBRAINS_PLUGINMARKET_SMITH_PRINTED_THE_PLUGINS_IN))
                return false
            }
        if (plugins.isEmpty()) {
            ports.notice.info(say(JETBRAINS_PLUGINMARKET_NO_PLUGIN_MATCHES, term.toString()))
            return false
        }
        val picked =
            ports.prompt
                .pickOne(
                    TITLE,
                    say(JETBRAINS_PLUGINMARKET_PICK_A_PLUGIN_THE_NEWEST),
                    PluginSearchMapper.rows(plugins),
                    "Next",
                )?.name
                ?: return false
        val plugin = plugins.first { it.name == picked }
        val answer =
            ports.prompt.plan(
                say(JETBRAINS_PLUGINMARKET_MONOSMITH_INSTALL, plugin.name.toString()),
                PluginSearchMapper.trustWords(plugin),
                TaskCommandAlgorithm.display(listOf("smith") + PluginSearchMapper.installArguments(plugin, trust = false, pin = false)),
                say(SETTINGSVIEW_GROUP_INSTALL),
                offerCommit = false,
                checks =
                    listOf(
                        say(JETBRAINS_PLUGINMARKET_ALSO_TRUST_IT_IN_THIS),
                        say(JETBRAINS_PLUGINMARKET_ALSO_PIN_THE_SHA_256),
                    ),
            ) ?: return false
        val args =
            PluginSearchMapper.installArguments(
                plugin,
                trust = answer.checks.getOrNull(0) == true,
                pin =
                    answer.checks.getOrNull(1) == true,
            )
        val done = ports.smith.run(args, show = true)
        ports.changed()
        if (!done.ok) {
            ports.notice.warn(
                say(JETBRAINS_PLUGINMARKET_WAS_NOT_INSTALLED, plugin.name.toString(), FailureTextMapper.describe(done).toString()),
            )
            return false
        }
        ports.notice.info(say(JETBRAINS_PLUGINMARKET_IS_INSTALLED, plugin.name.toString(), plugin.newest.toString()))
        return true
    }
}
