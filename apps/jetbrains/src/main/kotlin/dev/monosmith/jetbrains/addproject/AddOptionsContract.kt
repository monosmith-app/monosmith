package dev.monosmith.jetbrains.addproject

import dev.monosmith.jetbrains.addchoices.ChoiceOptionContract

/** A framework, a desktop shell or a cloud host that names an outcome. */
data class OutcomeItemContract(
    val name: String,
    val title: String,
    val description: String,
    val language: String,
    val layer: String,
    val aliases: List<String>,
)

/**
 * Something a person wants to build, in plain words: the first question of Add project.
 *
 * @property id `web`, `api`, `function`, `desktop`, `library` or `package`
 * @property role the kind smith adds: `app`, `lib` or `package`
 * @property languages the languages that fit it; one means no language is asked
 * @property typical what people usually pick for it
 * @property consequence one sentence: what the person gets
 * @property words the words that find it besides its title
 * @property items the frameworks, shells or hosts of its layer
 */
data class OutcomeContract(
    val id: String,
    val title: String,
    val role: String,
    val languages: List<String>,
    val typical: String,
    val consequence: String,
    val words: List<String>,
    val items: List<OutcomeItemContract>,
)

/** A kind `smith add` accepts, from `smith kinds --json`; [language] is the id `--lang` takes. */
data class KindOptionContract(
    val kind: String,
    val language: String,
    val label: String,
    val description: String,
) {
    /** What the kind adds to its language: `node-app` of `node` is `app`. */
    val role: String get() = kind.removePrefix("$language-")
}

/** A language smith can generate, with the words a person knows it by ("C# / .NET"). */
data class LanguageInfoContract(
    val name: String,
    val title: String,
    val description: String,
)

/**
 * What `smith compose --json` says about the stack: the outcomes, the languages, and for each layer its
 * title ("Desktop shell") and what it means, which the questions show as a one-line glossary.
 */
data class CatalogueContract(
    val outcomes: List<OutcomeContract>,
    val languages: List<LanguageInfoContract>,
    val layers: Map<String, String>,
    val about: Map<String, String>,
)

/** Everything the questions of Add project offer, all read from smith. */
data class AddOptionsContract(
    val kinds: List<KindOptionContract>,
    val choices: List<ChoiceOptionContract>,
    val outcomes: List<OutcomeContract>,
    val languages: List<LanguageInfoContract>,
    val layers: Map<String, String>,
    val layerAbout: Map<String, String> = emptyMap(),
)
