package dev.monosmith.jetbrains.smithcommand

/** The argument vector of a smith call, and how it reads on one line. */
object TaskCommandAlgorithm {
    fun runTask(
        smith: String,
        project: String,
        task: String,
    ): List<String> = listOf(smith, "run", task, "-p", project)

    fun listProjects(smith: String): List<String> = listOf(smith, "projects", "--json")

    /** The command as a terminal would show it: arguments with spaces are quoted. */
    fun display(command: List<String>): String = command.joinToString(" ") { if (it.any(Char::isWhitespace)) "\"$it\"" else it }
}
