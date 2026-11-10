package dev.monosmith.jetbrains.smithrun

/** What a finished smith command printed and how it ended. */
data class SmithRunContract(
    val code: Int,
    val stdout: String,
    val stderr: String,
) {
    /** True when the command ended with code 0. */
    val ok: Boolean get() = code == 0

    /** The lines smith printed on standard output, trimmed and without blank ones. */
    val lines: List<String> get() = stdout.lines().map { it.trim() }.filter { it.isNotEmpty() }
}
