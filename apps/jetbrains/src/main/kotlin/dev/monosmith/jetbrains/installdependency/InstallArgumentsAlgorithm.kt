package dev.monosmith.jetbrains.installdependency

/** The arguments of `smith install` and `smith uninstall`. */
object InstallArgumentsAlgorithm {
    /**
     * Adds a library of this workspace. `workspace:` names the library, so smith never asks which is meant.
     *
     * @param range the version range written for the library, or null to follow the workspace
     */
    fun library(
        project: String,
        library: String,
        range: String?,
    ): List<String> = listOf("install", "-w", project, "workspace:$library") + (range?.let { listOf("--range", it) } ?: emptyList())

    /** Adds registry packages. `--registry` names the source, so smith never asks which is meant. */
    fun registry(
        project: String,
        packages: List<String>,
        dev: Boolean,
    ): List<String> = listOf("install", "-w", project) + (if (dev) listOf("-D") else emptyList()) + "--registry" + packages

    /** Unlinks a library of this workspace from a project. `workspace:` names the library. */
    fun unlink(
        project: String,
        library: String,
    ): List<String> = listOf("uninstall", "-w", project, "workspace:$library", "--yes")

    /** Splits what was typed into package names; names are separated by spaces or commas. */
    fun parsePackages(text: String): List<String> = text.split(Regex("[\\s,]+")).filter { it.isNotEmpty() }
}
