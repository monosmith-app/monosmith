package dev.monosmith.jetbrains.smithaction

import com.intellij.openapi.project.Project
import com.intellij.openapi.ui.ComboBox
import com.intellij.openapi.ui.DialogWrapper
import com.intellij.openapi.ui.ValidationInfo
import com.intellij.ui.components.JBCheckBox
import com.intellij.ui.components.JBTextField
import com.intellij.util.ui.FormBuilder
import dev.monosmith.jetbrains.workspace.SmithKindContract
import javax.swing.JComponent

/**
 * Asks what to add: the kind (language and role), the name, and whether to commit the change. The
 * commit box is unticked: smith commits only when asked, and Undo then reverts the commit.
 */
class AddProjectDialog(
    ide: Project,
    kinds: List<SmithKindContract>,
    private val existing: Collection<String>,
) : DialogWrapper(ide) {
    private val kindBox = ComboBox(kinds.toTypedArray())
    private val nameField = JBTextField()
    private val commitBox = JBCheckBox("Commit this change so it can be undone", false)

    init {
        title = "MonoSmith: Add project"
        setOKButtonText("Add project")
        init()
    }

    /** The kind the person chose. */
    val kind: SmithKindContract get() = kindBox.selectedItem as SmithKindContract

    /** The name the person typed. */
    val projectName: String get() = nameField.text.trim()

    /** Whether the person ticked the commit box. */
    val commit: Boolean get() = commitBox.isSelected

    override fun createCenterPanel(): JComponent =
        FormBuilder
            .createFormBuilder()
            .addLabeledComponent("Kind:", kindBox)
            .addLabeledComponent("Name:", nameField)
            .addComponent(commitBox)
            .panel

    override fun getPreferredFocusedComponent(): JComponent = nameField

    override fun doValidate(): ValidationInfo? = ProjectNameValidator.problem(projectName, existing)?.let { ValidationInfo(it, nameField) }
}
