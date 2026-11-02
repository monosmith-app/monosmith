package dev.monosmith.jetbrains.commitchanges

/**
 * What the plan of `smith commit --dry-run` says.
 *
 * @property text the text smith printed: the files of the last operation, and the others git shows
 * @property extras the files git shows as changed that the last operation did not change; empty when there are none
 */
data class CommitPlanContract(
    val text: String,
    val extras: List<String>,
)

/** Reads the plan of `smith commit --dry-run` and builds its command lines. */
object CommitPlanAlgorithm {
    /** Reads the files that git shows and the last operation did not change from the plan [text]. */
    fun read(text: String): CommitPlanContract {
        val lines = text.lines()
        val start = lines.indexOfFirst { Regex("Git shows \\d+ more").containsMatchIn(it) }
        val after = if (start == -1) emptyList() else lines.drop(start + 1)
        val end = after.indexOfFirst { !Regex("^\\s+\\S").containsMatchIn(it) }
        val block = if (end == -1) after else after.take(end)
        return CommitPlanContract(text.trim(), block.map { it.trim() })
    }

    /**
     * The arguments of `smith commit`.
     *
     * @param operation the words that name the operation, such as "add app web"; empty for the last operation
     * @param scope null for the plan (`--dry-run`), `operation` for the files of the operation, `all` for everything
     */
    fun arguments(
        operation: String,
        scope: String?,
    ): List<String> =
        listOf("commit") + operation.split(" ").filter { it.isNotEmpty() } +
            when (scope) {
                null -> listOf("--dry-run")
                "all" -> listOf("--all", "--yes")
                else -> listOf("--only-operation", "--yes")
            }
}
