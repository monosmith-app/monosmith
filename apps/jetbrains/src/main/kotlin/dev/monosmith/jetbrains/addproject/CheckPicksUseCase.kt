package dev.monosmith.jetbrains.addproject

import dev.monosmith.jetbrains.flowports.SmithPort
import dev.monosmith.jetbrains.jsonread.JsonFieldsMapper
import dev.monosmith.jetbrains.smithcommand.SmithFlagsConfig
import dev.monosmith.jetbrains.words.SmithWordsConfig.GUIDE_ADD_PROJECT_PICKS_LAYER_FRAMEWORK
import dev.monosmith.jetbrains.words.SmithWordsConfig.GUIDE_ADD_PROJECT_PICKS_LAYER_HOST
import dev.monosmith.jetbrains.words.SmithWordsConfig.GUIDE_ADD_PROJECT_PICKS_LAYER_SHELL
import dev.monosmith.jetbrains.words.SmithWordsConfig.GUIDE_ADD_PROJECT_PICKS_UNTICK
import dev.monosmith.jetbrains.words.SmithWordsConfig.say

/**
 * Checks the ticked choices before the lists are left, as the terminal form does. smith checks the set
 * (`smith compose --json --with`, the solver of `smith add`): lint and test choices alone are complete,
 * because the stack's own scaffold is filled in. What smith cannot know without the kind is whether this
 * kind of project needs a framework, a desktop shell or a cloud host: the outcome says which layer, and the
 * choices of that layer that fit the kind are named. Nothing ticked is fine: the stack's defaults are used.
 */
object CheckPicksUseCase {
    private val layerWords =
        mapOf(
            "framework" to say(GUIDE_ADD_PROJECT_PICKS_LAYER_FRAMEWORK),
            "shell" to say(GUIDE_ADD_PROJECT_PICKS_LAYER_SHELL),
            "host" to say(GUIDE_ADD_PROJECT_PICKS_LAYER_HOST),
        )

    /**
     * Turns a sentence of `smith add` (which talks of `--with`) into one for a list where nothing is typed
     * as a flag; the same words as the terminal form.
     *
     * @param text what smith said is missing
     * @return the sentence without flags
     */
    fun plainWords(text: String): String =
        text.replace(" to --with", "").replace("give no --with to start", say(GUIDE_ADD_PROJECT_PICKS_UNTICK))

    /**
     * @param smith runs smith in the workspace folder
     * @param options what smith lists
     * @param asked the language, the picks, what is being built and the kind of project
     * @return what is wrong or missing in plain words, or an empty text when the set is fine or smith cannot
     *   say (`smith add` has the last word)
     */
    fun check(
        smith: SmithPort,
        options: AddOptionsContract,
        asked: PicksAskedContract,
    ): String {
        if (asked.picks.isEmpty()) return ""
        val run =
            smith.run(
                listOf(
                    "compose",
                    SmithFlagsConfig.JSON_LONG,
                    SmithFlagsConfig.LANG_LONG,
                    asked.language,
                    "--with",
                    asked.picks.joinToString(","),
                ),
            )
        if (!run.ok) return ""
        val check =
            runCatching { JsonFieldsMapper.record(JsonFieldsMapper.parseObject(run.stdout), "check") }.getOrNull() ?: return ""
        val conflict = JsonFieldsMapper.text(check, "conflict")
        if (conflict.isNotEmpty()) return conflict
        val missing = JsonFieldsMapper.text(check, "missing")
        if (missing.isNotEmpty()) return plainWords(missing)
        if (JsonFieldsMapper.record(check, "question") != null) return ""

        val layer =
            options.outcomes
                .firstOrNull { it.id == asked.outcomeId }
                ?.items
                ?.firstOrNull()
                ?.layer
        if (layer.isNullOrEmpty()) return ""
        val have =
            options.choices.filter {
                (it.language.isEmpty() || it.language == asked.language) &&
                    it.layer.ifEmpty { it.axis } == layer &&
                    it.fitsKind(asked.role)
            }
        if (have.any { it.name in asked.picks }) return ""
        return "this kind of project needs ${layerWords[layer] ?: "a $layer"}: tick one of ${have.joinToString(
            ", ",
        ) { it.name }}, or untick everything"
    }
}
