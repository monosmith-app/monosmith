package dev.monosmith.jetbrains.syncversions

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.planapply.PlanStepContract
import dev.monosmith.jetbrains.planapply.PlanThenApplyUseCase
import dev.monosmith.jetbrains.prompt.ChoiceRowContract
import dev.monosmith.jetbrains.prompt.PromptClient
import dev.monosmith.jetbrains.smithrun.BackgroundFlowClient
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
                "The installed version (smith's pick)",
                "(default)",
                "The version the package manager has installed.",
            ),
            ChoiceRowContract("highest", "The highest declared", "", "Projects on an older version move up."),
            ChoiceRowContract("lowest", "The lowest declared", "", "Projects on a newer version move down."),
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
    ) = BackgroundFlowClient.start(ide, "MonoSmith: sync dependency versions") {
        val rule =
            PromptClient
                .pickOne(
                    ide,
                    "MonoSmith: which version wins?",
                    "When projects declare a dependency at different versions, which one should every project use? The first answer is the default.",
                    ruleRows,
                )?.name ?: return@start
        PlanThenApplyUseCase.run(
            ide,
            root,
            PlanStepContract(
                title = "MonoSmith: sync dependency versions ($rule)?",
                planArgs = arguments(rule, apply = false),
                header = "Nothing was changed yet. This is the plan: every project that disagrees moves to one version.",
                okText = "Write it",
                offerCommit = true,
                apply = { arguments(rule, apply = true) },
                done = "The versions are in sync",
                operation = "sync",
            ),
        )
    }
}
