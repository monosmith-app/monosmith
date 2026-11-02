package dev.monosmith.jetbrains.addproject

import dev.monosmith.jetbrains.jsonread.JsonFieldsMapper

/** Reads what the first questions of Add project offer: `smith kinds --json` and the outcomes of `smith compose --json`. */
object OutcomesJsonMapper {
    /**
     * Reads `smith kinds --json`. The language of a kind is the id in front of its role (`node` of `node-app`),
     * the id that `--lang` takes and that `smith compose --json` uses.
     */
    fun toKinds(json: String): List<KindOptionContract> =
        JsonFieldsMapper
            .parseArray(json)
            .filter { it.isJsonObject }
            .map { it.asJsonObject }
            .map {
                val kind = JsonFieldsMapper.text(it, "kind")
                val dash = kind.lastIndexOf('-')
                KindOptionContract(
                    kind = kind,
                    language = if (dash > 0) kind.substring(0, dash) else JsonFieldsMapper.text(it, "language"),
                    label = JsonFieldsMapper.text(it, "label"),
                    description = JsonFieldsMapper.text(it, "description"),
                )
            }.filter { it.kind.isNotEmpty() && it.language.isNotEmpty() }

    /** The outcomes, the languages and the title of each layer of `smith compose --json`. */
    fun toCatalogue(json: String): CatalogueContract {
        val root = JsonFieldsMapper.parseObject(json)
        val outcomes =
            JsonFieldsMapper
                .records(root, "outcomes")
                .map { outcome ->
                    OutcomeContract(
                        id = JsonFieldsMapper.text(outcome, "id"),
                        title = JsonFieldsMapper.text(outcome, "title"),
                        role = JsonFieldsMapper.text(outcome, "kind"),
                        languages = JsonFieldsMapper.words(outcome, "languages"),
                        typical = JsonFieldsMapper.words(outcome, "typical").joinToString(", "),
                        consequence = JsonFieldsMapper.text(outcome, "consequence"),
                        words = JsonFieldsMapper.words(outcome, "words"),
                        items =
                            JsonFieldsMapper.records(outcome, "items").map { item ->
                                val name = JsonFieldsMapper.text(item, "name")
                                OutcomeItemContract(
                                    name = name,
                                    title = JsonFieldsMapper.text(item, "title").ifEmpty { name },
                                    description = JsonFieldsMapper.text(item, "description"),
                                    language = JsonFieldsMapper.text(item, "language"),
                                    layer = JsonFieldsMapper.text(item, "layer"),
                                    aliases = JsonFieldsMapper.words(item, "aliases"),
                                )
                            },
                    )
                }.filter { it.id.isNotEmpty() && it.title.isNotEmpty() }
        val languages =
            JsonFieldsMapper
                .records(root, "items")
                .filter { JsonFieldsMapper.text(it, "kind") == "language" && JsonFieldsMapper.text(it, "name").isNotEmpty() }
                .map {
                    val name = JsonFieldsMapper.text(it, "name")
                    LanguageInfoContract(
                        name,
                        JsonFieldsMapper.text(it, "title").ifEmpty { name },
                        JsonFieldsMapper.text(it, "description"),
                    )
                }
        val layers =
            JsonFieldsMapper
                .records(root, "layers")
                .filter { JsonFieldsMapper.text(it, "id").isNotEmpty() && JsonFieldsMapper.text(it, "title").isNotEmpty() }
                .associate { JsonFieldsMapper.text(it, "id") to JsonFieldsMapper.text(it, "title") }
        val about =
            JsonFieldsMapper
                .records(root, "layers")
                .filter { JsonFieldsMapper.text(it, "id").isNotEmpty() && JsonFieldsMapper.text(it, "about").isNotEmpty() }
                .associate { JsonFieldsMapper.text(it, "id") to JsonFieldsMapper.text(it, "about") }
        return CatalogueContract(outcomes, languages, layers, about)
    }

    /**
     * Keeps the languages of each outcome that have a kind of the outcome's role (a library needs a lib
     * kind), and leaves out an outcome with no language left.
     */
    fun fitOutcomes(
        outcomes: List<OutcomeContract>,
        kinds: List<KindOptionContract>,
    ): List<OutcomeContract> =
        outcomes.mapNotNull { outcome ->
            val languages = outcome.languages.filter { language -> kinds.any { it.language == language && it.role == outcome.role } }
            if (languages.isEmpty()) {
                null
            } else {
                outcome.copy(
                    languages = languages,
                    items = outcome.items.filter { it.language in languages },
                )
            }
        }
}
