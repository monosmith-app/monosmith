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
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_HEALTH_NOTHING_WAS_CHANGED
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_SWAPCHOICE_HAS_NO_CHOICE_TO_SWAP
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_SWAPCHOICE_MONOSMITH_SWAP_A_CHOICE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_SWAPCHOICE_MONOSMITH_SWAP_A_CHOICE_OF
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_SWAPCHOICE_MONOSMITH_SWAP_THE_OF
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_SWAPCHOICE_MONOSMITH_SWAP_THE_OF_TO
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_SWAPCHOICE_NOTHING_WAS_CHANGED
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_SWAPCHOICE_NOTHING_WAS_CHANGED_YET_BY
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_SWAPCHOICE_PICK_THE_ANSWER_GETS_THE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_SWAPCHOICE_SMITH_PRINTED_ITS_CHOICES_IN
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_SWAPCHOICE_SWAP
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_SWAPCHOICE_SWAP_A_CHOICE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_SWAPCHOICE_THE_OF_IS_NOW
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_SWAPCHOICE_WHICH_DECISION_DO_YOU_WANT
import dev.monosmith.jetbrains.words.SmithWordsConfig.say
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
    ) = BackgroundFlowClient.start(ide, say(JETBRAINS_SWAPCHOICE_MONOSMITH_SWAP_A_CHOICE)) {
        val nodes =
            ListProjectsUseCase
                .list(
                    root,
                ).getOrElse { return@start NoticeClient.warn(ide, say(JETBRAINS_SWAPCHOICE_NOTHING_WAS_CHANGED, it.message.toString())) }
        val node =
            if (project !=
                null
            ) {
                nodes.firstOrNull { it.name == project } ?: return@start
            } else {
                PickProjectUseCase.pick(ide, nodes, say(JETBRAINS_SWAPCHOICE_SWAP_A_CHOICE))
                    ?: return@start
            }
        val listed = SmithRunClient.run(ide, root, listOf("choices", "--json"))
        if (!listed.ok) {
            return@start NoticeClient.warn(
                ide,
                say(JETBRAINS_HEALTH_NOTHING_WAS_CHANGED, FailureTextMapper.describe(listed).toString()),
            )
        }
        val choices =
            runCatching { ChoicesJsonMapper.toChoices(listed.stdout) }
                .getOrElse {
                    return@start NoticeClient.warn(
                        ide,
                        say(JETBRAINS_SWAPCHOICE_SMITH_PRINTED_ITS_CHOICES_IN),
                    )
                }
        val axes = ChoiceRowsAlgorithm.byAxis(choices, node.language)
        if (axes.isEmpty()) return@start NoticeClient.info(ide, say(JETBRAINS_SWAPCHOICE_HAS_NO_CHOICE_TO_SWAP, node.name.toString()))
        val layers = emptyMap<String, String>()
        val axis =
            PromptClient
                .pickOne(
                    ide,
                    say(JETBRAINS_SWAPCHOICE_MONOSMITH_SWAP_A_CHOICE_OF, node.name.toString()),
                    say(JETBRAINS_SWAPCHOICE_WHICH_DECISION_DO_YOU_WANT),
                    SwapArgumentsAlgorithm.axisRows(axes, layers),
                )?.name
                ?: return@start
        val group = axes.first { it.axis == axis }
        val choice =
            PromptClient
                .pickOne(
                    ide,
                    say(JETBRAINS_SWAPCHOICE_MONOSMITH_SWAP_THE_OF, axis.toString(), node.name.toString()),
                    say(JETBRAINS_SWAPCHOICE_PICK_THE_ANSWER_GETS_THE, node.name.toString()),
                    SwapArgumentsAlgorithm.choiceRows(group),
                )?.name
                ?: return@start
        PlanThenApplyUseCase.run(
            ide,
            root,
            PlanStepContract(
                title = say(JETBRAINS_SWAPCHOICE_MONOSMITH_SWAP_THE_OF_TO, axis.toString(), node.name.toString(), choice.toString()),
                planArgs = SwapArgumentsAlgorithm.arguments(node.name, axis, choice, apply = false),
                header = say(JETBRAINS_SWAPCHOICE_NOTHING_WAS_CHANGED_YET_BY),
                okText = say(JETBRAINS_SWAPCHOICE_SWAP),
                offerCommit = true,
                apply = { SwapArgumentsAlgorithm.arguments(node.name, axis, choice, apply = true) },
                done = say(JETBRAINS_SWAPCHOICE_THE_OF_IS_NOW, axis.toString(), node.name.toString(), choice.toString()),
                operation = "swap ${node.name} $axis $choice",
            ),
        )
    }
}
