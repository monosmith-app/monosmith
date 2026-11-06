package dev.monosmith.jetbrains.installtools

import dev.monosmith.jetbrains.prompt.ChoiceRowContract
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDCHOICES_DEFAULT
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_INSTALLTOOLS_IT_NEEDS_ADMINISTRATOR_RIGHTS_AND
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_INSTALLTOOLS_RUNS
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_INSTALLTOOLS_SKIP
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_INSTALLTOOLS_SMITH_RUNS_THE_COMMAND_AND
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_INSTALLTOOLS_SMITH_SHOWS_THE_COMMAND
import dev.monosmith.jetbrains.words.SmithWordsConfig.TOOLSDIALOG_ANSWER_MYSELF
import dev.monosmith.jetbrains.words.SmithWordsConfig.TOOLSDIALOG_ANSWER_SMITH
import dev.monosmith.jetbrains.words.SmithWordsConfig.say

/** The lines of the questions about a tool, built from the options smith offers. */
object ToolRowsAlgorithm {
    /** The name of the row that lets smith run a command that needs administrator rights. */
    const val RUN_IT = "run"

    /** The name of the row that shows the command to run by hand. */
    const val MYSELF = "myself"

    /** The name of the row that skips the tool. */
    const val SKIP = "skip"

    /** The lines of "how do you want to get this tool?": the marked default first, each with its command and consequence. */
    fun wayRows(plan: ToolPlanContract): List<ChoiceRowContract> =
        plan.options.sortedByDescending { it.value == plan.default }.map { option ->
            val command =
                if (option.command.isEmpty()) "" else say(JETBRAINS_INSTALLTOOLS_RUNS, option.command.joinToString(" ").toString())
            val admin = if (option.admin) say(JETBRAINS_INSTALLTOOLS_IT_NEEDS_ADMINISTRATOR_RIGHTS_AND) else ""
            ChoiceRowContract(
                name = option.value,
                label = option.label,
                description = if (option.value == plan.default) say(JETBRAINS_ADDCHOICES_DEFAULT) else "",
                detail = "$command$admin${option.hint}${if (option.note.isEmpty()) "" else " ${option.note}"}",
            )
        }

    /** The question about a command that needs administrator rights. Skipping is the first row: it is the safe answer. */
    fun adminRows(
        plan: ToolPlanContract,
        option: ToolOptionContract,
    ): List<ChoiceRowContract> =
        listOf(
            ChoiceRowContract(
                SKIP,
                say(JETBRAINS_INSTALLTOOLS_SKIP, plan.title.toString()),
                say(JETBRAINS_ADDCHOICES_DEFAULT),
                plan.consequence,
            ),
            ChoiceRowContract(
                MYSELF,
                say(TOOLSDIALOG_ANSWER_MYSELF),
                "",
                say(JETBRAINS_INSTALLTOOLS_SMITH_SHOWS_THE_COMMAND, option.command.joinToString(" ").toString()),
            ),
            ChoiceRowContract(RUN_IT, say(TOOLSDIALOG_ANSWER_SMITH), "", say(JETBRAINS_INSTALLTOOLS_SMITH_RUNS_THE_COMMAND_AND)),
        )

    /** Every install command that would work, then the official page, one per line. */
    fun commands(plan: ToolPlanContract): List<String> =
        plan.options.filter { it.kind == "install" && it.command.isNotEmpty() }.map { it.command.joinToString(" ") } + plan.download
}
