package dev.monosmith.jetbrains.tauriapp

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.commitchanges.CommitChangesUseCase
import dev.monosmith.jetbrains.flowports.FlowPortsContract
import dev.monosmith.jetbrains.flowports.IdePortsClient
import dev.monosmith.jetbrains.planapply.PlanStepContract
import dev.monosmith.jetbrains.planapply.PlanThenApplyUseCase
import dev.monosmith.jetbrains.smithrun.BackgroundFlowClient
import java.nio.file.Path
import dev.monosmith.jetbrains.smithcommand.SmithFlags

/**
 * Two steps for a Tauri desktop app: make the icons from one square PNG (`smith tauri icon`), and make a
 * Node app a Tauri app (`smith tauri add`, which asks for the bundle identifier). Each shows its `--dry-run`
 * plan first. smith commits the second only when the unticked box is ticked.
 */
object TauriAppUseCase {
    private const val TITLE = "MonoSmith: Tauri"

    /** Starts the icons flow in a background task. */
    fun startIcons(
        ide: Project,
        root: Path,
    ) = BackgroundFlowClient.start(ide, TITLE) { icons(IdePortsClient.of(ide, root)) }

    /** Starts the flow that makes an app a Tauri app, in a background task. */
    fun startAdd(
        ide: Project,
        root: Path,
    ) = BackgroundFlowClient.start(ide, TITLE) { add(IdePortsClient.of(ide, root)) { CommitChangesUseCase.start(ide, root, it) } }

    /**
     * Asks for the PNG and, when more than one project has a src-tauri, for the app; then shows what the
     * icons would overwrite.
     *
     * @return true when smith made the icons
     */
    fun icons(ports: FlowPortsContract): Boolean {
        val logo =
            ports.prompt.text("$TITLE: the icons", "The square PNG the icons are made from (1024x1024 is best):") {
                if (it.endsWith(".png", ignoreCase = true)) null else "Name a PNG file."
            } ?: return false
        val app =
            ports.prompt.text("$TITLE: the icons", "The app, only when more than one project has a src-tauri (empty leaves it out):")
                ?: return false
        val args = listOf("tauri", "icon", logo) + if (app.isEmpty()) emptyList() else listOf(SmithFlags.PROJECT, app)
        return PlanThenApplyUseCase.run(
            ports,
            PlanStepContract(
                title = "$TITLE: make the icons?",
                planArgs = args + "--dry-run",
                header = "smith overwrites the icons in src-tauri/icons. Nothing was changed yet.",
                okText = "Make the icons",
                offerCommit = false,
                apply = { args + "--yes" },
                done = "The icons are made",
            ),
        ) {}
    }

    /**
     * Asks for the app and the bundle identifier, then shows what `smith tauri add` would write.
     *
     * @param commitLater starts the Commit action for an operation
     * @return true when smith made the app a Tauri app
     */
    fun add(
        ports: FlowPortsContract,
        commitLater: (String) -> Unit,
    ): Boolean {
        val app =
            ports.prompt.text("$TITLE: add Tauri", "The Node app to make a Tauri app (the name of a project):") {
                if (it.isEmpty()) "Name the app." else null
            } ?: return false
        val identifier =
            ports.prompt.text(
                "$TITLE: add Tauri to $app",
                "The bundle identifier, like com.yourcompany.yourapp. Empty takes com.<workspace>.<project>:",
            ) ?: return false
        val args = listOf("tauri", "add", app) + if (identifier.isEmpty()) emptyList() else listOf("--identifier", identifier)
        return PlanThenApplyUseCase.run(
            ports,
            PlanStepContract(
                title = "$TITLE: make $app a Tauri app?",
                planArgs = args + "--dry-run",
                header =
                    "smith writes src-tauri and the tauri tasks in the app, and never overwrites a file that exists. " +
                        "Nothing was changed yet.",
                okText = "Add Tauri",
                offerCommit = true,
                apply = { args + "--yes" },
                done = "$app is a Tauri app",
                operation = "tauri add $app",
            ),
            commitLater,
        )
    }
}
