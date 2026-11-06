package dev.monosmith.jetbrains.addproject

import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_AN_ENVIRONMENT_IS_LISTED_TWICE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_CANNOT_BE_AN_ENVIRONMENT_NAME
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_IS_NOT_AN_ENVIRONMENT_NAME
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_NAME_AT_LEAST_ONE_ENVIRONMENT
import dev.monosmith.jetbrains.words.SmithWordsConfig.say

/** The environments of a web app, built once for each: the names, the default one and the flags of `smith add`. */
object EnvironmentsAlgorithm {
    /** The environments a web app starts from: lcl is the developer's machine. */
    val defaults = listOf("lcl", "dev", "uat", "prd")

    private val reserved = setOf("development", "production", "test", "local", "example")

    /**
     * Whether the request makes a web app built with Vite (react, vue, svelte or a plain page), which is
     * built once per environment.
     */
    fun isWebApp(request: AddRequestContract): Boolean {
        if (request.role != "app" || request.language != "node") return false
        return (request.with + listOfNotNull(request.stack)).any { Regex("^(?:react|vue|svelte|html)(?:-|$)").containsMatchIn(it) }
    }

    /** The names typed, as a list: separated by commas, spaces or semicolons. */
    fun parse(text: String): List<String> = text.split(Regex("[\\s,;]+")).filter { it.isNotEmpty() }

    /** What is wrong with the names typed, in words that name the fix; null when they are fine. smith has the last word. */
    fun problem(text: String): String? {
        val names = parse(text)
        if (names.isEmpty()) return say(JETBRAINS_ADDPROJECT_NAME_AT_LEAST_ONE_ENVIRONMENT)
        for (name in names) {
            if (!Regex(
                    "^[a-z][a-z0-9-]*$",
                ).matches(name)
            ) {
                return say(JETBRAINS_ADDPROJECT_IS_NOT_AN_ENVIRONMENT_NAME, name.toString())
            }
            if (name in reserved) {
                return say(JETBRAINS_ADDPROJECT_CANNOT_BE_AN_ENVIRONMENT_NAME, name.toString())
            }
        }
        return if (names.toSet().size == names.size) null else say(JETBRAINS_ADDPROJECT_AN_ENVIRONMENT_IS_LISTED_TWICE)
    }

    /** The environment that `smith run build` builds when the person does not pick one: prd, else the last. */
    fun defaultOf(names: List<String>): String = if ("prd" in names) "prd" else names.lastOrNull() ?: "prd"

    /** The flags of `smith add` for the environments of a web app. */
    fun arguments(environments: EnvironmentsContract): List<String> =
        listOf(
            "--environments",
            environments.names.joinToString(","),
            "--default-environment",
            environments.default,
            "--environment-output",
            environments.output,
            "--env-files",
            environments.envFiles,
        )
}
