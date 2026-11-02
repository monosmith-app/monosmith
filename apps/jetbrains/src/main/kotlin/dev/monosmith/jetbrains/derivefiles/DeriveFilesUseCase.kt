package dev.monosmith.jetbrains.derivefiles

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.notice.NoticeClient
import dev.monosmith.jetbrains.planapply.PlanStepContract
import dev.monosmith.jetbrains.planapply.PlanThenApplyUseCase
import dev.monosmith.jetbrains.prompt.ChoiceRowContract
import dev.monosmith.jetbrains.prompt.PromptClient
import dev.monosmith.jetbrains.smithrun.BackgroundFlowClient
import dev.monosmith.jetbrains.smithrun.FailureTextMapper
import dev.monosmith.jetbrains.smithrun.SmithRunClient
import java.nio.file.Path

/**
 * Looks at the derived files (`smith derive --check`) and lets the person act on one: explain who owns it,
 * overwrite it with the derived content, keep the hand edit, eject it (own it) or reclaim it (give it back
 * to smith). Each change shows its effect first and commits nothing unless the box is ticked.
 */
object DeriveFilesUseCase {
    private const val TITLE = "MonoSmith: derived files"
    private const val ALL = "\u0000all"

    private val effects =
        mapOf(
            "apply" to
                "smith writes every derived file that is stale or missing and records it in smith.lock. It leaves the files someone edited alone.",
            "replace" to "smith overwrites the file with the derived content. Your edit is lost; git can bring it back.",
            "adopt" to
                "smith keeps your edit as the new baseline. This works for a file smith owns by keys that still has every derived key. If the inputs change later, the file is in drift again.",
            "eject" to "The file becomes yours. smith writes \"you\" as its owner in smith.lock and never touches it again.",
            "reclaim" to "smith gives the file back to itself: it shows the diff, rewrites the file and records smith as the owner.",
        )

    private fun actionRows(file: DerivedFileContract): List<ChoiceRowContract> {
        val write = if (DeriveReportMapper.isHeldBack(file.status)) "replace" else "apply"
        return listOf(
            ChoiceRowContract("explain", "Explain", "(default)", "Say who owns this file and why it is in its state. Changes nothing."),
            ChoiceRowContract(
                write,
                if (write ==
                    "replace"
                ) {
                    "Overwrite with the derived content"
                } else {
                    "Write the derived content"
                },
                "",
                effects.getValue(write),
            ),
            ChoiceRowContract("adopt", "Keep my edit", "", effects.getValue("adopt")),
            ChoiceRowContract("eject", "Eject: own it", "", effects.getValue("eject")),
            ChoiceRowContract("reclaim", "Reclaim: give it back to smith", "", effects.getValue("reclaim")),
        )
    }

    /** Checks the derived files in the Output tab (`smith derive --check`). */
    fun check(
        ide: Project,
        root: Path,
    ) = BackgroundFlowClient.start(ide, "MonoSmith: check the derived files") {
        SmithRunClient.run(ide, root, DeriveReportMapper.arguments("check"), show = true)
    }

    /** Previews what a plain `smith derive` would write (`--dry-run`) in the Output tab. */
    fun preview(
        ide: Project,
        root: Path,
    ) = BackgroundFlowClient.start(ide, "MonoSmith: preview the derived files") {
        SmithRunClient.run(ide, root, DeriveReportMapper.arguments("preview"), show = true)
    }

    /** Lets the person act on the derived files that need action. */
    fun start(
        ide: Project,
        root: Path,
    ) = BackgroundFlowClient.start(ide, TITLE) {
        val run = SmithRunClient.run(ide, root, DeriveReportMapper.arguments("check"))
        val files = DeriveReportMapper.files(run.stdout)
        if (files.isEmpty()) {
            return@start if (run.ok) {
                NoticeClient.info(ide, "Every derived file is current.")
            } else {
                NoticeClient.warn(ide, FailureTextMapper.describe(run))
            }
        }
        val rows =
            listOf(
                ChoiceRowContract(
                    ALL,
                    "Every file that is stale or missing",
                    "(default)",
                    "A plain smith derive: it leaves the files someone edited alone.",
                ),
            ) +
                files.map { ChoiceRowContract(it.path, it.path, "${it.status} · ${it.renderer}", it.detail) }
        val picked =
            PromptClient
                .pickOne(
                    ide,
                    TITLE,
                    "${files.size} derived files need action. Pick one. The first answer is the default.",
                    rows,
                )?.name
                ?: return@start
        if (picked == ALL) return@start applyAll(ide, root)
        val file = files.first { it.path == picked }
        val action =
            PromptClient.pickOne(ide, "$TITLE: ${file.path}", "${file.status}: ${file.detail}", actionRows(file))?.name ?: return@start
        if (action == "explain") {
            SmithRunClient.run(ide, root, DeriveReportMapper.arguments("explain", file.path), show = true)
            return@start
        }
        PlanThenApplyUseCase.run(
            ide,
            root,
            PlanStepContract(
                title = "MonoSmith: $action ${file.path}?",
                planArgs = null,
                fixedPlan = "smith ${DeriveReportMapper.arguments(action, file.path).joinToString(" ")}",
                header = "${effects.getValue(action)} By default smith leaves the files uncommitted: review them with the Commit window.",
                okText = "Do it",
                offerCommit = true,
                apply = { DeriveReportMapper.arguments(action, file.path) },
                done = "${file.path}: $action is done. Commit it together with smith.lock",
                operation = "derive",
            ),
        )
    }

    private fun applyAll(
        ide: Project,
        root: Path,
    ) {
        PlanThenApplyUseCase.run(
            ide,
            root,
            PlanStepContract(
                title = "MonoSmith: write the derived files?",
                planArgs = DeriveReportMapper.arguments("preview"),
                header = "${effects.getValue("apply")} By default smith leaves the files uncommitted.",
                okText = "Apply",
                offerCommit = true,
                apply = { DeriveReportMapper.arguments("apply") },
                done = "The derived files are written. Commit them together with smith.lock",
                operation = "derive",
            ),
        )
    }
}
