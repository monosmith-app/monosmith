package dev.monosmith.jetbrains.addproject

import dev.monosmith.jetbrains.prompt.ChoiceRowContract
import dev.monosmith.jetbrains.words.SmithWordsConfig.COMMANDS_FILE_COMPOSE_TITLE
import dev.monosmith.jetbrains.words.SmithWordsConfig.GUIDE_ADD_PROJECT_OUTCOMES_DECIDE_IN_THE_NEXT_QUESTION
import dev.monosmith.jetbrains.words.SmithWordsConfig.GUIDE_ADD_PROJECT_OUTCOMES_I_KNOW_THE_LANGUAGE_SHOW_EVERY
import dev.monosmith.jetbrains.words.SmithWordsConfig.GUIDE_ADD_PROJECT_SETUP_CUSTOMISE
import dev.monosmith.jetbrains.words.SmithWordsConfig.GUIDE_ADD_PROJECT_SETUP_START_WITH_THE_DEFAULTS
import dev.monosmith.jetbrains.words.SmithWordsConfig.GUIDE_WEB_ENVIRONMENTS_STEPS_IN_THE_PROJECT
import dev.monosmith.jetbrains.words.SmithWordsConfig.GUIDE_WEB_ENVIRONMENTS_STEPS_IN_THE_WORKSPACE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDCHOICES_DEFAULT
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_ANSWER_THE_QUESTIONS_ONE_AT
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_BUILT_BY_SMITH_RUN_BUILD
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_BUILT_BY_SMITH_RUN_BUILD_2
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_CODE_PUBLISHED_TO_A_REGISTRY
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_CODE_THE_OTHER_PROJECTS_OF
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_COMMIT_ONLY_ENV_EXAMPLE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_COMMIT_THE_ENV_FILES
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_COPY_THE_BUILD_TEST_LAYOUT
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_DIST_APPS_APP_NAME_AT
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_DIST_NAME_INSIDE_THE_APP
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_EACH_MACHINE_AND_CI_JOB
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_SEARCH_FRAMEWORKS_DESKTOP_SHELLS_AND
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_SMITH_FOLLOWS_THE_SIBLINGS_THAT
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_SOMETHING_YOU_RUN_OR_DEPLOY
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_THEY_HOLD_VITE_VALUES_WHICH
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_THE_LONG_WAY_THE_KIND
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_WHAT_MOST_PROJECTS_DO_DEFAULT
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_WHAT_THE_STACK_SUGGESTS_YOUR
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_YOU_CAN_STILL_PICK_A
import dev.monosmith.jetbrains.words.SmithWordsConfig.say

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
            "app" to say(JETBRAINS_ADDPROJECT_SOMETHING_YOU_RUN_OR_DEPLOY),
            "lib" to say(JETBRAINS_ADDPROJECT_CODE_THE_OTHER_PROJECTS_OF),
            "package" to say(JETBRAINS_ADDPROJECT_CODE_PUBLISHED_TO_A_REGISTRY),
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
                    say(GUIDE_ADD_PROJECT_OUTCOMES_I_KNOW_THE_LANGUAGE_SHOW_EVERY),
                    "",
                    say(JETBRAINS_ADDPROJECT_THE_LONG_WAY_THE_KIND),
                ),
                ChoiceRowContract(
                    COMPOSE,
                    say(COMMANDS_FILE_COMPOSE_TITLE),
                    "",
                    say(JETBRAINS_ADDPROJECT_SEARCH_FRAMEWORKS_DESKTOP_SHELLS_AND),
                ),
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
                say(GUIDE_ADD_PROJECT_SETUP_START_WITH_THE_DEFAULTS),
                say(JETBRAINS_ADDCHOICES_DEFAULT),
                say(JETBRAINS_ADDPROJECT_WHAT_THE_STACK_SUGGESTS_YOUR),
            ),
            ChoiceRowContract(
                "custom",
                say(GUIDE_ADD_PROJECT_SETUP_CUSTOMISE),
                "",
                say(JETBRAINS_ADDPROJECT_ANSWER_THE_QUESTIONS_ONE_AT),
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
                        say(GUIDE_ADD_PROJECT_OUTCOMES_DECIDE_IN_THE_NEXT_QUESTION),
                        say(JETBRAINS_ADDCHOICES_DEFAULT),
                        say(JETBRAINS_ADDPROJECT_YOU_CAN_STILL_PICK_A),
                    ),
                )
            }
        ) + items.map { ChoiceRowContract(it.name, it.title, it.name, it.description, search = it.aliases.joinToString(" ")) }

    /** The rows of "follow which project?": what most siblings do first, then each sibling. */
    fun likeRows(siblings: List<String>): List<ChoiceRowContract> =
        listOf(
            ChoiceRowContract(
                "",
                say(JETBRAINS_ADDPROJECT_WHAT_MOST_PROJECTS_DO_DEFAULT),
                "",
                say(JETBRAINS_ADDPROJECT_SMITH_FOLLOWS_THE_SIBLINGS_THAT),
            ),
        ) +
            siblings.map { ChoiceRowContract(it, it, "", say(JETBRAINS_ADDPROJECT_COPY_THE_BUILD_TEST_LAYOUT)) }

    /** The rows of "which environment is build?". */
    fun environmentRows(names: List<String>): List<ChoiceRowContract> {
        val fallback = EnvironmentsAlgorithm.defaultOf(names)
        return names
            .sortedByDescending { it == fallback }
            .map {
                ChoiceRowContract(
                    it,
                    it,
                    if (it == fallback) say(JETBRAINS_ADDCHOICES_DEFAULT) else "",
                    if (it ==
                        fallback
                    ) {
                        say(JETBRAINS_ADDPROJECT_BUILT_BY_SMITH_RUN_BUILD)
                    } else {
                        say(JETBRAINS_ADDPROJECT_BUILT_BY_SMITH_RUN_BUILD_2, it.toString())
                    },
                )
            }
    }

    /** The rows of "where do the builds go?"; the first is the default. */
    val outputRows =
        listOf(
            ChoiceRowContract(
                "workspace",
                say(GUIDE_WEB_ENVIRONMENTS_STEPS_IN_THE_WORKSPACE),
                say(JETBRAINS_ADDCHOICES_DEFAULT),
                say(JETBRAINS_ADDPROJECT_DIST_APPS_APP_NAME_AT),
            ),
            ChoiceRowContract(
                "project",
                say(GUIDE_WEB_ENVIRONMENTS_STEPS_IN_THE_PROJECT),
                "",
                say(JETBRAINS_ADDPROJECT_DIST_NAME_INSIDE_THE_APP),
            ),
        )

    /** The rows of "the .env files"; the first is the default. */
    val envFileRows =
        listOf(
            ChoiceRowContract(
                "commit",
                say(JETBRAINS_ADDPROJECT_COMMIT_THE_ENV_FILES),
                say(JETBRAINS_ADDCHOICES_DEFAULT),
                say(JETBRAINS_ADDPROJECT_THEY_HOLD_VITE_VALUES_WHICH),
            ),
            ChoiceRowContract(
                "example",
                say(JETBRAINS_ADDPROJECT_COMMIT_ONLY_ENV_EXAMPLE),
                "",
                say(JETBRAINS_ADDPROJECT_EACH_MACHINE_AND_CI_JOB),
            ),
        )
}
