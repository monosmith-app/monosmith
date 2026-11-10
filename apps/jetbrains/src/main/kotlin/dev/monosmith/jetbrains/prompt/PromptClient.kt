package dev.monosmith.jetbrains.prompt

import com.intellij.openapi.application.ApplicationManager
import com.intellij.openapi.application.ModalityState
import com.intellij.openapi.project.Project
import com.intellij.openapi.ui.InputValidatorEx
import com.intellij.openapi.ui.Messages

/**
 * The questions of the flows. A flow runs in a background task and blocks on these calls; each one opens
 * its dialog on the UI thread and returns what the person chose, or null when the person cancels.
 * Every question is a presented choice: the default is the first row and is marked in its text.
 */
object PromptClient {
    /** Runs [body] on the UI thread and waits for its result. */
    fun <T> onUi(body: () -> T): T {
        var result: T? = null
        ApplicationManager.getApplication().invokeAndWait({ result = body() }, ModalityState.defaultModalityState())
        @Suppress("UNCHECKED_CAST")
        return result as T
    }

    /** Asks a question with one answer; the first row is the default. Returns the row, or null on Cancel. */
    fun pickOne(
        ide: Project,
        title: String,
        header: String,
        rows: List<ChoiceRowContract>,
        okText: String = "Next",
    ): ChoiceRowContract? =
        onUi {
            val dialog = PickOneDialogHandler(ide, title, header, rows, okText)
            if (dialog.showAndGet()) dialog.selected else null
        }

    /** Asks a question with any number of answers. Returns the names ticked (maybe none), or null on Cancel. */
    fun pickMany(
        ide: Project,
        title: String,
        header: String,
        rows: List<ChoiceRowContract>,
        okText: String = "Next",
    ): List<String>? =
        onUi {
            val dialog = PickManyDialogHandler(ide, title, header, rows, okText)
            if (dialog.showAndGet()) dialog.selected else null
        }

    /** Shows a plan and asks; returns the answer, or null on Cancel. */
    fun plan(
        ide: Project,
        title: String,
        header: String,
        plan: String,
        okText: String,
        offerCommit: Boolean,
        checks: List<String> = emptyList(),
        others: List<String> = emptyList(),
    ): PlanAnswerContract? =
        onUi {
            val dialog = PlanDialogHandler(ide, title, header, plan, okText, offerCommit, checks, others)
            if (dialog.showAndGet()) dialog.answer else null
        }

    /** Asks for a line of text. [problem] says what is wrong with the text, or null when it is fine. */
    fun text(
        ide: Project,
        title: String,
        question: String,
        initial: String = "",
        problem: (String) -> String? = { null },
    ): String? =
        onUi {
            Messages
                .showInputDialog(
                    ide,
                    question,
                    title,
                    null,
                    initial,
                    object : InputValidatorEx {
                        override fun getErrorText(inputString: String): String? = problem(inputString.trim())

                        override fun checkInput(inputString: String): Boolean = problem(inputString.trim()) == null

                        override fun canClose(inputString: String): Boolean = checkInput(inputString)
                    },
                )?.trim()
        }
}
