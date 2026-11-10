package dev.monosmith.jetbrains.health

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.flowports.FlowPortsContract
import dev.monosmith.jetbrains.flowports.IdePortsClient
import dev.monosmith.jetbrains.flowports.WorkspaceChangedStore
import dev.monosmith.jetbrains.notice.NoticeButtonContract
import dev.monosmith.jetbrains.notice.NoticeClient
import dev.monosmith.jetbrains.prompt.ChoiceRowContract
import dev.monosmith.jetbrains.smithrun.BackgroundFlowClient
import dev.monosmith.jetbrains.smithrun.FailureTextMapper
import dev.monosmith.jetbrains.smithrun.OutputConsoleStore
import dev.monosmith.jetbrains.smithrun.SmithRunClient
import dev.monosmith.jetbrains.words.SmithWordsConfig.FIXWIZARD_STYLE_APPLY
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDCHOICES_SUGGESTED
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_OPEN_HEALTH
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_HEALTH_APPLY_THE_SAFE_FIXES
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_HEALTH_DOCTOR_DID_NOT_RUN
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_HEALTH_EACH_FIX_IS_SHOWN_AS
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_HEALTH_MONOSMITH_APPLY_THE_SAFE_FIXES
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_HEALTH_MONOSMITH_APPLY_THE_SAFE_FIXES_2
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_HEALTH_MONOSMITH_DOCTOR
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_HEALTH_MONOSMITH_FIX
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_HEALTH_MONOSMITH_FIX_2
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_HEALTH_NOTHING_WAS_CHANGED
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_HEALTH_OPEN_THE_HEALTH_TAB_EACH
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_HEALTH_SMITH_HAS_NO_FIX_FOR
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_HEALTH_SMITH_HAS_NO_SAFE_FIX
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_HEALTH_THE_FIX_IS_APPLIED_ANOTHER
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_HEALTH_THE_FIX_IS_APPLIED_REVIEW
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_HEALTH_THE_SAFE_FIXES_ARE_APPLIED
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_HEALTH_THIS_IS_WHAT_SMITH_WOULD
import dev.monosmith.jetbrains.words.SmithWordsConfig.say
import java.nio.file.Path

/** Reads the health of the workspace (`smith doctor --json`), fixes a check with its plan shown, and applies the safe fixes. */
object HealthUseCase {
    /**
     * Reads the checks. `smith doctor` ends non-zero while a check fails and still prints its JSON, so the
     * exit code is not the test. Blocks: call it off the UI thread.
     */
    fun read(root: Path): Result<List<DoctorFindingContract>> =
        runCatching {
            val run = SmithRunClient.run(null, root, listOf("doctor", "--json"))
            runCatching { DoctorReportJsonMapper.toFindings(run.stdout) }.getOrElse { error(FailureTextMapper.describe(run)) }
        }

    /** Runs doctor, tells the summary, and prints the checks that did not pass in the Output tab. */
    fun runDoctor(
        ide: Project,
        root: Path,
    ) = BackgroundFlowClient.start(ide, say(JETBRAINS_HEALTH_MONOSMITH_DOCTOR)) {
        val findings =
            read(root).getOrElse { return@start NoticeClient.warn(ide, say(JETBRAINS_HEALTH_DOCTOR_DID_NOT_RUN, it.message.toString())) }
        val summary = DoctorReportJsonMapper.summarize(findings)
        val console = OutputConsoleStore.consoleFor(ide)
        summary.lines.forEach { console.print("$it\n", com.intellij.execution.ui.ConsoleViewContentType.NORMAL_OUTPUT) }
        val buttons =
            listOf(
                NoticeButtonContract(say(JETBRAINS_ADDPROJECT_OPEN_HEALTH), say(JETBRAINS_HEALTH_OPEN_THE_HEALTH_TAB_EACH)) {
                    OutputConsoleStore.reveal(ide, "Health")
                },
            )
        val say = DoctorReportJsonMapper.say(summary)
        if (summary.failed == 0 && summary.warnings == 0) NoticeClient.info(ide, say) else NoticeClient.warn(ide, say, buttons)
        WorkspaceChangedStore.changed(ide)
    }

    /**
     * Fixes one failed check the way `smith doctor --fix --only` does: asks the questions of the fix, shows
     * smith's review (the diffs, the commands and the plan each command prints), and only on a yes applies it
     * with `--yes`. Nothing is decided silently: every answer is the person's.
     */
    fun fix(
        ide: Project,
        root: Path,
        finding: DoctorFindingContract,
    ) = BackgroundFlowClient.start(
        ide,
        say(JETBRAINS_HEALTH_MONOSMITH_FIX, finding.check.toString()),
    ) { runFix(IdePortsClient.of(ide, root), finding) }

    /**
     * Runs the fix of one check with the given ports.
     *
     * @return true when smith applied the fix
     */
    fun runFix(
        ports: FlowPortsContract,
        finding: DoctorFindingContract,
    ): Boolean {
        val fix =
            finding.fix ?: run {
                ports.notice.info(say(JETBRAINS_HEALTH_SMITH_HAS_NO_FIX_FOR, finding.remedy.toString()))
                return false
            }
        val answers = mutableListOf<Int>()
        for (question in fix.questions) {
            val rows =
                question.options.mapIndexed { index, label ->
                    ChoiceRowContract(
                        index.toString(),
                        label,
                        if (index ==
                            0
                        ) {
                            say(JETBRAINS_ADDCHOICES_SUGGESTED)
                        } else {
                            ""
                        },
                    )
                }
            val picked =
                ports.prompt.pickOne(say(JETBRAINS_HEALTH_MONOSMITH_FIX, finding.check.toString()), question.question, rows) ?: return false
            answers.add(picked.name.toInt() + 1)
        }
        val review = ports.smith.run(FixReviewMapper.arguments(fix, answers, apply = false))
        val shown = FixReviewMapper.review(review.stdout)
        if (shown.isEmpty()) {
            ports.notice.warn(say(JETBRAINS_HEALTH_NOTHING_WAS_CHANGED, FailureTextMapper.describe(review).toString()))
            return false
        }
        ports.prompt.plan(
            say(JETBRAINS_HEALTH_MONOSMITH_FIX_2, finding.check.toString()),
            say(JETBRAINS_HEALTH_THIS_IS_WHAT_SMITH_WOULD, fix.description.toString()),
            shown,
            say(FIXWIZARD_STYLE_APPLY),
            offerCommit = false,
        ) ?: return false
        val done = ports.smith.run(FixReviewMapper.arguments(fix, answers, apply = true), show = true)
        ports.changed()
        val incomplete = FixReviewMapper.stillFailing(done.code, done.lines)
        if (!done.ok && !incomplete) {
            ports.notice.warn(say(JETBRAINS_HEALTH_NOTHING_WAS_CHANGED, FailureTextMapper.describe(done).toString()))
            return false
        }
        val said =
            if (incomplete) {
                say(JETBRAINS_HEALTH_THE_FIX_IS_APPLIED_ANOTHER)
            } else {
                say(JETBRAINS_HEALTH_THE_FIX_IS_APPLIED_REVIEW)
            }
        ports.notice.info(said)
        return true
    }

    /** `smith doctor --fix`: shows the safe fixes as diffs and applies them after a yes. A fix with a question has its own Fix row. */
    fun fixSafe(
        ide: Project,
        root: Path,
    ) = BackgroundFlowClient.start(ide, say(JETBRAINS_HEALTH_MONOSMITH_APPLY_THE_SAFE_FIXES)) { runFixSafe(IdePortsClient.of(ide, root)) }

    /**
     * Runs the safe fixes with the given ports.
     *
     * @return true when smith applied the safe fixes
     */
    fun runFixSafe(ports: FlowPortsContract): Boolean {
        val review = ports.smith.run(listOf("doctor", "--fix"))
        val shown = FixReviewMapper.review(review.stdout)
        if (shown.isEmpty()) {
            ports.notice.info(say(JETBRAINS_HEALTH_SMITH_HAS_NO_SAFE_FIX, FailureTextMapper.describe(review).toString()))
            return false
        }
        ports.prompt.plan(
            say(JETBRAINS_HEALTH_MONOSMITH_APPLY_THE_SAFE_FIXES_2),
            say(JETBRAINS_HEALTH_EACH_FIX_IS_SHOWN_AS),
            shown,
            say(JETBRAINS_HEALTH_APPLY_THE_SAFE_FIXES),
            offerCommit = false,
        ) ?: return false
        ports.smith.run(listOf("doctor", "--fix", "--yes"), show = true)
        ports.changed()
        ports.notice.info(say(JETBRAINS_HEALTH_THE_SAFE_FIXES_ARE_APPLIED))
        return true
    }
}
