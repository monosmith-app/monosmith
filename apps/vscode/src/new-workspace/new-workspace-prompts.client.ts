import * as vscode from 'vscode'
import { checkProjectName } from '../add-project/index.ts'
import {
  ARCHITECTURE_CHOICES, CI_CHOICES, LANGUAGE_CHOICES, NODE_STACK_CHOICES,
} from './new-workspace-arguments.mapper.ts'
import type { ChoiceLabel, NewWorkspaceChoices, NewWorkspaceOutcome } from './new-workspace.contract.ts'
import type { SetupWay } from './new-workspace.use-case.ts'

const TITLE = 'MonoSmith: New workspace'

/**
 * Asks "How do you want to set it up?", with what the defaults choose in front
 * of the person. The third line keeps smith's own questions in a terminal.
 * @param defaults - the lines that say what "Start with the defaults" chooses
 * @returns the way, or undefined when the person pressed Escape
 */
export async function askSetupWay (defaults: readonly string[]): Promise<SetupWay | undefined> {
  const picked = await vscode.window.showQuickPick(
    [
      { label: '$(star) Start with the defaults', description: 'asks only the name', detail: `Chooses: ${defaults.join(' · ')}. You can change each one later.`, way: 'defaults' as const },
      { label: '$(settings-gear) Customise', description: 'CI, language, code layout, agent files, dev container, first commit', detail: 'Asks each question, with the defaults ticked.', way: 'customise' as const },
      { label: '$(terminal) Answer in a terminal', description: 'smith new --here asks its own questions', detail: 'Nothing is written before smith lists the files and you agree.', way: 'terminal' as const },
    ],
    { title: `${TITLE}: how do you want to set it up?`, placeHolder: 'Nothing is written until you have seen the list of files and said yes' },
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
    prompt:        `smith names the workspace after this folder, "${folderName}", unless you type another name here (kebab-case).`,
    value:         folderName,
    validateInput: (value) => checkProjectName(value),
  })
}

async function pickOne (choices: readonly ChoiceLabel[], current: string, title: string, hint: string): Promise<string | undefined> {
  const picked = await vscode.window.showQuickPick(
    choices.map((choice) => ({ label: choice.label, description: choice.value === current ? `${choice.description} (default)` : choice.description, value: choice.value })),
    { title: `${TITLE}: ${title}`, placeHolder: hint, matchOnDescription: true },
  )

  return picked?.value
}

async function pickYesNo (title: string, yes: string, no: string, current: boolean): Promise<boolean | undefined> {
  const picked = await vscode.window.showQuickPick(
    [{ label: `Yes${current ? ' (default)' : ''}`, description: yes, value: true }, { label: `No${current ? '' : ' (default)'}`, description: no, value: false }],
    { title: `${TITLE}: ${title}` },
  )

  return picked?.value
}

/**
 * Asks the questions of `smith new`, one at a time, with the default first.
 * @param start - the choices to start from (the defaults and the name)
 * @returns the choices, or undefined when the person pressed Escape at any question
 */
export async function askCustomChoices (start: NewWorkspaceChoices): Promise<NewWorkspaceChoices | undefined> {
  const ciPicked = await vscode.window.showQuickPick(
    CI_CHOICES.map((choice) => ({ label: choice.label, description: choice.description, picked: start.ci.includes(choice.value), value: choice.value })),
    { title: `${TITLE}: which CI?`, placeHolder: 'CI is where your tests run on every push. Tick one or more.', canPickMany: true },
  )
  if (ciPicked === undefined) return undefined
  if (ciPicked.length === 0) {
    void vscode.window.showWarningMessage('MonoSmith: pick at least one CI. The workspace was not created.')

    return undefined
  }
  const language = await pickOne(LANGUAGE_CHOICES, start.language, 'main language', 'Decide later is fine: each project you add picks its own')
  if (language === undefined) return undefined
  const nodeStack = language === 'node' ? await pickOne(NODE_STACK_CHOICES, start.nodeStack, 'tests and lint for Node projects', 'Used for every Node project you add') : start.nodeStack
  if (nodeStack === undefined) return undefined
  const architecture = await pickOne(ARCHITECTURE_CHOICES, start.architecture, 'code layout', 'The folder layout smith checks (smith architecture shows the rules)')
  if (architecture === undefined) return undefined
  const agent = await pickYesNo('add the MonoSmith agent?', 'files for Claude Code and VS Code to work here through smith', 'no agent files; smith agent install adds them later', start.agent)
  if (agent === undefined) return undefined
  const devcontainer = await pickYesNo('add a dev container?', 'writes .devcontainer/devcontainer.json', 'no dev container', start.devcontainer)
  if (devcontainer === undefined) return undefined
  const commit = await pickYesNo('make the first commit?', 'smith commits its own files as the first commit', 'smith writes the files and commits nothing', start.commit)

  return commit === undefined ? undefined : { ...start, ci: ciPicked.map((choice) => choice.value), language, nodeStack, architecture, agent, devcontainer, commit }
}

/**
 * Shows what will be chosen and the files `smith new --dry-run` lists, and asks.
 * Nothing was written yet.
 * @param summary - one line per choice
 * @param plan - the dry run, as smith printed it
 * @returns true when the person agrees
 */
export async function confirmNewWorkspace (summary: readonly string[], plan: string): Promise<boolean> {
  const create = 'Create the workspace'
  const open = 'Open the list of files'
  for (;;) {
    const choice = await vscode.window.showInformationMessage(
      'MonoSmith: this is what smith will create here.',
      { modal: true, detail: `${summary.join('\n')}\n\nNothing was written yet. The list of files is one click away.` },
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
      void vscode.window.showInformationMessage(`MonoSmith: the workspace ${outcome.name} is created. ${outcome.committed ? 'smith made the first commit of its own files.' : 'Nothing was committed: review the files with git.'} Add your first project from the Projects view.`)
      break
    }
    case 'refused': {
      void vscode.window.showWarningMessage(`MonoSmith: nothing was created. ${outcome.message}`)
      break
    }
    case 'terminal':
    case 'cancelled': {
      break
    }
  }
}
