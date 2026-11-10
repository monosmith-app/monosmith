package dev.monosmith.jetbrains.addproject

import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_EXISTS_ALREADY_PICK_ANOTHER_NAME
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_USE_LOWERCASE_LETTERS_DIGITS_AND
import dev.monosmith.jetbrains.words.SmithWordsConfig.say

/** The name of a new project, as far as the IDE can tell: smith has the last word. */
object ProjectNameValidator {
    private val kebab = Regex("^[a-z0-9]+(-[a-z0-9]+)*$")

    /** What is wrong with [name], with the fix; null when the name is fine. [existing] are the names projects have already. */
    fun problem(
        name: String,
        existing: Collection<String>,
    ): String? =
        when {
            !kebab.matches(name) -> say(JETBRAINS_ADDPROJECT_USE_LOWERCASE_LETTERS_DIGITS_AND)
            name in existing -> say(JETBRAINS_ADDPROJECT_EXISTS_ALREADY_PICK_ANOTHER_NAME, name.toString())
            else -> null
        }
}
