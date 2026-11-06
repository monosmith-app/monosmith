package dev.monosmith.jetbrains.derivefiles

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.commitchanges.CommitChangesUseCase
import dev.monosmith.jetbrains.flowports.FlowPortsContract
import dev.monosmith.jetbrains.flowports.IdePortsClient
import dev.monosmith.jetbrains.planapply.PlanStepContract
import dev.monosmith.jetbrains.planapply.PlanThenApplyUseCase
import dev.monosmith.jetbrains.prompt.ChoiceRowContract
import dev.monosmith.jetbrains.smithrun.BackgroundFlowClient
import dev.monosmith.jetbrains.smithrun.FailureTextMapper
import dev.monosmith.jetbrains.smithrun.SmithRunClient
import dev.monosmith.jetbrains.words.SmithWordsConfig.FIXWIZARD_STYLE_APPLY
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDCHOICES_DEFAULT
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_DERIVEFILES_A_PLAIN_SMITH_DERIVE_IT
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_DERIVEFILES_BY_DEFAULT_SMITH_LEAVES_THE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_DERIVEFILES_BY_DEFAULT_SMITH_LEAVES_THE_2
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_DERIVEFILES_DERIVED_FILES_NEED_ACTION_PICK
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_DERIVEFILES_DO_IT
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_DERIVEFILES_EJECT_OWN_IT
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_DERIVEFILES_EVERY_DERIVED_FILE_IS_CURRENT
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_DERIVEFILES_EVERY_FILE_THAT_IS_STALE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_DERIVEFILES_EXPLAIN
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_DERIVEFILES_IS_DONE_COMMIT_IT_TOGETHER
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_DERIVEFILES_KEEP_MY_EDIT
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_DERIVEFILES_MONOSMITH
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_DERIVEFILES_MONOSMITH_CHECK_THE_DERIVED_FILES
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_DERIVEFILES_MONOSMITH_DERIVED_FILES
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_DERIVEFILES_MONOSMITH_PREVIEW_THE_DERIVED_FILES
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_DERIVEFILES_MONOSMITH_WRITE_THE_DERIVED_FILES
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_DERIVEFILES_OVERWRITE_WITH_THE_DERIVED_CONTENT
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_DERIVEFILES_RECLAIM_GIVE_IT_BACK_TO
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_DERIVEFILES_SAY_WHO_OWNS_THIS_FILE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_DERIVEFILES_SMITH_GIVES_THE_FILE_BACK
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_DERIVEFILES_SMITH_KEEPS_YOUR_EDIT_AS
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_DERIVEFILES_SMITH_OVERWRITES_THE_FILE_WITH
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_DERIVEFILES_SMITH_WRITES_EVERY_DERIVED_FILE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_DERIVEFILES_THE_DERIVED_FILES_ARE_WRITTEN
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_DERIVEFILES_THE_FILE_BECOMES_YOURS_SMITH
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_DERIVEFILES_WRITE_THE_DERIVED_CONTENT
import dev.monosmith.jetbrains.words.SmithWordsConfig.say
import java.nio.file.Path

/**
 * Looks at the derived files (`smith derive --check`) and lets the person act on one: explain who owns it,
 * overwrite it with the derived content, keep the hand edit, eject it (own it) or reclaim it (give it back
 * to smith). Each change shows its effect first and commits nothing unless the box is ticked.
 */
object DeriveFilesUseCase {
    private val TITLE = say(JETBRAINS_DERIVEFILES_MONOSMITH_DERIVED_FILES)
    private const val ALL = "\u0000all"

    private val effects =
        mapOf(
            "apply" to
                say(JETBRAINS_DERIVEFILES_SMITH_WRITES_EVERY_DERIVED_FILE),
            "replace" to say(JETBRAINS_DERIVEFILES_SMITH_OVERWRITES_THE_FILE_WITH),
            "adopt" to
                say(JETBRAINS_DERIVEFILES_SMITH_KEEPS_YOUR_EDIT_AS),
            "eject" to say(JETBRAINS_DERIVEFILES_THE_FILE_BECOMES_YOURS_SMITH),
            "reclaim" to say(JETBRAINS_DERIVEFILES_SMITH_GIVES_THE_FILE_BACK),
        )

    private fun actionRows(file: DerivedFileContract): List<ChoiceRowContract> {
        val write = if (DeriveReportMapper.isHeldBack(file.status)) "replace" else "apply"
        return listOf(
            ChoiceRowContract(
                "explain",
                say(JETBRAINS_DERIVEFILES_EXPLAIN),
                say(JETBRAINS_ADDCHOICES_DEFAULT),
                say(JETBRAINS_DERIVEFILES_SAY_WHO_OWNS_THIS_FILE),
            ),
            ChoiceRowContract(
                write,
                if (write ==
                    "replace"
                ) {
                    say(JETBRAINS_DERIVEFILES_OVERWRITE_WITH_THE_DERIVED_CONTENT)
                } else {
                    say(JETBRAINS_DERIVEFILES_WRITE_THE_DERIVED_CONTENT)
                },
                "",
                effects.getValue(write),
            ),
            ChoiceRowContract("adopt", say(JETBRAINS_DERIVEFILES_KEEP_MY_EDIT), "", effects.getValue("adopt")),
            ChoiceRowContract("eject", say(JETBRAINS_DERIVEFILES_EJECT_OWN_IT), "", effects.getValue("eject")),
            ChoiceRowContract("reclaim", say(JETBRAINS_DERIVEFILES_RECLAIM_GIVE_IT_BACK_TO), "", effects.getValue("reclaim")),
        )
    }

    /** Checks the derived files in the Output tab (`smith derive --check`). */
    fun check(
        ide: Project,
        root: Path,
    ) = BackgroundFlowClient.start(ide, say(JETBRAINS_DERIVEFILES_MONOSMITH_CHECK_THE_DERIVED_FILES)) {
        SmithRunClient.run(ide, root, DeriveReportMapper.arguments("check"), show = true)
    }

    /** Previews what a plain `smith derive` would write (`--dry-run`) in the Output tab. */
    fun preview(
        ide: Project,
        root: Path,
    ) = BackgroundFlowClient.start(ide, say(JETBRAINS_DERIVEFILES_MONOSMITH_PREVIEW_THE_DERIVED_FILES)) {
        SmithRunClient.run(ide, root, DeriveReportMapper.arguments("preview"), show = true)
    }

    /** Lets the person act on the derived files that need action. */
    fun start(
        ide: Project,
        root: Path,
    ) = BackgroundFlowClient.start(ide, TITLE) { run(IdePortsClient.of(ide, root)) { CommitChangesUseCase.start(ide, root, it) } }

    /**
     * Runs the flow with the given ports.
     *
     * @param commitLater starts the Commit action for an operation
     * @return true when smith changed a derived file
     */
    fun run(
        ports: FlowPortsContract,
        commitLater: (String) -> Unit,
    ): Boolean {
        val run = ports.smith.run(DeriveReportMapper.arguments("check"))
        val files = DeriveReportMapper.files(run.stdout)
        if (files.isEmpty()) {
            if (run.ok) {
                ports.notice.info(say(JETBRAINS_DERIVEFILES_EVERY_DERIVED_FILE_IS_CURRENT))
            } else {
                ports.notice.warn(FailureTextMapper.describe(run))
            }
            return false
        }
        val rows =
            listOf(
                ChoiceRowContract(
                    ALL,
                    say(JETBRAINS_DERIVEFILES_EVERY_FILE_THAT_IS_STALE),
                    say(JETBRAINS_ADDCHOICES_DEFAULT),
                    say(JETBRAINS_DERIVEFILES_A_PLAIN_SMITH_DERIVE_IT),
                ),
            ) +
                files.map { ChoiceRowContract(it.path, it.path, "${it.status} · ${it.renderer}", it.detail) }
        val picked =
            ports.prompt
                .pickOne(
                    TITLE,
                    say(JETBRAINS_DERIVEFILES_DERIVED_FILES_NEED_ACTION_PICK, files.size.toString()),
                    rows,
                )?.name
                ?: return false
        if (picked == ALL) return applyAll(ports, commitLater)
        val file = files.first { it.path == picked }
        val action =
            ports.prompt.pickOne("$TITLE: ${file.path}", "${file.status}: ${file.detail}", actionRows(file))?.name ?: return false
        if (action == "explain") {
            ports.smith.run(DeriveReportMapper.arguments("explain", file.path), show = true)
            return false
        }
        return PlanThenApplyUseCase.run(
            ports,
            PlanStepContract(
                title = say(JETBRAINS_DERIVEFILES_MONOSMITH, action.toString(), file.path.toString()),
                planArgs = null,
                fixedPlan = "smith ${DeriveReportMapper.arguments(action, file.path).joinToString(" ")}",
                header = say(JETBRAINS_DERIVEFILES_BY_DEFAULT_SMITH_LEAVES_THE, effects.getValue(action).toString()),
                okText = say(JETBRAINS_DERIVEFILES_DO_IT),
                offerCommit = true,
                apply = { DeriveReportMapper.arguments(action, file.path) },
                done = say(JETBRAINS_DERIVEFILES_IS_DONE_COMMIT_IT_TOGETHER, file.path.toString(), action.toString()),
                operation = "derive",
            ),
            commitLater,
        )
    }

    private fun applyAll(
        ports: FlowPortsContract,
        commitLater: (String) -> Unit,
    ): Boolean =
        PlanThenApplyUseCase.run(
            ports,
            PlanStepContract(
                title = say(JETBRAINS_DERIVEFILES_MONOSMITH_WRITE_THE_DERIVED_FILES),
                planArgs = DeriveReportMapper.arguments("preview"),
                header = say(JETBRAINS_DERIVEFILES_BY_DEFAULT_SMITH_LEAVES_THE_2, effects.getValue("apply").toString()),
                okText = say(FIXWIZARD_STYLE_APPLY),
                offerCommit = true,
                apply = { DeriveReportMapper.arguments("apply") },
                done = say(JETBRAINS_DERIVEFILES_THE_DERIVED_FILES_ARE_WRITTEN),
                operation = "derive",
            ),
            commitLater,
        )
}
