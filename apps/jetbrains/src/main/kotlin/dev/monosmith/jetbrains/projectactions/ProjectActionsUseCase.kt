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
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_PROJECTACTIONS_DEPENDS_ON_NO_LIBRARY_OF
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_PROJECTACTIONS_IS_NOW
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_PROJECTACTIONS_IS_NOW_WHAT_SHOULD_IT
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_PROJECTACTIONS_IS_REMOVED
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_PROJECTACTIONS_IS_UNLINKED_FROM
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_PROJECTACTIONS_KEEP_A_COPY_MOVES_THE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_PROJECTACTIONS_MAKE_IT
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_PROJECTACTIONS_MAKE_IT_AN_APP_A
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_PROJECTACTIONS_MONOSMITH_CHANGE_THE_KIND_OF
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_PROJECTACTIONS_MONOSMITH_MAKE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_PROJECTACTIONS_MONOSMITH_MAKE_A
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_PROJECTACTIONS_MONOSMITH_REMOVE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_PROJECTACTIONS_MONOSMITH_REMOVE_A_PROJECT
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_PROJECTACTIONS_MONOSMITH_UNLINK_A_LIBRARY
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_PROJECTACTIONS_MONOSMITH_UNLINK_A_LIBRARY_FROM
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_PROJECTACTIONS_MONOSMITH_UNLINK_FROM
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_PROJECTACTIONS_NOTHING_WAS_CHANGED
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_PROJECTACTIONS_NOTHING_WAS_CHANGED_YET_SMITH
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_PROJECTACTIONS_REMOVE_A_PROJECT
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_PROJECTACTIONS_REMOVE_IT_AND_DELETE_THE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_PROJECTACTIONS_REMOVE_IT_AND_KEEP_A
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_PROJECTACTIONS_STOPS_DEPENDING_ON_THE_LIBRARY
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_PROJECTACTIONS_UNLINK_A_LIBRARY
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_PROJECTACTIONS_UNLINK_IT
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_PROJECTACTIONS_WHICH_LIBRARY_SHOULD_THE_PROJECT
import dev.monosmith.jetbrains.words.SmithWordsConfig.say
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
            ports.notice.warn(say(JETBRAINS_PROJECTACTIONS_NOTHING_WAS_CHANGED, listed.exceptionOrNull()?.message.toString()))
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
    ) = BackgroundFlowClient.start(ide, say(JETBRAINS_PROJECTACTIONS_MONOSMITH_REMOVE_A_PROJECT)) {
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
        val node = target(ports, project, say(JETBRAINS_PROJECTACTIONS_REMOVE_A_PROJECT)) ?: return false
        return PlanThenApplyUseCase.run(
            ports,
            PlanStepContract(
                title = say(JETBRAINS_PROJECTACTIONS_MONOSMITH_REMOVE, node.name.toString()),
                planArgs = ProjectActionArgumentsAlgorithm.removePlan(node.name),
                header =
                    say(JETBRAINS_PROJECTACTIONS_KEEP_A_COPY_MOVES_THE),
                okText = say(JETBRAINS_PROJECTACTIONS_REMOVE_IT_AND_KEEP_A),
                offerCommit = true,
                others = listOf(say(JETBRAINS_PROJECTACTIONS_REMOVE_IT_AND_DELETE_THE)),
                apply = { ProjectActionArgumentsAlgorithm.remove(node.name, delete = it.other == 0) },
                done = say(JETBRAINS_PROJECTACTIONS_IS_REMOVED, node.name.toString()),
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
    ) = BackgroundFlowClient.start(ide, say(JETBRAINS_PROJECTACTIONS_MONOSMITH_CHANGE_THE_KIND_OF)) {
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
        val node = target(ports, project, say(JETBRAINS_PROJECTACTIONS_MAKE_IT_AN_APP_A)) ?: return false
        val kind =
            ports.prompt
                .pickOne(
                    say(JETBRAINS_PROJECTACTIONS_MONOSMITH_MAKE_A, node.name.toString()),
                    say(
                        JETBRAINS_PROJECTACTIONS_IS_NOW_WHAT_SHOULD_IT,
                        node.name.toString(),
                        ProjectActionArgumentsAlgorithm.withArticle(node.kind).toString(),
                    ),
                    ProjectActionArgumentsAlgorithm.kindRows(node.kind),
                )?.name ?: return false
        return PlanThenApplyUseCase.run(
            ports,
            PlanStepContract(
                title =
                    say(
                        JETBRAINS_PROJECTACTIONS_MONOSMITH_MAKE,
                        node.name.toString(),
                        ProjectActionArgumentsAlgorithm.withArticle(kind).toString(),
                    ),
                planArgs = ProjectActionArgumentsAlgorithm.kindPlan(node.name, kind),
                header = say(JETBRAINS_PROJECTACTIONS_NOTHING_WAS_CHANGED_YET_SMITH),
                okText = say(JETBRAINS_PROJECTACTIONS_MAKE_IT, ProjectActionArgumentsAlgorithm.withArticle(kind).toString()),
                offerCommit = true,
                apply = { ProjectActionArgumentsAlgorithm.kind(node.name, kind) },
                done =
                    say(
                        JETBRAINS_PROJECTACTIONS_IS_NOW,
                        node.name.toString(),
                        ProjectActionArgumentsAlgorithm.withArticle(kind).toString(),
                    ),
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
    ) = BackgroundFlowClient.start(ide, say(JETBRAINS_PROJECTACTIONS_MONOSMITH_UNLINK_A_LIBRARY)) {
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
        val node = target(ports, project, say(JETBRAINS_PROJECTACTIONS_UNLINK_A_LIBRARY)) ?: return false
        if (node.deps.isEmpty()) {
            ports.notice.info(say(JETBRAINS_PROJECTACTIONS_DEPENDS_ON_NO_LIBRARY_OF, node.name.toString()))
            return false
        }
        val library =
            ports.prompt
                .pickOne(
                    say(JETBRAINS_PROJECTACTIONS_MONOSMITH_UNLINK_A_LIBRARY_FROM, node.name.toString()),
                    say(JETBRAINS_PROJECTACTIONS_WHICH_LIBRARY_SHOULD_THE_PROJECT),
                    node.deps.map { ChoiceRowContract(it, it) },
                )?.name ?: return false
        return PlanThenApplyUseCase.run(
            ports,
            PlanStepContract(
                title = say(JETBRAINS_PROJECTACTIONS_MONOSMITH_UNLINK_FROM, library.toString(), node.name.toString()),
                planArgs = null,
                fixedPlan = "smith uninstall -p ${node.name} workspace:$library",
                header =
                    say(JETBRAINS_PROJECTACTIONS_STOPS_DEPENDING_ON_THE_LIBRARY, node.name.toString(), library.toString()),
                okText = say(JETBRAINS_PROJECTACTIONS_UNLINK_IT),
                offerCommit = true,
                apply = { InstallArgumentsAlgorithm.unlink(node.name, library) },
                done = say(JETBRAINS_PROJECTACTIONS_IS_UNLINKED_FROM, library.toString(), node.name.toString()),
                operation = "uninstall $library from ${node.name}",
            ),
            commitLater,
        )
    }
}
