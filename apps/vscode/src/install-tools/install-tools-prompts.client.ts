import * as vscode from 'vscode'
import type { AdminAnswer, InstallToolsOutcome } from './install-tools.use-case.ts'
import type { ToolOption, ToolPlan } from './tool-plan.contract.ts'

/**
 * Asks how to get a tool. Each way shows what it does and its command; the
 * marked default is first. Nothing runs from here.
 * @param plan - the tool and the options smith offers
 * @returns the `value` of the option, or undefined when the person cancels
 */
export async function chooseToolWay (plan: ToolPlan): Promise<string | undefined> {
  const ordered = [...plan.options].sort((a, b) => Number(b.value === plan.default) - Number(a.value === plan.default))
  const items = ordered.map((option) => ({
    label:       option.label + (option.value === plan.default ? ' (default)' : ''),
    description: option.command === undefined ? undefined : option.command.join(' '),
    detail:      option.note === undefined ? option.hint : `${option.hint}. ${option.note}.`,
    value:       option.value,
  }))
  const picked = await vscode.window.showQuickPick(items, {
    title:          `MonoSmith: ${plan.title} is ${plan.state === 'outdated' ? 'too old' : 'not installed'}`,
    placeHolder:    `${plan.why}. How do you want to get it?`,
    ignoreFocusOut: true,
  })

  return picked?.value
}

/**
 * Asks what to do with a command that needs administrator rights.
 * @param plan - the tool
 * @param option - the install option, with its command
 * @returns the answer, or undefined when the person cancels
 */
export async function chooseToolAdmin (plan: ToolPlan, option: ToolOption): Promise<AdminAnswer | undefined> {
  const command = (option.command ?? []).join(' ')
  const choice = await vscode.window.showInformationMessage(
    'MonoSmith: this command needs administrator rights.',
    { modal: true, detail: `${command}\n\n"Let smith run it" opens a terminal where the system or sudo asks for your permission. "I will run it myself" shows the command and goes on. Skipping means: ${plan.consequence}.` },
    'I will run it myself',
    'Let smith run it',
    'Skip it',
  )
  switch (choice) {
    case 'Let smith run it': { return 'run' }
    case 'I will run it myself': { return 'myself' }
    case 'Skip it': { return 'skip' }
    default: { return undefined }
  }
}

/**
 * Opens a page in the browser.
 * @param url - the official download address
 */
export async function openToolPage (url: string): Promise<void> {
  await vscode.env.openExternal(vscode.Uri.parse(url))
}

/**
 * Shows the commands that would work and offers to copy them.
 * @param plan - the tool
 * @param lines - one command or address per line
 */
export async function showToolCommands (plan: ToolPlan, lines: string[]): Promise<void> {
  const text = lines.join('\n')
  const choice = await vscode.window.showInformationMessage(`MonoSmith: ways to get ${plan.title}`, { modal: true, detail: `${text}\n\nThen open a new terminal.` }, 'Copy')
  if (choice === 'Copy') await vscode.env.clipboard.writeText(text)
}

/**
 * Says how installing ended. It does not wait for the person to read it.
 * @param outcome - what `installTools` returned
 */
export function tellToolsOutcome (outcome: InstallToolsOutcome): void {
  switch (outcome.kind) {
    case 'nothing': {
      void vscode.window.showInformationMessage('MonoSmith: every tool the workspace needs is installed.')
      break
    }
    case 'refused': {
      void vscode.window.showWarningMessage(`MonoSmith: the tools were not checked. ${outcome.message}`)
      break
    }
    case 'handled': {
      const lines = outcome.results.flatMap((result) => result.lines)
      if (lines.length > 0) void vscode.window.showInformationMessage(`MonoSmith: ${lines.join(' ')}`)
      break
    }
  }
}
