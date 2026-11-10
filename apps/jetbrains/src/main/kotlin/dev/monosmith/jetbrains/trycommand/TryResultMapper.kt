package dev.monosmith.jetbrains.trycommand

import dev.monosmith.jetbrains.jsonread.JsonFieldsMapper
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_TRYCOMMAND_FILES_IT_CHANGED
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_TRYCOMMAND_NONE_WERE_ASKED_FOR
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_TRYCOMMAND_SMITH_PRINTED_NO_RESULT
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_TRYCOMMAND_SOMETHING_FAILED_IN_THE_COPY
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_TRYCOMMAND_THE_COMMAND_AND_EVERY_CHECK
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_TRYCOMMAND_TRY_SMITH
import dev.monosmith.jetbrains.words.SmithWordsConfig.say

/** Lays out the answer of `smith try <command> --json` as text: what the command did in the copy, the files it changed and how each check went. */
object TryResultMapper {
    /**
     * @param passed true when the command and every check passed
     * @param page the text to show
     */
    data class Page(
        val passed: Boolean,
        val page: String,
    )

    /**
     * @param json what smith printed
     * @return the page, and whether the command and every check passed
     * @throws IllegalStateException when the text is not a result
     */
    fun toPage(json: String): Page {
        val result = JsonFieldsMapper.parseObject(json)
        val command = JsonFieldsMapper.words(result, "command")
        check(command.isNotEmpty()) { say(JETBRAINS_TRYCOMMAND_SMITH_PRINTED_NO_RESULT) }
        val commandOk = JsonFieldsMapper.flag(result, "commandOk")
        val checks = JsonFieldsMapper.records(result, "checks")
        val changed = JsonFieldsMapper.words(result, "changed")
        val passed = commandOk && checks.all { JsonFieldsMapper.flag(it, "ok") }

        fun tail(text: String) = if (text.isBlank()) emptyList() else listOf("", text.trim())
        val lines =
            buildList {
                add(say(JETBRAINS_TRYCOMMAND_TRY_SMITH, command.joinToString(" ").toString()))
                add("")
                add(
                    if (passed) {
                        say(JETBRAINS_TRYCOMMAND_THE_COMMAND_AND_EVERY_CHECK)
                    } else {
                        say(JETBRAINS_TRYCOMMAND_SOMETHING_FAILED_IN_THE_COPY)
                    },
                )
                add("")
                add("command: ${if (commandOk) "ran" else "failed"}")
                if (!commandOk) addAll(tail(JsonFieldsMapper.text(result, "commandTail")))
                add("")
                add(say(JETBRAINS_TRYCOMMAND_FILES_IT_CHANGED))
                if (changed.isEmpty()) add("  none") else changed.forEach { add("  $it") }
                add("")
                add("Checks:")
                if (checks.isEmpty()) add(say(JETBRAINS_TRYCOMMAND_NONE_WERE_ASKED_FOR))
                checks.forEach {
                    val ok = JsonFieldsMapper.flag(it, "ok")
                    add("  ${JsonFieldsMapper.text(it, "name")}: ${if (ok) "passed" else "failed"}")
                    if (!ok) addAll(tail(JsonFieldsMapper.text(it, "tail")))
                }
            }
        return Page(passed, lines.joinToString("\n"))
    }
}
