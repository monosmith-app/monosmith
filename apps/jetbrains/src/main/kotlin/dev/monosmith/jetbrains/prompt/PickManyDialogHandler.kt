package dev.monosmith.jetbrains.prompt

import com.intellij.openapi.project.Project
import com.intellij.openapi.ui.DialogWrapper
import com.intellij.openapi.util.text.StringUtil
import com.intellij.ui.components.JBCheckBox
import com.intellij.ui.components.JBLabel
import com.intellij.ui.components.JBScrollPane
import com.intellij.util.ui.JBUI
import java.awt.BorderLayout
import java.awt.Dimension
import javax.swing.BoxLayout
import javax.swing.JComponent
import javax.swing.JPanel

/** A question with any number of answers, as check boxes. The rows marked `picked` start ticked. */
class PickManyDialogHandler(
    ide: Project,
    title: String,
    private val header: String,
    private val rows: List<ChoiceRowContract>,
    okText: String = "Next",
) : DialogWrapper(ide) {
    private val boxes =
        rows.map { row ->
            val detail = if (row.detail.isBlank()) "" else "<br><small>${StringUtil.escapeXmlEntities(row.detail)}</small>"
            val description = if (row.description.isBlank()) "" else " ${StringUtil.escapeXmlEntities(row.description)}"
            JBCheckBox("<html><b>${StringUtil.escapeXmlEntities(row.label)}</b>$description$detail</html>", row.picked)
        }

    init {
        this.title = title
        setOKButtonText(okText)
        init()
    }

    /** The names of the ticked rows, in the order of the list. */
    val selected: List<String> get() = rows.filterIndexed { index, _ -> boxes[index].isSelected }.map { it.name }

    override fun createCenterPanel(): JComponent {
        val list = JPanel().apply { layout = BoxLayout(this, BoxLayout.Y_AXIS) }
        boxes.forEach {
            it.border = JBUI.Borders.empty(2, 0)
            list.add(it)
        }
        val panel = JPanel(BorderLayout(0, JBUI.scale(6)))
        panel.add(JBLabel("<html>${StringUtil.escapeXmlEntities(header)}</html>"), BorderLayout.NORTH)
        panel.add(JBScrollPane(list).apply { preferredSize = Dimension(JBUI.scale(560), JBUI.scale(320)) }, BorderLayout.CENTER)
        return panel
    }
}
