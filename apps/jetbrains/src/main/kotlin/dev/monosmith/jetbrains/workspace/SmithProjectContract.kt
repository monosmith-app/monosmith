package dev.monosmith.jetbrains.workspace

/** One project of the workspace, as `smith projects --json` reports it. */
data class SmithProjectContract(
    val name: String,
    val dir: String,
    val kind: String,
    val language: String,
    val tasks: List<String>,
)
