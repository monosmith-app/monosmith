package dev.monosmith.jetbrains.newrecipe

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.notice.NoticeClient
import dev.monosmith.jetbrains.prompt.ChoiceRowContract
import dev.monosmith.jetbrains.prompt.PromptClient
import dev.monosmith.jetbrains.smithcommand.TaskCommandAlgorithm
import dev.monosmith.jetbrains.smithrun.BackgroundFlowClient
import dev.monosmith.jetbrains.smithrun.FailureTextMapper
import dev.monosmith.jetbrains.smithrun.SmithRunClient
import dev.monosmith.jetbrains.workspace.WorkspaceChangedStore
import java.nio.file.Path

/**
 * Writes a recipe that fixes plugins. The levels and plugins come from smith (no name is held here). The
 * person gives the basics and answers one question for each level (radio buttons for a level of one
 * plugin, check boxes for a level that takes several). The exact command and the file it writes are shown,
 * and only on a yes smith runs it. A level the users of the recipe choose is left out.
 */
object NewRecipeUseCase {
    private const val TITLE = "MonoSmith: new recipe"
    private const val LEAVE = "\u0000leave"

    /** Starts the flow in a background task. */
    fun start(
        ide: Project,
        root: Path,
    ) = BackgroundFlowClient.start(ide, TITLE) {
        val listed = SmithRunClient.run(ide, root, listOf("recipe", "new", "--list-levels", "--json"))
        if (!listed.ok) return@start NoticeClient.warn(ide, "No recipe was written. ${FailureTextMapper.describe(listed)}")
        val levels =
            runCatching { RecipeLevelsMapper.toLevels(listed.stdout) }
                .getOrElse {
                    return@start NoticeClient.warn(
                        ide,
                        "smith printed its levels in a form this plugin does not read. Update the plugin, or run smith recipe new in a terminal.",
                    )
                }
        if (levels.levels.isEmpty()) {
            return@start NoticeClient.warn(
                ide,
                "smith knows no plugin a recipe can fix. Update smith, or run smith recipe new in a terminal.",
            )
        }

        val name =
            PromptClient.text(ide, "$TITLE: name", "The name of the recipe (for example team):", problem = RecipeLevelsMapper::nameProblem)
                ?: return@start
        val version =
            PromptClient.text(
                ide,
                "$TITLE: version",
                "The first version of the recipe (empty takes 0.1.0):",
                "0.1.0",
                RecipeLevelsMapper::versionProblem,
            )
                ?: return@start
        val parent = PromptClient.text(ide, "$TITLE: parent", "The recipe this one extends (empty for none):") ?: return@start
        val fixed = linkedMapOf<String, List<String>>()
        var language = ""
        for (level in levels.levels) {
            val options = RecipeLevelsMapper.optionsFor(levels, level, language)
            if (options.isEmpty()) continue
            val rows = RecipeLevelsMapper.rows(options)
            val picked =
                if (level.additive) {
                    PromptClient.pickMany(
                        ide,
                        "$TITLE: ${level.level}",
                        "Tick the plugins the recipe fixes at the ${level.level} level. Tick none to leave it to the users of the recipe.",
                        rows,
                    )
                        ?: return@start
                } else {
                    val leave =
                        ChoiceRowContract(
                            LEAVE,
                            "Leave it to the users of the recipe",
                            "(default)",
                            "The recipe fixes no plugin at the ${level.level} level.",
                        )
                    val answer =
                        PromptClient.pickOne(
                            ide,
                            "$TITLE: ${level.level}",
                            "Pick the plugin the recipe fixes at the ${level.level} level.",
                            listOf(leave) + rows,
                        )
                            ?: return@start
                    if (answer.name == LEAVE) emptyList() else listOf(answer.name)
                }
            if (picked.isNotEmpty()) fixed[level.level] = picked
            if (level.level == "language") language = picked.firstOrNull().orEmpty()
        }
        val request = RecipeRequestContract(name, version, parent, fixed)
        val args = RecipeLevelsMapper.arguments(request)
        PromptClient.plan(
            ide,
            "$TITLE: write ${RecipeLevelsMapper.file(request)}?",
            "smith runs this command and writes ${RecipeLevelsMapper.file(request)} in the workspace folder. Nothing is committed.",
            TaskCommandAlgorithm.display(listOf("smith") + args),
            "Write the recipe",
            offerCommit = false,
        ) ?: return@start
        val done = SmithRunClient.run(ide, root, args, show = true)
        WorkspaceChangedStore.changed(ide)
        if (done.ok) {
            NoticeClient.info(
                ide,
                "${RecipeLevelsMapper.file(request)} is written.",
            )
        } else {
            NoticeClient.warn(ide, "No recipe was written. ${FailureTextMapper.describe(done)}")
        }
    }
}
