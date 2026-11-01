package dev.monosmith.jetbrains.workspace

/**
 * Turns what smith printed on stderr into the one line a tool window shows. smith words a problem as
 * a first line that names it (for an unreadable `smith.toml`: the file, the line and the column) and
 * then lines that start with an arrow and say what to do.
 */
object ProblemTextMapper {
    /** The problem line and the first remedy, joined: `smith.toml:3:9 ... Try: ...`. */
    fun toLine(stderr: String): String {
        val lines =
            stderr
                .lines()
                .map { it.trim().removePrefix("✖").trim() }
                .filter { it.isNotEmpty() }
        val problem = lines.firstOrNull { !it.startsWith("→") }.orEmpty()
        val remedy = lines.firstOrNull { it.startsWith("→") }?.removePrefix("→")?.trim()
        return if (remedy == null) problem else "$problem Try: $remedy"
    }
}
