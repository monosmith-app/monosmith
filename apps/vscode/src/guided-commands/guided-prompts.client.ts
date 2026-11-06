import * as vscode from 'vscode'
import type { GuidedOutcome, ResolvedStep } from './guided-command.contract.ts'

/**
 * Asks one step with the editor's own controls: an input box (starting on the value smith or the
 * workspace gives, to edit), a pick list whose first answer is the default, tick boxes, or the file dialog.
 * @param step - the question, with its options read
 * @returns the answer (a list for tick boxes), or undefined when the person pressed Escape
 */
export async function askGuidedStep (step: ResolvedStep): Promise<string | string[] | undefined> {
  switch (step.kind) {
    case 'text': {
      const typed = await vscode.window.showInputBox({ title: step.title, prompt: step.prompt, placeHolder: step.placeholder, value: step.value, validateInput: (value) => value.trim() === '' && step.optional === true ? undefined : step.validate?.(value) ?? (value.trim() === '' ? 'Write an answer, or press Escape to cancel.' : undefined) })

      return typed?.trim()
    }
    case 'pick': {
      const picked = await vscode.window.showQuickPick(step.options.map((option) => ({ label: option.label, description: option.description, value: option.value })), { title: step.title, placeHolder: step.placeholder ?? 'Pick one', matchOnDescription: true })

      return picked?.value
    }
    case 'multi': {
      const picked = await vscode.window.showQuickPick(step.options.map((option) => ({ label: option.label, description: option.description, value: option.value })), { title: step.title, placeHolder: step.placeholder, canPickMany: true, matchOnDescription: true })
      if (picked === undefined) return undefined
      if (picked.length === 0) {
        void vscode.window.showInformationMessage('MonoSmith: nothing was ticked, so nothing changes.')

        return undefined
      }

      return picked.map((item) => item.value)
    }
    case 'file': {
      const files = await vscode.window.showOpenDialog({ title: step.title, canSelectMany: false, filters: { Images: ['png'] }, openLabel: 'Use this file' })

      return files?.[0]?.fsPath
    }
  }
}

/**
 * Shows what will change and asks. The detail is the command, smith's plan and what it costs; nothing is changed before a yes.
 * @param title - the question
 * @param detail - the command, the plan and the consequence
 * @param verb - the button that confirms
 * @returns true to go on
 */
export async function confirmGuided (title: string, detail: string, verb: string): Promise<boolean> {
  return await vscode.window.showInformationMessage(title, { modal: true, detail }, verb) === verb
}

/**
 * Says how a guided command ended. A refusal carries smith's words; it never throws.
 * @param title - the command's title
 * @param outcome - what `runGuidedCommand` returned
 * @param onOutput - gets every line smith printed, for the output channel
 */
export function tellGuidedOutcome (title: string, outcome: GuidedOutcome, onOutput: (line: string) => void): void {
  switch (outcome.kind) {
    case 'done': {
      for (const line of outcome.lines) onOutput(line)
      void vscode.window.showInformationMessage(`MonoSmith: ${title.replace(/…$/, '')}: done. ${outcome.lines.at(-1) ?? ''}`.trim())
      break
    }
    case 'refused': {
      void vscode.window.showWarningMessage(`MonoSmith: ${title.replace(/…$/, '')}: nothing changed. ${outcome.message}`)
      break
    }
    case 'shown':
    case 'cancelled': {
      break
    }
  }
}

/**
 * Says smith could not list the divergences of the workspace.
 * @param reason - smith's words for why
 */
export function tellDivergencesFailure (reason: string): void {
  void vscode.window.showWarningMessage(`MonoSmith: smith could not list the divergences. ${reason}`)
}
