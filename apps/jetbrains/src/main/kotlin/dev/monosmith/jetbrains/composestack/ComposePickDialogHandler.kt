package dev.monosmith.jetbrains.composestack

import com.intellij.openapi.project.Project
import com.intellij.openapi.ui.DialogWrapper
import com.intellij.openapi.util.text.StringUtil
import com.intellij.ui.SearchTextField
import com.intellij.ui.components.JBCheckBox
import com.intellij.ui.components.JBLabel
import com.intellij.ui.components.JBScrollPane
import com.intellij.util.ui.JBUI
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_COMPOSESTACK_MONOSMITH_COMPOSE_A_STACK
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_COMPOSESTACK_NOTHING_IS_SELECTED_YET
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_COMPOSESTACK_SELECTED
import dev.monosmith.jetbrains.words.SmithWordsConfig.say
import java.awt.BorderLayout
import java.awt.Dimension
import javax.swing.BoxLayout
import javax.swing.JComponent
import javax.swing.JPanel
import javax.swing.event.DocumentEvent
import javax.swing.event.DocumentListener

/**
 * The catalogue of `smith compose --json` as a searchable list of check boxes, grouped by layer. What is
 * ticked stays visible whatever the search says, and its titles are named in the message line.
 */
class ComposePickDialogHandler(
    ide: Project,
    private val layers: List<ComposeLayerContract>,
    private val items: List<ComposeItemContract>,
    ticked: List<String>,
    private val message: String,
    okText: String,
) : DialogWrapper(ide) {
    private val ticks = ticked.toMutableSet()
    private val list = JPanel().apply { layout = BoxLayout(this, BoxLayout.Y_AXIS) }
    private val selection = JBLabel()
    private val search = SearchTextField()

    init {
        title = say(JETBRAINS_COMPOSESTACK_MONOSMITH_COMPOSE_A_STACK)
        setOKButtonText(okText)
        init()
        rebuild()
    }

    /** The ids ticked, in the order of the catalogue. */
    val selected: List<String> get() = items.filter { it.id in ticks }.map { it.id }

    override fun createCenterPanel(): JComponent {
        search.textEditor.document.addDocumentListener(
            object : DocumentListener {
                override fun insertUpdate(event: DocumentEvent) = rebuild()

                override fun removeUpdate(event: DocumentEvent) = rebuild()

                override fun changedUpdate(event: DocumentEvent) = rebuild()
            },
        )
        val top = JPanel(BorderLayout(0, JBUI.scale(4)))
        top.add(
            JBLabel(
                "<html>${StringUtil.escapeXmlEntities(
                    message.ifEmpty { "Tick what the project is made of: a framework, a desktop shell, a cloud host, a tool." },
                )}</html>",
            ),
            BorderLayout.NORTH,
        )
        top.add(search, BorderLayout.CENTER)
        top.add(selection, BorderLayout.SOUTH)
        val panel = JPanel(BorderLayout(0, JBUI.scale(6)))
        panel.add(top, BorderLayout.NORTH)
        panel.add(JBScrollPane(list).apply { preferredSize = Dimension(JBUI.scale(640), JBUI.scale(380)) }, BorderLayout.CENTER)
        return panel
    }

    override fun getPreferredFocusedComponent(): JComponent = search

    private fun rebuild() {
        val query = search.text.trim()
        list.removeAll()
        for (layer in layers) {
            val shown = items.filter { it.layer == layer.id && (it.id in ticks || ComposeStackAlgorithm.matches(it, query)) }
            if (shown.isEmpty()) continue
            list.add(
                JBLabel(
                    "<html><b>${StringUtil.escapeXmlEntities(
                        layer.title,
                    )}</b> <small>${StringUtil.escapeXmlEntities(layer.about)}</small></html>",
                ).apply {
                    border =
                        JBUI.Borders.empty(6, 0, 2, 0)
                },
            )
            shown.forEach { item ->
                val (label, where, detail) = ComposeStackAlgorithm.labels(item)
                val box =
                    JBCheckBox(
                        "<html><b>${StringUtil.escapeXmlEntities(
                            label,
                        )}</b> ${StringUtil.escapeXmlEntities(where)}<br><small>${StringUtil.escapeXmlEntities(detail)}</small></html>",
                        item.id in ticks,
                    )
                box.addActionListener {
                    if (box.isSelected) ticks.add(item.id) else ticks.remove(item.id)
                    selection.text = selectionText()
                }
                list.add(box)
            }
        }
        selection.text = selectionText()
        list.revalidate()
        list.repaint()
    }

    private fun selectionText(): String =
        ComposeStackAlgorithm.tickedTitles(items, ticks).let {
            if (it.isEmpty()) say(JETBRAINS_COMPOSESTACK_NOTHING_IS_SELECTED_YET) else say(JETBRAINS_COMPOSESTACK_SELECTED, it.toString())
        }
}
