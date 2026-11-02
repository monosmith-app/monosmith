package dev.monosmith.jetbrains.releaseflow

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.flowports.FlowPortsContract
import dev.monosmith.jetbrains.flowports.IdePortsClient
import dev.monosmith.jetbrains.notice.NoticeButtonContract
import dev.monosmith.jetbrains.smithrun.BackgroundFlowClient
import dev.monosmith.jetbrains.smithrun.FailureTextMapper
import dev.monosmith.jetbrains.smithrun.SmithRunClient
import java.nio.file.Path

/**
 * Releases the workspace with the choices of the person. It plans (`--dry-run`), asks which packages to
 * hold back and which bumps to override, offers a release without a releasing commit as a choice (never a
 * default), plans again when a choice was made, and only after the last question runs `smith release --yes`.
 * A release tags and publishes, and Undo does not revert it.
 */
object ReleaseWorkspaceUseCase {
    private const val TITLE = "MonoSmith: Release"

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
    ) = BackgroundFlowClient.start(ide, "MonoSmith: release preview") {
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
            ports.notice.warn("Nothing was released. ${FailureTextMapper.describe(text)}")
            return false
        }
        val json = ports.smith.run(listOf("release", "--dry-run", "--json"))
        if (!json.ok) {
            ports.notice.warn("Nothing was released. ${FailureTextMapper.describe(json)}")
            return false
        }
        val packages =
            runCatching { ReleaseArgumentsAlgorithm.planned(json.stdout) }
                .getOrElse {
                    ports.notice.warn(
                        "smith printed the release plan in a form this plugin does not read. Update the plugin, or run smith release in a terminal.",
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
                "$TITLE: this is what a release would do now",
                "Nothing is published yet. Next you can hold packages back and change a bump. After that comes the last question.",
                text.stdout.trim(),
                "Continue",
                offerCommit = false,
            ) ?: return false
            hold =
                ports.prompt.pickMany(
                    "$TITLE: hold packages back?",
                    "Tick the packages that must not release this time. Nothing ticked releases them all.",
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
            ports.notice.warn("Nothing was released. ${FailureTextMapper.describe(final)}")
            return false
        }
        val held = if (hold.isEmpty()) "" else "\nHeld back: ${hold.joinToString(", ")}."
        val bumped = if (bump.isEmpty()) "" else "\nChanged bumps: ${bump.joinToString(", ")}."
        ports.prompt.plan(
            "$TITLE: release now?",
            "This cannot be undone. smith creates the tags, pushes them and publishes the packages to the registries. Undo does not revert it.",
            "${final.stdout.trim()}$held$bumped${ReleaseArgumentsAlgorithm.forcedConsequences(forced)}",
            "Release now",
            offerCommit = false,
        ) ?: return false
        val done = ports.smith.run(ReleaseArgumentsAlgorithm.arguments(choices, publish = true), show = true)
        ports.changed()
        if (!done.ok) {
            ports.notice.warn("Nothing was released. ${FailureTextMapper.describe(done)}")
            return false
        }
        ports.notice.info(done.lines.lastOrNull() ?: "The release is done.")
        return true
    }

    private fun askForced(
        ports: FlowPortsContract,
        unchanged: List<String>,
        nothingPlanned: Boolean,
    ): List<String>? {
        val header =
            if (nothingPlanned) {
                "No commit releases anything now. You can release a package anyway: the same code under a new version. Tick none to release nothing."
            } else {
                "No commit releases these packages. Tick one to release it anyway; tick none to release only what the plan lists."
            }
        val ticked =
            ports.prompt.pickMany(
                "$TITLE: release a package without a releasing commit?",
                header,
                ReleaseArgumentsAlgorithm.forcedRows(unchanged),
            )
                ?: return null
        return ticked.map { name ->
            val level =
                ports.prompt.pickOne(
                    "$TITLE: which level for $name?",
                    "The first answer is the default.",
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
                    "$TITLE: change a bump?",
                    "Pick a package to change its level, or keep them all.",
                    ReleaseArgumentsAlgorithm.bumpRows(packages, changed),
                ) ?: return null
            if (picked.name.isEmpty()) return changed.map { "${it.key}=${it.value}" }
            val level =
                ports.prompt.pickOne(
                    "$TITLE: which level for ${picked.name}?",
                    "The first answer is the default.",
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
        val first = lines.firstOrNull { it.startsWith("Nothing to release") } ?: "Nothing to release."
        val why =
            if (unchanged.isEmpty()) {
                "A package releases when a commit since its last tag asks for it."
            } else {
                "No commit releases ${unchanged.joinToString(
                    ", ",
                )}. \"Without a commit…\" releases one anyway: the same code under a new version."
            }
        val buttons =
            listOfNotNull(
                NoticeButtonContract(
                    "Open Preview",
                    "Show the release preview: it says why nothing releases and what to do.",
                ) { openPreview() },
                if (unchanged.isEmpty()) {
                    null
                } else {
                    NoticeButtonContract(
                        "Without a commit…",
                        "Choose a package to release although no commit releases it. It creates a tag at HEAD, publishes, and cannot be undone.",
                    ) {
                        restart()
                    }
                },
            )
        ports.notice.info("$first $why", buttons)
    }
}
