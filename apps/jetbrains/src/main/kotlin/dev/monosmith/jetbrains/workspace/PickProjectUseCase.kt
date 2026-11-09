package dev.monosmith.jetbrains.workspace

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.prompt.ChoiceRowContract
import dev.monosmith.jetbrains.prompt.PromptClient
import dev.monosmith.jetbrains.words.SmithWordsConfig.GUIDE_ACTIONS_WHICH_PROJECT
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_WORKSPACE_MONOSMITH
import dev.monosmith.jetbrains.words.SmithWordsConfig.say

/** Asks which project an action works on, when the action was not started on one. Blocks: call it from a background flow. */
object PickProjectUseCase {
    /**
     * @param projects the projects to choose from
     * @param title what the action does, for the title of the dialog
     * @return the project, or null when the person cancels
     */
    fun pick(
        ide: Project,
        projects: List<SmithProjectContract>,
        title: String,
    ): SmithProjectContract? {
        val rows = projects.map { ChoiceRowContract(it.name, it.name, "${it.kind} · ${it.language}") }
        val answer =
            PromptClient.pickOne(
                ide,
                say(JETBRAINS_WORKSPACE_MONOSMITH, title.toString()),
                say(GUIDE_ACTIONS_WHICH_PROJECT),
                rows,
            ) ?: return null
        return projects.first { it.name == answer.name }
    }
}
