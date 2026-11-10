package dev.monosmith.jetbrains.workspace

import dev.monosmith.jetbrains.jsonread.JsonFieldsMapper

/**
 * What `smith affected --json` says.
 *
 * @property projects the projects the changes affect
 * @property all true when a change affects every project (a workspace file changed)
 * @property reason why, in smith's words
 */
data class AffectedContract(
    val projects: Set<String>,
    val all: Boolean,
    val reason: String,
)

/** Reads `smith affected --json`. */
object AffectedJsonMapper {
    /**
     * @param json what smith printed
     * @return the affected projects
     * @throws IllegalStateException when the text is not a JSON object
     */
    fun toAffected(json: String): AffectedContract {
        val root = JsonFieldsMapper.parseObject(json)
        return AffectedContract(
            JsonFieldsMapper.words(root, "projects").toSet(),
            JsonFieldsMapper.flag(root, "all"),
            JsonFieldsMapper.text(root, "reason"),
        )
    }
}
