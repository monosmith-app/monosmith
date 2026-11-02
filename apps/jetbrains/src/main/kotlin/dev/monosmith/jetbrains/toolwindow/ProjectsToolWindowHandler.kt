package dev.monosmith.jetbrains.toolwindow

import com.intellij.icons.AllIcons
import com.intellij.openapi.actionSystem.AnAction
import com.intellij.openapi.actionSystem.AnActionEvent
import com.intellij.openapi.application.ApplicationManager
import com.intellij.openapi.project.DumbAware
import com.intellij.openapi.project.DumbAwareAction
import com.intellij.openapi.project.Project
import com.intellij.openapi.wm.ToolWindow
import com.intellij.openapi.wm.ToolWindowFactory
import com.intellij.ui.components.JBScrollPane
import com.intellij.ui.content.ContentFactory
import com.intellij.ui.treeStructure.Tree
import dev.monosmith.jetbrains.addproject.AddProjectUseCase
import dev.monosmith.jetbrains.composestack.ComposeStackUseCase
import dev.monosmith.jetbrains.health.HealthUseCase
import dev.monosmith.jetbrains.installdependency.InstallDependencyUseCase
import dev.monosmith.jetbrains.newworkspace.NewWorkspaceUseCase
import dev.monosmith.jetbrains.projectactions.ProjectActionsUseCase
import dev.monosmith.jetbrains.smithrun.OutputConsoleStore
import dev.monosmith.jetbrains.smithversion.SmithPresenceUseCase
import dev.monosmith.jetbrains.swapchoice.SwapChoiceUseCase
import dev.monosmith.jetbrains.taskrun.RunTaskUseCase
import dev.monosmith.jetbrains.workspace.FindWorkspaceRootUseCase
import dev.monosmith.jetbrains.workspace.ListProjectsUseCase
import dev.monosmith.jetbrains.workspace.WorkspaceChangedStore
import dev.monosmith.jetbrains.workspacetools.WorkspaceToolsUseCase
import java.awt.event.MouseAdapter
import java.awt.event.MouseEvent
import java.nio.file.Files
import java.nio.file.Path
import javax.swing.JMenu
import javax.swing.JMenuItem
import javax.swing.JPopupMenu
import javax.swing.tree.DefaultMutableTreeNode
import javax.swing.tree.DefaultTreeModel

/**
 * The MonoSmith tool window: a Projects view (the projects and their tasks, a row to add a project, a menu
 * on each project), a Health view (the doctor's findings with a Fix row that runs the structured fix with
 * its plan shown) and an Output tab for what smith prints.
 */
class ProjectsToolWindowHandler :
    ToolWindowFactory,
    DumbAware {
    private data class TaskNode(
        val project: String,
        val task: String,
    ) {
        override fun toString() = task
    }

    override fun createToolWindowContent(
        project: Project,
        toolWindow: ToolWindow,
    ) {
        val base = project.basePath?.let(Path::of)
        val root = base?.let { FindWorkspaceRootUseCase.find(it, Files::exists) }
        val projects = Tree()
        val health = Tree()
        val factory = ContentFactory.getInstance()
        toolWindow.contentManager.addContent(factory.createContent(JBScrollPane(projects), "Projects", false))
        toolWindow.contentManager.addContent(factory.createContent(JBScrollPane(health), "Health", false))
        toolWindow.contentManager.addContent(
            factory.createContent(OutputConsoleStore.consoleFor(project).component, OutputConsoleStore.TAB_TITLE, false),
        )

        projects.addMouseListener(clicks(projects, project, root, base))
        health.addMouseListener(clicks(health, project, root, base))
        val reload = { load(project, root, projects, health) }
        toolWindow.setTitleActions(
            listOf<AnAction>(
                object : DumbAwareAction("Refresh", "Read the projects and the health of the workspace again", AllIcons.Actions.Refresh) {
                    override fun actionPerformed(event: AnActionEvent) = reload()
                },
            ),
        )
        WorkspaceChangedStore.listen { changed -> if (changed == project) reload() }
        reload()
        if (root != null) ApplicationManager.getApplication().executeOnPooledThread { SmithPresenceUseCase.check(project, root) }
    }

    private fun load(
        ide: Project,
        root: Path?,
        projects: Tree,
        health: Tree,
    ) {
        show(projects, ToolWindowRowsMapper.projectRows(root != null, null))
        show(health, ToolWindowRowsMapper.healthRows(root != null, null))
        if (root == null) return
        ApplicationManager.getApplication().executeOnPooledThread {
            val listed = ListProjectsUseCase.list(root)
            ApplicationManager.getApplication().invokeLater { show(projects, ToolWindowRowsMapper.projectRows(true, listed)) }
        }
        ApplicationManager.getApplication().executeOnPooledThread {
            val read = HealthUseCase.read(root)
            ApplicationManager.getApplication().invokeLater { show(health, ToolWindowRowsMapper.healthRows(true, read)) }
        }
    }

    private fun show(
        tree: Tree,
        rows: List<ToolWindowRowContract>,
    ) {
        val top = DefaultMutableTreeNode("MonoSmith")
        rows.forEach { row ->
            val node = DefaultMutableTreeNode(row)
            when (row) {
                is ToolWindowRowContract.Project -> {
                    row.project.tasks.forEach {
                        node.add(
                            DefaultMutableTreeNode(TaskNode(row.project.name, it)),
                        )
                    }
                }

                is ToolWindowRowContract.Finding -> {
                    ToolWindowRowsMapper
                        .findingChildren(
                            row.finding,
                        ).forEach { node.add(DefaultMutableTreeNode(it)) }
                }

                else -> {}
            }
            top.add(node)
        }
        tree.model = DefaultTreeModel(top)
        tree.isRootVisible = false
        for (index in 0 until tree.rowCount) tree.expandRow(index)
    }

    private fun clicks(
        tree: Tree,
        ide: Project,
        root: Path?,
        base: Path?,
    ) = object : MouseAdapter() {
        override fun mouseClicked(event: MouseEvent) {
            val path = tree.getPathForLocation(event.x, event.y) ?: return
            tree.selectionPath = path
            val value = (path.lastPathComponent as? DefaultMutableTreeNode)?.userObject
            if (event.isPopupTrigger || event.button == MouseEvent.BUTTON3) {
                if (value is ToolWindowRowContract.Project && root != null) projectMenu(ide, root, value).show(tree, event.x, event.y)
                return
            }
            if (event.clickCount != 2) return
            when (value) {
                is TaskNode -> {
                    if (root != null) RunTaskUseCase.run(ide, root, value.project, value.task)
                }

                is ToolWindowRowContract.AddProject -> {
                    if (root !=
                        null
                    ) {
                        AddProjectUseCase.start(ide, root) { ComposeStackUseCase.start(ide, root) }
                    }
                }

                is ToolWindowRowContract.NewWorkspace -> {
                    if (base != null) NewWorkspaceUseCase.start(ide, base)
                }

                is ToolWindowRowContract.Fix -> {
                    if (root != null) HealthUseCase.fix(ide, root, value.finding)
                }

                else -> {}
            }
        }
    }

    private fun projectMenu(
        ide: Project,
        root: Path,
        row: ToolWindowRowContract.Project,
    ): JPopupMenu {
        val name = row.project.name
        val menu = JPopupMenu()

        fun item(
            label: String,
            action: () -> Unit,
        ) = menu.add(JMenuItem(label).apply { addActionListener { action() } })
        item("Add library…") { InstallDependencyUseCase.start(ide, root, name, "workspace") }
        item("Install package…") { InstallDependencyUseCase.start(ide, root, name, "registry") }
        item("Unlink library…") { ProjectActionsUseCase.unlink(ide, root, name) }
        item("Swap a choice…") { SwapChoiceUseCase.start(ide, root, name) }
        item("Make app, library or package…") { ProjectActionsUseCase.changeKind(ide, root, name) }
        item("Remove…") { ProjectActionsUseCase.remove(ide, root, name) }
        menu.addSeparator()
        item("Why…") { WorkspaceToolsUseCase.why(ide, root, name) }
        item("Explain") { WorkspaceToolsUseCase.explain(ide, root, name) }
        if (row.project.tasks.isNotEmpty()) {
            val run = JMenu("Run task")
            row.project.tasks.forEach { task ->
                run.add(JMenuItem(task).apply { addActionListener { RunTaskUseCase.run(ide, root, name, task) } })
            }
            menu.add(run)
        }
        return menu
    }
}
