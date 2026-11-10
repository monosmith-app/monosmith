package dev.monosmith.jetbrains.notice

/**
 * One thing a notification can do: a short verb on the button, the long words in [detail].
 *
 * @property label the verb on the button ("Install tools"): short, because the balloon cuts what does not fit
 * @property detail what it does, in a full sentence; shown in the More... popup, never on the button
 * @property action what happens on a click; it runs on the UI thread, so a flow starts its own background task
 */
data class NoticeButtonContract(
    val label: String,
    val detail: String = "",
    val action: () -> Unit,
)
