package dev.monosmith.jetbrains.toolwindow

import dev.monosmith.jetbrains.dependencies.UpdatesJsonMapper
import dev.monosmith.jetbrains.dependencies.UpdatesReportContract
import dev.monosmith.jetbrains.health.DoctorFindingContract
import dev.monosmith.jetbrains.health.DoctorReportJsonMapper
import dev.monosmith.jetbrains.releaseflow.ReleasePreviewContract
import dev.monosmith.jetbrains.releaseflow.ReleasePreviewMapper
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
        affected: Set<String> = emptySet(),
        query: String = "",
        affectedOnly: Boolean = false,
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
        val shown = list.filter { matches(it, query) && (!affectedOnly || it.name in affected) }
        val rows = shown.map { ToolWindowRowContract.Project(it, it.name in affected) }
        val none =
            if (shown.isEmpty()) {
                listOf(
                    ToolWindowRowContract.Message(
                        if (affectedOnly &&
                            query.isEmpty()
                        ) {
                            "No project is affected by the changes."
                        } else {
                            "No project matches the filter."
                        },
                    ),
                )
            } else {
                emptyList()
            }
        return rows + none + ToolWindowRowContract.AddProject("New project…")
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

    /** Whether a project passes the filter: every word of [query] is in its name, kind, language or one of its tasks. */
    fun matches(
        project: SmithProjectContract,
        query: String,
    ): Boolean {
        val words = query.lowercase().split(Regex("\\s+")).filter { it.isNotEmpty() }
        val haystack = "${project.name} ${project.kind} ${project.language} ${project.tasks.joinToString(" ")}".lowercase()
        return words.all { it in haystack }
    }

    /**
     * The rows of the Dependencies view: two actions, then the updates grouped by risk (the riskiest
     * group first, each with its lines under it), then the tools that have a newer version.
     */
    fun depsRows(report: Result<UpdatesReportContract>?): List<ToolWindowRowContract> {
        val read =
            report?.getOrNull()
                ?: return listOf(
                    ToolWindowRowContract.Message(
                        report?.exceptionOrNull()?.let { "smith failed: ${it.message}" } ?: "Reading the updates…",
                    ),
                )
        val actions =
            listOf(ToolWindowRowContract.Link("Update dependencies…", "update"), ToolWindowRowContract.Link("Align versions…", "align"))
        if (read.updates.isEmpty() &&
            read.tools.isEmpty()
        ) {
            return actions + ToolWindowRowContract.Message("Every dependency is up to date.")
        }
        val groups =
            UpdatesJsonMapper.byRisk(read.updates).map { (level, list) ->
                ToolWindowRowContract.Group("${UpdatesJsonMapper.levelWords(level)} (${list.size})", list.map(UpdatesJsonMapper::line))
            }
        val tools =
            if (read.tools.isEmpty()) {
                emptyList()
            } else {
                listOf(
                    ToolWindowRowContract
                        .Group(
                            "Tools with a newer version (${read.tools.size})",
                            read.tools.map {
                                "${it.name} ${it.pinned} → ${it.latest}"
                            },
                        ),
                )
            }
        return actions + groups + tools
    }

    /** The rows of the Release view: a Release… action, then the plan with the reason and commits of each package, or why nothing releases. */
    fun releaseRows(preview: Result<ReleasePreviewContract>?): List<ToolWindowRowContract> {
        val read =
            preview?.getOrNull()
                ?: return listOf(
                    ToolWindowRowContract.Message(
                        preview?.exceptionOrNull()?.let { "smith failed: ${it.message}" } ?: "Reading the release plan…",
                    ),
                )
        return listOf<ToolWindowRowContract>(ToolWindowRowContract.Link("Release… (hold, bump, force)", "release")) +
            ReleasePreviewMapper.rows(read).map { ToolWindowRowContract.Group(it.text, it.children) }
    }

    /** The lines under a finding: its detail, the fix smith suggests, and the structured fix when there is one. */
    fun findingChildren(finding: DoctorFindingContract): List<ToolWindowRowContract> =
        listOfNotNull(
            finding.detail.takeIf { it.isNotEmpty() }?.let { ToolWindowRowContract.Message(it) },
            finding.remedy.takeIf { it.isNotEmpty() }?.let { ToolWindowRowContract.Message("fix: $it") },
            if (finding.fix != null) ToolWindowRowContract.Fix(finding) else null,
        )
}
