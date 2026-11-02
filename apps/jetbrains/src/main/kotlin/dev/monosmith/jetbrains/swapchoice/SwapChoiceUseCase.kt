package dev.monosmith.jetbrains.swapchoice

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.addchoices.ChoiceRowsAlgorithm
import dev.monosmith.jetbrains.addchoices.ChoicesJsonMapper
import dev.monosmith.jetbrains.notice.NoticeClient
import dev.monosmith.jetbrains.planapply.PlanStepContract
import dev.monosmith.jetbrains.planapply.PlanThenApplyUseCase
import dev.monosmith.jetbrains.prompt.PromptClient
import dev.monosmith.jetbrains.smithrun.BackgroundFlowClient
import dev.monosmith.jetbrains.smithrun.FailureTextMapper
import dev.monosmith.jetbrains.smithrun.SmithRunClient
import dev.monosmith.jetbrains.workspace.ListProjectsUseCase
import dev.monosmith.jetbrains.workspace.PickProjectUseCase
import java.nio.file.Path

/**
 * Swaps one choice of a project, such as its test runner: asks the project, the decision and the answer
 * from `smith choices --json`, shows the plan of `smith swap --dry-run`, and only on a yes runs it.
 */
object SwapChoiceUseCase {
    /**
     * Starts the flow in a background task.
     *
     * @param project the project; null asks which
     */
    fun start(
        ide: Project,
        root: Path,
        project: String?,
    ) = BackgroundFlowClient.start(ide, "MonoSmith: swap a choice") {
        val nodes = ListProjectsUseCase.list(root).getOrElse { return@start NoticeClient.warn(ide, "Nothing was changed. ${it.message}") }
        val node =
            if (project !=
                null
            ) {
                nodes.firstOrNull { it.name == project } ?: return@start
            } else {
                PickProjectUseCase.pick(ide, nodes, "swap a choice")
                    ?: return@start
            }
        val listed = SmithRunClient.run(ide, root, listOf("choices", "--json"))
        if (!listed.ok) return@start NoticeClient.warn(ide, "Nothing was changed. ${FailureTextMapper.describe(listed)}")
        val choices =
            runCatching { ChoicesJsonMapper.toChoices(listed.stdout) }
                .getOrElse {
                    return@start NoticeClient.warn(
                        ide,
                        "smith printed its choices in a form this plugin does not read. Update the plugin, or run smith swap in a terminal.",
                    )
                }
        val axes = ChoiceRowsAlgorithm.byAxis(choices, node.language)
        if (axes.isEmpty()) return@start NoticeClient.info(ide, "${node.name} has no choice to swap.")
        val layers = emptyMap<String, String>()
        val axis =
            PromptClient
                .pickOne(
                    ide,
                    "MonoSmith: swap a choice of ${node.name}",
                    "Which decision do you want to change?",
                    SwapArgumentsAlgorithm.axisRows(axes, layers),
                )?.name
                ?: return@start
        val group = axes.first { it.axis == axis }
        val choice =
            PromptClient
                .pickOne(
                    ide,
                    "MonoSmith: swap the $axis of ${node.name}",
                    "Pick the answer ${node.name} gets. The default is marked.",
                    SwapArgumentsAlgorithm.choiceRows(group),
                )?.name
                ?: return@start
        PlanThenApplyUseCase.run(
            ide,
            root,
            PlanStepContract(
                title = "MonoSmith: swap the $axis of ${node.name} to $choice?",
                planArgs = SwapArgumentsAlgorithm.arguments(node.name, axis, choice, apply = false),
                header = "Nothing was changed yet. By default smith leaves the files uncommitted.",
                okText = "Swap",
                offerCommit = true,
                apply = { SwapArgumentsAlgorithm.arguments(node.name, axis, choice, apply = true) },
                done = "The $axis of ${node.name} is now $choice",
                operation = "swap ${node.name} $axis $choice",
            ),
        )
    }
}
