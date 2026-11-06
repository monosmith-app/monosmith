package dev.monosmith.jetbrains.prompt

import com.intellij.openapi.project.Project
import com.intellij.openapi.ui.DialogWrapper
import com.intellij.openapi.util.text.StringUtil
import com.intellij.ui.components.JBCheckBox
import com.intellij.ui.components.JBLabel
import com.intellij.ui.components.JBScrollPane
import com.intellij.ui.components.JBTextArea
import com.intellij.util.ui.JBUI
import dev.monosmith.jetbrains.words.SmithWordsConfig.DIALOG_COMMIT_LABEL
import dev.monosmith.jetbrains.words.SmithWordsConfig.say
import java.awt.BorderLayout
import java.awt.Dimension
import java.awt.Font
import javax.swing.Action
import javax.swing.BoxLayout
import javax.swing.JComponent
import javax.swing.JPanel

/**
 * Shows what smith would do (the text of a `--dry-run`) and asks. It can carry the unticked "Commit this
 * change so it can be undone" box, other boxes, and other buttons beside the main one. smith commits
 * only when the box is ticked.
 *
 * @param header the sentence above the plan: what the answer decides, and what is not done yet
 * @param plan the plan text, shown as it is
 * @param okText the verb of the main button
 * @param offerCommit show the commit box (unticked)
 * @param checks the label of each extra box, unticked
 * @param others the label of each extra button
 */
class PlanDialogHandler(
    ide: Project,
    title: String,
    private val header: String,
    private val plan: String,
    okText: String,
    private val offerCommit: Boolean,
    private val checks: List<String> = emptyList(),
    private val others: List<String> = emptyList(),
) : DialogWrapper(ide) {
    private val commitBox = JBCheckBox(COMMIT_LABEL, false)
    private val extraBoxes = checks.map { JBCheckBox(it, false) }
    private var other = -1

    init {
        this.title = title
        setOKButtonText(okText)
        init()
    }

    /** What the person answered; read it after [showAndGet] returned true. */
    val answer: PlanAnswerContract get() = PlanAnswerContract(commitBox.isSelected, extraBoxes.map { it.isSelected }, other)

    override fun createCenterPanel(): JComponent {
        val text =
            JBTextArea(plan).apply {
                isEditable = false
                font = Font(Font.MONOSPACED, Font.PLAIN, font.size)
                caretPosition = 0
            }
        val boxes = JPanel().apply { layout = BoxLayout(this, BoxLayout.Y_AXIS) }
        extraBoxes.forEach(boxes::add)
        if (offerCommit) boxes.add(commitBox)
        val panel = JPanel(BorderLayout(0, JBUI.scale(6)))
        panel.add(JBLabel("<html>${StringUtil.escapeXmlEntities(header)}</html>"), BorderLayout.NORTH)
        panel.add(JBScrollPane(text).apply { preferredSize = Dimension(JBUI.scale(640), JBUI.scale(320)) }, BorderLayout.CENTER)
        panel.add(boxes, BorderLayout.SOUTH)
        return panel
    }

    override fun createActions(): Array<Action> {
        val extra = others.mapIndexed { index, label -> OtherAction(label, index) }
        return (listOf(okAction) + extra + listOf(cancelAction)).toTypedArray()
    }

    private inner class OtherAction(
        label: String,
        private val index: Int,
    ) : DialogWrapperAction(label) {
        override fun doAction(event: java.awt.event.ActionEvent) {
            other = index
            close(OK_EXIT_CODE)
        }
    }

    /** The text of the commit box, shared with the flows that tell the person what it did. */
    companion object {
        /** The words of the unticked commit box. */
        val COMMIT_LABEL = say(DIALOG_COMMIT_LABEL)
    }
}
