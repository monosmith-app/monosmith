package dev.monosmith.jetbrains.notice

import com.intellij.notification.NotificationAction
import com.intellij.notification.NotificationGroupManager
import com.intellij.notification.NotificationType
import com.intellij.openapi.project.Project
import com.intellij.openapi.ui.popup.JBPopupFactory
import com.intellij.ui.SimpleListCellRenderer

/** Shows balloon notifications in the MonoSmith notification group. */
object NoticeClient {
    private const val GROUP = "MonoSmith"

    /** Tells [message] as information. */
    fun info(
        ide: Project,
        message: String,
        buttons: List<NoticeButtonContract> = emptyList(),
    ) = show(ide, NotificationType.INFORMATION, message, buttons)

    /** Tells [message] as a warning. */
    fun warn(
        ide: Project,
        message: String,
        buttons: List<NoticeButtonContract> = emptyList(),
    ) = show(ide, NotificationType.WARNING, message, buttons)

    /**
     * Shows a notification with at most three short buttons. When there are more, the last slot is a
     * More... button that opens a list with the long text of each remaining button.
     */
    fun show(
        ide: Project,
        type: NotificationType,
        message: String,
        buttons: List<NoticeButtonContract>,
    ) {
        val (shown, more) = NoticeButtonsMapper.split(buttons)
        val notification =
            NotificationGroupManager
                .getInstance()
                .getNotificationGroup(GROUP)
                .createNotification("MonoSmith", message, type)
        shown.forEach { button -> notification.addAction(NotificationAction.createSimpleExpiring(button.label) { button.action() }) }
        if (more.isNotEmpty()) {
            notification.addAction(NotificationAction.createSimple(NoticeButtonsMapper.MORE_LABEL) { openMore(more) })
        }
        notification.notify(ide)
    }

    private fun openMore(more: List<NoticeButtonContract>) {
        JBPopupFactory
            .getInstance()
            .createPopupChooserBuilder(more)
            .setTitle("MonoSmith")
            .setRenderer(
                SimpleListCellRenderer.create<NoticeButtonContract> { label, value, _ ->
                    label.text = "<html><b>${value.label}</b>${if (value.detail.isEmpty()) "" else "<br>${value.detail}"}</html>"
                },
            ).setItemChosenCallback { it.action() }
            .createPopup()
            .showInFocusCenter()
    }
}
