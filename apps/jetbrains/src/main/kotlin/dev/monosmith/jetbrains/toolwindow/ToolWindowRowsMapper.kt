package dev.monosmith.jetbrains.toolwindow

import dev.monosmith.jetbrains.health.DoctorFindingContract
import dev.monosmith.jetbrains.health.DoctorReportJsonMapper
import dev.monosmith.jetbrains.workspace.SmithProjectContract

/** Builds the rows of the Projects and Health views from what smith reported. */
object ToolWindowRowsMapper {
    /**
     * The rows of the Projects view.
     *
     * @param hasWorkspace false when the folder has no smith.toml
     * @param projects what `smith projects --json` gave, or why it failed
     * @return a "New workspace" row with no workspace; "Add your first project" with no project; the projects and a "New project…" row otherwise
     */
    fun projectRows(
        hasWorkspace: Boolean,
        projects: Result<List<SmithProjectContract>>?,
    ): List<ToolWindowRowContract> {
        if (!hasWorkspace) {
            return listOf(
                ToolWindowRowContract.Message("No smith.toml in this project"),
                ToolWindowRowContract.NewWorkspace("New workspace…"),
            )
        }
        val list =
            projects?.getOrNull()
                ?: return listOf(
                    ToolWindowRowContract.Message(
                        projects?.exceptionOrNull()?.let { "smith failed: ${it.message}" } ?: "Reading the workspace…",
                    ),
                )
        if (list.isEmpty()) return listOf(ToolWindowRowContract.AddProject("Add your first project"))
        return list.map { ToolWindowRowContract.Project(it) } + ToolWindowRowContract.AddProject("New project…")
    }

    /**
     * The rows of the Health view: a summary line, then each check that did not pass (the failed ones first).
     *
     * @param findings what `smith doctor --json` gave, or why it failed
     */
    fun healthRows(
        hasWorkspace: Boolean,
        findings: Result<List<DoctorFindingContract>>?,
    ): List<ToolWindowRowContract> {
        if (!hasWorkspace) return listOf(ToolWindowRowContract.Message("No smith.toml in this project"))
        val list =
            findings?.getOrNull()
                ?: return listOf(
                    ToolWindowRowContract.Message(findings?.exceptionOrNull()?.let { "smith failed: ${it.message}" } ?: "Running doctor…"),
                )
        val summary = DoctorReportJsonMapper.summarize(list)
        val problems = list.filter { !it.ok }.sortedBy { it.warning }
        return listOf<ToolWindowRowContract>(ToolWindowRowContract.Message(DoctorReportJsonMapper.say(summary))) +
            problems.map { ToolWindowRowContract.Finding(it) }
    }

    /** The lines under a finding: its detail, the fix smith suggests, and the structured fix when there is one. */
    fun findingChildren(finding: DoctorFindingContract): List<ToolWindowRowContract> =
        listOfNotNull(
            finding.detail.takeIf { it.isNotEmpty() }?.let { ToolWindowRowContract.Message(it) },
            finding.remedy.takeIf { it.isNotEmpty() }?.let { ToolWindowRowContract.Message("fix: $it") },
            if (finding.fix != null) ToolWindowRowContract.Fix(finding) else null,
        )
}
