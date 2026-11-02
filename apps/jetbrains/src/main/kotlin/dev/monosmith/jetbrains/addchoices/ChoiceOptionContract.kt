package dev.monosmith.jetbrains.addchoices

/**
 * One choice of `smith choices --json`: a framework, a test runner, a style.
 *
 * @property axis the question it answers (`framework`, `test`, `style`)
 * @property name the name `--with` takes
 * @property language the language it fits; empty when it fits every language
 * @property requires the choices it needs (`chakra` needs `react`)
 * @property isDefault an axis left out takes this choice
 * @property many the axis allows several choices at once
 * @property needs the axes it needs an answer on; an axis with no default is asked (a function host)
 * @property suggested the answer a question pre-selects; smith never adds it by itself
 * @property conflicts the choices it does not go with
 * @property title what a person calls it ("Electron"); empty when the name is enough
 * @property layer where it sits in the stack (`framework`, `shell`, `host`); empty means the layer of its axis
 */
data class ChoiceOptionContract(
    val axis: String,
    val name: String,
    val description: String,
    val language: String,
    val requires: List<String>,
    val isDefault: Boolean,
    val many: Boolean,
    val needs: List<String>,
    val suggested: Boolean,
    val conflicts: List<String>,
    val title: String,
    val layer: String,
) {
    /** What a person reads: the title, else the name. */
    val label: String get() = title.ifEmpty { name }
}

/**
 * The choices of one axis for one language.
 *
 * @property axis the axis name
 * @property many the axis allows several choices
 * @property choices its choices, in the order smith lists them
 */
data class AxisGroupContract(
    val axis: String,
    val many: Boolean,
    val choices: List<ChoiceOptionContract>,
)
