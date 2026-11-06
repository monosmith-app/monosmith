import * as vscode from 'vscode'
import { checkProjectName } from '../add-project/index.ts'
import { askCustomAnswers } from './custom-answers.use-case.ts'
import type { SmithQuestion } from '../smith-questions/index.ts'
import type { NewWorkspaceChoices, NewWorkspaceOutcome } from './new-workspace.contract.ts'
import type { SetupWay } from './new-workspace.use-case.ts'
import { say } from '../words/index.ts'

const TITLE = say('vscode.new-workspace.monosmith-new-workspace')

/**
 * Asks "How do you want to set it up?", with what the defaults choose in front
 * of the person. The third line keeps smith's own questions in a terminal.
 * @param defaults - the lines that say what "Start with the defaults" chooses
 * @returns the way, or undefined when the person pressed Escape
 */
export async function askSetupWay (defaults: readonly string[]): Promise<SetupWay | undefined> {
  const picked = await vscode.window.showQuickPick(
    [
      { label: say('vscode.new-workspace.start-with-the-defaults'), description: say('vscode.new-workspace.asks-only-the-name'), detail: say('vscode.new-workspace.chooses-you-can-change-each-one', { join: defaults.join(' · ') }), way: 'defaults' as const },
      { label: say('vscode.new-workspace.customise'), description: say('vscode.new-workspace.ci-language-code-layout-agent-files'), detail: say('vscode.new-workspace.asks-each-question-with-the-defaults'), way: 'customise' as const },
      { label: say('vscode.new-workspace.answer-in-a-terminal'), description: say('vscode.new-workspace.smith-new-here-asks-its-own'), detail: say('vscode.new-workspace.nothing-is-written-before-smith-lists'), way: 'terminal' as const },
    ],
    { title: say('vscode.add-project.how-do-you-want-to-set', { title: TITLE }), placeHolder: say('vscode.new-workspace.nothing-is-written-until-you-have') },
  )

  return picked?.way
}

/**
 * Asks the name of the workspace. The prompt says that the name of the folder
 * is used unless another is typed; the box starts with it.
 * @param folderName - the name of the open folder
 * @returns the name, or undefined when the person pressed Escape
 */
export async function askWorkspaceName (folderName: string): Promise<string | undefined> {
  return await vscode.window.showInputBox({
    title:         `${TITLE}: name`,
    prompt:        say('vscode.new-workspace.smith-names-the-workspace-after-this', { folderName }),
    value:         folderName,
    validateInput: (value) => checkProjectName(value),
  })
}

/**
 * Shows one question of `smith new` or `smith convert` as a quick pick: the answers with smith's
 * hints, the default marked, and a tick box when the question takes many.
 * @param question - the question
 * @param current - the values picked at first
 * @returns the values picked, or undefined when the person pressed Escape or ticked nothing
 */
export async function pickQuestion (question: SmithQuestion, current: readonly string[]): Promise<string[] | undefined> {
  const items = question.answers.map((answer) => {
    const hint = answer.hint.replace(/\s*\(?default\)?$/, '').trim()

    return { label: answer.label, description: [hint, answer.value === question.default ? '(default)' : ''].filter(Boolean).join(' '), picked: current.includes(answer.value), value: answer.value }
  })
  const picked = await vscode.window.showQuickPick(items, { title: `${TITLE}: ${question.prompt}`, canPickMany: question.many, matchOnDescription: true })
  if (picked === undefined) return undefined
  const values = Array.isArray(picked) ? picked.map((item) => item.value) : [picked.value]
  if (values.length === 0) {
    void vscode.window.showWarningMessage(say('vscode.new-workspace.monosmith-pick-at-least-one-answer'))

    return undefined
  }

  return values
}

async function pickYesNo (title: string, yes: string, no: string, current: boolean): Promise<boolean | undefined> {
  const picked = await vscode.window.showQuickPick(
    [{ label: say('vscode.new-workspace.yes-2', { part: current ? say('vscode.add-project.default-2') : '' }), description: yes, value: true }, { label: say('vscode.new-workspace.no-2', { part: current ? '' : say('vscode.add-project.default-2') }), description: no, value: false }],
    { title: `${TITLE}: ${title}` },
  )

  return picked?.value
}

/**
 * Asks the questions of `smith new`, one at a time, with the default first.
 * The questions come from `smith new --questions --json`; the extension adds only the first commit,
 * which is not a question of smith.
 * @param start - the choices to start from (the defaults and the name)
 * @param questions - the questions of `smith new`
 * @returns the choices, or undefined when the person pressed Escape at any question
 */
export async function askCustomChoices (start: NewWorkspaceChoices, questions: readonly SmithQuestion[]): Promise<NewWorkspaceChoices | undefined> {
  const answered = await askCustomAnswers(questions, start, pickQuestion)
  if (answered === undefined) return undefined
  const commit = await pickYesNo(say('vscode.new-workspace.make-the-first-commit'), say('vscode.new-workspace.smith-commits-its-own-files-as'), say('vscode.new-workspace.smith-writes-the-files-and-commits'), start.commit)

  return commit === undefined ? undefined : { ...answered, commit }
}

/**
 * Shows what will be chosen and the files `smith new --dry-run` lists, and asks.
 * Nothing was written yet.
 * @param summary - one line per choice
 * @param plan - the dry run, as smith printed it
 * @returns true when the person agrees
 */
export async function confirmNewWorkspace (summary: readonly string[], plan: string): Promise<boolean> {
  const create = say('vscode.new-workspace.create-the-workspace')
  const open = say('vscode.new-workspace.open-the-list-of-files')
  for (;;) {
    const choice = await vscode.window.showInformationMessage(
      say('vscode.new-workspace.monosmith-this-is-what-smith-will'),
      { modal: true, detail: say('vscode.new-workspace.nothing-was-written-yet-the-list', { join: summary.join('\n') }) },
      create,
      open,
    )
    if (choice !== open) return choice === create
    const document = await vscode.workspace.openTextDocument({ language: 'plaintext', content: `${plan}\n` })
    await vscode.window.showTextDocument(document, { preview: true })
  }
}

/**
 * Says how creating the workspace ended. It does not wait for the person to read it.
 * @param outcome - what `createWorkspace` returned
 * @param onOutput - gets every line smith printed, for the output channel
 */
export function tellNewWorkspaceOutcome (outcome: NewWorkspaceOutcome, onOutput: (line: string) => void): void {
  switch (outcome.kind) {
    case 'created': {
      for (const line of outcome.lines) onOutput(line)
      void vscode.window.showInformationMessage(say('vscode.new-workspace.monosmith-the-workspace-is-created-add', { name: outcome.name, part: say(outcome.committed ? 'vscode.new-workspace.smith-made-the-first-commit-of' : 'vscode.convert-workspace.nothing-was-committed-review-the-files') }))
      break
    }
    case 'refused': {
      void vscode.window.showWarningMessage(say('vscode.new-workspace.monosmith-nothing-was-created', { message: outcome.message }))
      break
    }
    case 'terminal':
    case 'cancelled': {
      break
    }
  }
}
