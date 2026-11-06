import * as vscode from 'vscode'
import type { AdminAnswer, InstallToolsOutcome } from './install-tools.use-case.ts'
import type { ToolOption, ToolPlan } from './tool-plan.contract.ts'
import { say } from '../words/index.ts'

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
    title:          say('vscode.install-tools.monosmith-is', { title: plan.title, say: say(plan.state === 'outdated' ? 'vscode.install-tools.too-old' : 'vscode.install-tools.not-installed') }),
    placeHolder:    say('vscode.install-tools.how-do-you-want-to-get', { why: plan.why }),
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
    say('vscode.install-tools.monosmith-this-command-needs-administrator-rights'),
    { modal: true, detail: say('vscode.install-tools.let-smith-run-it-opens-a', { command, consequence: plan.consequence }) },
    say('toolsdialog.answer.myself'),
    say('toolsdialog.answer.smith'),
    say('toolsdialog.answer.skip'),
  )
  switch (choice) {
    case say('toolsdialog.answer.smith'): { return 'run' }
    case say('toolsdialog.answer.myself'): { return 'myself' }
    case say('toolsdialog.answer.skip'): { return 'skip' }
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
  const choice = await vscode.window.showInformationMessage(say('vscode.install-tools.monosmith-ways-to-get', { title: plan.title }), { modal: true, detail: say('vscode.install-tools.then-open-a-new-terminal', { text }) }, say('vscode.install-tools.copy'))
  if (choice === say('vscode.install-tools.copy')) await vscode.env.clipboard.writeText(text)
}

/**
 * Says how installing ended. It does not wait for the person to read it.
 * @param outcome - what `installTools` returned
 */
export function tellToolsOutcome (outcome: InstallToolsOutcome): void {
  switch (outcome.kind) {
    case 'nothing': {
      void vscode.window.showInformationMessage(say('vscode.install-tools.monosmith-every-tool-the-workspace-needs'))
      break
    }
    case 'refused': {
      void vscode.window.showWarningMessage(say('vscode.install-tools.monosmith-the-tools-were-not-checked', { message: outcome.message }))
      break
    }
    case 'handled': {
      const lines = outcome.results.flatMap((result) => result.lines)
      if (lines.length > 0) void vscode.window.showInformationMessage(say('vscode.install-tools.monosmith', { join: lines.join(' ') }))
      break
    }
  }
}
