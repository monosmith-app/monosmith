package dev.monosmith.jetbrains.composestack

/**
 * One thing the catalogue of `smith compose --json` lists.
 *
 * @property id unique: language/name
 * @property kind `language`, `choice` or `plugin`
 * @property axis the question a choice answers (`framework`); it can differ from the layer (Electron sits on `shell`)
 * @property many the axis takes several choices at once; false means one answer, like a radio button
 * @property problem why the item cannot be used here; empty when it can
 * @property route the command that turns a workspace plugin on; empty for an item of `--with`
 */
data class ComposeItemContract(
    val id: String,
    val kind: String,
    val name: String,
    val title: String,
    val description: String,
    val aliases: List<String>,
    val layer: String,
    val axis: String,
    val many: Boolean,
    val language: String,
    val suggested: Boolean,
    val isDefault: Boolean,
    val problem: String,
    val route: String,
)

/** One layer of the stack, from the language up to the pipelines. */
data class ComposeLayerContract(
    val id: String,
    val title: String,
    val about: String,
)

/** One answer to a question smith does not answer for the person. */
data class ComposeOptionContract(
    val name: String,
    val title: String,
    val description: String,
    val suggested: Boolean,
)

/** A pick needs an answer on an axis (a function host needs a trigger). */
data class ComposeQuestionContract(
    val choice: String,
    val about: String,
    val axis: String,
    val options: List<ComposeOptionContract>,
)

/**
 * What the picks add up to, as smith checked them.
 *
 * @property conflict the picks cannot go together: one sentence that names both items and the fix
 * @property missing not wrong, but not finished: the next thing to add
 * @property selection the picks with the defaults added; set when complete
 */
data class ComposeCheckContract(
    val language: String,
    val complete: Boolean,
    val conflict: String,
    val missing: String,
    val question: ComposeQuestionContract?,
    val selection: List<String>,
)

/**
 * A program a selection needs on the computer, and whether it is there.
 *
 * @property state `installed`, `outdated` or `missing`
 * @property command the line that offers the install: `smith tools install <name>`
 */
data class ComposeToolContract(
    val choice: String,
    val name: String,
    val title: String,
    val state: String,
    val command: String,
)

/** A file the plan writes: the action, the path and a note. */
data class ComposeFileContract(
    val action: String,
    val path: String,
    val note: String,
)

/**
 * The plan a complete selection ends in.
 *
 * @property command the exact `smith add` line
 * @property argv the arguments of that line, with `smith` first
 * @property recipeArgv the arguments that keep the stack as a recipe, with `smith` first; empty when there are none
 */
data class ComposePlanContract(
    val command: String,
    val argv: List<String>,
    val files: List<ComposeFileContract>,
    val tools: List<ComposeToolContract>,
    val notes: List<String>,
    val recipeArgv: List<String>,
)

/** What `smith compose --json` prints. */
data class ComposeReportContract(
    val layers: List<ComposeLayerContract>,
    val items: List<ComposeItemContract>,
    val check: ComposeCheckContract?,
    val plan: ComposePlanContract?,
)
