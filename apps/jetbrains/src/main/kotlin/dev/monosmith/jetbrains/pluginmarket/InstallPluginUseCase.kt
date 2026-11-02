package dev.monosmith.jetbrains.pluginmarket

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.flowports.FlowPortsContract
import dev.monosmith.jetbrains.flowports.IdePortsClient
import dev.monosmith.jetbrains.smithcommand.TaskCommandAlgorithm
import dev.monosmith.jetbrains.smithrun.BackgroundFlowClient
import dev.monosmith.jetbrains.smithrun.FailureTextMapper
import java.nio.file.Path

/**
 * Searches the plugin marketplace (`smith plugin search --json`) and installs the plugin the person picks.
 * The trust question comes before the install: the two boxes (trust it in this workspace, pin its program's
 * SHA-256) are unticked, so a plugin is never trusted by default.
 */
object InstallPluginUseCase {
    private const val TITLE = "MonoSmith: plugins"

    /** Starts the flow in a background task. */
    fun start(
        ide: Project,
        root: Path,
    ) = BackgroundFlowClient.start(ide, TITLE) { run(IdePortsClient.of(ide, root)) }

    /**
     * @return true when a plugin was installed
     */
    fun run(ports: FlowPortsContract): Boolean {
        val term = ports.prompt.text(TITLE, "Words to search for (empty lists every plugin):") ?: return false
        val found = ports.smith.run(PluginSearchMapper.searchArguments(term))
        if (!found.ok) {
            ports.notice.warn("No plugin was installed. ${FailureTextMapper.describe(found)}")
            return false
        }
        val plugins =
            runCatching { PluginSearchMapper.toPlugins(found.stdout) }.getOrElse {
                ports.notice.warn("smith printed the plugins in a form this plugin does not read. Update the plugin.")
                return false
            }
        if (plugins.isEmpty()) {
            ports.notice.info("No plugin matches \"$term\".")
            return false
        }
        val picked =
            ports.prompt.pickOne(TITLE, "Pick a plugin. The newest version is installed.", PluginSearchMapper.rows(plugins), "Next")?.name
                ?: return false
        val plugin = plugins.first { it.name == picked }
        val answer =
            ports.prompt.plan(
                "MonoSmith: install ${plugin.name}?",
                PluginSearchMapper.trustWords(plugin),
                TaskCommandAlgorithm.display(listOf("smith") + PluginSearchMapper.installArguments(plugin, trust = false, pin = false)),
                "Install",
                offerCommit = false,
                checks =
                    listOf(
                        "Also trust it in this workspace (--trust)",
                        "Also pin the SHA-256 of its program in smith.toml, which CI checks (--pin)",
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
            ports.notice.warn("${plugin.name} was not installed. ${FailureTextMapper.describe(done)}")
            return false
        }
        ports.notice.info("${plugin.name} ${plugin.newest} is installed.")
        return true
    }
}
