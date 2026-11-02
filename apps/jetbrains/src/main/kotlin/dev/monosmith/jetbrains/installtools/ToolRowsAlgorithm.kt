package dev.monosmith.jetbrains.installtools

import dev.monosmith.jetbrains.prompt.ChoiceRowContract

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
            val command = if (option.command.isEmpty()) "" else "Runs: ${option.command.joinToString(" ")}. "
            val admin = if (option.admin) "It needs administrator rights, and smith asks again before it runs. " else ""
            ChoiceRowContract(
                name = option.value,
                label = option.label,
                description = if (option.value == plan.default) "(default)" else "",
                detail = "$command$admin${option.hint}${if (option.note.isEmpty()) "" else " ${option.note}"}",
            )
        }

    /** The question about a command that needs administrator rights. Skipping is the first row: it is the safe answer. */
    fun adminRows(
        plan: ToolPlanContract,
        option: ToolOptionContract,
    ): List<ChoiceRowContract> =
        listOf(
            ChoiceRowContract(SKIP, "Skip ${plan.title}", "(default)", plan.consequence),
            ChoiceRowContract(MYSELF, "I will run it myself", "", "smith shows the command: ${option.command.joinToString(" ")}"),
            ChoiceRowContract(RUN_IT, "Let smith run it", "", "smith runs the command and the system asks for your password or approval."),
        )

    /** Every install command that would work, then the official page, one per line. */
    fun commands(plan: ToolPlanContract): List<String> =
        plan.options.filter { it.kind == "install" && it.command.isNotEmpty() }.map { it.command.joinToString(" ") } + plan.download
}
