package dev.monosmith.jetbrains.releaseflow

import dev.monosmith.jetbrains.jsonread.JsonFieldsMapper
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_RELEASEFLOW_CHANNEL_ITS_VERSIONS_ARE_PRERELEASES
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_RELEASEFLOW_CHANNEL_STABLE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_RELEASEFLOW_NOTHING_TO_RELEASE_A_PACKAGE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_RELEASEFLOW_NO_COMMIT_RELEASES
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_RELEASEFLOW_NO_COMMIT_RELEASES_IT_RELEASE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_RELEASEFLOW_THE_PLAN_STOPPED
import dev.monosmith.jetbrains.words.SmithWordsConfig.say

/**
 * One package a release would publish, with why.
 *
 * @property reason why the package releases
 * @property commits the commits that ask for it
 * @property forced the person asked for it although no commit does
 */
data class ReleaseResultContract(
    val project: String,
    val current: String,
    val next: String,
    val level: String,
    val reason: String,
    val commits: List<String>,
    val forced: Boolean,
    val error: String,
)

/**
 * What `smith release --dry-run --json` says.
 *
 * @property channel the release channel of the branch (`next`, `beta`); empty on the stable line, whose versions are not prereleases
 * @property unchanged the packages with nothing to release
 * @property error set when the plan stopped early
 */
data class ReleasePreviewContract(
    val channel: String,
    val results: List<ReleaseResultContract>,
    val unchanged: List<String>,
    val error: String,
)

/** One row of the Release view: a line of text with its detail lines under it. */
data class ReleaseRowContract(
    val text: String,
    val children: List<String> = emptyList(),
)

/** Reads the preview of a release and lays it out as rows: the plan, why nothing releases, and what to do about it. */
object ReleasePreviewMapper {
    /**
     * @param json what smith printed
     * @return the preview
     * @throws IllegalStateException when the text is not a JSON object
     */
    fun toPreview(json: String): ReleasePreviewContract {
        val root = JsonFieldsMapper.parseObject(json)
        return ReleasePreviewContract(
            channel = JsonFieldsMapper.text(root, "channel"),
            results =
                JsonFieldsMapper
                    .records(root, "results")
                    .map {
                        ReleaseResultContract(
                            JsonFieldsMapper.text(it, "project"),
                            JsonFieldsMapper.text(it, "current"),
                            JsonFieldsMapper.text(it, "next"),
                            JsonFieldsMapper.text(it, "level"),
                            JsonFieldsMapper.text(it, "reason"),
                            JsonFieldsMapper.words(it, "commits"),
                            JsonFieldsMapper.flag(it, "forced"),
                            JsonFieldsMapper.text(it, "error"),
                        )
                    }.filter { it.project.isNotEmpty() },
            unchanged = JsonFieldsMapper.words(root, "unchanged").filter { it.isNotEmpty() },
            error = JsonFieldsMapper.text(root, "error"),
        )
    }

    /** The rows of the Release view. Each package has its reason and commits under it. */
    fun rows(preview: ReleasePreviewContract): List<ReleaseRowContract> {
        val rows = mutableListOf<ReleaseRowContract>()
        if (preview.error.isNotEmpty()) rows.add(ReleaseRowContract(say(JETBRAINS_RELEASEFLOW_THE_PLAN_STOPPED, preview.error.toString())))
        rows.add(
            ReleaseRowContract(
                if (preview.channel.isEmpty()) {
                    say(
                        JETBRAINS_RELEASEFLOW_CHANNEL_STABLE,
                    )
                } else {
                    say(JETBRAINS_RELEASEFLOW_CHANNEL_ITS_VERSIONS_ARE_PRERELEASES, preview.channel.toString())
                },
            ),
        )
        if (preview.results.isEmpty()) {
            rows.add(
                ReleaseRowContract(
                    say(JETBRAINS_RELEASEFLOW_NOTHING_TO_RELEASE_A_PACKAGE),
                    preview.unchanged.map {
                        say(JETBRAINS_RELEASEFLOW_NO_COMMIT_RELEASES_IT_RELEASE, it.toString())
                    },
                ),
            )
        }
        preview.results.forEach {
            val head = "${it.project} ${it.current.ifEmpty {
                "new"
            }} → ${it.next}${if (it.level.isEmpty()) "" else " · ${it.level}"}${if (it.forced) " · forced" else ""}"
            rows.add(ReleaseRowContract(head, listOfNotNull(it.reason.ifEmpty { null }, it.error.ifEmpty { null }) + it.commits))
        }
        if (preview.results.isNotEmpty() && preview.unchanged.isNotEmpty()) {
            rows.add(ReleaseRowContract(say(JETBRAINS_RELEASEFLOW_NO_COMMIT_RELEASES, preview.unchanged.joinToString(", ").toString())))
        }
        return rows
    }
}
