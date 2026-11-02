package dev.monosmith.jetbrains.dependencies

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.commitchanges.CommitChangesUseCase
import dev.monosmith.jetbrains.flowports.FlowPortsContract
import dev.monosmith.jetbrains.flowports.IdePortsClient
import dev.monosmith.jetbrains.planapply.PlanStepContract
import dev.monosmith.jetbrains.planapply.PlanThenApplyUseCase
import dev.monosmith.jetbrains.prompt.ChoiceRowContract
import dev.monosmith.jetbrains.smithrun.BackgroundFlowClient
import dev.monosmith.jetbrains.smithrun.FailureTextMapper
import java.nio.file.Path

/**
 * Updates dependencies (`smith up`) and aligns the versions the projects declare (`smith up --align`).
 * The updates come from `smith up --json`, grouped by risk. The person picks what to update, sees the plan
 * (each line says what moves and where), and only then smith writes with `--yes`; a major update needs a
 * pick of its own because `--major` lets it through. smith commits only when the box is ticked.
 */
object UpdateDependenciesUseCase {
    private const val TITLE = "MonoSmith: update dependencies"
    private const val SAFE = "safe"
    private const val CHOOSE = "choose"

    /** The rules `smith up --align --prefer` takes; smith's own pick is first. */
    val ruleRows =
        listOf(
            ChoiceRowContract(
                "installed",
                "The installed version (smith's pick)",
                "(default)",
                "The version the package manager has installed.",
            ),
            ChoiceRowContract("highest", "The highest declared", "", "Projects on an older version move up."),
            ChoiceRowContract("lowest", "The lowest declared", "", "Projects on a newer version move down."),
        )

    /** Reads what `smith up --json` says, for a view. Blocks: call it off the UI thread. */
    fun read(ports: FlowPortsContract): Result<UpdatesReportContract> {
        val run = ports.smith.run(listOf("up", "--json"))
        if (!run.ok) return Result.failure(IllegalStateException(FailureTextMapper.describe(run)))
        return runCatching { UpdatesJsonMapper.toReport(run.stdout) }
    }

    /** Starts the update flow in a background task. */
    fun start(
        ide: Project,
        root: Path,
    ) = BackgroundFlowClient.start(ide, TITLE) { update(IdePortsClient.of(ide, root)) { CommitChangesUseCase.start(ide, root, it) } }

    /** Starts the align flow in a background task. */
    fun startAlign(
        ide: Project,
        root: Path,
    ) = BackgroundFlowClient.start(ide, "MonoSmith: align dependency versions") {
        align(IdePortsClient.of(ide, root)) { CommitChangesUseCase.start(ide, root, it) }
    }

    /**
     * The update flow.
     *
     * @return true when smith wrote the update
     */
    fun update(
        ports: FlowPortsContract,
        commitLater: (String) -> Unit,
    ): Boolean {
        val report = read(ports).getOrElse { return false.also { _ -> ports.notice.warn("Nothing was updated. ${it.message}") } }
        if (report.updates.isEmpty()) {
            ports.notice.info("Every dependency is up to date." + tools(report))
            return false
        }
        val safe = report.updates.filter { it.level != "major" }
        val rows = mutableListOf<ChoiceRowContract>()
        if (safe.isNotEmpty()) {
            rows.add(
                ChoiceRowContract(
                    SAFE,
                    "Every patch and minor update (${safe.size})",
                    "(default)",
                    "The risk is low: fixes and new features, no breaking change.",
                ),
            )
        }
        rows.add(
            ChoiceRowContract(
                CHOOSE,
                "Choose the packages…",
                "",
                "Pick among ${report.updates.size} updates, grouped by risk. Major updates are never ticked.",
            ),
        )
        val way =
            ports.prompt
                .pickOne(
                    TITLE,
                    "${UpdatesJsonMapper.byRisk(
                        report.updates,
                    ).joinToString("; ") { (level, list) -> "${list.size} $level" }}. ${tools(report)}".trim(),
                    rows,
                )?.name
                ?: return false
        val picked =
            if (way == SAFE) {
                safe
            } else {
                val names =
                    ports.prompt.pickMany(
                        TITLE,
                        "Tick what to update. A major update may break your code.",
                        UpdatesJsonMapper.rows(report.updates),
                    )
                        ?: return false
                report.updates.filter { it.name in names }
            }
        if (picked.isEmpty()) return false
        val named = if (way == SAFE) emptyList() else picked.map { it.name }
        val args = UpdatesJsonMapper.arguments(named, picked.any { it.level == "major" })
        return PlanThenApplyUseCase.run(
            ports,
            PlanStepContract(
                title = "$TITLE?",
                planArgs = null,
                fixedPlan = UpdatesJsonMapper.describe(picked) + "\n\nsmith ${args.joinToString(" ")}",
                header = "Nothing was changed yet. smith edits the manifests and installs. By default smith leaves the files uncommitted.",
                okText = "Update",
                offerCommit = true,
                apply = { args },
                done = "${picked.size} dependencies are updated",
                operation = "up",
            ),
            commitLater,
        )
    }

    /**
     * The align flow: which version wins, the plan of `up --align --dry-run`, then `--yes`.
     *
     * @return true when smith wrote the change
     */
    fun align(
        ports: FlowPortsContract,
        commitLater: (String) -> Unit,
    ): Boolean {
        val rule =
            ports.prompt
                .pickOne(
                    "MonoSmith: align dependency versions",
                    "When projects declare a dependency at different versions, which one wins? The first answer is the default.",
                    ruleRows,
                )?.name
                ?: return false
        return PlanThenApplyUseCase.run(
            ports,
            PlanStepContract(
                title = "MonoSmith: align the versions ($rule)?",
                planArgs = UpdatesJsonMapper.alignArguments(rule, apply = false),
                header = "Nothing was changed yet. Every project that disagrees moves to one version.",
                okText = "Align",
                offerCommit = true,
                apply = { UpdatesJsonMapper.alignArguments(rule, apply = true) },
                done = "The versions are aligned",
                operation = "up align",
            ),
            commitLater,
        )
    }

    private fun tools(report: UpdatesReportContract): String =
        if (report.tools.isEmpty()) "" else " Newer tools: ${report.tools.joinToString(", ") { "${it.name} ${it.pinned} → ${it.latest}" }}."
}
