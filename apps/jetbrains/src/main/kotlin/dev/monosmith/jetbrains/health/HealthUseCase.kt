package dev.monosmith.jetbrains.health

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.notice.NoticeButtonContract
import dev.monosmith.jetbrains.notice.NoticeClient
import dev.monosmith.jetbrains.prompt.ChoiceRowContract
import dev.monosmith.jetbrains.prompt.PromptClient
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
    ) = BackgroundFlowClient.start(ide, "MonoSmith: fix ${finding.check}") {
        val fix = finding.fix ?: return@start NoticeClient.info(ide, "smith has no fix for this check. ${finding.remedy}")
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
            val picked = PromptClient.pickOne(ide, "MonoSmith: fix ${finding.check}", question.question, rows) ?: return@start
            answers.add(picked.name.toInt() + 1)
        }
        val review = SmithRunClient.run(ide, root, FixReviewMapper.arguments(fix, answers, apply = false))
        val shown = FixReviewMapper.review(review.stdout)
        if (shown.isEmpty()) return@start NoticeClient.warn(ide, "Nothing was changed. ${FailureTextMapper.describe(review)}")
        PromptClient.plan(
            ide,
            "MonoSmith: fix ${finding.check}?",
            "${fix.description}. This is what smith would change. Nothing is changed until you apply it. Nothing is committed.",
            shown,
            "Apply",
            offerCommit = false,
        ) ?: return@start
        val done = SmithRunClient.run(ide, root, FixReviewMapper.arguments(fix, answers, apply = true), show = true)
        WorkspaceChangedStore.changed(ide)
        val incomplete = FixReviewMapper.stillFailing(done.code, done.lines)
        if (done.ok || incomplete) {
            NoticeClient.info(
                ide,
                if (incomplete) {
                    "The fix is applied. Another check still fails."
                } else {
                    "The fix is applied. Review it with the Commit window."
                },
            )
        } else {
            NoticeClient.warn(ide, "Nothing was changed. ${FailureTextMapper.describe(done)}")
        }
    }

    /** `smith doctor --fix`: shows the safe fixes as diffs and applies them after a yes. A fix with a question has its own Fix row. */
    fun fixSafe(
        ide: Project,
        root: Path,
    ) = BackgroundFlowClient.start(ide, "MonoSmith: apply the safe fixes") {
        val review = SmithRunClient.run(ide, root, listOf("doctor", "--fix"))
        val shown = FixReviewMapper.review(review.stdout)
        if (shown.isEmpty()) return@start NoticeClient.info(ide, "smith has no safe fix to apply. ${FailureTextMapper.describe(review)}")
        PromptClient.plan(
            ide,
            "MonoSmith: apply the safe fixes?",
            "Each fix is shown as a diff. Nothing is changed until you apply. A fix with a question has its own Fix row in the Health tab. Nothing is committed.",
            shown,
            "Apply the safe fixes",
            offerCommit = false,
        ) ?: return@start
        SmithRunClient.run(ide, root, listOf("doctor", "--fix", "--yes"), show = true)
        WorkspaceChangedStore.changed(ide)
        NoticeClient.info(ide, "The safe fixes are applied. Review them with the Commit window.")
    }
}
