package dev.monosmith.jetbrains.installdependency

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.flowports.FlowPortsContract
import dev.monosmith.jetbrains.flowports.IdePortsClient
import dev.monosmith.jetbrains.prompt.ChoiceRowContract
import dev.monosmith.jetbrains.smithcommand.TaskCommandAlgorithm
import dev.monosmith.jetbrains.smithrun.BackgroundFlowClient
import dev.monosmith.jetbrains.smithrun.FailureTextMapper
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDCHOICES_DEFAULT
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_NOTHING_WAS_ADDED
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_THE_FIRST_ANSWER_IS_THE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADOPTFOLDER_ADD_IT
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_INSTALLDEPENDENCY_ADD_A_DEPENDENCY
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_INSTALLDEPENDENCY_A_PACKAGE_FROM_THE_REGISTRY
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_INSTALLDEPENDENCY_DOWNLOADED_FROM_THE_LANGUAGE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_INSTALLDEPENDENCY_FOLLOW_THE_WORKSPACE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_INSTALLDEPENDENCY_LIBRARIES_MAY_USE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_INSTALLDEPENDENCY_LINKED_NOTHING_IS_DOWNLOADED
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_INSTALLDEPENDENCY_MONOSMITH_ADD_A_DEPENDENCY
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_INSTALLDEPENDENCY_NOTHING_WAS_ADDED
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_INSTALLDEPENDENCY_NO_LIBRARY_OF_THIS_WORKSPACE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_INSTALLDEPENDENCY_PACKAGE_NAMES_SEPARATED_BY
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_INSTALLDEPENDENCY_SMITH_RUNS_THIS_COMMAND_IT
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_INSTALLDEPENDENCY_THE_DEPENDENCY_IS_ADDED
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_INSTALLDEPENDENCY_THE_DEPENDENCY_IS_ADDED_NOT
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_INSTALLDEPENDENCY_TYPE_AT_LEAST_ONE_PACKAGE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_INSTALLDEPENDENCY_USES_THE_CODE_OF_ANOTHER
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_INSTALLDEPENDENCY_WHAT_DO_YOU_ADD_TO
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_INSTALLDEPENDENCY_WHICH_LIBRARY
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_INSTALLDEPENDENCY_WHICH_VERSION_OF_DOES_ASK
import dev.monosmith.jetbrains.words.SmithWordsConfig.PACKAGEPROMPT_RANGE_ANY_LABEL
import dev.monosmith.jetbrains.words.SmithWordsConfig.PACKAGEPROMPT_RANGE_ANY_MEANING
import dev.monosmith.jetbrains.words.SmithWordsConfig.PACKAGEPROMPT_RANGE_FOLLOW_MEANING
import dev.monosmith.jetbrains.words.SmithWordsConfig.PACKAGEPROMPT_SOURCE_WORKSPACE
import dev.monosmith.jetbrains.words.SmithWordsConfig.say
import dev.monosmith.jetbrains.workspace.PortProjectsUseCase
import java.nio.file.Path

/**
 * Adds a dependency to a project: a library of this workspace (`workspace:<name>`) or packages from the
 * registry (`--registry`). The person picks the source, the valid libraries are listed, and the command
 * line is shown before it runs. smith never commits here.
 */
object InstallDependencyUseCase {
    private val TITLE = say(JETBRAINS_INSTALLDEPENDENCY_MONOSMITH_ADD_A_DEPENDENCY)

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
                ports.notice.warn(say(JETBRAINS_INSTALLDEPENDENCY_NOTHING_WAS_ADDED, it.message.toString()))
                return false
            }
        val target =
            if (project != null) {
                nodes.firstOrNull { it.name == project } ?: return false
            } else {
                PortProjectsUseCase.pick(ports, nodes, say(JETBRAINS_INSTALLDEPENDENCY_ADD_A_DEPENDENCY)) ?: return false
            }
        val source = only ?: askSource(ports, target.name) ?: return false
        val args =
            if (source == "registry") {
                val typed =
                    ports.prompt.text("$TITLE: packages", say(JETBRAINS_INSTALLDEPENDENCY_PACKAGE_NAMES_SEPARATED_BY)) {
                        if (InstallArgumentsAlgorithm
                                .parsePackages(
                                    it,
                                ).isEmpty()
                        ) {
                            say(JETBRAINS_INSTALLDEPENDENCY_TYPE_AT_LEAST_ONE_PACKAGE)
                        } else {
                            null
                        }
                    } ?: return false
                InstallArgumentsAlgorithm.registry(target.name, InstallArgumentsAlgorithm.parsePackages(typed), dev = false)
            } else {
                val libraries = LinkableLibrariesPolicy.linkable(nodes, target.name)
                if (libraries.isEmpty()) {
                    ports.notice.warn(
                        say(JETBRAINS_INSTALLDEPENDENCY_NO_LIBRARY_OF_THIS_WORKSPACE, target.name.toString()),
                    )
                    return false
                }
                val rows = libraries.map { ChoiceRowContract(it.name, it.name, "${it.kind} · ${it.language}") }
                val library =
                    ports.prompt.pickOne(
                        say(JETBRAINS_INSTALLDEPENDENCY_WHICH_LIBRARY, TITLE.toString()),
                        say(JETBRAINS_INSTALLDEPENDENCY_LIBRARIES_MAY_USE, target.name.toString()),
                        rows,
                    ) ?: return false
                val range = if (target.language == "node") askRange(ports, target.name, library.name) ?: return false else null
                InstallArgumentsAlgorithm.library(target.name, library.name, range.takeUnless { it == "follow" })
            }
        val shown = TaskCommandAlgorithm.display(listOf("smith") + args)
        val answer =
            ports.prompt.plan(
                TITLE,
                say(JETBRAINS_INSTALLDEPENDENCY_SMITH_RUNS_THIS_COMMAND_IT),
                shown,
                say(JETBRAINS_ADOPTFOLDER_ADD_IT),
                offerCommit = true,
            )
                ?: return false
        val done = ports.smith.run(if (answer.commit) args + "--commit" else args, show = true)
        ports.changed()
        if (!done.ok) {
            ports.notice.warn(say(JETBRAINS_ADDPROJECT_NOTHING_WAS_ADDED, FailureTextMapper.describe(done).toString()))
            return false
        }
        ports.notice.info(
            if (answer.commit) {
                say(JETBRAINS_INSTALLDEPENDENCY_THE_DEPENDENCY_IS_ADDED)
            } else {
                say(JETBRAINS_INSTALLDEPENDENCY_THE_DEPENDENCY_IS_ADDED_NOT)
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
                say(JETBRAINS_INSTALLDEPENDENCY_WHAT_DO_YOU_ADD_TO, TITLE.toString(), project.toString()),
                say(JETBRAINS_ADDPROJECT_THE_FIRST_ANSWER_IS_THE),
                listOf(
                    ChoiceRowContract(
                        "registry",
                        say(JETBRAINS_INSTALLDEPENDENCY_A_PACKAGE_FROM_THE_REGISTRY),
                        say(JETBRAINS_ADDCHOICES_DEFAULT),
                        say(JETBRAINS_INSTALLDEPENDENCY_DOWNLOADED_FROM_THE_LANGUAGE),
                    ),
                    ChoiceRowContract(
                        "workspace",
                        say(PACKAGEPROMPT_SOURCE_WORKSPACE),
                        say(JETBRAINS_INSTALLDEPENDENCY_LINKED_NOTHING_IS_DOWNLOADED),
                        say(JETBRAINS_INSTALLDEPENDENCY_USES_THE_CODE_OF_ANOTHER, project.toString()),
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
                say(JETBRAINS_INSTALLDEPENDENCY_WHICH_VERSION_OF_DOES_ASK, TITLE.toString(), library.toString(), project.toString()),
                say(JETBRAINS_ADDPROJECT_THE_FIRST_ANSWER_IS_THE),
                listOf(
                    ChoiceRowContract(
                        "follow",
                        say(JETBRAINS_INSTALLDEPENDENCY_FOLLOW_THE_WORKSPACE),
                        say(JETBRAINS_ADDCHOICES_DEFAULT),
                        say(PACKAGEPROMPT_RANGE_FOLLOW_MEANING),
                    ),
                    ChoiceRowContract(
                        "*",
                        say(PACKAGEPROMPT_RANGE_ANY_LABEL),
                        "",
                        say(PACKAGEPROMPT_RANGE_ANY_MEANING),
                    ),
                ),
            )?.name
}
