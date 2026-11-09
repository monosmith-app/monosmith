package dev.monosmith.jetbrains.tabcompletion

import dev.monosmith.jetbrains.smithrun.SmithRunContract

/** What the offer of tab completion needs from the IDE. Tests give fakes. */
interface TabCompletionPortsContract {
    /** What the person answered before: [TabCompletionOfferPolicy.DONE], [TabCompletionOfferPolicy.NEVER] or null. */
    fun remembered(): String?

    /** Keeps an answer for the person, across projects. */
    fun remember(answer: String)

    /** True when the offer was shown since the IDE started. */
    fun askedThisSession(): Boolean

    /** Notes that the offer was shown. */
    fun markAsked()

    /** Runs `smith completion status --json`. Blocks: call it off the UI thread. */
    fun status(): SmithRunContract

    /**
     * Shows the offer with three buttons: "Set up" ([setUp]), "Not now" (nothing) and "Don't ask again" ([never]).
     * The callbacks run on the UI thread.
     */
    fun ask(
        setUp: () -> Unit,
        never: () -> Unit,
    )

    /** Opens the terminal on `smith completion install`; the person answers its question there. */
    fun openTerminal()
}
