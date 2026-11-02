package dev.monosmith.jetbrains.addproject

import dev.monosmith.jetbrains.prompt.ChoiceRowContract

/** The lines of the first questions of Add project, built from what smith lists. No outcome, language or framework is named here. */
object AddRowsAlgorithm {
    /** The name of the row that opens the long way: the kind first, then every language. */
    const val EVERY = "\u0000every"

    /** The name of the row that opens the stack composer. */
    const val COMPOSE = "\u0000compose"

    /** The name of the row that leaves the framework for the next question. */
    const val LATER = ""

    /** What the three kinds of project mean, one sentence each. */
    val roleWords =
        mapOf(
            "app" to "Something you run or deploy. Other projects do not import it.",
            "lib" to "Code the other projects of this workspace import. It is not published.",
            "package" to "Code published to a registry on release, for other people to install.",
        )

    /** The first question: what is being built. The two last lines lead to the long way and to the composer. */
    fun outcomeRows(outcomes: List<OutcomeContract>): List<ChoiceRowContract> =
        outcomes.map {
            ChoiceRowContract(
                name = it.id,
                label = it.title,
                description = it.typical,
                detail = it.consequence,
                search = (it.words + it.items.flatMap { item -> listOf(item.name, item.title) + item.aliases }).joinToString(" "),
            )
        } +
            listOf(
                ChoiceRowContract(
                    EVERY,
                    "I know the language: show every language",
                    "",
                    "The long way: the kind first, then the language from the full list.",
                ),
                ChoiceRowContract(COMPOSE, "Compose a stack…", "", "Search frameworks, desktop shells and cloud hosts in one place."),
            )

    /**
     * The rows of the language question in words a person knows: the language by its name, and next to it
     * what it offers for the thing being built (the clouds of a function app, the shells of a desktop app).
     */
    fun languageRows(
        languages: List<LanguageInfoContract>,
        kinds: List<KindOptionContract>,
        role: String,
        fitting: List<String>,
        outcome: OutcomeContract?,
    ): List<ChoiceRowContract> =
        fitting.mapNotNull { name ->
            val kind = kinds.firstOrNull { it.language == name && it.role == role } ?: return@mapNotNull null
            val info = languages.firstOrNull { it.name == name }
            val offers =
                outcome
                    ?.items
                    .orEmpty()
                    .filter { it.language == name }
                    .map { it.title }
                    .distinct()
            ChoiceRowContract(
                name = name,
                label = info?.title ?: name,
                description = offers.joinToString(", "),
                detail = info?.description?.ifEmpty { null } ?: kind.description.ifEmpty { kind.label },
            )
        }

    /** The answer of "Start with the defaults". */
    const val SETUP_DEFAULTS = "defaults"

    /** The two answers after the name: the defaults, which go straight to the plan, or the questions one by one. */
    val setupRows =
        listOf(
            ChoiceRowContract(
                SETUP_DEFAULTS,
                "Start with the defaults",
                "(default)",
                "What the stack suggests: your pick, the shape of the other projects, and for a website lcl, dev, uat, prd with prd as the build.",
            ),
            ChoiceRowContract(
                "custom",
                "Customise",
                "",
                "Answer the questions one at a time: choices, the project to follow, the environments.",
            ),
        )

    /** The kinds the long way offers: `app`, `lib`, `package` as smith lists them. */
    fun roleRows(kinds: List<KindOptionContract>): List<ChoiceRowContract> =
        kinds.map { it.role }.distinct().map { ChoiceRowContract(it, it, "", roleWords[it].orEmpty()) }

    /**
     * The rows of the framework, shell or host an outcome uses. A website or a desktop app may decide in
     * the next question (the last row); a function app may not.
     */
    fun presetRows(
        items: List<OutcomeItemContract>,
        outcomeId: String,
    ): List<ChoiceRowContract> =
        (
            if (outcomeId == "function") {
                emptyList()
            } else {
                // First, so Enter on the list picks nothing: a framework is the person's pick.
                listOf(
                    ChoiceRowContract(
                        LATER,
                        "Decide in the next question",
                        "(default)",
                        "You can still pick a framework, a shell or a tool there.",
                    ),
                )
            }
        ) + items.map { ChoiceRowContract(it.name, it.title, it.name, it.description, search = it.aliases.joinToString(" ")) }

    /** The rows of "follow which project?": what most siblings do first, then each sibling. */
    fun likeRows(siblings: List<String>): List<ChoiceRowContract> =
        listOf(ChoiceRowContract("", "What most projects do (default)", "", "smith follows the siblings that agree.")) +
            siblings.map { ChoiceRowContract(it, it, "", "Copy the build, test, layout and lint of this project.") }

    /** The rows of "which environment is build?". */
    fun environmentRows(names: List<String>): List<ChoiceRowContract> {
        val fallback = EnvironmentsAlgorithm.defaultOf(names)
        return names
            .sortedByDescending { it == fallback }
            .map {
                ChoiceRowContract(
                    it,
                    it,
                    if (it == fallback) "(default)" else "",
                    if (it == fallback) "Built by `smith run build`: what CI and the release build." else "Built by `smith run build:$it`.",
                )
            }
    }

    /** The rows of "where do the builds go?"; the first is the default. */
    val outputRows =
        listOf(
            ChoiceRowContract(
                "workspace",
                "In the workspace",
                "(default)",
                "dist/apps/<app>/<name> at the workspace root, with every other build.",
            ),
            ChoiceRowContract("project", "In the project", "", "dist/<name> inside the app's own folder; smith doctor then accepts it."),
        )

    /** The rows of "the .env files"; the first is the default. */
    val envFileRows =
        listOf(
            ChoiceRowContract(
                "commit",
                "Commit the .env files",
                "(default)",
                "They hold VITE_ values, which end up in the browser, so they hold no secret; CI builds with them.",
            ),
            ChoiceRowContract(
                "example",
                "Commit only .env.example",
                "",
                "Each machine and CI job makes its own .env.<name>; git ignores them.",
            ),
        )
}
