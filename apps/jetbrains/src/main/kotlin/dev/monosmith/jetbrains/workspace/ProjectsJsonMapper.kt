package dev.monosmith.jetbrains.workspace

import com.google.gson.JsonParser

/** Reads the documented `smith projects --json` contract; it never parses human output. */
object ProjectsJsonMapper {
    fun toProjects(json: String): List<SmithProjectContract> =
        JsonParser.parseString(json).asJsonArray.map { entry ->
            val project = entry.asJsonObject
            SmithProjectContract(
                name = project["name"].asString,
                dir = project["dir"].asString,
                kind = project["kind"]?.asString.orEmpty(),
                language = project["language"]?.asString.orEmpty(),
                tasks = project["tasks"]?.asJsonArray?.map { it.asString }.orEmpty(),
            )
        }
}
