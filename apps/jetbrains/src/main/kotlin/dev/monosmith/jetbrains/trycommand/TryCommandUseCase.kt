package dev.monosmith.jetbrains.trycommand

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.flowports.FlowPortsContract
import dev.monosmith.jetbrains.flowports.IdePortsClient
import dev.monosmith.jetbrains.smithrun.BackgroundFlowClient
import dev.monosmith.jetbrains.smithrun.FailureTextMapper
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_HEALTH_NOTHING_WAS_CHANGED
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_TRYCOMMAND_IT_FAILED_IN_THE_COPY
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_TRYCOMMAND_IT_PASSED_IN_A_COPY
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_TRYCOMMAND_MONOSMITH_TRY_A_COMMAND
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_TRYCOMMAND_MONOSMITH_TRY_A_COMMAND_IN
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_TRYCOMMAND_MONOSMITH_TRY_SMITH
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_TRYCOMMAND_NOTHING_WAS_TRIED
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_TRYCOMMAND_RUN_IT_FOR_REAL
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_TRYCOMMAND_THE_COMMAND_RAN_FOR_REAL
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_TRYCOMMAND_THE_SMITH_COMMAND_TO_TRY
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_TRYCOMMAND_TYPE_THE_COMMAND_TO_TRY
import dev.monosmith.jetbrains.words.SmithWordsConfig.say
import java.nio.file.Path

/**
 * Runs a smith command that changes files in a temporary copy (`smith try <command> --json`), shows what
 * it did and which checks passed, and, only after a try that passed and a yes, runs it for real
 * (`smith try <command> --yes`). The try itself leaves the files of the person as they are.
 */
object TryCommandUseCase {
    /** Starts the flow in a background task. */
    fun start(
        ide: Project,
        root: Path,
    ) = BackgroundFlowClient.start(ide, say(JETBRAINS_TRYCOMMAND_MONOSMITH_TRY_A_COMMAND)) { run(IdePortsClient.of(ide, root)) }

    /**
     * Runs the flow with the given ports.
     *
     * @return true when the command ran for real
     */
    fun run(ports: FlowPortsContract): Boolean {
        val typed =
            ports.prompt.text(
                say(JETBRAINS_TRYCOMMAND_MONOSMITH_TRY_A_COMMAND_IN),
                say(JETBRAINS_TRYCOMMAND_THE_SMITH_COMMAND_TO_TRY),
            ) {
                if (it.isEmpty()) say(JETBRAINS_TRYCOMMAND_TYPE_THE_COMMAND_TO_TRY) else null
            } ?: return false
        val command = typed.removePrefix("smith ").split(Regex("\\s+")).filter { it.isNotEmpty() }
        val tried = ports.smith.run(listOf("try") + command + "--json")
        val shown =
            runCatching {
                TryResultMapper.toPage(tried.stdout)
            }.getOrElse {
                ports.notice.warn(say(JETBRAINS_TRYCOMMAND_NOTHING_WAS_TRIED, FailureTextMapper.describe(tried).toString()))
                return false
            }
        val answer =
            ports.prompt.plan(
                say(JETBRAINS_TRYCOMMAND_MONOSMITH_TRY_SMITH, command.joinToString(" ").toString()),
                if (shown.passed) {
                    say(JETBRAINS_TRYCOMMAND_IT_PASSED_IN_A_COPY)
                } else {
                    say(JETBRAINS_TRYCOMMAND_IT_FAILED_IN_THE_COPY)
                },
                shown.page,
                if (shown.passed) say(JETBRAINS_TRYCOMMAND_RUN_IT_FOR_REAL) else "Close",
                offerCommit = false,
            )
        if (answer == null || !shown.passed) return false
        val done = ports.smith.run(listOf("try") + command + "--yes", show = true)
        ports.changed()
        if (!done.ok) {
            ports.notice.warn(say(JETBRAINS_HEALTH_NOTHING_WAS_CHANGED, FailureTextMapper.describe(done).toString()))
            return false
        }
        ports.notice.info(say(JETBRAINS_TRYCOMMAND_THE_COMMAND_RAN_FOR_REAL))
        return true
    }
}
