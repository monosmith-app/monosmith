package dev.monosmith.jetbrains.tabcompletion

/**
 * Offers tab completion once. The IDE writes nothing: "Set up" opens the terminal on
 * `smith completion install`, which shows the files and asks before it writes.
 */
object OfferTabCompletionUseCase {
    /**
     * @param ports the IDE
     * @return true when the offer was shown
     */
    fun offer(ports: TabCompletionPortsContract): Boolean {
        if (ports.remembered() != null || ports.askedThisSession()) return false
        val run = ports.status()
        val wanted = if (run.ok) CompletionStatusMapper.wanted(run.stdout) else null
        if (!TabCompletionOfferPolicy.shouldOffer(null, false, wanted == true)) return false
        ports.markAsked()
        ports.ask(
            setUp = {
                ports.remember(TabCompletionOfferPolicy.DONE)
                ports.openTerminal()
            },
            never = { ports.remember(TabCompletionOfferPolicy.NEVER) },
        )
        return true
    }
}
