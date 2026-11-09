package dev.monosmith.jetbrains.releaseflow

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.flowports.FlowPortsContract
import dev.monosmith.jetbrains.flowports.IdePortsClient
import dev.monosmith.jetbrains.notice.NoticeButtonContract
import dev.monosmith.jetbrains.smithrun.BackgroundFlowClient
import dev.monosmith.jetbrains.smithrun.FailureTextMapper
import dev.monosmith.jetbrains.smithrun.SmithRunClient
import dev.monosmith.jetbrains.words.SmithWordsConfig.ADOPT_FOUND_CONTINUE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_THE_FIRST_ANSWER_IS_THE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_RELEASEFLOW_A_PACKAGE_RELEASES_WHEN_A
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_RELEASEFLOW_CHANGED_BUMPS
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_RELEASEFLOW_CHANGE_A_BUMP
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_RELEASEFLOW_CHOOSE_A_PACKAGE_TO_RELEASE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_RELEASEFLOW_HELD_BACK
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_RELEASEFLOW_HOLD_PACKAGES_BACK
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_RELEASEFLOW_MONOSMITH_RELEASE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_RELEASEFLOW_MONOSMITH_RELEASE_PREVIEW
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_RELEASEFLOW_NOTHING_IS_PUBLISHED_YET_NEXT
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_RELEASEFLOW_NOTHING_TO_RELEASE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_RELEASEFLOW_NOTHING_WAS_RELEASED
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_RELEASEFLOW_NO_COMMIT_RELEASES_ANYTHING_NOW
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_RELEASEFLOW_NO_COMMIT_RELEASES_THESE_PACKAGES
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_RELEASEFLOW_OPEN_PREVIEW
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_RELEASEFLOW_PICK_A_PACKAGE_TO_CHANGE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_RELEASEFLOW_RELEASE_A_PACKAGE_WITHOUT_A
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_RELEASEFLOW_RELEASE_NOW
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_RELEASEFLOW_RELEASE_NOW_2
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_RELEASEFLOW_SHOW_THE_RELEASE_PREVIEW_IT
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_RELEASEFLOW_SMITH_PRINTED_THE_RELEASE_PLAN
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_RELEASEFLOW_THE_RELEASE_IS_DONE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_RELEASEFLOW_THIS_CANNOT_BE_UNDONE_SMITH
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_RELEASEFLOW_THIS_IS_WHAT_A_RELEASE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_RELEASEFLOW_TICK_THE_PACKAGES_THAT_MUST
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_RELEASEFLOW_WHICH_LEVEL_FOR
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_RELEASEFLOW_WITHOUT_A_COMMIT
import dev.monosmith.jetbrains.words.SmithWordsConfig.say
import java.nio.file.Path

/**
 * Releases the workspace with the choices of the person. It plans (`--dry-run`), asks which packages to
 * hold back and which bumps to override, offers a release without a releasing commit as a choice (never a
 * default), plans again when a choice was made, and only after the last question runs `smith release --yes`.
 * A release tags and publishes, and Undo does not revert it.
 */
object ReleaseWorkspaceUseCase {
    private val TITLE = say(JETBRAINS_RELEASEFLOW_MONOSMITH_RELEASE)

    /** Starts the flow in a background task. */
    fun start(
        ide: Project,
        root: Path,
    ) {
        BackgroundFlowClient.start(ide, TITLE) {
            run(IdePortsClient.of(ide, root), { preview(ide, root) }, { start(ide, root) })
        }
    }

    /** Shows what a release would publish, changing nothing (`smith release --dry-run`), in the Output tab. */
    fun preview(
        ide: Project,
        root: Path,
    ) = BackgroundFlowClient.start(ide, say(JETBRAINS_RELEASEFLOW_MONOSMITH_RELEASE_PREVIEW)) {
        SmithRunClient.run(ide, root, listOf("release", "--dry-run"), show = true)
    }

    /**
     * Runs the flow with the given ports.
     *
     * @param openPreview starts the release preview; the "Open Preview" button of the notice calls it
     * @param restart starts the flow again; the "Without a commit…" button of the notice calls it
     * @return true when smith published the release
     */
    fun run(
        ports: FlowPortsContract,
        openPreview: () -> Unit,
        restart: () -> Unit,
    ): Boolean {
        val text = ports.smith.run(listOf("release", "--dry-run"))
        if (!text.ok) {
            ports.notice.warn(say(JETBRAINS_RELEASEFLOW_NOTHING_WAS_RELEASED, FailureTextMapper.describe(text).toString()))
            return false
        }
        val json = ports.smith.run(listOf("release", "--dry-run", "--json"))
        if (!json.ok) {
            ports.notice.warn(say(JETBRAINS_RELEASEFLOW_NOTHING_WAS_RELEASED, FailureTextMapper.describe(json).toString()))
            return false
        }
        val packages =
            runCatching { ReleaseArgumentsAlgorithm.planned(json.stdout) }
                .getOrElse {
                    ports.notice.warn(
                        say(JETBRAINS_RELEASEFLOW_SMITH_PRINTED_THE_RELEASE_PLAN),
                    )
                    return false
                }
        val unchanged = runCatching { ReleaseArgumentsAlgorithm.unchanged(json.stdout) }.getOrDefault(emptyList())
        val forced = if (unchanged.isEmpty()) emptyList() else askForced(ports, unchanged, packages.isEmpty()) ?: return false
        if (packages.isEmpty() && forced.isEmpty()) {
            nothing(ports, text.lines, unchanged, openPreview, restart)
            return false
        }

        var hold = emptyList<String>()
        var bump = emptyList<String>()
        if (packages.isNotEmpty()) {
            ports.prompt.plan(
                say(JETBRAINS_RELEASEFLOW_THIS_IS_WHAT_A_RELEASE, TITLE.toString()),
                say(JETBRAINS_RELEASEFLOW_NOTHING_IS_PUBLISHED_YET_NEXT),
                text.stdout.trim(),
                say(ADOPT_FOUND_CONTINUE),
                offerCommit = false,
            ) ?: return false
            hold =
                ports.prompt.pickMany(
                    say(JETBRAINS_RELEASEFLOW_HOLD_PACKAGES_BACK, TITLE.toString()),
                    say(JETBRAINS_RELEASEFLOW_TICK_THE_PACKAGES_THAT_MUST),
                    ReleaseArgumentsAlgorithm.holdRows(packages),
                ) ?: return false
            bump = askBumps(ports, packages.filter { it.name !in hold }) ?: return false
        }
        val choices = ReleaseChoicesContract(hold, bump, forced)
        val final =
            if (hold.isEmpty() && bump.isEmpty() && forced.isEmpty()) {
                text
            } else {
                ports.smith.run(ReleaseArgumentsAlgorithm.arguments(choices, publish = false))
            }
        if (!final.ok) {
            ports.notice.warn(say(JETBRAINS_RELEASEFLOW_NOTHING_WAS_RELEASED, FailureTextMapper.describe(final).toString()))
            return false
        }
        val held = if (hold.isEmpty()) "" else say(JETBRAINS_RELEASEFLOW_HELD_BACK, hold.joinToString(", ").toString())
        val bumped = if (bump.isEmpty()) "" else say(JETBRAINS_RELEASEFLOW_CHANGED_BUMPS, bump.joinToString(", ").toString())
        ports.prompt.plan(
            say(JETBRAINS_RELEASEFLOW_RELEASE_NOW, TITLE.toString()),
            say(JETBRAINS_RELEASEFLOW_THIS_CANNOT_BE_UNDONE_SMITH),
            "${final.stdout.trim()}$held$bumped${ReleaseArgumentsAlgorithm.forcedConsequences(forced)}",
            say(JETBRAINS_RELEASEFLOW_RELEASE_NOW_2),
            offerCommit = false,
        ) ?: return false
        val done = ports.smith.run(ReleaseArgumentsAlgorithm.arguments(choices, publish = true), show = true)
        ports.changed()
        if (!done.ok) {
            ports.notice.warn(say(JETBRAINS_RELEASEFLOW_NOTHING_WAS_RELEASED, FailureTextMapper.describe(done).toString()))
            return false
        }
        ports.notice.info(done.lines.lastOrNull() ?: say(JETBRAINS_RELEASEFLOW_THE_RELEASE_IS_DONE))
        return true
    }

    private fun askForced(
        ports: FlowPortsContract,
        unchanged: List<String>,
        nothingPlanned: Boolean,
    ): List<String>? {
        val header =
            if (nothingPlanned) {
                say(JETBRAINS_RELEASEFLOW_NO_COMMIT_RELEASES_ANYTHING_NOW)
            } else {
                say(JETBRAINS_RELEASEFLOW_NO_COMMIT_RELEASES_THESE_PACKAGES)
            }
        val ticked =
            ports.prompt.pickMany(
                say(JETBRAINS_RELEASEFLOW_RELEASE_A_PACKAGE_WITHOUT_A, TITLE.toString()),
                header,
                ReleaseArgumentsAlgorithm.forcedRows(unchanged),
            )
                ?: return null
        return ticked.map { name ->
            val level =
                ports.prompt.pickOne(
                    say(JETBRAINS_RELEASEFLOW_WHICH_LEVEL_FOR, TITLE.toString(), name.toString()),
                    say(JETBRAINS_ADDPROJECT_THE_FIRST_ANSWER_IS_THE),
                    ReleaseArgumentsAlgorithm.levelRows,
                )
                    ?: return null
            "$name=${level.name}"
        }
    }

    private fun askBumps(
        ports: FlowPortsContract,
        packages: List<PlannedPackageContract>,
    ): List<String>? {
        val changed = linkedMapOf<String, String>()
        if (packages.none { it.current.isNotEmpty() }) return emptyList()
        while (true) {
            val picked =
                ports.prompt.pickOne(
                    say(JETBRAINS_RELEASEFLOW_CHANGE_A_BUMP, TITLE.toString()),
                    say(JETBRAINS_RELEASEFLOW_PICK_A_PACKAGE_TO_CHANGE),
                    ReleaseArgumentsAlgorithm.bumpRows(packages, changed),
                ) ?: return null
            if (picked.name.isEmpty()) return changed.map { "${it.key}=${it.value}" }
            val level =
                ports.prompt.pickOne(
                    say(JETBRAINS_RELEASEFLOW_WHICH_LEVEL_FOR, TITLE.toString(), picked.name.toString()),
                    say(JETBRAINS_ADDPROJECT_THE_FIRST_ANSWER_IS_THE),
                    ReleaseArgumentsAlgorithm.levelRows,
                )
                    ?: return null
            changed[picked.name] = level.name
        }
    }

    private fun nothing(
        ports: FlowPortsContract,
        lines: List<String>,
        unchanged: List<String>,
        openPreview: () -> Unit,
        restart: () -> Unit,
    ) {
        val first = lines.firstOrNull { it.startsWith("Nothing to release") } ?: say(JETBRAINS_RELEASEFLOW_NOTHING_TO_RELEASE)
        val why =
            if (unchanged.isEmpty()) {
                say(JETBRAINS_RELEASEFLOW_A_PACKAGE_RELEASES_WHEN_A)
            } else {
                "No commit releases ${unchanged.joinToString(
                    ", ",
                )}. \"Without a commit…\" releases one anyway: the same code under a new version."
            }
        val buttons =
            listOfNotNull(
                NoticeButtonContract(
                    say(JETBRAINS_RELEASEFLOW_OPEN_PREVIEW),
                    say(JETBRAINS_RELEASEFLOW_SHOW_THE_RELEASE_PREVIEW_IT),
                ) { openPreview() },
                if (unchanged.isEmpty()) {
                    null
                } else {
                    NoticeButtonContract(
                        say(JETBRAINS_RELEASEFLOW_WITHOUT_A_COMMIT),
                        say(JETBRAINS_RELEASEFLOW_CHOOSE_A_PACKAGE_TO_RELEASE),
                    ) {
                        restart()
                    }
                },
            )
        ports.notice.info("$first $why", buttons)
    }
}
