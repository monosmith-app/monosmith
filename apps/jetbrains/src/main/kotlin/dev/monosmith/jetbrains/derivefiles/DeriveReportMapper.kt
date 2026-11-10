package dev.monosmith.jetbrains.derivefiles

/**
 * One derived file that needs action, as a line of `smith derive --check`.
 *
 * @property path relative to the workspace folder
 * @property status `stale`, `missing`, `unrecorded`, `drift`, `diverged` or `broken`
 * @property renderer the renderer that writes it, such as `tsconfig-references`
 * @property detail why the file is in that state, in smith's words
 */
data class DerivedFileContract(
    val path: String,
    val status: String,
    val renderer: String,
    val detail: String,
)

/** The command lines of `smith derive` and the reader of `smith derive --check`. smith has no JSON for it yet. */
object DeriveReportMapper {
    private val statuses = setOf("stale", "missing", "unrecorded", "drift", "diverged", "broken")
    private val fileLine = Regex("^\\s+([A-Z][a-z]+)\\s+(\\S.*?)\\s{2}\\((.*?)\\)\\s{2}(.*)$")
    private val held = setOf("drift", "diverged", "broken")

    /** The files that need action, read from what `smith derive --check` printed; a line this does not know is ignored. */
    fun files(stdout: String): List<DerivedFileContract> =
        stdout.lines().mapNotNull { line ->
            val match = fileLine.find(line) ?: return@mapNotNull null
            val status = match.groupValues[1].lowercase()
            if (status in statuses) DerivedFileContract(match.groupValues[2], status, match.groupValues[3], match.groupValues[4]) else null
        }

    /** Whether a plain `smith derive` holds the file back until the person chooses. */
    fun isHeldBack(status: String): Boolean = status in held

    /**
     * The arguments of `smith derive` for an action. The file goes in `--flag=file` form, so a name that
     * starts with a dash is never read as an option.
     *
     * @param action `check`, `preview`, `apply`, `explain`, `replace`, `adopt`, `eject` or `reclaim`
     * @param file the file, relative to the folder; needed by every action but `check`, `preview` and `apply`
     */
    fun arguments(
        action: String,
        file: String = "",
    ): List<String> =
        when (action) {
            "check" -> listOf("derive", "--check")
            "preview" -> listOf("derive", "--dry-run")
            "apply" -> listOf("derive", "--yes")
            "explain" -> listOf("derive", "--explain=$file")
            else -> listOf("derive", "--$action=$file", "--yes")
        }
}
