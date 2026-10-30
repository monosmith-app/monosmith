package dev.monosmith.jetbrains.toolwindow

import com.intellij.openapi.application.ApplicationManager
import com.intellij.openapi.project.DumbAware
import com.intellij.openapi.project.Project
import com.intellij.openapi.wm.ToolWindow
import com.intellij.openapi.wm.ToolWindowFactory
import com.intellij.ui.components.JBScrollPane
import com.intellij.ui.content.ContentFactory
import com.intellij.ui.treeStructure.Tree
import dev.monosmith.jetbrains.taskrun.RunTaskUseCase
import dev.monosmith.jetbrains.workspace.FindWorkspaceRootUseCase
import dev.monosmith.jetbrains.workspace.ListProjectsUseCase
import java.awt.event.MouseAdapter
import java.awt.event.MouseEvent
import java.nio.file.Files
import java.nio.file.Path
import javax.swing.tree.DefaultMutableTreeNode
import javax.swing.tree.DefaultTreeModel

/** The MonoSmith tool window: the workspace's projects and their tasks; double-click a task to run it. */
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
        val tree = Tree(DefaultTreeModel(DefaultMutableTreeNode("Looking for a MonoSmith workspace...")))
        val base = project.basePath?.let(Path::of)
        val root = base?.let { FindWorkspaceRootUseCase.find(it, Files::exists) }
        tree.addMouseListener(
            object : MouseAdapter() {
                override fun mouseClicked(event: MouseEvent) {
                    if (event.clickCount != 2 || root == null) return
                    val node = tree.lastSelectedPathComponent as? DefaultMutableTreeNode
                    (node?.userObject as? TaskNode)?.let { RunTaskUseCase.run(project, root, it.project, it.task) }
                }
            },
        )
        toolWindow.contentManager.addContent(
            ContentFactory.getInstance().createContent(JBScrollPane(tree), "Projects", false),
        )
        if (root == null) {
            tree.model = DefaultTreeModel(DefaultMutableTreeNode("No smith.toml in this project"))
            return
        }
        ApplicationManager.getApplication().executeOnPooledThread {
            val result = ListProjectsUseCase.list(root)
            ApplicationManager.getApplication().invokeLater {
                tree.model = DefaultTreeModel(toNodes(result))
            }
        }
    }

    private fun toNodes(result: Result<List<dev.monosmith.jetbrains.workspace.SmithProjectContract>>): DefaultMutableTreeNode {
        val top = DefaultMutableTreeNode("Projects")
        result.fold(
            onSuccess = { projects ->
                projects.forEach { smithProject ->
                    val node = DefaultMutableTreeNode("${smithProject.name} (${smithProject.language} ${smithProject.kind})")
                    smithProject.tasks.forEach { node.add(DefaultMutableTreeNode(TaskNode(smithProject.name, it))) }
                    top.add(node)
                }
            },
            onFailure = { top.add(DefaultMutableTreeNode("smith failed: ${it.message}")) },
        )
        return top
    }
}
