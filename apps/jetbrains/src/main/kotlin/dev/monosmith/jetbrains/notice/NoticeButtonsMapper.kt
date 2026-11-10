package dev.monosmith.jetbrains.notice

/** Splits the buttons of a notification so the balloon never shows more than three. */
object NoticeButtonsMapper {
    /** The label of the button that opens the rest of the buttons. */
    const val MORE_LABEL = "More…"

    /** The most buttons a notification shows, the More... button included. */
    const val LIMIT = 3

    /**
     * Splits the buttons of a notification: it shows at most [limit] of them, and the rest go behind a
     * More... button that opens a list.
     *
     * @param buttons every button, the most likely first
     * @param limit how many buttons the notification may show, the More... button included
     * @return the buttons to show, and those behind More... (empty when all fit)
     */
    fun <T> split(
        buttons: List<T>,
        limit: Int = LIMIT,
    ): Pair<List<T>, List<T>> = if (buttons.size <= limit) buttons to emptyList() else buttons.take(limit - 1) to buttons.drop(limit - 1)
}
