package dev.monosmith.jetbrains.prompt

import com.intellij.openapi.project.Project
import com.intellij.openapi.ui.DialogWrapper
import com.intellij.openapi.util.text.StringUtil
import com.intellij.ui.SearchTextField
import com.intellij.ui.components.JBLabel
import com.intellij.ui.components.JBRadioButton
import com.intellij.ui.components.JBScrollPane
import com.intellij.util.ui.JBUI
import java.awt.BorderLayout
import java.awt.Dimension
import javax.swing.BoxLayout
import javax.swing.ButtonGroup
import javax.swing.JComponent
import javax.swing.JPanel
import javax.swing.event.DocumentEvent
import javax.swing.event.DocumentListener

/**
 * A question with one answer, as radio buttons. The first row is selected and is what Enter takes, so
 * the default always comes first. A long list gets a filter box.
 */
class PickOneDialogHandler(
    ide: Project,
    title: String,
    private val header: String,
    private val rows: List<ChoiceRowContract>,
    okText: String = "Next",
) : DialogWrapper(ide) {
    private val buttons = rows.map { JBRadioButton(label(it)) }
    private val group = ButtonGroup()

    init {
        this.title = title
        setOKButtonText(okText)
        buttons.forEach(group::add)
        buttons.firstOrNull()?.isSelected = true
        init()
    }

    /** The row the person selected. */
    val selected: ChoiceRowContract get() = rows[buttons.indexOfFirst { it.isSelected }.coerceAtLeast(0)]

    override fun createCenterPanel(): JComponent {
        val list = JPanel().apply { layout = BoxLayout(this, BoxLayout.Y_AXIS) }
        buttons.forEach { it.border = JBUI.Borders.empty(2, 0) }
        buttons.forEach(list::add)
        val panel = JPanel(BorderLayout(0, JBUI.scale(6)))
        panel.add(JBLabel("<html>${StringUtil.escapeXmlEntities(header)}</html>"), BorderLayout.NORTH)
        if (rows.size > FILTER_FROM) {
            val filter = SearchTextField()
            filter.textEditor.document.addDocumentListener(
                object : DocumentListener {
                    override fun insertUpdate(event: DocumentEvent) = apply()

                    override fun removeUpdate(event: DocumentEvent) = apply()

                    override fun changedUpdate(event: DocumentEvent) = apply()

                    private fun apply() {
                        val query = filter.text.trim().lowercase()
                        buttons.forEachIndexed { index, button -> button.isVisible = matches(rows[index], query) }
                        list.revalidate()
                    }
                },
            )
            panel.add(filter, BorderLayout.SOUTH)
        }
        panel.add(JBScrollPane(list).apply { preferredSize = Dimension(JBUI.scale(560), JBUI.scale(320)) }, BorderLayout.CENTER)
        return panel
    }

    override fun getPreferredFocusedComponent(): JComponent? = buttons.firstOrNull()

    private companion object {
        const val FILTER_FROM = 6

        fun label(row: ChoiceRowContract): String {
            val detail = if (row.detail.isBlank()) "" else "<br><small>${StringUtil.escapeXmlEntities(row.detail)}</small>"
            val description = if (row.description.isBlank()) "" else " ${StringUtil.escapeXmlEntities(row.description)}"
            return "<html><b>${StringUtil.escapeXmlEntities(row.label)}</b>$description$detail</html>"
        }

        fun matches(
            row: ChoiceRowContract,
            query: String,
        ): Boolean =
            query.isEmpty() || "${row.name} ${row.label} ${row.description} ${row.detail} ${row.search}".lowercase().contains(query)
    }
}
