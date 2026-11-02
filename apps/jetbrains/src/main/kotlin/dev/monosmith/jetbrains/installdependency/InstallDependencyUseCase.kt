package dev.monosmith.jetbrains.installdependency

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.notice.NoticeClient
import dev.monosmith.jetbrains.prompt.ChoiceRowContract
import dev.monosmith.jetbrains.prompt.PromptClient
import dev.monosmith.jetbrains.smithcommand.TaskCommandAlgorithm
import dev.monosmith.jetbrains.smithrun.BackgroundFlowClient
import dev.monosmith.jetbrains.smithrun.FailureTextMapper
import dev.monosmith.jetbrains.smithrun.SmithRunClient
import dev.monosmith.jetbrains.workspace.ListProjectsUseCase
import dev.monosmith.jetbrains.workspace.PickProjectUseCase
import dev.monosmith.jetbrains.workspace.WorkspaceChangedStore
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
    ) = BackgroundFlowClient.start(ide, TITLE) {
        val nodes =
            ListProjectsUseCase.list(root).getOrElse { return@start NoticeClient.warn(ide, "Nothing was added. ${it.message}") }
        val target =
            if (project !=
                null
            ) {
                nodes.first { it.name == project }
            } else {
                PickProjectUseCase.pick(ide, nodes, "add a dependency") ?: return@start
            }
        val source = only ?: askSource(ide, target.name) ?: return@start
        val args =
            if (source == "registry") {
                val typed =
                    PromptClient.text(ide, "$TITLE: packages", "Package names, separated by spaces (for example zod react@19):") {
                        if (InstallArgumentsAlgorithm.parsePackages(it).isEmpty()) "Type at least one package." else null
                    } ?: return@start
                InstallArgumentsAlgorithm.registry(target.name, InstallArgumentsAlgorithm.parsePackages(typed), dev = false)
            } else {
                val libraries = LinkableLibrariesPolicy.linkable(nodes, target.name)
                if (libraries.isEmpty()) {
                    return@start NoticeClient.warn(
                        ide,
                        "No library of this workspace can be linked to ${target.name}. A library must be a lib or a package, " +
                            "in the same language, and must not already depend on ${target.name}.",
                    )
                }
                val rows = libraries.map { ChoiceRowContract(it.name, it.name, "${it.kind} · ${it.language}") }
                val library = PromptClient.pickOne(ide, "$TITLE: which library?", "Libraries ${target.name} may use.", rows) ?: return@start
                val range = if (target.language == "node") askRange(ide, target.name, library.name) ?: return@start else null
                InstallArgumentsAlgorithm.library(target.name, library.name, range.takeUnless { it == "follow" })
            }
        val shown = TaskCommandAlgorithm.display(listOf("smith") + args)
        PromptClient.plan(ide, TITLE, "smith runs this command. Nothing is committed.", shown, "Add it", offerCommit = false)
            ?: return@start
        val done = SmithRunClient.run(ide, root, args, show = true)
        WorkspaceChangedStore.changed(ide)
        if (done.ok) {
            NoticeClient.info(ide, "The dependency is added. Not committed: review it with the Commit window.")
        } else {
            NoticeClient.warn(ide, "Nothing was added. ${FailureTextMapper.describe(done)}")
        }
    }

    private fun askSource(
        ide: Project,
        project: String,
    ): String? =
        PromptClient
            .pickOne(
                ide,
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
        ide: Project,
        project: String,
        library: String,
    ): String? =
        PromptClient
            .pickOne(
                ide,
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
