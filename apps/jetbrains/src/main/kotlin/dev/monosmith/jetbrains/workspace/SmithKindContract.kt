package dev.monosmith.jetbrains.workspace

/** One thing `smith add` accepts, as `smith kinds --json` lists it. */
data class SmithKindContract(
    /** The combined kind, such as `node-app` or `go-lib`. */
    val kind: String,
    val language: String,
    val label: String,
) {
    /** What the kind adds to its language: `node-app` of `node` is `app`. */
    fun role(): String = kind.removePrefix("$language-")

    override fun toString() = label
}
