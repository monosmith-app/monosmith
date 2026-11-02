package dev.monosmith.jetbrains.toolwindow

import dev.monosmith.jetbrains.health.DoctorFindingContract
import dev.monosmith.jetbrains.workspace.SmithProjectContract

/** One row of a view of the MonoSmith tool window. */
sealed interface ToolWindowRowContract {
    /** The text of the row. */
    val text: String

    /** A line of information with no action. */
    data class Message(
        override val text: String,
    ) : ToolWindowRowContract

    /** A row that starts the Add project flow: "Add your first project" or "New project…". */
    data class AddProject(
        override val text: String,
    ) : ToolWindowRowContract

    /** A row that starts the New workspace flow, shown when the folder has no smith.toml. */
    data class NewWorkspace(
        override val text: String,
    ) : ToolWindowRowContract

    /** A project, with its tasks as children. */
    data class Project(
        val project: SmithProjectContract,
    ) : ToolWindowRowContract {
        override val text: String get() = "${project.name} (${project.language} ${project.kind})"
    }

    /** A check of the doctor that did not pass, with its fix as a child when it has one. */
    data class Finding(
        val finding: DoctorFindingContract,
    ) : ToolWindowRowContract {
        override val text: String get() = "${if (finding.warning) "warning" else "failed"}: ${finding.check}"
    }

    /** The row under a finding that runs its structured fix. */
    data class Fix(
        val finding: DoctorFindingContract,
    ) : ToolWindowRowContract {
        override val text: String get() = "Fix this check: ${finding.fix?.description.orEmpty()}"
    }
}
