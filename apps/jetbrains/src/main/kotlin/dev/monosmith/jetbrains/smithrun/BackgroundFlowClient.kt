package dev.monosmith.jetbrains.smithrun

import com.intellij.openapi.application.ApplicationManager
import com.intellij.openapi.progress.ProcessCanceledException
import com.intellij.openapi.progress.ProgressIndicator
import com.intellij.openapi.progress.ProgressManager
import com.intellij.openapi.progress.Task
import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.notice.NoticeClient

/** Runs a multi-step flow (questions and smith commands in turn) in a background task with a progress bar. */
object BackgroundFlowClient {
    /**
     * Starts [body] in a background task. The body may block on smith and ask questions through the
     * prompt client. An error it throws becomes a notice, never a crash.
     *
     * @param ide the IDE project
     * @param title the title of the progress bar
     * @param body the flow
     */
    fun start(
        ide: Project,
        title: String,
        body: () -> Unit,
    ) {
        ApplicationManager.getApplication().invokeLater {
            ProgressManager.getInstance().run(
                object : Task.Backgroundable(ide, title, true) {
                    override fun run(indicator: ProgressIndicator) {
                        indicator.isIndeterminate = true
                        try {
                            body()
                        } catch (cancelled: ProcessCanceledException) {
                            throw cancelled
                        } catch (problem: Exception) {
                            NoticeClient.warn(ide, "Nothing was changed. ${problem.message ?: problem.javaClass.simpleName}")
                        }
                    }
                },
            )
        }
    }
}
