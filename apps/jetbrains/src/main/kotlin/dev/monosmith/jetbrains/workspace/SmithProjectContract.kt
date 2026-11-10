package dev.monosmith.jetbrains.workspace

/**
 * One project of the workspace, as `smith projects --json` reports it.
 *
 * @property deps the names of the projects of this workspace it depends on
 */
data class SmithProjectContract(
    val name: String,
    val dir: String,
    val kind: String,
    val language: String,
    val tasks: List<String>,
    val deps: List<String> = emptyList(),
)
