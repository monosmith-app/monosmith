package dev.monosmith.jetbrains.installdependency

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.flowports.FlowPortsContract
import dev.monosmith.jetbrains.flowports.IdePortsClient
import dev.monosmith.jetbrains.prompt.ChoiceRowContract
import dev.monosmith.jetbrains.smithcommand.TaskCommandAlgorithm
import dev.monosmith.jetbrains.smithrun.BackgroundFlowClient
import dev.monosmith.jetbrains.smithrun.FailureTextMapper
import dev.monosmith.jetbrains.workspace.PortProjectsUseCase
import java.nio.file.Path

/**
 * Adds a dependency to a project: a library of this workspace (`workspace:<name>`) or packages from the
 * registry (`--registry`). The person picks the source, the valid libraries are listed, and the command
 * line is shown before it runs. smith never commits here.
 */
object InstallDependencyUseCase {
    private const val TITLE = "MonoSmith: add a dependency"

    /**
     * Starts the flow in a background task.
     *
     * @param project the project that gains the dependency; null asks which
     * @param only `workspace` or `registry` to skip the source question; null asks
     */
    fun start(
        ide: Project,
        root: Path,
        project: String? = null,
        only: String? = null,
    ) = BackgroundFlowClient.start(ide, TITLE) { run(IdePortsClient.of(ide, root), project, only) }

    /**
     * Runs the flow with the given ports.
     *
     * @param project the project that gains the dependency; null asks which
     * @param only `workspace` or `registry` to skip the source question; null asks
     * @return true when smith added the dependency
     */
    fun run(
        ports: FlowPortsContract,
        project: String? = null,
        only: String? = null,
    ): Boolean {
        val nodes =
            PortProjectsUseCase.list(ports).getOrElse {
                ports.notice.warn("Nothing was added. ${it.message}")
                return false
            }
        val target =
            if (project != null) {
                nodes.firstOrNull { it.name == project } ?: return false
            } else {
                PortProjectsUseCase.pick(ports, nodes, "add a dependency") ?: return false
            }
        val source = only ?: askSource(ports, target.name) ?: return false
        val args =
            if (source == "registry") {
                val typed =
                    ports.prompt.text("$TITLE: packages", "Package names, separated by spaces (for example zod react@19):") {
                        if (InstallArgumentsAlgorithm.parsePackages(it).isEmpty()) "Type at least one package." else null
                    } ?: return false
                InstallArgumentsAlgorithm.registry(target.name, InstallArgumentsAlgorithm.parsePackages(typed), dev = false)
            } else {
                val libraries = LinkableLibrariesPolicy.linkable(nodes, target.name)
                if (libraries.isEmpty()) {
                    ports.notice.warn(
                        "No library of this workspace can be linked to ${target.name}. A library must be a lib or a package, " +
                            "in the same language, and must not already depend on ${target.name}.",
                    )
                    return false
                }
                val rows = libraries.map { ChoiceRowContract(it.name, it.name, "${it.kind} · ${it.language}") }
                val library = ports.prompt.pickOne("$TITLE: which library?", "Libraries ${target.name} may use.", rows) ?: return false
                val range = if (target.language == "node") askRange(ports, target.name, library.name) ?: return false else null
                InstallArgumentsAlgorithm.library(target.name, library.name, range.takeUnless { it == "follow" })
            }
        val shown = TaskCommandAlgorithm.display(listOf("smith") + args)
        val answer =
            ports.prompt.plan(TITLE, "smith runs this command. It commits only if you tick the box.", shown, "Add it", offerCommit = true)
                ?: return false
        val done = ports.smith.run(if (answer.commit) args + "--commit" else args, show = true)
        ports.changed()
        if (!done.ok) {
            ports.notice.warn("Nothing was added. ${FailureTextMapper.describe(done)}")
            return false
        }
        ports.notice.info(
            if (answer.commit) {
                "The dependency is added. Committed as one change: Undo reverts it."
            } else {
                "The dependency is added. Not committed: review it with the Commit window."
            },
        )
        return true
    }

    private fun askSource(
        ports: FlowPortsContract,
        project: String,
    ): String? =
        ports.prompt
            .pickOne(
                "$TITLE: what do you add to $project?",
                "The first answer is the default.",
                listOf(
                    ChoiceRowContract(
                        "registry",
                        "A package from the registry",
                        "(default)",
                        "Downloaded from the language registry (npm, PyPI, crates.io...), a copy outside this workspace.",
                    ),
                    ChoiceRowContract(
                        "workspace",
                        "A library of this workspace",
                        "linked, nothing is downloaded",
                        "$project uses the code of another project here, always its current copy.",
                    ),
                ),
            )?.name

    private fun askRange(
        ports: FlowPortsContract,
        project: String,
        library: String,
    ): String? =
        ports.prompt
            .pickOne(
                "$TITLE: which version of $library does $project ask for?",
                "The first answer is the default.",
                listOf(
                    ChoiceRowContract(
                        "follow",
                        "Follow the workspace",
                        "(default)",
                        "Use the range the workspace already uses for its own libraries. When it has none yet, smith writes * and says so.",
                    ),
                    ChoiceRowContract(
                        "*",
                        "Any version (*)",
                        "",
                        "npm uses the library of this workspace whatever its version. You never edit the line when the library changes.",
                    ),
                ),
            )?.name
}
