package dev.monosmith.jetbrains.runacross

import dev.monosmith.jetbrains.prompt.ChoiceRowContract
import dev.monosmith.jetbrains.workspace.SmithProjectContract

/** The command lines and the lists of running one task in many projects. */
object RunAcrossAlgorithm {
    /**
     * `smith run <task>` in every project, or `--affected` in the projects the changes affect.
     * A task is one argument and never an option, so a name that looks like a flag is refused.
     *
     * @return the arguments after `smith`, or null when the task is not a word
     */
    fun arguments(
        task: String,
        affected: Boolean,
    ): List<String>? =
        if (task.isEmpty() || task.startsWith("-") ||
            task.any(Char::isWhitespace)
        ) {
            null
        } else {
            listOf("run", task) + if (affected) listOf("--affected") else emptyList()
        }

    /**
     * The tasks to choose from: those of the projects the changes affect when [affectedOnly], else of every
     * project; each with how many projects run it.
     *
     * @param affected the names of the affected projects
     */
    fun taskRows(
        projects: List<SmithProjectContract>,
        affected: Set<String>,
        affectedOnly: Boolean,
    ): List<ChoiceRowContract> {
        val scope = if (affectedOnly) projects.filter { it.name in affected } else projects
        val byTask = linkedMapOf<String, MutableList<String>>()
        scope.forEach { project -> project.tasks.forEach { byTask.getOrPut(it) { mutableListOf() }.add(project.name) } }
        return byTask.map { (task, names) ->
            ChoiceRowContract(task, task, "${names.size} ${if (names.size == 1) "project" else "projects"}", names.joinToString(", "))
        }
    }
}
