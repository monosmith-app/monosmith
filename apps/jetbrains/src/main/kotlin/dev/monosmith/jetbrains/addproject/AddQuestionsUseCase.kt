package dev.monosmith.jetbrains.addproject

import dev.monosmith.jetbrains.addchoices.ChoiceRowsAlgorithm
import dev.monosmith.jetbrains.addchoices.GlossaryAlgorithm
import dev.monosmith.jetbrains.addchoices.PickChoicesUseCase
import dev.monosmith.jetbrains.flowports.PromptPort

/**
 * The questions of Add project, outcome first: what is being built, the language only where several fit,
 * the framework, the choices of each axis, the project to follow, the name and, for a web app, its
 * environments. Every list is a presented choice with the default first. Blocks: call it from a background flow.
 */
object AddQuestionsUseCase {
    private const val TITLE = "MonoSmith: Add project"

    /** A project of the workspace, as far as these questions read it. */
    data class Sibling(
        val name: String,
        val language: String,
    )

    /**
     * @param options what smith offers
     * @param siblings the projects the workspace has; their names are taken, and those of the language can be followed
     * @param check checks the picks with smith before the choices are left; what is wrong in plain words, or an empty text
     * @return the answers, or null when the person cancelled
     */
    fun ask(
        prompt: PromptPort,
        options: AddOptionsContract,
        siblings: List<Sibling>,
        check: (PicksAskedContract) -> String = { "" },
    ): AddAnswerContract? {
        val first =
            prompt.pickOne(
                "$TITLE: what are you building?",
                "Pick what you want. Type in the box to narrow the list: try react, azure function, electron, tauri, api.",
                AddRowsAlgorithm.outcomeRows(options.outcomes),
            ) ?: return null
        if (first.name == AddRowsAlgorithm.COMPOSE) return AddAnswerContract.Compose
        val found = if (first.name == AddRowsAlgorithm.EVERY) askEveryLanguage(prompt, options) else askOutcome(prompt, options, first.name)
        found ?: return null

        // The name comes first, then one question: the defaults (straight to the plan) or the questions one by one.
        // A function app keeps its questions: smith does not choose its trigger.
        var typedName: String? = null
        if (found.outcomeId != "function") {
            val chosen = askName(prompt, found.role, siblings) ?: return null
            typedName = chosen
            val setup =
                prompt.pickOne(
                    "$TITLE: how do you want to set it up?",
                    "The defaults are the answers the stack suggests. The plan shows every default as a flag.",
                    AddRowsAlgorithm.setupRows,
                ) ?: return null
            if (setup.name == AddRowsAlgorithm.SETUP_DEFAULTS) {
                val request = AddRequestContract(found.role, chosen, found.language, with = found.preset)
                val environments =
                    if (EnvironmentsAlgorithm.isWebApp(request)) {
                        EnvironmentsContract(
                            EnvironmentsAlgorithm.defaults,
                            EnvironmentsAlgorithm.defaultOf(EnvironmentsAlgorithm.defaults),
                            "workspace",
                            "commit",
                        )
                    } else {
                        null
                    }
                return AddAnswerContract.Request(
                    request.copy(environments = environments),
                    ChoiceRowsAlgorithm.inWords(found.preset, null, found.language, options.choices, options.layers),
                )
            }
        }

        val picked =
            PickChoicesUseCase.pick(
                prompt,
                options.choices,
                found.language,
                found.preset,
                found.outcomeId,
                options.layers,
                options.layerAbout,
                found.role,
            ) { names ->
                check(PicksAskedContract(found.language, names, found.outcomeId, found.role))
            }
                ?: return null
        val like = askLike(prompt, siblings.filter { it.language == found.language }.map { it.name }) ?: return null
        val name = typedName ?: askName(prompt, found.role, siblings) ?: return null
        val request = AddRequestContract(found.role, name, found.language, with = picked, like = like.ifEmpty { null })
        val environments =
            if (EnvironmentsAlgorithm.isWebApp(request)) askEnvironments(prompt, name) ?: return null else null
        return AddAnswerContract.Request(
            request.copy(environments = environments),
            ChoiceRowsAlgorithm.inWords(picked, null, found.language, options.choices, options.layers),
        )
    }

    private fun askName(
        prompt: PromptPort,
        role: String,
        siblings: List<Sibling>,
    ): String? =
        prompt.text("$TITLE: name", "The name of the new $role (for example billing-api):") {
            ProjectNameValidator.problem(it, siblings.map { sibling -> sibling.name })
        }

    private data class Found(
        val role: String,
        val language: String,
        val preset: List<String>,
        val outcomeId: String?,
    )

    private fun askOutcome(
        prompt: PromptPort,
        options: AddOptionsContract,
        id: String,
    ): Found? {
        val outcome = options.outcomes.first { it.id == id }
        val language =
            if (outcome.languages.size == 1) {
                outcome.languages.first()
            } else {
                askLanguage(prompt, options, outcome.role, outcome.languages, outcome) ?: return null
            }
        val items = outcome.items.filter { it.language == language }
        if (items.isEmpty()) return Found(outcome.role, language, emptyList(), outcome.id)
        val layer = options.layers[items.first().layer]?.lowercase()
        val answer =
            prompt.pickOne(
                "$TITLE: which ${layer ?: "one"}?",
                GlossaryAlgorithm.withLine(
                    "Your pick is kept: the next questions do not ask it again. The first answer is the default.",
                    GlossaryAlgorithm.line(items.first().layer, options.layers, options.layerAbout),
                ),
                AddRowsAlgorithm.presetRows(items, outcome.id),
            ) ?: return null
        return Found(outcome.role, language, listOfNotNull(answer.name.ifEmpty { null }), outcome.id)
    }

    private fun askEveryLanguage(
        prompt: PromptPort,
        options: AddOptionsContract,
    ): Found? {
        val role =
            prompt.pickOne("$TITLE: what kind?", "app, lib or package.", AddRowsAlgorithm.roleRows(options.kinds)) ?: return null
        val language = askLanguage(prompt, options, role.name, options.kinds.map { it.language }.distinct(), null) ?: return null
        return Found(role.name, language, emptyList(), null)
    }

    private fun askLanguage(
        prompt: PromptPort,
        options: AddOptionsContract,
        role: String,
        fitting: List<String>,
        outcome: OutcomeContract?,
    ): String? {
        val rows = AddRowsAlgorithm.languageRows(options.languages, options.kinds, role, fitting, outcome)
        return prompt
            .pickOne(
                "$TITLE: which language?",
                "Pick the language. What each one offers for this project is next to it.",
                rows,
            )?.name
    }

    /** Asks which project to copy; returns the name, an empty text for "what most do", or null on Cancel. */
    private fun askLike(
        prompt: PromptPort,
        siblings: List<String>,
    ): String? {
        if (siblings.size < 2) return ""
        return prompt
            .pickOne(
                "$TITLE: follow which project?",
                "The new project copies the build, test, layout and lint of one project (smith add --like).",
                AddRowsAlgorithm.likeRows(siblings),
            )?.name
    }

    private fun askEnvironments(
        prompt: PromptPort,
        app: String,
    ): EnvironmentsContract? {
        val typed =
            prompt.text(
                "$TITLE: environments of $app",
                "Vite writes its VITE_ values into the bundle when it builds, so the app is built once per environment.\nEdit the names (lowercase words, separated by commas):",
                EnvironmentsAlgorithm.defaults.joinToString(", "),
                EnvironmentsAlgorithm::problem,
            ) ?: return null
        val names = EnvironmentsAlgorithm.parse(typed)
        val build =
            prompt.pickOne(
                "$TITLE: which environment is \"build\"?",
                "CI and releases build this one; the others are build:<name>.",
                AddRowsAlgorithm.environmentRows(names),
            )
                ?: return null
        val output =
            prompt.pickOne("$TITLE: where do the builds go?", "The first answer is the default.", AddRowsAlgorithm.outputRows)
                ?: return null
        val files =
            prompt.pickOne("$TITLE: the .env files", "The first answer is the default.", AddRowsAlgorithm.envFileRows)
                ?: return null
        return EnvironmentsContract(names, build.name, output.name, files.name)
    }
}
