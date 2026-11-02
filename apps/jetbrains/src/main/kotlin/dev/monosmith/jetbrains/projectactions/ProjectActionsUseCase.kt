package dev.monosmith.jetbrains.projectactions

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.commitchanges.CommitChangesUseCase
import dev.monosmith.jetbrains.flowports.FlowPortsContract
import dev.monosmith.jetbrains.flowports.IdePortsClient
import dev.monosmith.jetbrains.installdependency.InstallArgumentsAlgorithm
import dev.monosmith.jetbrains.planapply.PlanStepContract
import dev.monosmith.jetbrains.planapply.PlanThenApplyUseCase
import dev.monosmith.jetbrains.prompt.ChoiceRowContract
import dev.monosmith.jetbrains.smithrun.BackgroundFlowClient
import dev.monosmith.jetbrains.workspace.PortProjectsUseCase
import dev.monosmith.jetbrains.workspace.SmithProjectContract
import java.nio.file.Path

/** The actions of the project menu that change one project: remove it, make it an app, a lib or a package, unlink a library. */
object ProjectActionsUseCase {
    private fun target(
        ports: FlowPortsContract,
        project: String?,
        title: String,
    ): SmithProjectContract? {
        val listed = PortProjectsUseCase.list(ports)
        val nodes = listed.getOrNull()
        if (nodes == null) {
            ports.notice.warn("Nothing was changed. ${listed.exceptionOrNull()?.message}")
            return null
        }
        return if (project != null) nodes.firstOrNull { it.name == project } else PortProjectsUseCase.pick(ports, nodes, title)
    }

    /**
     * Removes a project: shows the plan of `smith remove --dry-run`, then asks the one thing smith asks:
     * keep a copy of the folder, or delete it. Cancel leaves everything as it is.
     *
     * @param project the project; null asks which
     */
    fun remove(
        ide: Project,
        root: Path,
        project: String?,
    ) = BackgroundFlowClient.start(ide, "MonoSmith: remove a project") {
        runRemove(IdePortsClient.of(ide, root), { CommitChangesUseCase.start(ide, root, it) }, project)
    }

    /**
     * Runs the removal with the given ports.
     *
     * @param commitLater starts the Commit action for an operation
     * @return true when smith removed the project
     */
    fun runRemove(
        ports: FlowPortsContract,
        commitLater: (String) -> Unit,
        project: String?,
    ): Boolean {
        val node = target(ports, project, "remove a project") ?: return false
        return PlanThenApplyUseCase.run(
            ports,
            PlanStepContract(
                title = "MonoSmith: remove ${node.name}?",
                planArgs = ProjectActionArgumentsAlgorithm.removePlan(node.name),
                header =
                    "\"Keep a copy\" moves the folder to .smith/removed/ so you can get it back. \"Delete\" removes the folder for good, " +
                        "files that git cannot bring back included. Nothing is removed until you answer.",
                okText = "Remove it and keep a copy",
                offerCommit = true,
                others = listOf("Remove it and delete the folder"),
                apply = { ProjectActionArgumentsAlgorithm.remove(node.name, delete = it.other == 0) },
                done = "${node.name} is removed",
                operation = "remove ${node.name}",
            ),
            commitLater,
        )
    }

    /**
     * Changes the kind of a project: asks the kind with what each means, shows `smith kind --dry-run`, then runs it.
     *
     * @param project the project; null asks which
     */
    fun changeKind(
        ide: Project,
        root: Path,
        project: String?,
    ) = BackgroundFlowClient.start(ide, "MonoSmith: change the kind of a project") {
        runChangeKind(IdePortsClient.of(ide, root), { CommitChangesUseCase.start(ide, root, it) }, project)
    }

    /**
     * Runs the kind change with the given ports.
     *
     * @param commitLater starts the Commit action for an operation
     * @return true when smith changed the kind
     */
    fun runChangeKind(
        ports: FlowPortsContract,
        commitLater: (String) -> Unit,
        project: String?,
    ): Boolean {
        val node = target(ports, project, "make it an app, a library or a package") ?: return false
        val kind =
            ports.prompt
                .pickOne(
                    "MonoSmith: make ${node.name} a ...",
                    "${node.name} is ${ProjectActionArgumentsAlgorithm.withArticle(node.kind)} now. What should it become?",
                    ProjectActionArgumentsAlgorithm.kindRows(node.kind),
                )?.name ?: return false
        return PlanThenApplyUseCase.run(
            ports,
            PlanStepContract(
                title = "MonoSmith: make ${node.name} ${ProjectActionArgumentsAlgorithm.withArticle(kind)}?",
                planArgs = ProjectActionArgumentsAlgorithm.kindPlan(node.name, kind),
                header = "Nothing was changed yet. smith moves the folder and rewrites every place that names it.",
                okText = "Make it ${ProjectActionArgumentsAlgorithm.withArticle(kind)}",
                offerCommit = true,
                apply = { ProjectActionArgumentsAlgorithm.kind(node.name, kind) },
                done = "${node.name} is now ${ProjectActionArgumentsAlgorithm.withArticle(kind)}",
                operation = "kind ${node.name} $kind",
            ),
            commitLater,
        )
    }

    /**
     * Unlinks a library of this workspace from a project: the project stops depending on it, the library stays.
     *
     * @param project the project; null asks which
     */
    fun unlink(
        ide: Project,
        root: Path,
        project: String?,
    ) = BackgroundFlowClient.start(ide, "MonoSmith: unlink a library") {
        runUnlink(IdePortsClient.of(ide, root), project) { CommitChangesUseCase.start(ide, root, it) }
    }

    /**
     * Runs the unlink with the given ports.
     *
     * @param commitLater starts the Commit action for an operation
     * @return true when smith unlinked the library
     */
    fun runUnlink(
        ports: FlowPortsContract,
        project: String?,
        commitLater: (String) -> Unit,
    ): Boolean {
        val node = target(ports, project, "unlink a library") ?: return false
        if (node.deps.isEmpty()) {
            ports.notice.info("${node.name} depends on no library of this workspace.")
            return false
        }
        val library =
            ports.prompt
                .pickOne(
                    "MonoSmith: unlink a library from ${node.name}",
                    "Which library should the project stop using?",
                    node.deps.map { ChoiceRowContract(it, it) },
                )?.name ?: return false
        return PlanThenApplyUseCase.run(
            ports,
            PlanStepContract(
                title = "MonoSmith: unlink $library from ${node.name}?",
                planArgs = null,
                fixedPlan = "smith uninstall -w ${node.name} workspace:$library",
                header =
                    "${node.name} stops depending on $library. The library stays in this workspace, and other projects keep using it. " +
                        "smith edits the manifest of ${node.name} and runs the install.",
                okText = "Unlink it",
                offerCommit = true,
                apply = { InstallArgumentsAlgorithm.unlink(node.name, library) },
                done = "$library is unlinked from ${node.name}",
                operation = "uninstall $library from ${node.name}",
            ),
            commitLater,
        )
    }
}
