package dev.monosmith.jetbrains.tabcompletion

/** When to offer the setup of tab completion, and what the IDE remembers about the answer. */
object TabCompletionOfferPolicy {
    /** Remembered after "Set up": the terminal took over, so the IDE never asks again. */
    const val DONE = "done"

    /** Remembered after "Don't ask again". */
    const val NEVER = "never"

    /**
     * Offers only when smith says tab completion is not set up, the person never answered "Set up" or
     * "Don't ask again", and the offer was not shown in this session ("Not now" asks again next time).
     *
     * @param remembered what the person answered before, or null
     * @param askedThisSession true when the offer was shown since the IDE started
     * @param wanted what `smith completion status --json` said
     */
    fun shouldOffer(
        remembered: String?,
        askedThisSession: Boolean,
        wanted: Boolean,
    ): Boolean = wanted && remembered == null && !askedThisSession
}
