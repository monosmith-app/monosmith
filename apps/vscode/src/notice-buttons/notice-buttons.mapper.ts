import { say } from '../words/index.ts'
/** One thing a notification can do: a short verb on the button, the long words in `detail`. */
export interface NoticeButton {
  /** The verb on the button ("Install tools"): short, because the notification cuts what does not fit. */
  label:   string
  /** What it does, in a full sentence; shown in the "More…" list, never on the button. */
  detail?: string
}

/** The label of the button that opens the rest of the buttons. */
export const MORE_LABEL = say('vscode.notice-buttons.more')

/**
 * Splits the buttons of a notification: it shows at most `limit` of them, and the
 * rest go behind a "More…" button that opens a list. A notification with more
 * buttons gets cut at the edge of the window.
 * @param buttons - every button, the most likely first
 * @param limit - how many buttons the notification may show, "More…" included
 * @returns the buttons to show, and those behind "More…" (empty when all fit)
 */
export function splitNoticeButtons (buttons: readonly NoticeButton[], limit = 3): { shown: NoticeButton[], more: NoticeButton[] } {
  return buttons.length <= limit ? { shown: [...buttons], more: [] } : { shown: buttons.slice(0, limit - 1), more: buttons.slice(limit - 1) }
}
