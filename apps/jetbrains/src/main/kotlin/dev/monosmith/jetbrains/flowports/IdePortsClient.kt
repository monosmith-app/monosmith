package dev.monosmith.jetbrains.flowports

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.notice.NoticeButtonContract
import dev.monosmith.jetbrains.notice.NoticeClient
import dev.monosmith.jetbrains.prompt.ChoiceRowContract
import dev.monosmith.jetbrains.prompt.PlanAnswerContract
import dev.monosmith.jetbrains.prompt.PromptClient
import dev.monosmith.jetbrains.smithrun.SmithRunClient
import dev.monosmith.jetbrains.smithrun.SmithRunContract
import dev.monosmith.jetbrains.workspace.WorkspaceChangedStore
import java.nio.file.Path

/** Gives a flow its ports in the IDE: smith runs in [root], dialogs and balloons belong to [ide]. */
object IdePortsClient {
    /** The ports of the IDE project [ide] for the workspace at [root]. */
    fun of(
        ide: Project,
        root: Path,
    ): FlowPortsContract =
        FlowPortsContract(
            smith =
                object : SmithPort {
                    override fun run(
                        args: List<String>,
                        show: Boolean,
                    ): SmithRunContract = SmithRunClient.run(ide, root, args, show)
                },
            prompt =
                object : PromptPort {
                    override fun pickOne(
                        title: String,
                        header: String,
                        rows: List<ChoiceRowContract>,
                        okText: String,
                    ) = PromptClient.pickOne(ide, title, header, rows, okText)

                    override fun pickMany(
                        title: String,
                        header: String,
                        rows: List<ChoiceRowContract>,
                        okText: String,
                    ) = PromptClient.pickMany(ide, title, header, rows, okText)

                    override fun plan(
                        title: String,
                        header: String,
                        plan: String,
                        okText: String,
                        offerCommit: Boolean,
                        checks: List<String>,
                        others: List<String>,
                    ): PlanAnswerContract? = PromptClient.plan(ide, title, header, plan, okText, offerCommit, checks, others)

                    override fun text(
                        title: String,
                        question: String,
                        initial: String,
                        problem: (String) -> String?,
                    ) = PromptClient.text(ide, title, question, initial, problem)
                },
            notice =
                object : NoticePort {
                    override fun info(
                        message: String,
                        buttons: List<NoticeButtonContract>,
                    ) = NoticeClient.info(ide, message, buttons)

                    override fun warn(
                        message: String,
                        buttons: List<NoticeButtonContract>,
                    ) = NoticeClient.warn(ide, message, buttons)
                },
            changed = { WorkspaceChangedStore.changed(ide) },
        )
}
