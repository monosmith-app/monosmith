package dev.monosmith.jetbrains.mcpinstall

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.flowports.FlowPortsContract
import dev.monosmith.jetbrains.flowports.IdePortsClient
import dev.monosmith.jetbrains.planapply.PlanStepContract
import dev.monosmith.jetbrains.planapply.PlanThenApplyUseCase
import dev.monosmith.jetbrains.prompt.ChoiceRowContract
import dev.monosmith.jetbrains.smithrun.BackgroundFlowClient
import java.nio.file.Path

/**
 * Connects smith to the coding assistants that read an MCP server file (`smith mcp install`): Claude Code
 * (.mcp.json) and VS Code (.vscode/mcp.json). The person ticks the assistants; smith merges into an
 * existing file, keeps the other servers and never overwrites a smith entry the team edited. smith has no
 * `--commit` here, so nothing is committed.
 */
object McpInstallUseCase {
    private const val TITLE = "MonoSmith: connect an assistant"

    /** The two assistants, both unticked: the person decides which files smith writes. */
    val rows =
        listOf(
            ChoiceRowContract("--claude", "Claude Code", "", "Writes .mcp.json."),
            ChoiceRowContract("--vscode", "VS Code", "", "Writes .vscode/mcp.json."),
        )

    /** Starts the flow in a background task. */
    fun start(
        ide: Project,
        root: Path,
    ) = BackgroundFlowClient.start(ide, TITLE) { run(IdePortsClient.of(ide, root)) }

    /**
     * @return true when smith wrote the files
     */
    fun run(ports: FlowPortsContract): Boolean {
        val flags = ports.prompt.pickMany(TITLE, "Tick the assistants to connect.", rows, "Next")
        if (flags.isNullOrEmpty()) return false
        val args = listOf("mcp", "install") + flags
        val plan =
            "smith ${args.joinToString(" ")}\n\n" +
                "It writes the file that lets the assistant start smith's MCP server, merges into a file that exists " +
                "and keeps your other servers."
        return PlanThenApplyUseCase.run(
            ports,
            PlanStepContract(
                title = "$TITLE?",
                planArgs = null,
                fixedPlan = plan,
                header = "Nothing was written yet.",
                okText = "Write the files",
                offerCommit = false,
                apply = { args + "--yes" },
                done = "smith is connected",
            ),
        ) {}
    }
}
