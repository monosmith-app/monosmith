package dev.monosmith.jetbrains.flowports

import dev.monosmith.jetbrains.notice.NoticeButtonContract
import dev.monosmith.jetbrains.prompt.ChoiceRowContract
import dev.monosmith.jetbrains.prompt.PlanAnswerContract
import dev.monosmith.jetbrains.smithrun.SmithRunContract

/**
 * Ports for a unit test of a flow. Smith answers from [answers] (the exact command line, else the first
 * key that starts it, else a success with no output); the person answers from the queues, in order, and a
 * missing answer is a Cancel. Every call and message is recorded.
 */
class FakePortsHarness(
    val answers: MutableMap<String, SmithRunContract> = mutableMapOf(),
) {
    val calls = mutableListOf<List<String>>()
    val shown = mutableListOf<Boolean>()
    val infos = mutableListOf<String>()
    val warnings = mutableListOf<String>()
    val buttons = mutableListOf<List<NoticeButtonContract>>()
    val plansSeen = mutableListOf<String>()
    val rowsSeen = mutableListOf<List<ChoiceRowContract>>()
    val headersSeen = mutableListOf<String>()
    val ones = ArrayDeque<String?>()
    val manies = ArrayDeque<List<String>?>()
    val plans = ArrayDeque<PlanAnswerContract?>()
    val texts = ArrayDeque<String?>()
    var changedCount = 0

    val ports =
        FlowPortsContract(
            smith =
                object : SmithPort {
                    override fun run(
                        args: List<String>,
                        show: Boolean,
                    ): SmithRunContract {
                        calls.add(args)
                        shown.add(show)
                        val line = args.joinToString(" ")
                        return answers[line] ?: answers.entries.firstOrNull { line.startsWith(it.key) }?.value
                            ?: SmithRunContract(0, "", "")
                    }
                },
            prompt =
                object : PromptPort {
                    override fun pickOne(
                        title: String,
                        header: String,
                        rows: List<ChoiceRowContract>,
                        okText: String,
                    ): ChoiceRowContract? {
                        rowsSeen.add(rows)
                        headersSeen.add(header)
                        val name = ones.removeFirstOrNull() ?: return null
                        return rows.first { it.name == name }
                    }

                    override fun pickMany(
                        title: String,
                        header: String,
                        rows: List<ChoiceRowContract>,
                        okText: String,
                    ): List<String>? {
                        rowsSeen.add(rows)
                        headersSeen.add(header)
                        return manies.removeFirstOrNull()
                    }

                    override fun plan(
                        title: String,
                        header: String,
                        plan: String,
                        okText: String,
                        offerCommit: Boolean,
                        checks: List<String>,
                        others: List<String>,
                    ): PlanAnswerContract? {
                        plansSeen.add(plan)
                        return plans.removeFirstOrNull()
                    }

                    override fun text(
                        title: String,
                        question: String,
                        initial: String,
                        problem: (String) -> String?,
                    ): String? = texts.removeFirstOrNull()
                },
            notice =
                object : NoticePort {
                    override fun info(
                        message: String,
                        buttons: List<NoticeButtonContract>,
                    ) {
                        infos.add(message)
                        this@FakePortsHarness.buttons.add(buttons)
                    }

                    override fun warn(
                        message: String,
                        buttons: List<NoticeButtonContract>,
                    ) {
                        warnings.add(message)
                        this@FakePortsHarness.buttons.add(buttons)
                    }
                },
            changed = { changedCount++ },
        )

    /** A plan answer: the main button, with the commit box as given. */
    fun yes(commit: Boolean = false) = PlanAnswerContract(commit, emptyList(), -1)

    /** A success of smith that printed [stdout]. */
    fun ok(stdout: String) = SmithRunContract(0, stdout, "")

    /** A failure of smith that said [stderr]. */
    fun fail(stderr: String) = SmithRunContract(1, "", stderr)
}
