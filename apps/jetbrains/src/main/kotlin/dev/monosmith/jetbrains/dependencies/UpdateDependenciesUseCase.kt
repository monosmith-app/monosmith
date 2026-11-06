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
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDCHOICES_DEFAULT
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_DEPENDENCIES_ALIGN
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_DEPENDENCIES_CHOOSE_THE_PACKAGES
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_DEPENDENCIES_DEPENDENCIES_ARE_UPDATED
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_DEPENDENCIES_EVERY_DEPENDENCY_IS_UP_TO
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_DEPENDENCIES_EVERY_PATCH_AND_MINOR_UPDATE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_DEPENDENCIES_MONOSMITH_ALIGN_DEPENDENCY_VERSIONS
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_DEPENDENCIES_MONOSMITH_ALIGN_THE_VERSIONS
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_DEPENDENCIES_MONOSMITH_UPDATE_DEPENDENCIES
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_DEPENDENCIES_NEWER_TOOLS
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_DEPENDENCIES_NOTHING_WAS_CHANGED_YET_EVERY
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_DEPENDENCIES_NOTHING_WAS_CHANGED_YET_SMITH
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_DEPENDENCIES_NOTHING_WAS_UPDATED
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_DEPENDENCIES_PICK_AMONG_UPDATES_GROUPED_BY
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_DEPENDENCIES_PROJECTS_ON_AN_OLDER_VERSION
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_DEPENDENCIES_PROJECTS_ON_A_NEWER_VERSION
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_DEPENDENCIES_THE_RISK_IS_LOW_FIXES
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_DEPENDENCIES_THE_VERSIONS_ARE_ALIGNED
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_DEPENDENCIES_THE_VERSION_THE_PACKAGE_MANAGER
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_DEPENDENCIES_TICK_WHAT_TO_UPDATE_A
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_DEPENDENCIES_UPDATE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_DEPENDENCIES_WHEN_PROJECTS_DECLARE_A_DEPENDENCY
import dev.monosmith.jetbrains.words.SmithWordsConfig.RESULTDIALOG_VERSION_HIGHEST_LABEL
import dev.monosmith.jetbrains.words.SmithWordsConfig.RESULTDIALOG_VERSION_INSTALLED_LABEL
import dev.monosmith.jetbrains.words.SmithWordsConfig.RESULTDIALOG_VERSION_LOWEST_LABEL
import dev.monosmith.jetbrains.words.SmithWordsConfig.say
import java.nio.file.Path

/**
 * Updates dependencies (`smith up`) and aligns the versions the projects declare (`smith up --align`).
 * The updates come from `smith up --json`, grouped by risk. The person picks what to update, sees the plan
 * (each line says what moves and where), and only then smith writes with `--yes`; a major update needs a
 * pick of its own because `--major` lets it through. smith commits only when the box is ticked.
 */
object UpdateDependenciesUseCase {
    private val TITLE = say(JETBRAINS_DEPENDENCIES_MONOSMITH_UPDATE_DEPENDENCIES)
    private const val SAFE = "safe"
    private const val CHOOSE = "choose"

    /** The rules `smith up --align --prefer` takes; smith's own pick is first. */
    val ruleRows =
        listOf(
            ChoiceRowContract(
                "installed",
                say(RESULTDIALOG_VERSION_INSTALLED_LABEL),
                say(JETBRAINS_ADDCHOICES_DEFAULT),
                say(JETBRAINS_DEPENDENCIES_THE_VERSION_THE_PACKAGE_MANAGER),
            ),
            ChoiceRowContract(
                "highest",
                say(RESULTDIALOG_VERSION_HIGHEST_LABEL),
                "",
                say(JETBRAINS_DEPENDENCIES_PROJECTS_ON_AN_OLDER_VERSION),
            ),
            ChoiceRowContract(
                "lowest",
                say(RESULTDIALOG_VERSION_LOWEST_LABEL),
                "",
                say(JETBRAINS_DEPENDENCIES_PROJECTS_ON_A_NEWER_VERSION),
            ),
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
    ) = BackgroundFlowClient.start(ide, say(JETBRAINS_DEPENDENCIES_MONOSMITH_ALIGN_DEPENDENCY_VERSIONS)) {
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
        val report =
            read(ports).getOrElse {
                return false.also { _ ->
                    ports.notice.warn(say(JETBRAINS_DEPENDENCIES_NOTHING_WAS_UPDATED, it.message.toString()))
                }
            }
        if (report.updates.isEmpty()) {
            ports.notice.info(say(JETBRAINS_DEPENDENCIES_EVERY_DEPENDENCY_IS_UP_TO) + tools(report))
            return false
        }
        val safe = report.updates.filter { it.level != "major" }
        val rows = mutableListOf<ChoiceRowContract>()
        if (safe.isNotEmpty()) {
            rows.add(
                ChoiceRowContract(
                    SAFE,
                    say(JETBRAINS_DEPENDENCIES_EVERY_PATCH_AND_MINOR_UPDATE, safe.size.toString()),
                    say(JETBRAINS_ADDCHOICES_DEFAULT),
                    say(JETBRAINS_DEPENDENCIES_THE_RISK_IS_LOW_FIXES),
                ),
            )
        }
        rows.add(
            ChoiceRowContract(
                CHOOSE,
                say(JETBRAINS_DEPENDENCIES_CHOOSE_THE_PACKAGES),
                "",
                say(JETBRAINS_DEPENDENCIES_PICK_AMONG_UPDATES_GROUPED_BY, report.updates.size.toString()),
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
                        say(JETBRAINS_DEPENDENCIES_TICK_WHAT_TO_UPDATE_A),
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
                header = say(JETBRAINS_DEPENDENCIES_NOTHING_WAS_CHANGED_YET_SMITH),
                okText = say(JETBRAINS_DEPENDENCIES_UPDATE),
                offerCommit = true,
                apply = { args },
                done = say(JETBRAINS_DEPENDENCIES_DEPENDENCIES_ARE_UPDATED, picked.size.toString()),
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
                    say(JETBRAINS_DEPENDENCIES_MONOSMITH_ALIGN_DEPENDENCY_VERSIONS),
                    say(JETBRAINS_DEPENDENCIES_WHEN_PROJECTS_DECLARE_A_DEPENDENCY),
                    ruleRows,
                )?.name
                ?: return false
        return PlanThenApplyUseCase.run(
            ports,
            PlanStepContract(
                title = say(JETBRAINS_DEPENDENCIES_MONOSMITH_ALIGN_THE_VERSIONS, rule.toString()),
                planArgs = UpdatesJsonMapper.alignArguments(rule, apply = false),
                header = say(JETBRAINS_DEPENDENCIES_NOTHING_WAS_CHANGED_YET_EVERY),
                okText = say(JETBRAINS_DEPENDENCIES_ALIGN),
                offerCommit = true,
                apply = { UpdatesJsonMapper.alignArguments(rule, apply = true) },
                done = say(JETBRAINS_DEPENDENCIES_THE_VERSIONS_ARE_ALIGNED),
                operation = "up align",
            ),
            commitLater,
        )
    }

    private fun tools(report: UpdatesReportContract): String =
        if (report.tools.isEmpty()) {
            ""
        } else {
            say(
                JETBRAINS_DEPENDENCIES_NEWER_TOOLS,
                (
                    report.tools.joinToString(", ") {
                        "${it.name} ${it.pinned} → ${it.latest}"
                    }
                ).toString(),
            )
        }
}
