package dev.monosmith.jetbrains.cachecontrol

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.flowports.FlowPortsContract
import dev.monosmith.jetbrains.flowports.IdePortsClient
import dev.monosmith.jetbrains.planapply.PlanStepContract
import dev.monosmith.jetbrains.planapply.PlanThenApplyUseCase
import dev.monosmith.jetbrains.smithrun.BackgroundFlowClient
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_CACHECONTROL_DELETE_THE_CACHE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_CACHECONTROL_IT_DELETES_CORRUPT_BLOBS_THE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_CACHECONTROL_IT_DELETES_THE_LOCAL_CACHE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_CACHECONTROL_IT_REMOVES_THE_LEAST_RECENTLY
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_CACHECONTROL_MONOSMITH_CACHE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_CACHECONTROL_NOTHING_WAS_CHANGED_YET
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_CACHECONTROL_REMOVE_WHAT_IS_BROKEN
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_CACHECONTROL_THE_CACHE_IS_TRIMMED
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_CACHECONTROL_THE_LOCAL_CACHE_IS_DELETED
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_CACHECONTROL_TRIM_THE_CACHE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_CACHECONTROL_WHAT_WAS_BROKEN_IS_REMOVED
import dev.monosmith.jetbrains.words.SmithWordsConfig.say
import java.nio.file.Path

/**
 * The three ways to change the local cache of smith. Each tells what it costs and asks before it runs:
 * trimming shows the plan of `smith cache prune --dry-run`; removing what is broken and deleting everything
 * show the consequence, because smith has no plan for them. The remote cache is never touched. smith has no
 * `--commit` here, so nothing is committed.
 */
object CacheControlUseCase {
    private val TITLE = say(JETBRAINS_CACHECONTROL_MONOSMITH_CACHE)

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
            say(JETBRAINS_CACHECONTROL_REMOVE_WHAT_IS_BROKEN),
            say(JETBRAINS_CACHECONTROL_IT_DELETES_CORRUPT_BLOBS_THE),
            listOf("cache", "verify", "--fix"),
            listOf("cache", "verify", "--fix"),
            null,
            say(JETBRAINS_CACHECONTROL_WHAT_WAS_BROKEN_IS_REMOVED),
        ),

        /** `smith cache prune`: removes the least recently used entries. */
        PRUNE(
            say(JETBRAINS_CACHECONTROL_TRIM_THE_CACHE),
            say(JETBRAINS_CACHECONTROL_IT_REMOVES_THE_LEAST_RECENTLY),
            listOf("cache", "prune"),
            listOf("cache", "prune"),
            listOf("cache", "prune", "--dry-run"),
            say(JETBRAINS_CACHECONTROL_THE_CACHE_IS_TRIMMED),
        ),

        /** `smith cache clear --yes`: deletes the local cache. */
        CLEAR(
            say(JETBRAINS_CACHECONTROL_DELETE_THE_CACHE),
            say(JETBRAINS_CACHECONTROL_IT_DELETES_THE_LOCAL_CACHE),
            listOf("cache", "clear"),
            listOf("cache", "clear", "--yes"),
            null,
            say(JETBRAINS_CACHECONTROL_THE_LOCAL_CACHE_IS_DELETED),
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
                header = say(JETBRAINS_CACHECONTROL_NOTHING_WAS_CHANGED_YET, kind.consequence.toString()),
                okText = kind.verb,
                offerCommit = false,
                apply = { kind.applyArgs },
                done = kind.done,
            ),
        ) {}
}
