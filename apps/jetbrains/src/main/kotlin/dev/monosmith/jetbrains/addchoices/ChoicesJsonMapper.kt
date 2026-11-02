package dev.monosmith.jetbrains.addchoices

import dev.monosmith.jetbrains.jsonread.JsonFieldsMapper

/** Reads `smith choices --json`. */
object ChoicesJsonMapper {
    /**
     * @param json what smith printed
     * @return the choices that have an axis and a name
     * @throws IllegalStateException when the text is not a JSON list
     */
    fun toChoices(json: String): List<ChoiceOptionContract> =
        JsonFieldsMapper
            .parseArray(json)
            .filter { it.isJsonObject }
            .map { it.asJsonObject }
            .map {
                ChoiceOptionContract(
                    axis = JsonFieldsMapper.text(it, "axis"),
                    name = JsonFieldsMapper.text(it, "name"),
                    description = JsonFieldsMapper.text(it, "description"),
                    language = JsonFieldsMapper.text(it, "language"),
                    requires = JsonFieldsMapper.words(it, "requires"),
                    isDefault = JsonFieldsMapper.flag(it, "default"),
                    many = JsonFieldsMapper.flag(it, "many"),
                    needs = JsonFieldsMapper.words(it, "needs"),
                    suggested = JsonFieldsMapper.flag(it, "suggested"),
                    conflicts = JsonFieldsMapper.words(it, "conflicts"),
                    title = JsonFieldsMapper.text(it, "title"),
                    layer = JsonFieldsMapper.text(it, "layer"),
                )
            }.filter { it.axis.isNotEmpty() && it.name.isNotEmpty() }
}
