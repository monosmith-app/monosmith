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
import com.intellij.ui.SearchTextField
import com.intellij.ui.components.JBCheckBox
import com.intellij.ui.components.JBScrollPane
import com.intellij.ui.content.ContentFactory
import com.intellij.ui.treeStructure.Tree
import dev.monosmith.jetbrains.addproject.AddProjectUseCase
import dev.monosmith.jetbrains.composestack.ComposeStackUseCase
import dev.monosmith.jetbrains.dependencies.UpdateDependenciesUseCase
import dev.monosmith.jetbrains.flowports.IdePortsClient
import dev.monosmith.jetbrains.flowports.WorkspaceChangedStore
import dev.monosmith.jetbrains.health.HealthUseCase
import dev.monosmith.jetbrains.installdependency.InstallDependencyUseCase
import dev.monosmith.jetbrains.newworkspace.NewWorkspaceUseCase
import dev.monosmith.jetbrains.projectactions.ProjectActionsUseCase
import dev.monosmith.jetbrains.releaseflow.ReadReleasePreviewUseCase
import dev.monosmith.jetbrains.releaseflow.ReleaseWorkspaceUseCase
import dev.monosmith.jetbrains.smithrun.OutputConsoleStore
import dev.monosmith.jetbrains.smithrun.SmithRunClient
import dev.monosmith.jetbrains.smithversion.SmithPresenceUseCase
import dev.monosmith.jetbrains.swapchoice.SwapChoiceUseCase
import dev.monosmith.jetbrains.taskrun.RunTaskUseCase
import dev.monosmith.jetbrains.words.SmithWordsConfig.COMMANDS_GROUP_RELEASE
import dev.monosmith.jetbrains.words.SmithWordsConfig.COMMANDS_PROJ_SWAP_TITLE
import dev.monosmith.jetbrains.words.SmithWordsConfig.COMMANDS_VIEW_DEPS_TITLE
import dev.monosmith.jetbrains.words.SmithWordsConfig.DASHBOARD_ROW_HEALTH
import dev.monosmith.jetbrains.words.SmithWordsConfig.DASHBOARD_ROW_PROJECTS
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_TOOLWINDOW_ADD_LIBRARY
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_TOOLWINDOW_AFFECTED_ONLY
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_TOOLWINDOW_INSTALL_PACKAGE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_TOOLWINDOW_MAKE_APP_LIBRARY_OR_PACKAGE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_TOOLWINDOW_READ_THE_PROJECTS_THE_UPDATES
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_TOOLWINDOW_REFRESH
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_TOOLWINDOW_RUN_TASK
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_TOOLWINDOW_UNLINK_LIBRARY
import dev.monosmith.jetbrains.words.SmithWordsConfig.say
import dev.monosmith.jetbrains.workspace.AffectedJsonMapper
import dev.monosmith.jetbrains.workspace.FindWorkspaceRootUseCase
import dev.monosmith.jetbrains.workspace.ListProjectsUseCase
import dev.monosmith.jetbrains.workspace.SmithProjectContract
import dev.monosmith.jetbrains.workspacetools.WorkspaceToolsUseCase
import java.awt.BorderLayout
import java.awt.event.MouseAdapter
import java.awt.event.MouseEvent
import java.nio.file.Files
import java.nio.file.Path
import javax.swing.JMenu
import javax.swing.JMenuItem
import javax.swing.JPanel
import javax.swing.JPopupMenu
import javax.swing.event.DocumentEvent
import javax.swing.event.DocumentListener
import javax.swing.tree.DefaultMutableTreeNode
import javax.swing.tree.DefaultTreeModel

/**
 * The MonoSmith tool window. Projects: the projects and their tasks with the affected ones marked, a filter
 * box and an "Affected only" box, a row to add a project and a menu on each project. Dependencies: the
 * updates grouped by risk, with the update and align flows. Release: the plan, why nothing releases, and
 * the Release… flow. Health: the doctor's findings, each fixable one with a Fix row that shows its plan.
 * Output: what smith prints.
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

    private class Views(
        val projects: Tree = Tree(),
        val deps: Tree = Tree(),
        val release: Tree = Tree(),
        val health: Tree = Tree(),
        val filter: SearchTextField = SearchTextField(),
        val affectedOnly: JBCheckBox = JBCheckBox(say(JETBRAINS_TOOLWINDOW_AFFECTED_ONLY)),
    ) {
        var listed: Result<List<SmithProjectContract>>? = null
        var affected: Set<String> = emptySet()
    }

    override fun createToolWindowContent(
        project: Project,
        toolWindow: ToolWindow,
    ) {
        val base = project.basePath?.let(Path::of)
        val root = base?.let { FindWorkspaceRootUseCase.find(it, Files::exists) }
        val views = Views()
        val factory = ContentFactory.getInstance()
        val top =
            JPanel(BorderLayout()).apply {
                add(views.filter, BorderLayout.CENTER)
                add(views.affectedOnly, BorderLayout.EAST)
            }
        val projectsPanel =
            JPanel(BorderLayout()).apply {
                add(top, BorderLayout.NORTH)
                add(JBScrollPane(views.projects), BorderLayout.CENTER)
            }
        toolWindow.contentManager.addContent(factory.createContent(projectsPanel, say(DASHBOARD_ROW_PROJECTS), false))
        toolWindow.contentManager.addContent(factory.createContent(JBScrollPane(views.deps), say(COMMANDS_VIEW_DEPS_TITLE), false))
        toolWindow.contentManager.addContent(factory.createContent(JBScrollPane(views.release), say(COMMANDS_GROUP_RELEASE), false))
        toolWindow.contentManager.addContent(factory.createContent(JBScrollPane(views.health), say(DASHBOARD_ROW_HEALTH), false))
        toolWindow.contentManager.addContent(
            factory.createContent(OutputConsoleStore.consoleFor(project).component, OutputConsoleStore.TAB_TITLE, false),
        )

        listOf(views.projects, views.deps, views.release, views.health).forEach { it.addMouseListener(clicks(it, project, root, base)) }
        val redraw = {
            show(
                views.projects,
                ToolWindowRowsMapper.projectRows(
                    root != null,
                    views.listed,
                    views.affected,
                    views.filter.text.trim(),
                    views.affectedOnly.isSelected,
                ),
            )
        }
        views.filter.textEditor.document.addDocumentListener(
            object : DocumentListener {
                override fun insertUpdate(event: DocumentEvent) = redraw()

                override fun removeUpdate(event: DocumentEvent) = redraw()

                override fun changedUpdate(event: DocumentEvent) = redraw()
            },
        )
        views.affectedOnly.addActionListener { redraw() }
        val reload = { load(project, root, views, redraw) }
        toolWindow.setTitleActions(
            listOf<AnAction>(
                object : DumbAwareAction(
                    say(JETBRAINS_TOOLWINDOW_REFRESH),
                    say(JETBRAINS_TOOLWINDOW_READ_THE_PROJECTS_THE_UPDATES),
                    AllIcons.Actions.Refresh,
                ) {
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
        views: Views,
        redraw: () -> Unit,
    ) {
        views.listed = null
        redraw()
        show(views.deps, ToolWindowRowsMapper.depsRows(null))
        show(views.release, ToolWindowRowsMapper.releaseRows(null))
        show(views.health, ToolWindowRowsMapper.healthRows(root != null, null))
        if (root == null) return
        val pool = ApplicationManager.getApplication()
        pool.executeOnPooledThread {
            val listed = ListProjectsUseCase.list(root)
            val affected =
                runCatching {
                    val run = SmithRunClient.run(null, root, listOf("affected", "--json"))
                    if (run.ok) AffectedJsonMapper.toAffected(run.stdout).projects else emptySet()
                }.getOrDefault(emptySet())
            pool.invokeLater {
                views.listed = listed
                views.affected = affected
                redraw()
            }
        }
        pool.executeOnPooledThread {
            val read = UpdateDependenciesUseCase.read(IdePortsClient.of(ide, root))
            pool.invokeLater { show(views.deps, ToolWindowRowsMapper.depsRows(read)) }
        }
        pool.executeOnPooledThread {
            val read = ReadReleasePreviewUseCase.read(root)
            pool.invokeLater { show(views.release, ToolWindowRowsMapper.releaseRows(read)) }
        }
        pool.executeOnPooledThread {
            val read = HealthUseCase.read(root)
            pool.invokeLater { show(views.health, ToolWindowRowsMapper.healthRows(true, read)) }
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
                    row.project.tasks.forEach { node.add(DefaultMutableTreeNode(TaskNode(row.project.name, it))) }
                }

                is ToolWindowRowContract.Finding -> {
                    ToolWindowRowsMapper.findingChildren(row.finding).forEach { node.add(DefaultMutableTreeNode(it)) }
                }

                is ToolWindowRowContract.Group -> {
                    row.children.forEach { node.add(DefaultMutableTreeNode(it)) }
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

                is ToolWindowRowContract.Link -> {
                    if (root != null) link(ide, root, value.action)
                }

                else -> {}
            }
        }
    }

    private fun link(
        ide: Project,
        root: Path,
        action: String,
    ) = when (action) {
        "update" -> UpdateDependenciesUseCase.start(ide, root)
        "align" -> UpdateDependenciesUseCase.startAlign(ide, root)
        else -> ReleaseWorkspaceUseCase.start(ide, root)
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
        item(say(JETBRAINS_TOOLWINDOW_ADD_LIBRARY)) { InstallDependencyUseCase.start(ide, root, name, "workspace") }
        item(say(JETBRAINS_TOOLWINDOW_INSTALL_PACKAGE)) { InstallDependencyUseCase.start(ide, root, name, "registry") }
        item(say(JETBRAINS_TOOLWINDOW_UNLINK_LIBRARY)) { ProjectActionsUseCase.unlink(ide, root, name) }
        item(say(COMMANDS_PROJ_SWAP_TITLE)) { SwapChoiceUseCase.start(ide, root, name) }
        item(say(JETBRAINS_TOOLWINDOW_MAKE_APP_LIBRARY_OR_PACKAGE)) { ProjectActionsUseCase.changeKind(ide, root, name) }
        item("Remove…") { ProjectActionsUseCase.remove(ide, root, name) }
        menu.addSeparator()
        item("Why…") { WorkspaceToolsUseCase.why(ide, root, name) }
        item("Explain") { WorkspaceToolsUseCase.explain(ide, root, name) }
        if (row.project.tasks.isNotEmpty()) {
            val run = JMenu(say(JETBRAINS_TOOLWINDOW_RUN_TASK))
            row.project.tasks.forEach { task ->
                run.add(JMenuItem(task).apply { addActionListener { RunTaskUseCase.run(ide, root, name, task) } })
            }
            menu.add(run)
        }
        return menu
    }
}
