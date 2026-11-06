package dev.monosmith.jetbrains.mcpinstall

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.flowports.FlowPortsContract
import dev.monosmith.jetbrains.flowports.IdePortsClient
import dev.monosmith.jetbrains.planapply.PlanStepContract
import dev.monosmith.jetbrains.planapply.PlanThenApplyUseCase
import dev.monosmith.jetbrains.prompt.ChoiceRowContract
import dev.monosmith.jetbrains.smithrun.BackgroundFlowClient
import dev.monosmith.jetbrains.words.SmithWordsConfig.GUIDE_PARITY_FORMS_CLAUDE_CODE
import dev.monosmith.jetbrains.words.SmithWordsConfig.GUIDE_PARITY_FORMS_VS_CODE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_AGENTFILE_WRITE_THE_FILES
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_MCPINSTALL_MONOSMITH_CONNECT_AN_ASSISTANT
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_MCPINSTALL_NOTHING_WAS_WRITTEN_YET
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_MCPINSTALL_SMITH_IS_CONNECTED
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_MCPINSTALL_SMITH_IT_WRITES_THE_FILE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_MCPINSTALL_TICK_THE_ASSISTANTS_TO_CONNECT
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_MCPINSTALL_WRITES_MCP_JSON
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_MCPINSTALL_WRITES_VSCODE_MCP_JSON
import dev.monosmith.jetbrains.words.SmithWordsConfig.say
import java.nio.file.Path

/**
 * Connects smith to the coding assistants that read an MCP server file (`smith mcp install`): Claude Code
 * (.mcp.json) and VS Code (.vscode/mcp.json). The person ticks the assistants; smith merges into an
 * existing file, keeps the other servers and never overwrites a smith entry the team edited. smith has no
 * `--commit` here, so nothing is committed.
 */
object McpInstallUseCase {
    private val TITLE = say(JETBRAINS_MCPINSTALL_MONOSMITH_CONNECT_AN_ASSISTANT)

    /** The two assistants, both unticked: the person decides which files smith writes. */
    val rows =
        listOf(
            ChoiceRowContract("--claude", say(GUIDE_PARITY_FORMS_CLAUDE_CODE), "", say(JETBRAINS_MCPINSTALL_WRITES_MCP_JSON)),
            ChoiceRowContract("--vscode", say(GUIDE_PARITY_FORMS_VS_CODE), "", say(JETBRAINS_MCPINSTALL_WRITES_VSCODE_MCP_JSON)),
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
        val flags = ports.prompt.pickMany(TITLE, say(JETBRAINS_MCPINSTALL_TICK_THE_ASSISTANTS_TO_CONNECT), rows, "Next")
        if (flags.isNullOrEmpty()) return false
        val args = listOf("mcp", "install") + flags
        val plan =
            say(JETBRAINS_MCPINSTALL_SMITH_IT_WRITES_THE_FILE, args.joinToString(" ").toString())
        return PlanThenApplyUseCase.run(
            ports,
            PlanStepContract(
                title = "$TITLE?",
                planArgs = null,
                fixedPlan = plan,
                header = say(JETBRAINS_MCPINSTALL_NOTHING_WAS_WRITTEN_YET),
                okText = say(JETBRAINS_AGENTFILE_WRITE_THE_FILES),
                offerCommit = false,
                apply = { args + "--yes" },
                done = say(JETBRAINS_MCPINSTALL_SMITH_IS_CONNECTED),
            ),
        ) {}
    }
}
