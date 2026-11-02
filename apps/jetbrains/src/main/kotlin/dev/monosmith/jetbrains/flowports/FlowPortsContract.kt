package dev.monosmith.jetbrains.flowports

import dev.monosmith.jetbrains.notice.NoticeButtonContract
import dev.monosmith.jetbrains.prompt.ChoiceRowContract
import dev.monosmith.jetbrains.prompt.PlanAnswerContract
import dev.monosmith.jetbrains.smithrun.SmithRunContract

/** Runs smith for a flow. The IDE one runs the installed smith; a test gives a fake that answers from a script. */
interface SmithPort {
    /** Runs `smith <args>` and waits; [show] prints the command and its output in the Output tab. */
    fun run(
        args: List<String>,
        show: Boolean = false,
    ): SmithRunContract
}

/** Asks the person. The IDE one opens dialogs; a test gives a fake that answers from a script. Every method returns null on Cancel. */
interface PromptPort {
    /** One answer; the first row is the default. */
    fun pickOne(
        title: String,
        header: String,
        rows: List<ChoiceRowContract>,
        okText: String = "Next",
    ): ChoiceRowContract?

    /** Any number of answers; the names ticked. */
    fun pickMany(
        title: String,
        header: String,
        rows: List<ChoiceRowContract>,
        okText: String = "Next",
    ): List<String>?

    /** A plan and a question, with the unticked commit box when [offerCommit]. */
    fun plan(
        title: String,
        header: String,
        plan: String,
        okText: String,
        offerCommit: Boolean,
        checks: List<String> = emptyList(),
        others: List<String> = emptyList(),
    ): PlanAnswerContract?

    /** A line of text; [problem] says what is wrong with it, or null. */
    fun text(
        title: String,
        question: String,
        initial: String = "",
        problem: (String) -> String? = { null },
    ): String?
}

/** Tells the person what happened, with at most three short buttons (the rest go behind More…). */
interface NoticePort {
    /** Tells good news. */
    fun info(
        message: String,
        buttons: List<NoticeButtonContract> = emptyList(),
    )

    /** Tells that nothing was changed, or that something needs attention. */
    fun warn(
        message: String,
        buttons: List<NoticeButtonContract> = emptyList(),
    )
}

/**
 * What a flow needs from the IDE, as three small ports and a hook. A flow written against these runs in
 * the IDE and in a unit test.
 *
 * @property changed tells the views that the workspace changed
 */
data class FlowPortsContract(
    val smith: SmithPort,
    val prompt: PromptPort,
    val notice: NoticePort,
    val changed: () -> Unit = {},
)
