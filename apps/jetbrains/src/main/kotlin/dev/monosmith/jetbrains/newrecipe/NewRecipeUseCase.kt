package dev.monosmith.jetbrains.newrecipe

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.flowports.FlowPortsContract
import dev.monosmith.jetbrains.flowports.IdePortsClient
import dev.monosmith.jetbrains.prompt.ChoiceRowContract
import dev.monosmith.jetbrains.smithcommand.TaskCommandAlgorithm
import dev.monosmith.jetbrains.smithrun.BackgroundFlowClient
import dev.monosmith.jetbrains.smithrun.FailureTextMapper
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDCHOICES_DEFAULT
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_NEWRECIPE_IS_WRITTEN
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_NEWRECIPE_LEAVE_IT_TO_THE_USERS
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_NEWRECIPE_MONOSMITH_NEW_RECIPE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_NEWRECIPE_NO_RECIPE_WAS_WRITTEN
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_NEWRECIPE_PICK_THE_PLUGIN_THE_RECIPE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_NEWRECIPE_SMITH_KNOWS_NO_PLUGIN_A
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_NEWRECIPE_SMITH_PRINTED_ITS_LEVELS_IN
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_NEWRECIPE_SMITH_RUNS_THIS_COMMAND_AND
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_NEWRECIPE_THE_FIRST_VERSION_OF_THE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_NEWRECIPE_THE_NAME_OF_THE_RECIPE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_NEWRECIPE_THE_RECIPE_FIXES_NO_PLUGIN
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_NEWRECIPE_THE_RECIPE_THIS_ONE_EXTENDS
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_NEWRECIPE_TICK_THE_PLUGINS_THE_RECIPE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_NEWRECIPE_WRITE_THE_RECIPE
import dev.monosmith.jetbrains.words.SmithWordsConfig.say
import java.nio.file.Path

/**
 * Writes a recipe that fixes plugins. The levels and plugins come from smith (no name is held here). The
 * person gives the basics and answers one question for each level (radio buttons for a level of one
 * plugin, check boxes for a level that takes several). The exact command and the file it writes are shown,
 * and only on a yes smith runs it. A level the users of the recipe choose is left out.
 */
object NewRecipeUseCase {
    private val TITLE = say(JETBRAINS_NEWRECIPE_MONOSMITH_NEW_RECIPE)
    private const val LEAVE = "\u0000leave"

    /** Starts the flow in a background task. */
    fun start(
        ide: Project,
        root: Path,
    ) = BackgroundFlowClient.start(ide, TITLE) { run(IdePortsClient.of(ide, root)) }

    /**
     * Runs the flow with the given ports.
     *
     * @return true when smith wrote the recipe
     */
    fun run(ports: FlowPortsContract): Boolean {
        val listed = ports.smith.run(listOf("recipe", "new", "--list-levels", "--json"))
        if (!listed.ok) {
            ports.notice.warn(say(JETBRAINS_NEWRECIPE_NO_RECIPE_WAS_WRITTEN, FailureTextMapper.describe(listed).toString()))
            return false
        }
        val levels =
            runCatching { RecipeLevelsMapper.toLevels(listed.stdout) }
                .getOrElse {
                    ports.notice.warn(
                        say(JETBRAINS_NEWRECIPE_SMITH_PRINTED_ITS_LEVELS_IN),
                    )
                    return false
                }
        if (levels.levels.isEmpty()) {
            ports.notice.warn(say(JETBRAINS_NEWRECIPE_SMITH_KNOWS_NO_PLUGIN_A))
            return false
        }

        val name =
            ports.prompt.text("$TITLE: name", say(JETBRAINS_NEWRECIPE_THE_NAME_OF_THE_RECIPE), problem = RecipeLevelsMapper::nameProblem)
                ?: return false
        val version =
            ports.prompt.text(
                "$TITLE: version",
                say(JETBRAINS_NEWRECIPE_THE_FIRST_VERSION_OF_THE),
                "0.1.0",
                RecipeLevelsMapper::versionProblem,
            )
                ?: return false
        val parent = ports.prompt.text("$TITLE: parent", say(JETBRAINS_NEWRECIPE_THE_RECIPE_THIS_ONE_EXTENDS)) ?: return false
        val fixed = linkedMapOf<String, List<String>>()
        var language = ""
        for (level in levels.levels) {
            val options = RecipeLevelsMapper.optionsFor(levels, level, language)
            if (options.isEmpty()) continue
            val rows = RecipeLevelsMapper.rows(options)
            val picked =
                if (level.additive) {
                    ports.prompt.pickMany(
                        "$TITLE: ${level.level}",
                        say(JETBRAINS_NEWRECIPE_TICK_THE_PLUGINS_THE_RECIPE, level.level.toString()),
                        rows,
                    )
                        ?: return false
                } else {
                    val leave =
                        ChoiceRowContract(
                            LEAVE,
                            say(JETBRAINS_NEWRECIPE_LEAVE_IT_TO_THE_USERS),
                            say(JETBRAINS_ADDCHOICES_DEFAULT),
                            say(JETBRAINS_NEWRECIPE_THE_RECIPE_FIXES_NO_PLUGIN, level.level.toString()),
                        )
                    val answer =
                        ports.prompt.pickOne(
                            "$TITLE: ${level.level}",
                            say(JETBRAINS_NEWRECIPE_PICK_THE_PLUGIN_THE_RECIPE, level.level.toString()),
                            listOf(leave) + rows,
                        )
                            ?: return false
                    if (answer.name == LEAVE) emptyList() else listOf(answer.name)
                }
            if (picked.isNotEmpty()) fixed[level.level] = picked
            if (level.level == "language") language = picked.firstOrNull().orEmpty()
        }
        val request = RecipeRequestContract(name, version, parent, fixed)
        val args = RecipeLevelsMapper.arguments(request)
        ports.prompt.plan(
            "$TITLE: write ${RecipeLevelsMapper.file(request)}?",
            say(JETBRAINS_NEWRECIPE_SMITH_RUNS_THIS_COMMAND_AND, RecipeLevelsMapper.file(request).toString()),
            TaskCommandAlgorithm.display(listOf("smith") + args),
            say(JETBRAINS_NEWRECIPE_WRITE_THE_RECIPE),
            offerCommit = false,
        ) ?: return false
        val done = ports.smith.run(args, show = true)
        ports.changed()
        if (!done.ok) {
            ports.notice.warn(say(JETBRAINS_NEWRECIPE_NO_RECIPE_WAS_WRITTEN, FailureTextMapper.describe(done).toString()))
            return false
        }
        ports.notice.info(say(JETBRAINS_NEWRECIPE_IS_WRITTEN, RecipeLevelsMapper.file(request).toString()))
        return true
    }
}
