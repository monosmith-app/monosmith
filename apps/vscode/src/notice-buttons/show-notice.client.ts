import * as vscode from 'vscode'
import { MORE_LABEL, splitNoticeButtons, type NoticeButton } from './notice-buttons.mapper.ts'
import { say } from '../words/index.ts'

/**
 * Shows a notification with at most three buttons. Buttons beyond that sit
 * behind "More…", which opens a list that names each one in a full sentence.
 * @param severity - `info` or `warning`
 * @param message - what happened and what the buttons do, in words: the long text belongs here, not on a button
 * @param buttons - the buttons, the most likely first
 * @returns the label of the button that was pressed or picked; undefined when the notification was closed or the list was dismissed
 */
export async function showNotice (severity: 'info' | 'warning', message: string, buttons: readonly NoticeButton[]): Promise<string | undefined> {
  const { shown, more } = splitNoticeButtons(buttons)
  const labels = [...shown.map((button) => button.label), ...(more.length > 0 ? [MORE_LABEL] : [])]
  const choice = await (severity === 'warning' ? vscode.window.showWarningMessage(message, ...labels) : vscode.window.showInformationMessage(message, ...labels))
  if (choice !== MORE_LABEL) return choice
  const picked = await vscode.window.showQuickPick(
    more.map((button) => ({ label: button.label, detail: button.detail ?? '' })),
    { title: say('vscode.notice-buttons.monosmith-more-choices'), placeHolder: message, ignoreFocusOut: true },
  )

  return picked?.label
}
