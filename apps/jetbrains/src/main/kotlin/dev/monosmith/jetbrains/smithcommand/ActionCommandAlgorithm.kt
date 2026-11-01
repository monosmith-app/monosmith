package dev.monosmith.jetbrains.smithcommand

/** The argument vectors of the smith calls the menu actions make, one function per action. */
object ActionCommandAlgorithm {
    /** `smith doctor`; with [fix] also `--fix`, which applies the safe fixes and shows the others. */
    fun doctor(
        smith: String,
        fix: Boolean,
    ): List<String> = listOf(smith, "doctor") + if (fix) listOf("--fix") else emptyList()

    /** `smith kinds --json`: what `smith add` accepts, as JSON. */
    fun listKinds(smith: String): List<String> = listOf(smith, "kinds", "--json")

    /**
     * `smith add <role> <name> --lang <language> --yes`. The person agreed in the dialog, so `--yes`
     * skips the question. [commit] adds `--commit`, which makes the change one commit that
     * `smith undo` reverts; without it smith commits nothing.
     */
    fun add(
        smith: String,
        role: String,
        name: String,
        language: String,
        commit: Boolean,
    ): List<String> = listOf(smith, "add", role, name, "--lang", language, "--yes") + if (commit) listOf("--commit") else emptyList()

    /** `smith undo --yes`: reverts the last operation smith committed. */
    fun undo(smith: String): List<String> = listOf(smith, "undo", "--yes")

    /** `smith why <file>`: why the file is as it is. [file] is relative to the workspace root, with forward slashes. */
    fun why(
        smith: String,
        file: String,
    ): List<String> = listOf(smith, "why", file)

    /** `smith release --dry-run`: what a release would publish, changing nothing. */
    fun releasePreview(smith: String): List<String> = listOf(smith, "release", "--dry-run")
}
