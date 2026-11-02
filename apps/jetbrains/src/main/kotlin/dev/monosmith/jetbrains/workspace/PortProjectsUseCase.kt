package dev.monosmith.jetbrains.workspace

import dev.monosmith.jetbrains.flowports.FlowPortsContract
import dev.monosmith.jetbrains.prompt.ChoiceRowContract
import dev.monosmith.jetbrains.smithrun.FailureTextMapper

/**
 * Reads and picks the projects of the workspace through the ports of a flow, so a flow that needs a
 * project runs in a unit test with a fake smith and a fake person.
 */
object PortProjectsUseCase {
    /**
     * Asks smith for the projects (`smith projects --json`).
     *
     * @return the projects, or the words of the failure
     */
    fun list(ports: FlowPortsContract): Result<List<SmithProjectContract>> {
        val listed = ports.smith.run(listOf("projects", "--json"))
        if (!listed.ok) return Result.failure(IllegalStateException(FailureTextMapper.describe(listed)))
        return runCatching { ProjectsJsonMapper.toProjects(listed.stdout) }
            .recoverCatching {
                throw IllegalStateException(
                    "smith printed its projects in a form this plugin does not read. Update the plugin.",
                )
            }
    }

    /**
     * Asks which project an action works on.
     *
     * @param projects the projects to choose from
     * @param title what the action does, for the title of the dialog
     * @return the project, or null when the person cancels
     */
    fun pick(
        ports: FlowPortsContract,
        projects: List<SmithProjectContract>,
        title: String,
    ): SmithProjectContract? {
        val rows = projects.map { ChoiceRowContract(it.name, it.name, "${it.kind} · ${it.language}") }
        val answer = ports.prompt.pickOne("MonoSmith: $title", "Which project?", rows) ?: return null
        return projects.first { it.name == answer.name }
    }
}
