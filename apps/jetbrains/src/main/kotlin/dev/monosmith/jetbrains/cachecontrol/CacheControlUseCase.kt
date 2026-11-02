package dev.monosmith.jetbrains.cachecontrol

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.flowports.FlowPortsContract
import dev.monosmith.jetbrains.flowports.IdePortsClient
import dev.monosmith.jetbrains.planapply.PlanStepContract
import dev.monosmith.jetbrains.planapply.PlanThenApplyUseCase
import dev.monosmith.jetbrains.smithrun.BackgroundFlowClient
import java.nio.file.Path

/**
 * The three ways to change the local cache of smith. Each tells what it costs and asks before it runs:
 * trimming shows the plan of `smith cache prune --dry-run`; removing what is broken and deleting everything
 * show the consequence, because smith has no plan for them. The remote cache is never touched. smith has no
 * `--commit` here, so nothing is committed.
 */
object CacheControlUseCase {
    private const val TITLE = "MonoSmith: cache"

    /** The kinds of change. */
    enum class Kind(
        val verb: String,
        val consequence: String,
        val args: List<String>,
        val applyArgs: List<String>,
        val planArgs: List<String>?,
        val done: String,
    ) {
        /** `smith cache verify --fix`: deletes what the check finds broken. */
        REPAIR(
            "Remove what is broken",
            "It deletes corrupt blobs, the entries that name them, unused blobs and leftover files. The next run of an affected task is slower.",
            listOf("cache", "verify", "--fix"),
            listOf("cache", "verify", "--fix"),
            null,
            "What was broken is removed from the cache",
        ),

        /** `smith cache prune`: removes the least recently used entries. */
        PRUNE(
            "Trim the cache",
            "It removes the least recently used entries, then the blobs no entry uses.",
            listOf("cache", "prune"),
            listOf("cache", "prune"),
            listOf("cache", "prune", "--dry-run"),
            "The cache is trimmed",
        ),

        /** `smith cache clear --yes`: deletes the local cache. */
        CLEAR(
            "Delete the cache",
            "It deletes the local cache. The remote cache stays, and the next run of every task is slower.",
            listOf("cache", "clear"),
            listOf("cache", "clear", "--yes"),
            null,
            "The local cache is deleted",
        ),
    }

    /** Starts the change in a background task. */
    fun start(
        ide: Project,
        root: Path,
        kind: Kind,
    ) = BackgroundFlowClient.start(ide, TITLE) { run(IdePortsClient.of(ide, root), kind) }

    /**
     * @return true when smith changed the cache
     */
    fun run(
        ports: FlowPortsContract,
        kind: Kind,
    ): Boolean =
        PlanThenApplyUseCase.run(
            ports,
            PlanStepContract(
                title = "$TITLE: ${kind.verb.replaceFirstChar { it.lowercase() }}?",
                planArgs = kind.planArgs,
                fixedPlan = "smith ${kind.args.joinToString(" ")}\n\n${kind.consequence}",
                header = "${kind.consequence} Nothing was changed yet.",
                okText = kind.verb,
                offerCommit = false,
                apply = { kind.applyArgs },
                done = kind.done,
            ),
        ) {}
}
