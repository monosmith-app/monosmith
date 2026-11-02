package dev.monosmith.jetbrains.installtools

import dev.monosmith.jetbrains.jsonread.JsonFieldsMapper

/** Reads `smith tools install --dry-run --json`. */
object ToolPlansJsonMapper {
    /**
     * @param json what smith printed; `null` or an empty list means nothing is missing
     * @return one plan for each tool that is missing or too old
     * @throws IllegalStateException when the text is neither a list nor `null`
     */
    fun toPlans(json: String): List<ToolPlanContract> {
        if (json.trim() == "null") return emptyList()
        return JsonFieldsMapper
            .parseArray(json)
            .filter { it.isJsonObject }
            .map { it.asJsonObject }
            .map { plan ->
                ToolPlanContract(
                    name = JsonFieldsMapper.text(plan, "name"),
                    title = JsonFieldsMapper.text(plan, "title").ifEmpty { JsonFieldsMapper.text(plan, "name") },
                    why = JsonFieldsMapper.text(plan, "why"),
                    consequence = JsonFieldsMapper.text(plan, "consequence"),
                    download = JsonFieldsMapper.text(plan, "download"),
                    state = JsonFieldsMapper.text(plan, "state"),
                    options =
                        JsonFieldsMapper.records(plan, "options").map {
                            ToolOptionContract(
                                kind = JsonFieldsMapper.text(it, "kind"),
                                value = JsonFieldsMapper.text(it, "value"),
                                manager = JsonFieldsMapper.text(it, "manager"),
                                command = JsonFieldsMapper.words(it, "command"),
                                admin = JsonFieldsMapper.flag(it, "admin"),
                                label = JsonFieldsMapper.text(it, "label"),
                                hint = JsonFieldsMapper.text(it, "hint"),
                                note = JsonFieldsMapper.text(it, "note"),
                            )
                        },
                    default = JsonFieldsMapper.text(plan, "default"),
                )
            }.filter { it.name.isNotEmpty() }
    }
}
