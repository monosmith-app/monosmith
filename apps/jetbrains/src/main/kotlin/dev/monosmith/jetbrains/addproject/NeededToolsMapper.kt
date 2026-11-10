package dev.monosmith.jetbrains.addproject

/**
 * The tools a new project needs that the machine lacks or has in another version, as `smith add` reports them.
 *
 * @property names the names `smith tools install` takes: `golangci-lint`, `python`
 * @property problems what is wrong with each, in plain words
 */
data class NeededToolsContract(
    val names: List<String>,
    val problems: List<String>,
)

/**
 * Reads the tools `smith add` says the new project needs and does not have: the `! <tool> is not installed: ...`
 * lines and the `-> Run smith tools install <names> to choose how to get them.` line that ends them. smith
 * installs nothing when the IDE runs it (no terminal), so this is how the IDE learns what to offer.
 */
object NeededToolsMapper {
    private val command = Regex("^→ Run smith tools install (.+?) to choose how to get (?:it|them)\\.$")

    /**
     * Reads the tools from what smith printed.
     *
     * @param output what smith printed
     * @return the tools and what is wrong with each, or null when none is lacking
     */
    fun parse(output: String): NeededToolsContract? {
        val lines = output.lines().map { it.trim() }
        val names = lines.firstNotNullOfOrNull { command.find(it)?.groupValues?.get(1) } ?: return null
        return NeededToolsContract(
            names = names.split(' '),
            problems = lines.filter { it.startsWith("! ") }.map { it.removePrefix("! ") },
        )
    }
}
