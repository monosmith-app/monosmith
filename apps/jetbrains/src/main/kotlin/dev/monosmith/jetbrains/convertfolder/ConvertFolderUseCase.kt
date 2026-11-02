package dev.monosmith.jetbrains.convertfolder

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.commitchanges.CommitChangesUseCase
import dev.monosmith.jetbrains.flowports.FlowPortsContract
import dev.monosmith.jetbrains.flowports.IdePortsClient
import dev.monosmith.jetbrains.planapply.PlanStepContract
import dev.monosmith.jetbrains.planapply.PlanThenApplyUseCase
import dev.monosmith.jetbrains.prompt.ChoiceRowContract
import dev.monosmith.jetbrains.smithrun.BackgroundFlowClient
import java.nio.file.Path

/**
 * Converts an Nx or Turborepo workspace (`smith convert`): the level first, at the full level the four
 * choices behind it, then the plan of `--dry-run` that lists the files it deletes, and only then `--yes`.
 * smith commits only when the box is ticked.
 */
object ConvertFolderUseCase {
    private const val TITLE = "MonoSmith: convert this workspace"

    /** Starts the flow in a background task. */
    fun start(
        ide: Project,
        root: Path,
    ) = BackgroundFlowClient.start(ide, TITLE) { run(IdePortsClient.of(ide, root)) { CommitChangesUseCase.start(ide, root, it) } }

    /**
     * @return true when smith converted the workspace
     */
    fun run(
        ports: FlowPortsContract,
        commitLater: (String) -> Unit,
    ): Boolean {
        val level = ask(ports, "how far to go", "The first answer changes the least.", ConvertArgumentsAlgorithm.levelRows) ?: return false
        var choices = ConvertChoicesContract(level = level)
        if (level == "full") {
            choices = choices.copy(lint = ask(ports, "lint", "", ConvertArgumentsAlgorithm.lintRows) ?: return false)
            choices = choices.copy(test = ask(ports, "tests", "", ConvertArgumentsAlgorithm.testRows) ?: return false)
            choices = choices.copy(build = ask(ports, "build", "", ConvertArgumentsAlgorithm.buildRows) ?: return false)
            choices = choices.copy(ci = ask(ports, "pipelines", "", ConvertArgumentsAlgorithm.ciRows) ?: return false)
        }
        val chosen = choices
        return PlanThenApplyUseCase.run(
            ports,
            PlanStepContract(
                title = "$TITLE?",
                planArgs = ConvertArgumentsAlgorithm.arguments(chosen, apply = false),
                header = "Nothing was changed yet. The plan lists the files smith deletes and the changes it makes.",
                okText = "Convert",
                offerCommit = true,
                apply = { ConvertArgumentsAlgorithm.arguments(chosen, apply = true) },
                done = "The workspace is converted",
                operation = "convert",
            ),
            commitLater,
        )
    }

    private fun ask(
        ports: FlowPortsContract,
        what: String,
        header: String,
        rows: List<ChoiceRowContract>,
    ): String? = ports.prompt.pickOne("$TITLE: $what", "${header.ifEmpty { "The first answer is the default." }}", rows)?.name
}
