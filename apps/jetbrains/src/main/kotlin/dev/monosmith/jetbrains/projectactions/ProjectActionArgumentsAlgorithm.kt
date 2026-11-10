package dev.monosmith.jetbrains.projectactions

import dev.monosmith.jetbrains.prompt.ChoiceRowContract
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_CODE_PUBLISHED_TO_A_REGISTRY
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_CODE_THE_OTHER_PROJECTS_OF
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_SOMETHING_YOU_RUN_OR_DEPLOY
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_PROJECTACTIONS_AN_APP
import dev.monosmith.jetbrains.words.SmithWordsConfig.say

/** The command lines and the questions of the actions on one project: remove, change the kind, unlink. */
object ProjectActionArgumentsAlgorithm {
    /** The kinds of a project in the order app, lib, package, each with what it means. */
    val kindWords =
        mapOf(
            "app" to say(JETBRAINS_ADDPROJECT_SOMETHING_YOU_RUN_OR_DEPLOY),
            "lib" to say(JETBRAINS_ADDPROJECT_CODE_THE_OTHER_PROJECTS_OF),
            "package" to say(JETBRAINS_ADDPROJECT_CODE_PUBLISHED_TO_A_REGISTRY),
        )

    /** The plan of removing a project. */
    fun removePlan(project: String): List<String> = listOf("remove", project, "--dry-run")

    /**
     * Removing a project for the answer of the person. smith never deletes files git cannot bring back by
     * surprise, so the answer is always named.
     *
     * @param delete true deletes the folder for good; false moves it to .smith/removed
     */
    fun remove(
        project: String,
        delete: Boolean,
    ): List<String> = listOf("remove", project, if (delete) "--delete" else "--keep-copy", "--yes")

    /** The plan of changing the kind of a project. */
    fun kindPlan(
        project: String,
        kind: String,
    ): List<String> = listOf("kind", project, kind, "--dry-run")

    /** Changing the kind of a project after the plan was shown. */
    fun kind(
        project: String,
        kind: String,
    ): List<String> = listOf("kind", project, kind, "--yes")

    /** The kinds a project can become: all but its own, in the order app, lib, package. */
    fun otherKinds(current: String): List<String> = kindWords.keys.filter { it != current }

    /** The lines of "what should it become?". */
    fun kindRows(current: String): List<ChoiceRowContract> =
        otherKinds(current).map { ChoiceRowContract(it, it, "", kindWords.getValue(it)) }

    /** The word for a kind with its article: "an app", "a lib". */
    fun withArticle(kind: String): String = if (kind == "app") say(JETBRAINS_PROJECTACTIONS_AN_APP) else "a $kind"
}
