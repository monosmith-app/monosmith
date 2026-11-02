package dev.monosmith.jetbrains.addproject

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
        if (names.isEmpty()) return "Name at least one environment, for example lcl, dev, uat, prd."
        for (name in names) {
            if (!Regex(
                    "^[a-z][a-z0-9-]*$",
                ).matches(name)
            ) {
                return "\"$name\" is not an environment name: use a lowercase word, for example uat."
            }
            if (name in reserved) {
                return "\"$name\" cannot be an environment name: Vite or the env files already give it a meaning. " +
                    "Use another word, for example lcl for local."
            }
        }
        return if (names.toSet().size == names.size) null else "An environment is listed twice."
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
