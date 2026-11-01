package dev.monosmith.jetbrains.smithaction

/** The name of a new project, as far as the IDE can tell: smith has the last word. */
object ProjectNameValidator {
    private val kebab = Regex("^[a-z0-9]+(-[a-z0-9]+)*$")

    /** What is wrong with [name], with the fix; null when the name is fine. [existing] are the names projects have already. */
    fun problem(
        name: String,
        existing: Collection<String>,
    ): String? =
        when {
            !kebab.matches(name) -> "Use lowercase letters, digits and single dashes, for example \"web\" or \"billing-api\"."
            name in existing -> "$name exists already: pick another name."
            else -> null
        }
}
