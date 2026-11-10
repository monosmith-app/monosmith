package dev.monosmith.jetbrains.workspace

import com.google.gson.JsonParser

/** Reads the documented `smith kinds --json` contract; it never parses human output. */
object KindsJsonMapper {
    fun toKinds(json: String): List<SmithKindContract> =
        JsonParser.parseString(json).asJsonArray.map { entry ->
            val kind = entry.asJsonObject
            SmithKindContract(
                kind = kind["kind"].asString,
                language = kind["language"]?.asString.orEmpty(),
                label = kind["label"]?.asString ?: kind["kind"].asString,
            )
        }
}
