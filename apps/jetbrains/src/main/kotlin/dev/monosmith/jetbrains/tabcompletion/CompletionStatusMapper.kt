package dev.monosmith.jetbrains.tabcompletion

import com.google.gson.JsonParseException
import dev.monosmith.jetbrains.jsonread.JsonFieldsMapper

/** Reads what `smith completion status --json` prints. */
object CompletionStatusMapper {
    /**
     * Says whether an offer is in order.
     *
     * @param stdout what smith printed
     * @return `wanted` of the status: true when nothing is installed for the shell, no package ships the
     *   file and the person did not say no on the command line; null when the text is not that JSON (an
     *   older smith has no such command)
     */
    fun wanted(stdout: String): Boolean? =
        try {
            val status = JsonFieldsMapper.parseObject(stdout)
            val field = status.get("wanted")
            if (field != null && field.isJsonPrimitive && field.asJsonPrimitive.isBoolean) field.asBoolean else null
        } catch (_: JsonParseException) {
            null
        } catch (_: IllegalStateException) {
            null
        }
}
