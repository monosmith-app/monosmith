package dev.monosmith.jetbrains.newworkspace

import dev.monosmith.jetbrains.jsonread.JsonFieldsMapper
import dev.monosmith.jetbrains.prompt.ChoiceRowContract
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDCHOICES_DEFAULT
import dev.monosmith.jetbrains.words.SmithWordsConfig.say

/** One answer of a question of `smith new`. */
data class NewOptionContract(
    val value: String,
    val label: String,
    val hint: String,
)

/**
 * One question of `smith new`, as `smith new --questions --json` prints it: the data the CLI and the TUI
 * ask from.
 *
 * @property flag the flag that answers it
 * @property many the answer may name several options
 * @property advanced asked only after "more options"
 * @property atEnd asked after the summary
 * @property when the answers that make it apply (for example `language` is `node`)
 * @property options the answers on offer, the default first; none means free text
 */
data class NewQuestionContract(
    val key: String,
    val prompt: String,
    val default: String,
    val flag: String,
    val many: Boolean,
    val advanced: Boolean,
    val atEnd: Boolean,
    val `when`: Map<String, String>,
    val options: List<NewOptionContract>,
)

/** Reads `smith new --questions --json` and words its questions as lists. */
object NewQuestionsMapper {
    /**
     * @param json what smith printed
     * @throws IllegalStateException when the text is not a JSON list
     */
    fun toQuestions(json: String): List<NewQuestionContract> =
        JsonFieldsMapper
            .parseArray(json)
            .filter { it.isJsonObject }
            .map { it.asJsonObject }
            .map { q ->
                NewQuestionContract(
                    key = JsonFieldsMapper.text(q, "key"),
                    prompt = JsonFieldsMapper.text(q, "prompt"),
                    default = JsonFieldsMapper.text(q, "default"),
                    flag = JsonFieldsMapper.text(q, "flag"),
                    many = JsonFieldsMapper.flag(q, "many"),
                    advanced = JsonFieldsMapper.flag(q, "advanced"),
                    atEnd = JsonFieldsMapper.flag(q, "atEnd"),
                    `when` =
                        JsonFieldsMapper
                            .record(q, "when")
                            ?.entrySet()
                            .orEmpty()
                            .associate { it.key to it.value.asString },
                    options =
                        JsonFieldsMapper.records(q, "options").map {
                            NewOptionContract(
                                JsonFieldsMapper.text(it, "value"),
                                JsonFieldsMapper.text(it, "label"),
                                JsonFieldsMapper.text(it, "hint"),
                            )
                        },
                )
            }.filter { it.key.isNotEmpty() }

    /** The lines of a question; the default is marked, and ticked first in a list of many. */
    fun rows(question: NewQuestionContract): List<ChoiceRowContract> =
        question.options.map {
            val isDefault = it.value == question.default || (question.many && it.value in question.default.split(","))
            val hint =
                it.hint
                    .removeSuffix("(default)")
                    .trim()
                    .let { text -> if (text == "default") "" else text }
            ChoiceRowContract(it.value, it.label, if (isDefault) say(JETBRAINS_ADDCHOICES_DEFAULT) else "", hint, picked = isDefault)
        }

    /** The questions a first answer settles: not advanced, not asked at the end. */
    fun main(questions: List<NewQuestionContract>): List<NewQuestionContract> =
        questions.filter {
            !it.advanced && !it.atEnd &&
                it.options.isNotEmpty()
        }

    /** Whether a question applies given the answers so far (key to value). */
    fun applies(
        question: NewQuestionContract,
        answers: Map<String, String>,
    ): Boolean = question.`when`.all { (key, value) -> answers[key] == value }

    /** The label of an answer, or the answer itself when the question does not list it. */
    fun labelOf(
        question: NewQuestionContract?,
        value: String,
    ): String = question?.options?.firstOrNull { it.value == value }?.label ?: value
}
