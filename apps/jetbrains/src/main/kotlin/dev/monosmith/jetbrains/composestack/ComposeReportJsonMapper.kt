package dev.monosmith.jetbrains.composestack

import com.google.gson.JsonObject
import dev.monosmith.jetbrains.jsonread.JsonFieldsMapper as Json

/** Reads `smith compose --json`. */
object ComposeReportJsonMapper {
    /**
     * @param json what smith printed
     * @return the layers, the items that have a name, and the check and plan when asked for
     * @throws IllegalStateException when the text is not a JSON object
     */
    fun toReport(json: String): ComposeReportContract {
        val root = Json.parseObject(json)
        return ComposeReportContract(
            layers =
                Json
                    .records(
                        root,
                        "layers",
                    ).map { ComposeLayerContract(Json.text(it, "id"), Json.text(it, "title"), Json.text(it, "about")) },
            items =
                Json
                    .records(root, "items")
                    .map {
                        ComposeItemContract(
                            id = Json.text(it, "id"),
                            kind = Json.text(it, "kind"),
                            name = Json.text(it, "name"),
                            title = Json.text(it, "title").ifEmpty { Json.text(it, "name") },
                            description = Json.text(it, "description"),
                            aliases = Json.words(it, "aliases"),
                            layer = Json.text(it, "layer"),
                            axis = Json.text(it, "axis"),
                            many = Json.flag(it, "many"),
                            language = Json.text(it, "language"),
                            suggested = Json.flag(it, "suggested"),
                            isDefault = Json.flag(it, "default"),
                            problem = Json.text(it, "problem"),
                            route = Json.text(it, "route"),
                        )
                    }.filter { it.id.isNotEmpty() && it.name.isNotEmpty() },
            check = Json.record(root, "check")?.let(::toCheck),
            plan = Json.record(root, "plan")?.let(::toPlan),
        )
    }

    private fun toCheck(check: JsonObject): ComposeCheckContract =
        ComposeCheckContract(
            language = Json.text(check, "language"),
            complete = Json.flag(check, "complete"),
            conflict = Json.text(check, "conflict"),
            missing = Json.text(check, "missing"),
            question =
                Json.record(check, "question")?.takeIf { Json.text(it, "axis").isNotEmpty() }?.let { question ->
                    ComposeQuestionContract(
                        choice = Json.text(question, "choice"),
                        about = Json.text(question, "about"),
                        axis = Json.text(question, "axis"),
                        options =
                            Json.records(question, "options").map {
                                ComposeOptionContract(
                                    Json.text(it, "name"),
                                    Json.text(it, "title").ifEmpty { Json.text(it, "name") },
                                    Json.text(it, "description"),
                                    Json.flag(it, "suggested"),
                                )
                            },
                    )
                },
            selection = Json.words(check, "selection"),
        )

    private fun toPlan(plan: JsonObject): ComposePlanContract? {
        if (Json.text(plan, "command").isEmpty() && Json.words(plan, "notes").isEmpty()) return null
        return ComposePlanContract(
            command = Json.text(plan, "command"),
            argv = Json.words(plan, "argv"),
            files =
                Json
                    .records(
                        plan,
                        "files",
                    ).map { ComposeFileContract(Json.text(it, "action"), Json.text(it, "path"), Json.text(it, "note")) },
            tools =
                Json.records(plan, "tools").map {
                    ComposeToolContract(
                        Json.text(it, "choice"),
                        Json.text(it, "name"),
                        Json.text(it, "title").ifEmpty { Json.text(it, "name") },
                        Json.text(it, "state"),
                        Json.text(it, "command"),
                    )
                },
            notes = Json.words(plan, "notes"),
            recipeArgv = Json.words(plan, "recipeArgv"),
        )
    }
}
