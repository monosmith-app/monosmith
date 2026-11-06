package dev.monosmith.jetbrains.syncversions

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.planapply.PlanStepContract
import dev.monosmith.jetbrains.planapply.PlanThenApplyUseCase
import dev.monosmith.jetbrains.prompt.ChoiceRowContract
import dev.monosmith.jetbrains.prompt.PromptClient
import dev.monosmith.jetbrains.smithrun.BackgroundFlowClient
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDCHOICES_DEFAULT
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_DEPENDENCIES_PROJECTS_ON_AN_OLDER_VERSION
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_DEPENDENCIES_PROJECTS_ON_A_NEWER_VERSION
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_DEPENDENCIES_THE_VERSION_THE_PACKAGE_MANAGER
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_SYNCVERSIONS_MONOSMITH_SYNC_DEPENDENCY_VERSIONS
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_SYNCVERSIONS_MONOSMITH_SYNC_DEPENDENCY_VERSIONS_2
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_SYNCVERSIONS_MONOSMITH_WHICH_VERSION_WINS
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_SYNCVERSIONS_NOTHING_WAS_CHANGED_YET_THIS
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_SYNCVERSIONS_THE_VERSIONS_ARE_IN_SYNC
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_SYNCVERSIONS_WHEN_PROJECTS_DECLARE_A_DEPENDENCY
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_SYNCVERSIONS_WRITE_IT
import dev.monosmith.jetbrains.words.SmithWordsConfig.RESULTDIALOG_VERSION_HIGHEST_LABEL
import dev.monosmith.jetbrains.words.SmithWordsConfig.RESULTDIALOG_VERSION_INSTALLED_LABEL
import dev.monosmith.jetbrains.words.SmithWordsConfig.RESULTDIALOG_VERSION_LOWEST_LABEL
import dev.monosmith.jetbrains.words.SmithWordsConfig.say
import java.nio.file.Path

/**
 * Makes every project declare a shared dependency at one version (`smith sync`). The person chooses which
 * version wins, smith prints the plan with `--dry-run`, and only on a yes is it written, with `--yes`.
 */
object SyncVersionsUseCase {
    /** The rules `smith sync --prefer` takes, each with what it does; smith's own pick is first. */
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

    /** The arguments of `smith sync`; [apply] false makes the plan. */
    fun arguments(
        rule: String,
        apply: Boolean,
    ): List<String> = listOf("sync", "--prefer", rule, if (apply) "--yes" else "--dry-run")

    /** Starts the flow in a background task. */
    fun start(
        ide: Project,
        root: Path,
    ) = BackgroundFlowClient.start(ide, say(JETBRAINS_SYNCVERSIONS_MONOSMITH_SYNC_DEPENDENCY_VERSIONS)) {
        val rule =
            PromptClient
                .pickOne(
                    ide,
                    say(JETBRAINS_SYNCVERSIONS_MONOSMITH_WHICH_VERSION_WINS),
                    say(JETBRAINS_SYNCVERSIONS_WHEN_PROJECTS_DECLARE_A_DEPENDENCY),
                    ruleRows,
                )?.name ?: return@start
        PlanThenApplyUseCase.run(
            ide,
            root,
            PlanStepContract(
                title = say(JETBRAINS_SYNCVERSIONS_MONOSMITH_SYNC_DEPENDENCY_VERSIONS_2, rule.toString()),
                planArgs = arguments(rule, apply = false),
                header = say(JETBRAINS_SYNCVERSIONS_NOTHING_WAS_CHANGED_YET_THIS),
                okText = say(JETBRAINS_SYNCVERSIONS_WRITE_IT),
                offerCommit = true,
                apply = { arguments(rule, apply = true) },
                done = say(JETBRAINS_SYNCVERSIONS_THE_VERSIONS_ARE_IN_SYNC),
                operation = "sync",
            ),
        )
    }
}
