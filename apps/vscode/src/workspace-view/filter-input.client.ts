import * as vscode from 'vscode'

/**
 * Asks for the text that narrows the view, and narrows it while it is typed.
 * Escape puts the view back as it was.
 * @param current - the text the view is narrowed by now
 * @param apply - narrows the view by a text
 */
export function askFilterText (current: string, apply: (text: string) => void): void {
  const box = vscode.window.createInputBox()
  let accepted = false
  box.title = 'Filter projects and tasks'
  box.prompt = 'Part of a project name, a task name or a language. Enter keeps it, Escape puts the view back.'
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
