package dev.monosmith.jetbrains.pluginswitch

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.flowports.FlowPortsContract
import dev.monosmith.jetbrains.flowports.IdePortsClient
import dev.monosmith.jetbrains.smithrun.BackgroundFlowClient
import dev.monosmith.jetbrains.smithrun.FailureTextMapper
import java.nio.file.Path

/**
 * Turns a built-in plugin of smith on or off in smith.toml (`smith plugin enable` and `disable`). The
 * plugin is picked from `smith plugin list --json`, and the command is shown before it runs. When smith
 * refuses to turn a plugin off because projects use it, the flow shows smith's list of those projects and
 * asks again; only a second yes runs `--force`. smith never commits here.
 */
object SwitchPluginUseCase {
    private const val TITLE = "MonoSmith: plugins"
    private const val HIDES_PROJECTS = "switching off hides projects"

    /** Starts the flow in a background task; [on] true turns a plugin on, false turns one off. */
    fun start(
        ide: Project,
        root: Path,
        on: Boolean,
    ) = BackgroundFlowClient.start(ide, TITLE) { run(IdePortsClient.of(ide, root), on) }

    /**
     * @param on true turns a plugin on, false turns one off
     * @return true when smith changed smith.toml
     */
    fun run(
        ports: FlowPortsContract,
        on: Boolean,
    ): Boolean {
        val listed = ports.smith.run(listOf("plugin", "list", "--json"))
        if (!listed.ok) {
            ports.notice.warn("No plugin was changed. ${FailureTextMapper.describe(listed)}")
            return false
        }
        val plugins =
            runCatching { BuiltinPluginsMapper.toPlugins(listed.stdout) }.getOrElse {
                ports.notice.warn("smith printed the plugins in a form this plugin does not read. Update the plugin.")
                return false
            }
        if (plugins.isEmpty()) {
            ports.notice.info("smith lists no built-in plugin.")
            return false
        }
        val word = if (on) "on" else "off"
        val name =
            ports.prompt.pickOne("$TITLE: turn a plugin $word", "Pick a built-in plugin.", BuiltinPluginsMapper.rows(plugins))?.name
                ?: return false
        val args = listOf("plugin", if (on) "enable" else "disable", name)
        ports.prompt.plan(
            "$TITLE: turn $name $word?",
            if (on) {
                "The projects $name finds appear in smith. Nothing is written until you answer."
            } else {
                "smith counts the projects $name finds first. If there are any, it stops and names them. Nothing is written until you answer."
            },
            "smith ${args.joinToString(" ")}",
            "Turn it $word",
            offerCommit = false,
        ) ?: return false
        val first = ports.smith.run(args, show = true)
        if (first.ok) return done(ports, name, word)
        if (on || !first.stderr.contains(HIDES_PROJECTS)) {
            ports.notice.warn("$name was not turned $word. ${FailureTextMapper.describe(first)}")
            return false
        }
        ports.prompt.plan(
            "$TITLE: turn $name off anyway?",
            "Nothing was changed. These projects disappear from smith until you turn $name on again.",
            first.stderr.trim(),
            "Turn it off anyway",
            offerCommit = false,
        ) ?: return false
        val forced = ports.smith.run(args + "--force", show = true)
        if (!forced.ok) {
            ports.notice.warn("$name was not turned off. ${FailureTextMapper.describe(forced)}")
            return false
        }
        return done(ports, name, word)
    }

    private fun done(
        ports: FlowPortsContract,
        name: String,
        word: String,
    ): Boolean {
        ports.changed()
        ports.notice.info("$name is $word. Nothing is committed.")
        return true
    }
}
