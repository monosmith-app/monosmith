import * as vscode from 'vscode'
import { say } from '../words/index.ts'

/**
 * Asks for the text that narrows the view, and narrows it while it is typed.
 * Escape puts the view back as it was.
 * @param current - the text the view is narrowed by now
 * @param apply - narrows the view by a text
 */
export function askFilterText (current: string, apply: (text: string) => void): void {
  const box = vscode.window.createInputBox()
  let accepted = false
  box.title = say('vscode.workspace-view.filter-projects-and-tasks')
  box.prompt = say('vscode.workspace-view.part-of-a-project-name-a')
  box.placeholder = 'web, lint, go'
  box.value = current
  box.onDidChangeValue((text) => { apply(text) })
  box.onDidAccept(() => {
    accepted = true
    box.hide()
  })
  box.onDidHide(() => {
    if (!accepted) apply(current)
    box.dispose()
  })
  box.show()
}
