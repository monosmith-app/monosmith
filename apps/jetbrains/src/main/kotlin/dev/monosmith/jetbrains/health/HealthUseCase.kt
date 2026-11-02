package dev.monosmith.jetbrains.health

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.flowports.FlowPortsContract
import dev.monosmith.jetbrains.flowports.IdePortsClient
import dev.monosmith.jetbrains.notice.NoticeButtonContract
import dev.monosmith.jetbrains.notice.NoticeClient
import dev.monosmith.jetbrains.prompt.ChoiceRowContract
import dev.monosmith.jetbrains.smithrun.BackgroundFlowClient
import dev.monosmith.jetbrains.smithrun.FailureTextMapper
import dev.monosmith.jetbrains.smithrun.OutputConsoleStore
import dev.monosmith.jetbrains.smithrun.SmithRunClient
import dev.monosmith.jetbrains.workspace.WorkspaceChangedStore
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
    ) = BackgroundFlowClient.start(ide, "MonoSmith: doctor") {
        val findings = read(root).getOrElse { return@start NoticeClient.warn(ide, "Doctor did not run. ${it.message}") }
        val summary = DoctorReportJsonMapper.summarize(findings)
        val console = OutputConsoleStore.consoleFor(ide)
        summary.lines.forEach { console.print("$it\n", com.intellij.execution.ui.ConsoleViewContentType.NORMAL_OUTPUT) }
        val buttons =
            listOf(
                NoticeButtonContract("Open Health", "Open the Health tab: each check with a fix has a Fix row.") {
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
    ) = BackgroundFlowClient.start(ide, "MonoSmith: fix ${finding.check}") { runFix(IdePortsClient.of(ide, root), finding) }

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
                ports.notice.info("smith has no fix for this check. ${finding.remedy}")
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
                            "(suggested)"
                        } else {
                            ""
                        },
                    )
                }
            val picked = ports.prompt.pickOne("MonoSmith: fix ${finding.check}", question.question, rows) ?: return false
            answers.add(picked.name.toInt() + 1)
        }
        val review = ports.smith.run(FixReviewMapper.arguments(fix, answers, apply = false))
        val shown = FixReviewMapper.review(review.stdout)
        if (shown.isEmpty()) {
            ports.notice.warn("Nothing was changed. ${FailureTextMapper.describe(review)}")
            return false
        }
        ports.prompt.plan(
            "MonoSmith: fix ${finding.check}?",
            "${fix.description}. This is what smith would change. Nothing is changed until you apply it. Nothing is committed.",
            shown,
            "Apply",
            offerCommit = false,
        ) ?: return false
        val done = ports.smith.run(FixReviewMapper.arguments(fix, answers, apply = true), show = true)
        ports.changed()
        val incomplete = FixReviewMapper.stillFailing(done.code, done.lines)
        if (!done.ok && !incomplete) {
            ports.notice.warn("Nothing was changed. ${FailureTextMapper.describe(done)}")
            return false
        }
        val said =
            if (incomplete) {
                "The fix is applied. Another check still fails."
            } else {
                "The fix is applied. Review it with the Commit window."
            }
        ports.notice.info(said)
        return true
    }

    /** `smith doctor --fix`: shows the safe fixes as diffs and applies them after a yes. A fix with a question has its own Fix row. */
    fun fixSafe(
        ide: Project,
        root: Path,
    ) = BackgroundFlowClient.start(ide, "MonoSmith: apply the safe fixes") { runFixSafe(IdePortsClient.of(ide, root)) }

    /**
     * Runs the safe fixes with the given ports.
     *
     * @return true when smith applied the safe fixes
     */
    fun runFixSafe(ports: FlowPortsContract): Boolean {
        val review = ports.smith.run(listOf("doctor", "--fix"))
        val shown = FixReviewMapper.review(review.stdout)
        if (shown.isEmpty()) {
            ports.notice.info("smith has no safe fix to apply. ${FailureTextMapper.describe(review)}")
            return false
        }
        ports.prompt.plan(
            "MonoSmith: apply the safe fixes?",
            "Each fix is shown as a diff. Nothing is changed until you apply. A fix with a question has its own Fix row in the Health tab. Nothing is committed.",
            shown,
            "Apply the safe fixes",
            offerCommit = false,
        ) ?: return false
        ports.smith.run(listOf("doctor", "--fix", "--yes"), show = true)
        ports.changed()
        ports.notice.info("The safe fixes are applied. Review them with the Commit window.")
        return true
    }
}
