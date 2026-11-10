package dev.monosmith.jetbrains.installtools

/**
 * One way to get a tool, as `smith tools install --dry-run --json` prints it.
 *
 * @property kind `install` (a package manager), `page` (the official page), `myself` or `skip`
 * @property value the stable name of the option: `install:winget`, `page`, `myself`, `skip`
 * @property manager the package manager of an install option
 * @property command the install command of an install option: the program and its arguments
 * @property admin true when the command needs administrator rights, or may ask the system for them
 * @property hint the option and its consequence in plain words
 * @property note a sentence the person needs before choosing
 */
data class ToolOptionContract(
    val kind: String,
    val value: String,
    val manager: String,
    val command: List<String>,
    val admin: Boolean,
    val label: String,
    val hint: String,
    val note: String,
)

/**
 * A tool that is missing or too old, with the options smith offers for it.
 *
 * @property default the `value` of the marked option
 */
data class ToolPlanContract(
    val name: String,
    val title: String,
    val why: String,
    val consequence: String,
    val download: String,
    val state: String,
    val options: List<ToolOptionContract>,
    val default: String,
)
