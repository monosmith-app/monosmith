package dev.monosmith.jetbrains.tauriapp

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.commitchanges.CommitChangesUseCase
import dev.monosmith.jetbrains.flowports.FlowPortsContract
import dev.monosmith.jetbrains.flowports.IdePortsClient
import dev.monosmith.jetbrains.planapply.PlanStepContract
import dev.monosmith.jetbrains.planapply.PlanThenApplyUseCase
import dev.monosmith.jetbrains.smithcommand.SmithFlagsConfig
import dev.monosmith.jetbrains.smithrun.BackgroundFlowClient
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_PIPELINETEMPLATE_NAME_THE_APP
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_TAURIAPP_ADD_TAURI
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_TAURIAPP_ADD_TAURI_2
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_TAURIAPP_ADD_TAURI_TO
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_TAURIAPP_IS_A_TAURI_APP
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_TAURIAPP_MAKE_A_TAURI_APP
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_TAURIAPP_MAKE_THE_ICONS
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_TAURIAPP_MAKE_THE_ICONS_2
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_TAURIAPP_MONOSMITH_TAURI
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_TAURIAPP_NAME_A_PNG_FILE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_TAURIAPP_SMITH_OVERWRITES_THE_ICONS_IN
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_TAURIAPP_SMITH_WRITES_SRC_TAURI_AND
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_TAURIAPP_THE_APP_ONLY_WHEN_MORE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_TAURIAPP_THE_BUNDLE_IDENTIFIER_LIKE_COM
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_TAURIAPP_THE_ICONS
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_TAURIAPP_THE_ICONS_ARE_MADE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_TAURIAPP_THE_NODE_APP_TO_MAKE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_TAURIAPP_THE_SQUARE_PNG_THE_ICONS
import dev.monosmith.jetbrains.words.SmithWordsConfig.say
import java.nio.file.Path

/**
 * Two steps for a Tauri desktop app: make the icons from one square PNG (`smith tauri icon`), and make a
 * Node app a Tauri app (`smith tauri add`, which asks for the bundle identifier). Each shows its `--dry-run`
 * plan first. smith commits the second only when the unticked box is ticked.
 */
object TauriAppUseCase {
    private val TITLE = say(JETBRAINS_TAURIAPP_MONOSMITH_TAURI)

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
            ports.prompt.text(say(JETBRAINS_TAURIAPP_THE_ICONS, TITLE.toString()), say(JETBRAINS_TAURIAPP_THE_SQUARE_PNG_THE_ICONS)) {
                if (it.endsWith(".png", ignoreCase = true)) null else say(JETBRAINS_TAURIAPP_NAME_A_PNG_FILE)
            } ?: return false
        val app =
            ports.prompt.text(say(JETBRAINS_TAURIAPP_THE_ICONS, TITLE.toString()), say(JETBRAINS_TAURIAPP_THE_APP_ONLY_WHEN_MORE))
                ?: return false
        val args = listOf("tauri", "icon", logo) + if (app.isEmpty()) emptyList() else listOf(SmithFlagsConfig.PROJECT, app)
        return PlanThenApplyUseCase.run(
            ports,
            PlanStepContract(
                title = say(JETBRAINS_TAURIAPP_MAKE_THE_ICONS, TITLE.toString()),
                planArgs = args + "--dry-run",
                header = say(JETBRAINS_TAURIAPP_SMITH_OVERWRITES_THE_ICONS_IN),
                okText = say(JETBRAINS_TAURIAPP_MAKE_THE_ICONS_2),
                offerCommit = false,
                apply = { args + "--yes" },
                done = say(JETBRAINS_TAURIAPP_THE_ICONS_ARE_MADE),
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
            ports.prompt.text(say(JETBRAINS_TAURIAPP_ADD_TAURI, TITLE.toString()), say(JETBRAINS_TAURIAPP_THE_NODE_APP_TO_MAKE)) {
                if (it.isEmpty()) say(JETBRAINS_PIPELINETEMPLATE_NAME_THE_APP) else null
            } ?: return false
        val identifier =
            ports.prompt.text(
                say(JETBRAINS_TAURIAPP_ADD_TAURI_TO, TITLE.toString(), app.toString()),
                say(JETBRAINS_TAURIAPP_THE_BUNDLE_IDENTIFIER_LIKE_COM),
            ) ?: return false
        val args = listOf("tauri", "add", app) + if (identifier.isEmpty()) emptyList() else listOf("--identifier", identifier)
        return PlanThenApplyUseCase.run(
            ports,
            PlanStepContract(
                title = say(JETBRAINS_TAURIAPP_MAKE_A_TAURI_APP, TITLE.toString(), app.toString()),
                planArgs = args + "--dry-run",
                header =
                    say(JETBRAINS_TAURIAPP_SMITH_WRITES_SRC_TAURI_AND),
                okText = say(JETBRAINS_TAURIAPP_ADD_TAURI_2),
                offerCommit = true,
                apply = { args + "--yes" },
                done = say(JETBRAINS_TAURIAPP_IS_A_TAURI_APP, app.toString()),
                operation = "tauri add $app",
            ),
            commitLater,
        )
    }
}
